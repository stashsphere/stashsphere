package operations

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries"
)

func GetFriendIds(ctx context.Context, exec boil.ContextExecutor, userId string) ([]string, error) {
	friendIds := make([]string, 0)
	type IdRow struct {
		Id string `boil:"id"`
	}
	var idRows []IdRow
	err := queries.Raw(
		`SELECT 
		CASE WHEN friend1_id=$1 THEN friend2_id ELSE friend1_id END AS id
		FROM friendships
		WHERE friend1_id=$1 OR friend2_id=$1`, userId,
	).Bind(ctx, exec, &idRows)
	if err != nil {
		return nil, err
	}
	for _, id := range idRows {
		friendIds = append(friendIds, id.Id)
	}
	return friendIds, nil
}
