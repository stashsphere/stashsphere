package operations

import (
	"bytes"
	"context"
	"image"
	"io"
	"os"
	"path/filepath"

	"image/jpeg"
	_ "image/jpeg"
	"image/png"
	_ "image/png"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/disintegration/imaging"
	exifremove "github.com/neurosnap/go-exif-remove"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/utils"
)

type Rotation int

const (
	Rotation90  Rotation = iota
	Rotation180          = iota
	Rotation270          = iota
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
		qm.Load(qm.Rels(models.ThingRels.ImagesThings)),
		models.ThingWhere.ID.IN(thingIds)).All(ctx, exec)
	if err != nil {
		return nil, err
	}
	imageIds := make(map[string]bool)
	for _, thing := range things {
		for _, image := range thing.R.ImagesThings {
			imageIds[image.ImageID] = true
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
		qm.Load(models.ImageRels.ImagesThings),
		qm.Load(models.ImageRels.Owner),
		qm.Load(models.ImageRels.Profiles),
	).One(ctx, exec)
	if err != nil {
		return nil, err
	}
	if image.OwnerID != userId {
		return nil, utils.EntityDoesNotBelongToUserError{}
	}
	if len(image.R.ImagesThings) > 0 {
		return nil, utils.EntityInUseError{}
	}
	if len(image.R.Profiles) > 0 {
		return nil, utils.EntityInUseError{}
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
		return utils.EntityInUseError{}
	}
	path := filepath.Join(storePath, contentId)
	err = os.Remove(path)
	if err != nil {
		return err
	}
	return nil
}

// input is current path
func ClearExifData(path string) ([]byte, error) {
	imgFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer imgFile.Close()
	imgBytes, err := io.ReadAll(imgFile)
	if err != nil {
		return nil, err
	}
	// TODO: The picture needs to be either rotated by the user
	// or this function needs to rotate according to the exif data,
	// however this might already be wrong, so the application
	// could provide such an endpoint /rotate?by=[90,180,270]
	// TODO: this does not remove jpeg comments, not important
	removed, err := exifremove.Remove(imgBytes)
	if err != nil {
		return nil, err
	}
	return removed, nil
}

func RotateImage(path string, rotation Rotation) ([]byte, error) {
	imgFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	img, codec, err := image.Decode(imgFile)
	if err != nil {
		return nil, err
	}
	var rotated *image.NRGBA
	switch rotation {
	case Rotation90:
		rotated = imaging.Rotate90(img)
	case Rotation180:
		rotated = imaging.Rotate180(img)
	case Rotation270:
		rotated = imaging.Rotate270(img)
	}
	var b bytes.Buffer
	if codec == "jpeg" {
		err = jpeg.Encode(&b, rotated, &jpeg.Options{
			Quality: 90,
		})
		if err != nil {
			return nil, err
		}
	} else {
		err = png.Encode(&b, rotated)
		if err != nil {
			return nil, err
		}
	}
	return b.Bytes(), nil
}

func ResizeImage(imgFile io.Reader, width int) (io.Reader, error) {
	img, codec, err := image.Decode(imgFile)
	if err != nil {
		return nil, err
	}
	resized := imaging.Resize(img, width, 0, imaging.CatmullRom)
	var b bytes.Buffer
	if codec == "png" {
		err = png.Encode(&b, resized)
		if err != nil {
			return nil, err
		}
	} else {
		err = jpeg.Encode(&b, resized, &jpeg.Options{
			Quality: 90,
		})
		if err != nil {
			return nil, err
		}
	}
	return &b, err
}
