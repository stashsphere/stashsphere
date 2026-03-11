package workers

import (
	"context"
	"database/sql"
	"path/filepath"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/rs/zerolog/log"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
)

type ImageSizeWorker struct {
	db             *sql.DB
	imageStorePath string
}

func NewImageSizeWorker(db *sql.DB, imageStorePath string) *ImageSizeWorker {
	return &ImageSizeWorker{
		db:             db,
		imageStorePath: imageStorePath,
	}
}

func (iw *ImageSizeWorker) Start() {
	go iw.run()
}

func (iw *ImageSizeWorker) run() {
	log.Info().Msg("Starting ImageSizeWorker")
	ctx := context.Background()

	images, err := models.Images(
		models.ImageWhere.Height.IsNull(),
		qm.Or2(
			models.ImageWhere.Width.IsNull(),
		),
	).All(ctx, iw.db)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get images without height / width")
		return
	}
	for _, image := range images {
		path := filepath.Join(iw.imageStorePath, image.Hash)
		width, height, err := operations.GetImageSizeFromPath(path)
		if err != nil {
			log.Info().Err(err).Msgf("Failed to determine height / width for image id %s / image hash %s", image.ID, image.Hash)
			continue
		}
		image.Height = null.IntFrom(height)
		image.Width = null.IntFrom(width)
		_, err = image.Update(ctx, iw.db, boil.Whitelist(models.ImageColumns.Height, models.ImageColumns.Width))
		if err != nil {
			log.Error().Err(err).Msgf("Failed to update width / height in database for image.ID %s", image.ID)
		}
	}
}
