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

	"github.com/labstack/gommon/log"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/rakyll/magicmime"
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

	hasher := sha256.New()

	firstChunk := make([]byte, 1024)
	chunk := 0
	for {
		buf := make([]byte, 1024)
		n, err := src.Read(buf)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if n == 0 {
			break
		}
		hasher.Write(buf[:n])
		_, err = tmp.Write(buf[:n])
		if err != nil {
			return nil, err
		}
		if chunk == 0 {
			copy(firstChunk, buf)
		}
		chunk += 1
	}
	hash := hasher.Sum(nil)
	mime, err := is.mimeDecoder.TypeByBuffer(firstChunk)
	if err != nil {
		defer os.Remove(tmp.Name())
		return nil, err
	}

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

	err = os.Rename(tmp.Name(), newPath)
	if err != nil {
		defer os.Remove(tmp.Name())
		return nil, err
	}
	log.Infof("Created %s", newPath)
	return &image, nil
}

func (is *ImageService) ImageGet(ctx context.Context, userId string, imageId string) (*os.File, *models.Image, error) {
	image, err := models.FindImage(ctx, is.db, imageId)
	if err != nil {
		return nil, nil, err
	}

	sharedImagesForUser, err := operations.GetSharedImageIdsForUser(ctx, is.db, userId)
	if err != nil {
		return nil, nil, err
	}
	authorized := func() bool {
		for _, id := range sharedImagesForUser {
			if id == imageId {
				return true
			}
		}
		if userId == image.OwnerID {
			return true
		}
		return false
	}()
	if !authorized {
		return nil, nil, utils.ErrUserHasNoAccessRights
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
		qm.Load(models.ImageRels.ImageThings),
		qm.Load(qm.Rels(models.ImageRels.ImageThings, models.ThingRels.Owner)),
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
			return err
		}
		if deletedImage == nil {
			return errors.New("Unexpected error: image is nil")
		}
		err = operations.DeleteContent(ctx, tx, is.storePath, deletedImage.Hash)
		if err != nil && !errors.Is(err, utils.ErrEntityInUse) {
			return err
		}
		image = deletedImage
		return nil
	})
	return image, err
}
