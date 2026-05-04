package services

import (
	"context"
	"database/sql"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/utils"
)

type ExportService struct {
	db *sql.DB
}

func NewExportService(db *sql.DB) *ExportService {
	return &ExportService{db}
}

func (es *ExportService) CreateExport(ctx context.Context, ownerId string) (*models.Export, error) {
	var result *models.Export
	err := utils.Tx(ctx, es.db, func(tx *sql.Tx) error {
		pending, err := models.Exports(
			models.ExportWhere.OwnerID.EQ(ownerId),
			models.ExportWhere.Status.IN([]models.ExportStatus{
				models.ExportStatusPending,
				models.ExportStatusProcessing,
			}),
		).Count(ctx, tx)
		if err != nil {
			return err
		}
		if pending > 0 {
			return utils.ExportAlreadyInProgressError{}
		}

		exportId, err := gonanoid.New()
		if err != nil {
			return err
		}

		export := models.Export{
			ID:      exportId,
			OwnerID: ownerId,
			Status:  models.ExportStatusPending,
		}
		if err := export.Insert(ctx, tx, boil.Infer()); err != nil {
			return err
		}
		result = &export
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (es *ExportService) GetExport(ctx context.Context, ownerId string) (*models.Export, error) {
	export, err := models.Exports(
		models.ExportWhere.OwnerID.EQ(ownerId),
		qm.OrderBy("created_at DESC"),
		qm.Limit(1),
	).One(ctx, es.db)
	if err == sql.ErrNoRows {
		return nil, utils.NotFoundError{EntityName: "export"}
	}
	return export, err
}

func (es *ExportService) GetDoneExport(ctx context.Context, ownerId string) (*models.Export, error) {
	export, err := models.Exports(
		models.ExportWhere.OwnerID.EQ(ownerId),
		models.ExportWhere.Status.EQ(models.ExportStatusDone),
		qm.OrderBy("created_at DESC"),
		qm.Limit(1),
	).One(ctx, es.db)
	if err == sql.ErrNoRows {
		return nil, utils.NotFoundError{EntityName: "export"}
	}
	return export, err
}
