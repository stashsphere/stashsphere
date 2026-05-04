package operations

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/stashsphere/backend/models"
)

func GetOwnedThingsForExport(ctx context.Context, exec boil.ContextExecutor, userId string) (models.ThingSlice, error) {
	return models.Things(
		models.ThingWhere.OwnerID.EQ(userId),
		qm.Load(models.ThingRels.Properties),
		qm.Load(models.ThingRels.QuantityEntries),
		qm.Load(qm.Rels(models.ThingRels.ImagesThings, models.ImagesThingRels.Image)),
	).All(ctx, exec)
}

func GetOwnedListsForExport(ctx context.Context, exec boil.ContextExecutor, userId string) (models.ListSlice, error) {
	return models.Lists(
		models.ListWhere.OwnerID.EQ(userId),
		qm.Load(models.ListRels.Things),
	).All(ctx, exec)
}
