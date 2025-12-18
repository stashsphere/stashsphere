package operations

import (
	"context"
	"database/sql"
	"errors"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/utils"
)

func FindUserByID(ctx context.Context, exec boil.ContextExecutor, userId string) (*models.User, error) {
	user, err := models.Users(models.UserWhere.ID.EQ(userId)).One(ctx, exec)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.NotFoundError{EntityName: "user"}
		}
		return nil, err
	}
	return user, nil
}

func FindUserWithProfileByID(ctx context.Context, exec boil.ContextExecutor, userId string) (*models.User, error) {
	user, err := models.Users(models.UserWhere.ID.EQ(userId),
		qm.Load(models.UserRels.Profile),
		qm.Load(qm.Rels(models.UserRels.Profile, models.ProfileRels.Image)),
	).One(ctx, exec)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.NotFoundError{EntityName: "user"}
		}
		return nil, err
	}
	return user, nil
}
