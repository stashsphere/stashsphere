package operations

import (
	"context"
	"os"
	"path/filepath"

	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/utils"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func ImageBelongsToUser(ctx context.Context, exec boil.ContextExecutor, userId string, imageId string) (bool, error) {
	image, err := models.Images(models.ImageWhere.ID.EQ(imageId)).One(ctx, exec)
	if err != nil {
		return false, err
	}
	return image.OwnerID == userId, nil
}

func GetSharedImageIdsForUser(ctx context.Context, exec boil.ContextExecutor, userId string) ([]string, error) {
	thingIds, err := GetSharedThingIdsForUser(ctx, exec, userId)
	if err != nil {
		return nil, err
	}
	things, err := models.Things(
		qm.Load(qm.Rels(models.ThingRels.ThingImages)),
		models.ThingWhere.ID.IN(thingIds)).All(ctx, exec)
	if err != nil {
		return nil, err
	}
	imageIds := make(map[string]bool)
	for _, thing := range things {
		for _, image := range thing.R.ThingImages {
			imageIds[image.ID] = true
		}
	}
	res := make([]string, len(imageIds))
	i := 0
	for key, _ := range imageIds {
		res[i] = key
		i++
	}
	return res, nil
}

func DeleteImage(ctx context.Context, exec boil.ContextExecutor, userId string, imageId string) (*models.Image, error) {
	image, err := models.Images(
		models.ImageWhere.ID.EQ(imageId),
		qm.Load(models.ImageRels.ImageThings),
		qm.Load(models.ImageRels.Owner),
	).One(ctx, exec)
	if err != nil {
		return nil, err
	}
	if image.OwnerID != userId {
		return nil, utils.ErrEntityDoesNotBelongToUser
	}
	if len(image.R.ImageThings) > 0 {
		return nil, utils.ErrEntityInUse
	}
	_, err = image.Delete(ctx, exec)
	if err != nil {
		return nil, err
	}
	return image, nil
}

func DeleteContent(ctx context.Context, exec boil.ContextExecutor, storePath string, contentId string) error {
	imagesWithHash, err := models.Images(models.ImageWhere.Hash.EQ(contentId)).Count(ctx, exec)
	if err != nil {
		return err
	}
	if imagesWithHash > 0 {
		return utils.ErrEntityInUse
	}
	path := filepath.Join(storePath, contentId)
	err = os.Remove(path)
	if err != nil {
		return err
	}
	return nil
}
