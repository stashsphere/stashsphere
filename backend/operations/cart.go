package operations

import (
	"context"
	"errors"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/utils"
)

func GetCart(ctx context.Context, exec boil.ContextExecutor, userId string) (models.CartEntrySlice, error) {
	return models.CartEntries(models.CartEntryWhere.UserID.EQ(userId), qm.Load(models.CartEntryRels.Thing)).All(ctx, exec)
}

func UpdateCart(ctx context.Context, exec boil.ContextExecutor, thingIds []string, userId string) error {
	_, err := models.CartEntries(models.CartEntryWhere.UserID.EQ(userId), models.CartEntryWhere.ThingID.NIN(thingIds)).DeleteAll(ctx, exec)
	if err != nil {
		return err
	}
	existingEntriesMap := make(map[string]bool)
	existingEntries, err := models.CartEntries(qm.Select(models.CartEntryColumns.ThingID), models.CartEntryWhere.UserID.EQ(userId)).All(ctx, exec)
	if err != nil {
		return err
	}
	for _, existingEntry := range existingEntries {
		existingEntriesMap[existingEntry.ThingID] = true
	}
	for _, thingId := range thingIds {
		_, err := GetThingChecked(ctx, exec, thingId, userId)
		if err != nil {
			return err
		}
		_, exists := existingEntriesMap[thingId]
		if !exists {
			entry := models.CartEntry{
				UserID:  userId,
				ThingID: thingId,
			}
			err = entry.Insert(ctx, exec, boil.Infer())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func RemoveForbiddenThingsFromCarts(ctx context.Context, exec boil.ContextExecutor, thingIds []string) error {
	entries, err := models.CartEntries(models.CartEntryWhere.ThingID.IN(thingIds)).All(ctx, exec)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		_, err := GetThingChecked(ctx, exec, entry.ThingID, entry.UserID)
		if err != nil {
			if errors.Is(err, utils.UserHasNoAccessRightsError{}) {
				_, err = entry.Delete(ctx, exec)
				if err != nil {
					return err
				}
				continue
			} else {
				return err
			}
		}
	}
	return nil
}
