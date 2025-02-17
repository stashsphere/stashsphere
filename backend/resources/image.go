package resources

import (
	"time"

	"github.com/stashsphere/backend/models"
)

type ReducedImage struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Mime      string    `json:"mime"`
	CreatedAt time.Time `json:"createdAt"`
	Owner     User      `json:"owner"`
	Hash      string    `json:"hash"`
}

func ReducedImageFromModel(image *models.Image) ReducedImage {
	return ReducedImage{
		ID:        image.ID,
		Name:      image.Name,
		Mime:      image.Mime,
		CreatedAt: image.CreatedAt,
		Hash:      image.Hash,
	}
}

func ReducedImagesFromModelSlice(images models.ImageSlice) []ReducedImage {
	res := make([]ReducedImage, len(images))
	for idx, image := range images {
		res[idx] = ReducedImageFromModel(image)
	}
	return res
}

type ImageActions struct {
	CanDelete bool `json:"canDelete"`
}

type Image struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Mime      string         `json:"mime"`
	CreatedAt time.Time      `json:"createdAt"`
	Owner     User           `json:"owner"`
	Hash      string         `json:"hash"`
	Actions   ImageActions   `json:"actions"`
	Things    []ReducedThing `json:"things"`
}

func ImageFromModel(image *models.Image, userId string) Image {
	canDelete := userId == image.OwnerID && len(image.R.ImageThings) == 0

	return Image{
		ID:        image.ID,
		Name:      image.Name,
		Mime:      image.Mime,
		CreatedAt: image.CreatedAt,
		Owner:     UserFromModel(image.R.Owner),
		Hash:      image.Hash,
		Things:    ReducedThingsFromModel(image.R.ImageThings, userId),
		Actions: ImageActions{
			CanDelete: canDelete,
		},
	}
}

func ImagesFromModelSlice(images models.ImageSlice, userId string) []Image {
	res := make([]Image, len(images))
	for idx, image := range images {
		res[idx] = ImageFromModel(image, userId)
	}
	return res
}

type PaginatedImages struct {
	Images         []Image `json:"images"`
	PerPage        uint64  `json:"perPage"`
	Page           uint64  `json:"page"`
	TotalPageCount uint64  `json:"totalPageCount"`
	TotalCount     uint64  `json:"totalCount"`
}
