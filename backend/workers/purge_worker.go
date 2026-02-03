package workers

import (
	"context"
	"database/sql"
	"time"

	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/rs/zerolog/log"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/utils"
)

type PurgeWorker struct {
	db             *sql.DB
	imageStorePath string
	pollInterval   time.Duration
	stopCh         chan struct{}
}

func NewPurgeWorker(db *sql.DB, imageStorePath string, pollInterval time.Duration) *PurgeWorker {
	return &PurgeWorker{
		db:             db,
		imageStorePath: imageStorePath,
		pollInterval:   pollInterval,
		stopCh:         make(chan struct{}),
	}
}

func (pw *PurgeWorker) Start() {
	go pw.run()
}

func (pw *PurgeWorker) Stop() {
	close(pw.stopCh)
}

func (pw *PurgeWorker) run() {
	ticker := time.NewTicker(pw.pollInterval)
	defer ticker.Stop()

	log.Info().Msgf("Purge worker started, polling every %s", pw.pollInterval)

	// Run immediately on start
	pw.processPendingPurges()

	for {
		select {
		case <-ticker.C:
			pw.processPendingPurges()
		case <-pw.stopCh:
			log.Info().Msg("Purge worker stopped")
			return
		}
	}
}

func (pw *PurgeWorker) processPendingPurges() {
	ctx := context.Background()

	users, err := models.Users(
		models.UserWhere.PurgeAt.IsNotNull(),
		qm.Where("purge_at <= CURRENT_TIMESTAMP"),
	).All(ctx, pw.db)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get users pending purge")
		return
	}

	for _, user := range users {
		log.Info().Str("userId", user.ID).Msg("Purging user account")

		err := utils.Tx(ctx, pw.db, func(tx *sql.Tx) error {
			return operations.PurgeUser(ctx, tx, user.ID, pw.imageStorePath)
		})
		if err != nil {
			log.Error().Err(err).Str("userId", user.ID).Msg("Failed to purge user")
			continue
		}

		log.Info().Str("userId", user.ID).Msg("User account purged successfully")
	}
}
