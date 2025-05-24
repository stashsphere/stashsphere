package operations

import (
	"context"

	"github.com/stashsphere/backend/models"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func GetListUnchecked(ctx context.Context, exec boil.ContextExecutor, listId string) (*models.List, error) {
	list, err := models.Lists(
		models.ListWhere.ID.EQ(listId),
		qm.Load(models.ListRels.Owner),
		qm.Load(qm.Rels(models.ListRels.Things, models.ThingRels.Owner)),
		qm.Load(qm.Rels(models.ListRels.Things, models.ThingRels.Images)),
		qm.Load(qm.Rels(models.ListRels.Things, models.ThingRels.QuantityEntries)),
		qm.Load(qm.Rels(models.ListRels.Things, models.ThingRels.Properties)),
		qm.Load(qm.Rels(models.ListRels.Shares, models.ShareRels.Owner)),
		qm.Load(qm.Rels(models.ListRels.Shares, models.ShareRels.TargetUser)),
	).One(ctx, exec)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func GetSharedListIdsForUser(ctx context.Context, exec boil.ContextExecutor, userId string) ([]string, error) {
	shares, err := models.Shares(
		qm.Load(qm.Rels(models.ShareRels.Lists)),
		models.ShareWhere.TargetUserID.EQ(userId)).All(ctx, exec)
	if err != nil {
		return nil, err
	}
	listIds := make(map[string]bool)
	for _, share := range shares {
		for _, list := range share.R.Lists {
			listIds[list.ID] = true
		}
	}
	res := make([]string, len(listIds))
	i := 0
	for key, _ := range listIds {
		res[i] = key
		i++
	}
	return res, nil
}
