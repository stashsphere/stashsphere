package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/services"
	"github.com/stashsphere/backend/utils"
)

type PropertyHandler struct {
	propertyService *services.PropertyService
}

func NewPropertyHandler(propertyService *services.PropertyService) *PropertyHandler {
	return &PropertyHandler{
		propertyService,
	}
}

func (ph *PropertyHandler) GetSchemaCollection(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	collection := ph.propertyService.SchemaCollection()

	etag, err := collection.Hash()
	if err != nil {
		return err
	}

	oldETag := c.Request().Header.Get("If-None-Match")
	if oldETag == etag {
		return c.String(http.StatusNotModified, "Collection not modified")
	}
	c.Response().Header().Set("ETag", etag)
	c.Response().Header().Set("Cache-Control", "no-cache")
	return c.JSON(http.StatusOK, collection)
}
