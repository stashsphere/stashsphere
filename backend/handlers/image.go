package handlers

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/binary"
	"io"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/resources"
	"github.com/stashsphere/backend/services"
	"github.com/stashsphere/backend/utils"
)

type ImageHandler struct {
	image_service *services.ImageService
	cache_service *services.CacheService
}

func NewImageHandler(image_service *services.ImageService, cache_service *services.CacheService) *ImageHandler {
	return &ImageHandler{image_service, cache_service}
}

func (is *ImageHandler) ImageHandlerPost(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	// Source
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	image, err := is.image_service.CreateImage(c.Request().Context(), authCtx.User.UserId, file.Filename, src)
	if err != nil {
		return err
	}
	resource := resources.ReducedImageFromModel(image)
	return c.JSON(http.StatusCreated, resource)
}

type ImageGetParams struct {
	Width uint16 `query:"width" validate:"min=20,max=8192"`
}

func (is *ImageHandler) ImageHandlerGet(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	var imageParams ImageGetParams
	if err := c.Bind(&imageParams); err != nil {
		return &utils.ParameterError{Err: err}
	}
	hash := c.Param("hash")
	file, image, err := is.image_service.ImageGet(c.Request().Context(), authCtx.User.UserId, hash)
	if err != nil {
		if os.IsNotExist(err) {
			return c.String(http.StatusNotFound, "Image Not Found")
		}
		return err
	}
	defer file.Close()

	var etag string
	var returnedImageReader io.Reader

	resize := imageParams.Width != 0 && (image.Mime == "image/jpeg" || image.Mime == "image/png")

	if resize {
		widthBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(widthBytes, imageParams.Width)

		hasher := sha256.New()
		hasher.Write([]byte(image.Hash))
		hasher.Write(widthBytes)
		hash := hasher.Sum(nil)
		encoding := base32.StdEncoding.WithPadding(base32.NoPadding)
		hash32 := encoding.EncodeToString(hash[:])
		etag = hash32
	} else {
		etag = image.Hash
	}

	if resize {
		if !is.cache_service.Exists(etag) {
			resized, err := operations.ResizeImage(file, int(imageParams.Width))
			if err != nil {
				return err
			}
			resizedContent, err := io.ReadAll(resized)
			if err != nil {
				return err
			}
			err = is.cache_service.Put(etag, resizedContent)
			if err != nil {
				return err
			}
		}
		returnedImageReader, err = is.cache_service.Get(etag)
		if err != nil {
			return err
		}
	} else {
		returnedImageReader = file
	}

	oldETag := c.Request().Header.Get("If-None-Match")
	if oldETag == etag {
		return c.String(http.StatusNotModified, "Image Not Modified")
	}
	c.Response().Header().Set("ETag", etag)
	c.Response().Header().Set("Cache-Control", "no-cache")
	return c.Stream(http.StatusOK, image.Mime, returnedImageReader)
}

type ImageModifyParams struct {
	Rotation uint16 `json:"rotation" validate:"oneof=90 180 270"`
}

func (is *ImageHandler) ImageHandlerPatch(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}

	imageId := c.Param("imageId")

	var params ImageModifyParams
	if err := c.Bind(&params); err != nil {
		return &utils.ParameterError{Err: err}
	}

	rotation := operations.Rotation90
	switch params.Rotation {
	case 90:
		rotation = operations.Rotation90
	case 180:
		rotation = operations.Rotation180
	case 270:
		rotation = operations.Rotation270
	}

	image, err := is.image_service.ModifyImage(c.Request().Context(), authCtx.User.UserId, imageId, services.ModifyImageParams{
		Rotation: rotation,
	})
	if err != nil {
		if os.IsNotExist(err) {
			return &utils.NotFoundError{EntityName: "Image"}
		}
		c.Logger().Error(err)
		return err
	}
	resource := resources.ReducedImageFromModel(image)
	return c.JSON(http.StatusCreated, resource)
}

type ImagesParams struct {
	Page           uint64 `query:"page"`
	PerPage        uint64 `query:"perPage"`
	OnlyUnassigned bool   `query:"onlyUnassigned"`
}

// this handler only lists own images to be able to create galleries and manage
// pictures independent of things and lists
func (is *ImageHandler) ImageHandlerIndex(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	var params ImagesParams
	if err := c.Bind(&params); err != nil {
		return &utils.ParameterError{Err: err}
	}
	if params.PerPage == 0 {
		params.PerPage = 50
	}

	totalCount, totalPageCount, images, err := is.image_service.ImageIndex(c.Request().Context(),
		services.ImageIndexParams{
			UserId:         authCtx.User.UserId,
			PerPage:        params.PerPage,
			Page:           params.Page,
			OnlyUnassigned: params.OnlyUnassigned,
		})

	if err != nil {
		return err
	}
	paginated := resources.PaginatedImages{
		Images:         resources.ImagesFromModelSlice(images, authCtx.User.UserId),
		PerPage:        uint64(params.PerPage),
		Page:           uint64(params.Page),
		TotalPageCount: totalPageCount,
		TotalCount:     totalCount,
	}
	return c.JSON(http.StatusOK, paginated)
}

func (is *ImageHandler) ImageHandlerDelete(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	imageId := c.Param("imageId")
	deletedImage, err := is.image_service.DeleteImage(c.Request().Context(), authCtx.User.UserId, imageId)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resources.ReducedImageFromModel(deletedImage))
}
