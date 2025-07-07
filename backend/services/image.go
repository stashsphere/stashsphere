package services

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"errors"
	"io"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/rakyll/magicmime"
	"github.com/rs/zerolog/log"
	"github.com/stashsphere/backend/models"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/utils"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type ImageService struct {
	db          *sql.DB
	storePath   string
	mimeDecoder *magicmime.Decoder
}

func NewImageService(db *sql.DB, storePath string) (*ImageService, error) {
	mimeDecoder, err := magicmime.NewDecoder(magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK | magicmime.MAGIC_ERROR)
	if err != nil {
		return nil, err
	}
	is := &ImageService{db, storePath, mimeDecoder}
	err = os.MkdirAll(storePath, 0755)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(is.tmpPath(), 0755)
	if err != nil {
		return nil, err
	}
	return is, nil
}

// / for testing purposes
func NewTmpImageService(db *sql.DB) (*ImageService, error) {
	dir, err := os.MkdirTemp("", "image_service")
	if err != nil {
		return nil, err
	}
	return NewImageService(db, dir)
}

func (is *ImageService) tmpPath() string {
	return path.Join(is.storePath, "tmp")
}

func (is *ImageService) StorePath() string {
	return is.storePath
}

type ImageFile interface {
	io.Reader
	io.Closer
}

func (is *ImageService) CreateImage(ctx context.Context, ownerId string, name string, src ImageFile) (*models.Image, error) {
	tmp, err := os.CreateTemp(is.tmpPath(), "tmpfile")
	if err != nil {
		return nil, err
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())

	_, err = io.Copy(tmp, src)
	if err != nil {
		return nil, err
	}

	exifRemoved, err := operations.ClearExifData(tmp.Name())
	if err != nil {
		return nil, err
	}

	var srcData []byte
	if len(exifRemoved) == 0 {
		imgFile, err := os.Open(tmp.Name())
		defer imgFile.Close()
		if err != nil {
			return nil, err
		}
		srcData, err = io.ReadAll(imgFile)
		if err != nil {
			return nil, err
		}
	} else {
		srcData = exifRemoved
	}

	srcDataLength := len(srcData)

	firstChunk := srcData[:min(srcDataLength, 1024)]
	mime, err := is.mimeDecoder.TypeByBuffer(firstChunk)
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(mime, "image/") {
		return nil, utils.IllegalMimeTypeError{}
	}

	hasher := sha256.New()
	hasher.Write(srcData)
	hash := hasher.Sum(nil)
	encoding := base32.StdEncoding.WithPadding(base32.NoPadding)
	hash32 := encoding.EncodeToString(hash[:])

	newPath := filepath.Join(is.storePath, string(hash32))
	imageID, err := gonanoid.New()
	if err != nil {
		return nil, err
	}
	image := models.Image{
		Name:    name,
		Mime:    mime,
		Hash:    string(hash32),
		OwnerID: ownerId,
		ID:      imageID,
	}
	err = image.Insert(ctx, is.db, boil.Infer())
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(newPath, srcData, 0640)
	if err != nil {
		return nil, err
	}
	log.Info().Msgf("Created %s", newPath)
	return &image, nil
}

func (is *ImageService) ImageGet(ctx context.Context, userId string, hash string) (*os.File, *models.Image, error) {
	image, err := models.Images(models.ImageWhere.Hash.EQ(hash),
		qm.Load(models.ImageRels.Profiles),
	).One(ctx, is.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, utils.NotFoundError{EntityName: "Image"}
		}
		return nil, nil, err
	}
	sharedImagesForUser, err := operations.GetSharedImageIdsForUser(ctx, is.db, userId)
	if err != nil {
		return nil, nil, err
	}
	authorized := func() bool {
		for _, id := range sharedImagesForUser {
			if id == image.ID {
				return true
			}
		}
		if userId == image.OwnerID {
			return true
		}
		// it's used as a profile picture
		if len(image.R.Profiles) > 0 {
			return true
		}
		return false
	}()
	if !authorized {
		return nil, nil, utils.UserHasNoAccessRightsError{}
	}

	path := filepath.Join(is.storePath, image.Hash)

	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	return file, image, nil
}

func (is *ImageService) ImageIndex(ctx context.Context, userId string, perPage uint64, page uint64) (uint64, uint64, models.ImageSlice, error) {
	searchCond := models.ImageWhere.OwnerID.EQ(userId)

	imageCount, err := models.Images(searchCond).Count(ctx, is.db)
	if err != nil {
		return 0, 0, models.ImageSlice{}, err
	}

	images, err := models.Images(
		qm.Load(models.ImageRels.Things),
		qm.Load(qm.Rels(models.ImageRels.Things, models.ThingRels.Owner)),
		qm.Load(models.ImageRels.Owner),
		searchCond,
		qm.OrderBy(models.ImageColumns.CreatedAt),
		qm.Offset(int(perPage*page)),
		qm.Limit(int(perPage)),
	).All(ctx, is.db)
	if err != nil {
		return 0, 0, models.ImageSlice{}, err
	}
	totalPages := uint64(math.Ceil(float64(imageCount) / float64(perPage)))
	return uint64(imageCount), totalPages, images, nil
}

func (is *ImageService) DeleteImage(ctx context.Context, userId string, imageId string) (*models.Image, error) {
	var image *models.Image
	err := utils.Tx(ctx, is.db, func(tx *sql.Tx) error {
		deletedImage, err := operations.DeleteImage(ctx, tx, userId, imageId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return utils.NotFoundError{EntityName: "Image"}
			}
			return err
		}
		if deletedImage == nil {
			return errors.New("Unexpected error: image is nil")
		}
		err = operations.DeleteContent(ctx, tx, is.storePath, deletedImage.Hash)
		if err != nil && !errors.Is(err, utils.EntityInUseError{}) {
			return err
		}
		image = deletedImage
		return nil
	})
	return image, err
}

type ModifyImageParams struct {
	Rotation operations.Rotation
}

func (is *ImageService) ModifyImage(ctx context.Context, userId string, imageId string, params ModifyImageParams) (*models.Image, error) {
	image, err := models.FindImage(ctx, is.db, imageId)
	if err != nil {
		return nil, err
	}
	if image.OwnerID != userId {
		return nil, utils.EntityDoesNotBelongToUserError{}
	}

	path := filepath.Join(is.storePath, image.Hash)
	rotatedBytes, err := operations.RotateImage(path, params.Rotation)
	if err != nil {
		return nil, err
	}
	hasher := sha256.New()
	hasher.Write(rotatedBytes)
	hash := hasher.Sum(nil)
	encoding := base32.StdEncoding.WithPadding(base32.NoPadding)
	hash32 := encoding.EncodeToString(hash[:])
	newPath := filepath.Join(is.storePath, string(hash32))
	err = os.WriteFile(newPath, rotatedBytes, 0640)
	if err != nil {
		return nil, err
	}
	image.Hash = hash32
	_, err = image.Update(ctx, is.db, boil.Infer())
	if err != nil {
		return nil, err
	}
	return image, nil
}
