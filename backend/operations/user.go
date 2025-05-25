package operations

import (
	"context"
	"database/sql"
	"errors"

	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/utils"
	"github.com/volatiletech/sqlboiler/v4/boil"
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
