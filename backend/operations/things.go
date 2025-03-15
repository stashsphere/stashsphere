package operations

import (
	"context"

	"github.com/stashsphere/backend/models"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func GetThingUnchecked(ctx context.Context, exec boil.ContextExecutor, thingId string) (*models.Thing, error) {
	thing, err := models.Things(
		qm.Load(models.ThingRels.Properties),
		qm.Load(qm.Rels(models.ThingRels.Lists, models.ListRels.Owner)),
		qm.Load(models.ThingRels.Owner),
		qm.Load(models.ThingRels.ThingImages),
		qm.Load(models.ThingRels.QuantityEntries),
		qm.Load(qm.Rels(models.ThingRels.Shares, models.ShareRels.Owner)),
		qm.Load(qm.Rels(models.ThingRels.Shares, models.ShareRels.TargetUser)),
		models.ThingWhere.ID.EQ(thingId)).One(ctx, exec)
	if err != nil {
		return nil, err
	}
	return thing, nil
}

func GetSharedThingIdsForUser(ctx context.Context, exec boil.ContextExecutor, userId string) ([]string, error) {
	shares, err := models.Shares(
		qm.Load(qm.Rels(models.ShareRels.Things)),
		qm.Load(qm.Rels(models.ShareRels.Lists, models.ListRels.Things)),
		models.ShareWhere.TargetUserID.EQ(userId)).All(ctx, exec)
	if err != nil {
		return nil, err
	}
	thingIds := make(map[string]bool)
	for _, share := range shares {
		for _, thing := range share.R.Things {
			thingIds[thing.ID] = true
		}
		for _, list := range share.R.Lists {
			for _, thing := range list.R.Things {
				thingIds[thing.ID] = true
			}
		}
	}
	res := make([]string, len(thingIds))
	i := 0
	for key, _ := range thingIds {
		res[i] = key
		i++
	}
	return res, nil
}

func SumQuantity(thing *models.Thing) int64 {
	currentQuantity := int64(0)
	for _, x := range thing.R.QuantityEntries {
		currentQuantity += int64(x.DeltaValue)
	}
	return currentQuantity
}

func DeltaQuantity(thing *models.Thing, target uint64) int64 {
	return int64(target) - SumQuantity(thing)
}
