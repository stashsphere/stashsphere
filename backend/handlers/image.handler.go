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
)

type ImageHandler struct {
	image_service *services.ImageService
}

func NewImageHandler(image_service *services.ImageService) *ImageHandler {
	return &ImageHandler{image_service}
}

func (is *ImageHandler) ImageHandlerPost(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return c.String(http.StatusInternalServerError, "No auth context")
	}
	if !authCtx.Authenticated {
		return c.String(http.StatusUnauthorized, "Not authorized")
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

	image, err := is.image_service.CreateImage(c.Request().Context(), authCtx.User.ID, file.Filename, src)
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
		return c.String(http.StatusInternalServerError, "No auth context")
	}
	if !authCtx.Authenticated {
		return c.String(http.StatusUnauthorized, "Not authorized")
	}
	var imageParams ImageGetParams
	if err := c.Bind(&imageParams); err != nil {
		c.Logger().Errorf("Bind error: %v", err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}
	hash := c.Param("hash")
	file, image, err := is.image_service.ImageGet(c.Request().Context(), authCtx.User.ID, hash)
	if err != nil {
		if os.IsNotExist(err) {
			return c.String(http.StatusNotFound, "Image Not Found")
		}
		c.Logger().Error(err)
		return c.String(http.StatusInternalServerError, "Internal Server Error")
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
		returnedImageReader, err = operations.ResizeImage(file, int(imageParams.Width))
		if err != nil {
			c.Logger().Errorf("Resize error: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
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
		return c.String(http.StatusInternalServerError, "No auth context")
	}
	if !authCtx.Authenticated {
		return c.String(http.StatusUnauthorized, "Not authorized")
	}

	imageId := c.Param("imageId")

	var params ImageModifyParams
	err := c.Bind(&params)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid parameters")
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

	image, err := is.image_service.ModifyImage(c.Request().Context(), authCtx.User.ID, imageId, services.ModifyImageParams{
		Rotation: rotation,
	})
	if err != nil {
		if os.IsNotExist(err) {
			return c.String(http.StatusNotFound, "Image Not Found")
		}
		c.Logger().Error(err)
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}
	resource := resources.ReducedImageFromModel(image)
	return c.JSON(http.StatusCreated, resource)
}

type ImagesParams struct {
	Page    uint64 `query:"page"`
	PerPage uint64 `query:"perPage"`
}

// this handler only lists own images to be able to create galleries and manage
// pictures independent of things and lists
func (is *ImageHandler) ImageHandlerIndex(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return c.String(http.StatusInternalServerError, "No auth context")
	}
	if !authCtx.Authenticated {
		return c.String(http.StatusUnauthorized, "Not authorized")
	}
	var params ImagesParams
	err := c.Bind(&params)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid parameters")
	}
	if params.PerPage == 0 {
		params.PerPage = 50
	}

	totalCount, totalPageCount, images, err := is.image_service.ImageIndex(c.Request().Context(), authCtx.User.ID, params.PerPage, params.Page)
	if err != nil {
		c.Logger().Warn("Could not retrieve images: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	paginated := resources.PaginatedImages{
		Images:         resources.ImagesFromModelSlice(images, authCtx.User.ID),
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
		return c.String(http.StatusInternalServerError, "No auth context")
	}
	if !authCtx.Authenticated {
		return c.String(http.StatusUnauthorized, "Not authorized")
	}
	imageId := c.Param("imageId")
	deletedImage, err := is.image_service.DeleteImage(c.Request().Context(), authCtx.User.ID, imageId)
	if err != nil {
		c.Logger().Warn("Could not delete images: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, resources.ReducedImageFromModel(deletedImage))
}
