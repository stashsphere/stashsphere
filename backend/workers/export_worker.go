package workers

import (
	"archive/zip"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/rs/zerolog/log"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
)

type ExportWorker struct {
	db           *sql.DB
	imageStore   string
	storePath    string
	tmpPath      string
	retention    time.Duration
	pollInterval time.Duration
	stopCh       chan struct{}
}

func NewExportWorker(db *sql.DB, imageStore, storePath, tmpPath string, retention, pollInterval time.Duration) *ExportWorker {
	return &ExportWorker{
		db:           db,
		imageStore:   imageStore,
		storePath:    storePath,
		tmpPath:      tmpPath,
		retention:    retention,
		pollInterval: pollInterval,
		stopCh:       make(chan struct{}),
	}
}

func (ew *ExportWorker) Start() {
	go ew.run()
}

func (ew *ExportWorker) Stop() {
	close(ew.stopCh)
}

func (ew *ExportWorker) run() {
	ticker := time.NewTicker(ew.pollInterval)
	defer ticker.Stop()

	log.Info().Msgf("Export worker started, polling every %s", ew.pollInterval)

	ew.process()

	for {
		select {
		case <-ticker.C:
			ew.process()
		case <-ew.stopCh:
			log.Info().Msg("Export worker stopped")
			return
		}
	}
}

func (ew *ExportWorker) process() {
	ctx := context.Background()

	// Reset stale processing rows (stuck > 1 hour).
	_, err := models.Exports(
		models.ExportWhere.Status.EQ(models.ExportStatusProcessing),
		qm.Where("created_at < NOW() - INTERVAL '1 hour'"),
	).UpdateAll(ctx, ew.db, models.M{
		"status":    models.ExportStatusError,
		"error_msg": "processing timed out",
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to reset stale export rows")
	}

	// Build pending exports (up to 10 per tick).
	pending, err := models.Exports(
		models.ExportWhere.Status.EQ(models.ExportStatusPending),
		qm.Limit(10),
	).All(ctx, ew.db)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch pending exports")
		return
	}

	for _, export := range pending {
		if err := ew.buildExport(ctx, export); err != nil {
			log.Error().Err(err).Str("exportId", export.ID).Msg("Failed to build export")
		}
	}

	// Purge expired done exports (file + row).
	doneExpired, err := models.Exports(
		models.ExportWhere.Status.EQ(models.ExportStatusDone),
		qm.Where("expires_at <= NOW()"),
	).All(ctx, ew.db)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch expired exports")
	} else {
		for _, export := range doneExpired {
			if export.FilePath.Valid {
				if err := os.Remove(filepath.Join(ew.storePath, export.FilePath.String)); err != nil && !os.IsNotExist(err) {
					log.Error().Err(err).Str("exportId", export.ID).Msg("Failed to remove export file")
					continue
				}
			}
			if _, err := export.Delete(ctx, ew.db); err != nil {
				log.Error().Err(err).Str("exportId", export.ID).Msg("Failed to delete expired export row")
			}
		}
	}

	// Purge error rows older than the retention period (no file to remove).
	cutoff := time.Now().UTC().Add(-ew.retention)
	errored, err := models.Exports(
		models.ExportWhere.Status.EQ(models.ExportStatusError),
		models.ExportWhere.CreatedAt.LTE(cutoff),
	).All(ctx, ew.db)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch errored exports")
	} else {
		for _, export := range errored {
			if _, err := export.Delete(ctx, ew.db); err != nil {
				log.Error().Err(err).Str("exportId", export.ID).Msg("Failed to delete errored export row")
			}
		}
	}
}

func (ew *ExportWorker) buildExport(ctx context.Context, export *models.Export) error {
	// Claim the row.
	n, err := models.Exports(
		models.ExportWhere.ID.EQ(export.ID),
		models.ExportWhere.Status.EQ(models.ExportStatusPending),
	).UpdateAll(ctx, ew.db, models.M{"status": models.ExportStatusProcessing})
	if err != nil {
		return err
	}
	if n == 0 {
		return nil // another worker claimed it
	}

	zipFile := export.ID + ".zip"
	zipPath := filepath.Join(ew.storePath, zipFile)

	buildErr := ew.doBuild(ctx, export, zipPath)
	if buildErr != nil {
		os.Remove(zipPath)
		_, _ = models.Exports(models.ExportWhere.ID.EQ(export.ID)).UpdateAll(ctx, ew.db, models.M{
			"status":    models.ExportStatusError,
			"error_msg": buildErr.Error(),
		})
		return buildErr
	}

	expiresAt := time.Now().UTC().Add(ew.retention)
	_, err = models.Exports(models.ExportWhere.ID.EQ(export.ID)).UpdateAll(ctx, ew.db, models.M{
		"status":     models.ExportStatusDone,
		"file_path":  zipFile,
		"expires_at": expiresAt,
	})
	return err
}

func (ew *ExportWorker) doBuild(ctx context.Context, export *models.Export, zipPath string) error {
	tmpDir, err := os.MkdirTemp(ew.tmpPath, "export-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	things, err := operations.GetOwnedThingsForExport(ctx, ew.db, export.OwnerID)
	if err != nil {
		return err
	}
	lists, err := operations.GetOwnedListsForExport(ctx, ew.db, export.OwnerID)
	if err != nil {
		return err
	}

	// Copy unique image files into tmpDir/images/.
	if err := os.MkdirAll(filepath.Join(tmpDir, "images"), 0750); err != nil {
		return err
	}
	seen := map[string]struct{}{}
	for _, thing := range things {
		for _, it := range thing.R.ImagesThings {
			hash := it.R.Image.Hash
			if _, ok := seen[hash]; ok {
				continue
			}
			seen[hash] = struct{}{}
			src := filepath.Join(ew.imageStore, hash)
			dst := filepath.Join(tmpDir, "images", hash)
			if err := copyFile(src, dst); err != nil {
				return err
			}
		}
	}

	// Build collection.json.
	collection := buildCollection(things, lists, export)
	collectionData, err := json.Marshal(collection)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "collection.json"), collectionData, 0644); err != nil {
		return err
	}

	// Assemble ZIP.
	zf, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zf.Close()

	zw := zip.NewWriter(zf)
	defer zw.Close()

	return filepath.WalkDir(tmpDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(tmpDir, path)
		if err != nil {
			return err
		}
		w, err := zw.Create(rel)
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		return err
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

type exportCollection struct {
	Version    int           `json:"version"`
	ExportedAt time.Time     `json:"exportedAt"`
	Things     []exportThing `json:"things"`
	Lists      []exportList  `json:"lists"`
}

type exportThing struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	PrivateNote  string           `json:"privateNote"`
	CreatedAt    time.Time        `json:"createdAt"`
	SharingState string           `json:"sharingState"`
	Quantity     int64            `json:"quantity"`
	QuantityUnit string           `json:"quantityUnit"`
	Images       []exportImage    `json:"images"`
	Properties   []exportProperty `json:"properties"`
}

type exportList struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"createdAt"`
	SharingState string    `json:"sharingState"`
	ThingIDs     []string  `json:"thingIds"`
}

type exportImage struct {
	Name string `json:"name"`
	Mime string `json:"mime"`
	File string `json:"file"`
}

type exportProperty struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value any    `json:"value"`
	Unit  string `json:"unit,omitempty"`
}

func buildCollection(things models.ThingSlice, lists models.ListSlice, export *models.Export) exportCollection {
	et := make([]exportThing, 0, len(things))
	for _, t := range things {
		images := make([]exportImage, 0, len(t.R.ImagesThings))
		for _, it := range t.R.ImagesThings {
			images = append(images, exportImage{
				Name: it.R.Image.Name,
				Mime: it.R.Image.Mime,
				File: "images/" + it.R.Image.Hash,
			})
		}
		props := make([]exportProperty, 0, len(t.R.Properties))
		for _, p := range t.R.Properties {
			var val any
			switch p.Type {
			case models.PropertyTypeFloat:
				val = p.ValueFloat.Float64
			case models.PropertyTypeDatetime:
				val = p.ValueDatetime.Time
			case models.PropertyTypeBoolean:
				val = p.ValueBoolean.Bool
			default:
				val = p.ValueString.String
			}
			ep := exportProperty{
				Type:  string(p.Type),
				Name:  p.Name,
				Value: val,
			}
			if p.Unit.Valid {
				ep.Unit = p.Unit.String
			}
			props = append(props, ep)
		}
		et = append(et, exportThing{
			ID:           t.ID,
			Name:         t.Name,
			Description:  t.Description,
			PrivateNote:  t.PrivateNote,
			CreatedAt:    t.CreatedAt,
			SharingState: string(t.SharingState),
			Quantity:     operations.SumQuantity(t),
			QuantityUnit: t.QuantityUnit,
			Images:       images,
			Properties:   props,
		})
	}

	el := make([]exportList, 0, len(lists))
	for _, l := range lists {
		thingIds := make([]string, 0, len(l.R.Things))
		for _, t := range l.R.Things {
			thingIds = append(thingIds, t.ID)
		}
		el = append(el, exportList{
			ID:           l.ID,
			Name:         l.Name,
			CreatedAt:    l.CreatedAt,
			SharingState: string(l.SharingState),
			ThingIDs:     thingIds,
		})
	}

	return exportCollection{
		Version:    1,
		ExportedAt: time.Now().UTC(),
		Things:     et,
		Lists:      el,
	}
}
