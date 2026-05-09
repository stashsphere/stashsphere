package workers

import (
	"archive/zip"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/rs/zerolog/log"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/services"
	"github.com/stashsphere/backend/utils"
)

const (
	importMaxEntries        = 50_000
	importMaxComprRatio     = uint64(100)
	importMaxJsonBytes      = int64(64 * 1024 * 1024)   // 64 MB
	importMaxFileBytes      = int64(200 * 1024 * 1024)  // 200 MB per image
	importMaxExtractedBytes = int64(1024 * 1024 * 1024) // 1 GB total
	importErrorRetention    = 7 * 24 * time.Hour
)

type ImportWorker struct {
	db           *sql.DB
	imageService *services.ImageService
	importDir    string
	pollInterval time.Duration
	stopCh       chan struct{}
}

func NewImportWorker(db *sql.DB, imageService *services.ImageService, importDir string, pollInterval time.Duration) *ImportWorker {
	return &ImportWorker{
		db:           db,
		imageService: imageService,
		importDir:    importDir,
		pollInterval: pollInterval,
		stopCh:       make(chan struct{}),
	}
}

func (iw *ImportWorker) Start() {
	go iw.run()
}

func (iw *ImportWorker) Stop() {
	close(iw.stopCh)
}

func (iw *ImportWorker) run() {
	ticker := time.NewTicker(iw.pollInterval)
	defer ticker.Stop()

	log.Info().Msgf("Import worker started, polling every %s", iw.pollInterval)

	iw.process()

	for {
		select {
		case <-ticker.C:
			iw.process()
		case <-iw.stopCh:
			log.Info().Msg("Import worker stopped")
			return
		}
	}
}

func (iw *ImportWorker) process() {
	ctx := context.Background()

	// Reset stale processing rows (stuck > 1 hour).
	_, err := models.Imports(
		models.ImportWhere.Status.EQ(models.ImportStatusProcessing),
		qm.Where("created_at < NOW() - INTERVAL '1 hour'"),
	).UpdateAll(ctx, iw.db, models.M{
		"status":    models.ImportStatusError,
		"error_msg": "processing timed out",
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to reset stale import rows")
	}

	// Process pending imports (up to 10 per tick).
	pending, err := models.Imports(
		models.ImportWhere.Status.EQ(models.ImportStatusPending),
		qm.Limit(10),
	).All(ctx, iw.db)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch pending imports")
		return
	}

	for _, imp := range pending {
		if err := iw.runImport(ctx, imp); err != nil {
			log.Error().Err(err).Str("importId", imp.ID).Msg("Failed to run import")
		}
	}

	// Purge error rows older than the retention period: delete any remaining file and the DB row.
	cutoff := time.Now().UTC().Add(-importErrorRetention)
	errored, err := models.Imports(
		models.ImportWhere.Status.EQ(models.ImportStatusError),
		models.ImportWhere.CreatedAt.LTE(cutoff),
	).All(ctx, iw.db)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch errored imports")
	} else {
		for _, imp := range errored {
			if imp.FilePath.Valid {
				zipPath := filepath.Join(iw.importDir, imp.FilePath.String)
				if err := os.Remove(zipPath); err != nil && !os.IsNotExist(err) {
					log.Error().Err(err).Str("importId", imp.ID).Msg("Failed to remove import zip")
					continue
				}
			}
			if _, err := imp.Delete(ctx, iw.db); err != nil {
				log.Error().Err(err).Str("importId", imp.ID).Msg("Failed to delete errored import row")
			}
		}
	}
}

func (iw *ImportWorker) runImport(ctx context.Context, imp *models.Import) error {
	// Claim the row.
	n, err := models.Imports(
		models.ImportWhere.ID.EQ(imp.ID),
		models.ImportWhere.Status.EQ(models.ImportStatusPending),
	).UpdateAll(ctx, iw.db, models.M{"status": models.ImportStatusProcessing})
	if err != nil {
		return err
	}
	if n == 0 {
		return nil // another worker claimed it
	}

	if !imp.FilePath.Valid {
		_, _ = models.Imports(models.ImportWhere.ID.EQ(imp.ID)).UpdateAll(ctx, iw.db, models.M{
			"status":    models.ImportStatusError,
			"error_msg": "missing file path",
		})
		return nil
	}

	zipPath := filepath.Join(iw.importDir, imp.FilePath.String)

	summary, buildErr := iw.doImport(ctx, imp.OwnerID, zipPath)

	// Always delete the ZIP after processing, regardless of outcome.
	if err := os.Remove(zipPath); err != nil && !os.IsNotExist(err) {
		log.Error().Err(err).Str("importId", imp.ID).Msg("Failed to remove import zip after processing")
	}

	if buildErr != nil {
		_, _ = models.Imports(models.ImportWhere.ID.EQ(imp.ID)).UpdateAll(ctx, iw.db, models.M{
			"status":    models.ImportStatusError,
			"error_msg": buildErr.Error(),
		})
		return buildErr
	}

	_, err = models.Imports(models.ImportWhere.ID.EQ(imp.ID)).UpdateAll(ctx, iw.db, models.M{
		"status":          models.ImportStatusDone,
		"completed_at":    time.Now().UTC(),
		"things_imported": summary.things,
		"lists_imported":  summary.lists,
		"images_imported": summary.images,
	})
	return err
}

type importSummary struct {
	things int
	lists  int
	images int
}

func (iw *ImportWorker) doImport(ctx context.Context, ownerID, zipPath string) (*importSummary, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("invalid zip: %w", err)
	}
	defer r.Close()

	if len(r.File) > importMaxEntries {
		return nil, fmt.Errorf("archive has too many entries (%d)", len(r.File))
	}

	// Validate all entries before extracting anything.
	hasCollection := false
	for _, f := range r.File {
		if err := validateImportEntry(f); err != nil {
			return nil, err
		}
		if f.Name == "collection.json" {
			hasCollection = true
		}
	}
	if !hasCollection {
		return nil, errors.New("archive is missing collection.json")
	}

	// Extract to a temp dir inside importDir.
	tmpDir, err := os.MkdirTemp(iw.importDir, "extract-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	imgDir := filepath.Join(tmpDir, "images")
	if err := os.MkdirAll(imgDir, 0750); err != nil {
		return nil, err
	}

	var collection exportCollection
	totalExtracted := int64(0)

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		switch {
		case f.Name == "collection.json":
			collection, err = extractCollection(f, &totalExtracted)
			if err != nil {
				return nil, err
			}
		case strings.HasPrefix(f.Name, "images/") && path.Dir(f.Name) == "images":
			dst := filepath.Join(imgDir, filepath.Base(f.Name))
			if err := extractImportFile(f, dst, importMaxFileBytes, &totalExtracted); err != nil {
				return nil, fmt.Errorf("extracting %s: %w", f.Name, err)
			}
		}
		if totalExtracted > importMaxExtractedBytes {
			return nil, errors.New("archive exceeds maximum decompressed size")
		}
	}

	if collection.Version != 1 {
		return nil, fmt.Errorf("unsupported collection version %d", collection.Version)
	}

	// Import images: map export file path ("images/<hash>") -> new image ID.
	imageIDMap := make(map[string]string)
	imagesImported := 0
	for _, et := range collection.Things {
		for _, ei := range et.Images {
			if _, ok := imageIDMap[ei.File]; ok {
				continue
			}
			imgPath := filepath.Join(imgDir, filepath.Base(ei.File))
			f, err := os.Open(imgPath)
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return nil, err
			}
			img, err := iw.imageService.CreateImage(ctx, ownerID, ei.Name, f)
			f.Close()
			if err != nil {
				return nil, fmt.Errorf("importing image %s: %w", ei.Name, err)
			}
			imageIDMap[ei.File] = img.ID
			imagesImported++
		}
	}

	// Import things, properties, quantity entries, image associations, and lists.
	thingIDMap := make(map[string]string, len(collection.Things))
	err = utils.Tx(ctx, iw.db, func(tx *sql.Tx) error {
		for _, et := range collection.Things {
			thingID, err := gonanoid.New()
			if err != nil {
				return err
			}

			thing := &models.Thing{
				ID:           thingID,
				Name:         et.Name,
				Description:  et.Description,
				PrivateNote:  et.PrivateNote,
				OwnerID:      ownerID,
				QuantityUnit: et.QuantityUnit,
				SharingState: parseSharingState(et.SharingState),
			}
			if err := thing.Insert(ctx, tx, boil.Infer()); err != nil {
				return err
			}

			for _, ep := range et.Properties {
				prop, err := importPropertyParams(ep)
				if err != nil {
					continue // skip properties with unrecognised types or malformed values
				}
				if _, err := operations.CreateProperty(ctx, tx, thingID, prop); err != nil {
					return err
				}
			}

			quantityID, err := gonanoid.New()
			if err != nil {
				return err
			}
			if err := thing.AddQuantityEntries(ctx, tx, true, &models.QuantityEntry{
				ID:         quantityID,
				DeltaValue: et.Quantity,
			}); err != nil {
				return err
			}

			imageThings := make([]*models.ImagesThing, 0, len(et.Images))
			for pos, ei := range et.Images {
				newID, ok := imageIDMap[ei.File]
				if !ok {
					continue
				}
				imageThings = append(imageThings, &models.ImagesThing{
					Pos:     pos,
					ImageID: newID,
				})
			}
			if len(imageThings) > 0 {
				if err := thing.AddImagesThings(ctx, tx, true, imageThings...); err != nil {
					return err
				}
			}

			thingIDMap[et.ID] = thingID
		}

		for _, el := range collection.Lists {
			listID, err := gonanoid.New()
			if err != nil {
				return err
			}

			list := &models.List{
				ID:           listID,
				Name:         el.Name,
				OwnerID:      ownerID,
				SharingState: parseSharingState(el.SharingState),
			}
			if err := list.Insert(ctx, tx, boil.Infer()); err != nil {
				return err
			}

			for _, oldThingID := range el.ThingIDs {
				newThingID, ok := thingIDMap[oldThingID]
				if !ok {
					continue
				}
				thing := &models.Thing{ID: newThingID}
				if err := thing.AddLists(ctx, tx, false, list); err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		for _, imageID := range imageIDMap {
			if _, delErr := iw.imageService.DeleteImage(ctx, ownerID, imageID); delErr != nil {
				log.Error().Err(delErr).Str("imageId", imageID).Msg("Failed to delete image after failed import transaction")
			}
		}
		return nil, err
	}

	return &importSummary{
		things: len(collection.Things),
		lists:  len(collection.Lists),
		images: imagesImported,
	}, nil
}

func validateImportEntry(f *zip.File) error {
	name := f.Name
	if strings.Contains(name, "..") || strings.HasPrefix(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("unsafe path in archive: %s", name)
	}
	if f.CompressedSize64 > 0 && f.UncompressedSize64/f.CompressedSize64 > importMaxComprRatio {
		return fmt.Errorf("suspicious compression ratio for %s", name)
	}
	return nil
}

func extractCollection(f *zip.File, totalExtracted *int64) (exportCollection, error) {
	rc, err := f.Open()
	if err != nil {
		return exportCollection{}, err
	}
	defer rc.Close()

	limited := io.LimitReader(rc, importMaxJsonBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return exportCollection{}, err
	}
	if int64(len(data)) > importMaxJsonBytes {
		return exportCollection{}, errors.New("collection.json exceeds maximum size")
	}
	*totalExtracted += int64(len(data))

	var col exportCollection
	if err := json.Unmarshal(data, &col); err != nil {
		return exportCollection{}, fmt.Errorf("invalid collection.json: %w", err)
	}
	return col, nil
}

func extractImportFile(f *zip.File, dst string, maxBytes int64, totalExtracted *int64) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	limited := io.LimitReader(rc, maxBytes+1)
	n, err := io.Copy(out, limited)
	if err != nil {
		os.Remove(dst)
		return err
	}
	if n > maxBytes {
		os.Remove(dst)
		return fmt.Errorf("file exceeds maximum size of %d bytes", maxBytes)
	}
	*totalExtracted += n
	return nil
}

func parseSharingState(s string) models.SharingState {
	switch s {
	case "friends":
		return models.SharingStateFriends
	case "friends-of-friends":
		return models.SharingStateFriendsOfFriends
	default:
		return models.SharingStatePrivate
	}
}

func importPropertyParams(ep exportProperty) (operations.CreatePropertyParams, error) {
	switch ep.Type {
	case "float":
		v, ok := ep.Value.(float64)
		if !ok {
			return nil, fmt.Errorf("invalid float value for property %q", ep.Name)
		}
		p := operations.CreatePropertyFloatParams{Name: ep.Name, Value: v}
		if ep.Unit != "" {
			p.Unit = &ep.Unit
		}
		return p, nil
	case "string":
		v, _ := ep.Value.(string)
		return operations.CreatePropertyStringParams{Name: ep.Name, Value: v}, nil
	case "datetime":
		s, ok := ep.Value.(string)
		if !ok {
			return nil, fmt.Errorf("invalid datetime value for property %q", ep.Name)
		}
		t, err := time.Parse(time.RFC3339Nano, s)
		if err != nil {
			t, err = time.Parse(time.RFC3339, s)
			if err != nil {
				return nil, fmt.Errorf("unparseable datetime for property %q: %w", ep.Name, err)
			}
		}
		return operations.CreatePropertyDatetimeParams{Name: ep.Name, Value: t}, nil
	case "boolean":
		v, _ := ep.Value.(bool)
		return operations.CreatePropertyBooleanParams{Name: ep.Name, Value: v}, nil
	}
	return nil, fmt.Errorf("unknown property type %q", ep.Type)
}
