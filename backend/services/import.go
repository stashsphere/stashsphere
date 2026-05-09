package services

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/utils"
)

type ImportService struct {
	db        *sql.DB
	importDir string // tmpPath/imports — holds both in-progress uploads and queued ZIPs
}

func NewImportService(db *sql.DB, importDir string) (*ImportService, error) {
	if _, err := os.Stat(importDir); err != nil {
		return nil, err
	}
	return &ImportService{db: db, importDir: importDir}, nil
}

// QueueImport copies src into a temp file inside importDir, enforces the upload
// size limit, checks for a concurrent import, then renames the file to its final
// name and inserts a pending import row.
func (is *ImportService) QueueImport(ctx context.Context, ownerId string, src io.Reader, maxBytes int64) (*models.Import, error) {
	tmp, err := os.CreateTemp(is.importDir, "upload-*.zip")
	if err != nil {
		return nil, err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	limited := io.LimitReader(src, maxBytes+1)
	n, err := io.Copy(tmp, limited)
	tmp.Close()
	if err != nil {
		return nil, utils.InvalidImportFileError{Msg: "upload failed: " + err.Error()}
	}
	if n > maxBytes {
		return nil, utils.InvalidImportFileError{Msg: fmt.Sprintf("uploaded file exceeds maximum size of %d MB", maxBytes/1024/1024)}
	}

	var result *models.Import
	err = utils.Tx(ctx, is.db, func(tx *sql.Tx) error {
		pending, err := models.Imports(
			models.ImportWhere.OwnerID.EQ(ownerId),
			models.ImportWhere.Status.IN([]models.ImportStatus{
				models.ImportStatusPending,
				models.ImportStatusProcessing,
			}),
		).Count(ctx, tx)
		if err != nil {
			return err
		}
		if pending > 0 {
			return utils.ImportAlreadyInProgressError{}
		}

		importID, err := gonanoid.New()
		if err != nil {
			return err
		}

		fileName := importID + ".zip"
		imp := models.Import{
			ID:       importID,
			OwnerID:  ownerId,
			Status:   models.ImportStatusPending,
			FilePath: null.StringFrom(fileName),
		}
		if err := imp.Insert(ctx, tx, boil.Infer()); err != nil {
			return err
		}
		result = &imp
		return nil
	})
	if err != nil {
		return nil, err
	}

	destPath := filepath.Join(is.importDir, result.FilePath.String)
	if err := os.Rename(tmpPath, destPath); err != nil {
		return nil, err
	}

	return result, nil
}

func (is *ImportService) GetImport(ctx context.Context, ownerId string) (*models.Import, error) {
	imp, err := models.Imports(
		models.ImportWhere.OwnerID.EQ(ownerId),
		qm.OrderBy("created_at DESC"),
		qm.Limit(1),
	).One(ctx, is.db)
	if err == sql.ErrNoRows {
		return nil, utils.NotFoundError{EntityName: "import"}
	}
	return imp, err
}
