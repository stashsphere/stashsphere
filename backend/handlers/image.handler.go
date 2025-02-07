package handlers

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/middleware"
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

func (is *ImageHandler) ImageHandlerGet(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return c.String(http.StatusInternalServerError, "No auth context")
	}
	if !authCtx.Authenticated {
		return c.String(http.StatusUnauthorized, "Not authorized")
	}

	imageId := c.Param("imageId")
	file, image, err := is.image_service.ImageGet(c.Request().Context(), authCtx.User.ID, imageId)
	if err != nil {
		if os.IsNotExist(err) {
			return c.String(http.StatusNotFound, "Image Not Found")
		}
		c.Logger().Error(err)
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}
	defer file.Close()
	oldETag := c.Request().Header.Get("If-None-Match")
	if oldETag == image.Hash {
		return c.String(http.StatusNotModified, "Image Not Modified")
	}
	c.Response().Header().Set("ETag", image.Hash)
	return c.Stream(http.StatusOK, image.Mime, file)
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
