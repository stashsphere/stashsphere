package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/resources"
	"github.com/stashsphere/backend/services"
)

type ShareHandler struct {
	share_service *services.ShareService
}

func NewShareHandler(share_service *services.ShareService) *ShareHandler {
	return &ShareHandler{
		share_service,
	}
}

type NewShareParams struct {
	TargetUserId string `json:"targetUserId"`
	ObjectId     string `json:"objectId"`
}

func NewShareParamsToCreateShareParams(params NewShareParams, ownerId string) *services.CreateShareParams {
	return &services.CreateShareParams{
		TargetUserId: params.TargetUserId,
		ObjectId:     params.ObjectId,
		OwnerId:      ownerId,
	}
}

func (sh *ShareHandler) ShareHandlerPost(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "No auth context")
	}
	if !authCtx.Authenticated {
		return echo.NewHTTPError(http.StatusForbidden, "Unauthorized")
	}
	shareParams := NewShareParams{}
	if err := c.Bind(&shareParams); err != nil {
		c.Logger().Errorf("Bind error: %v", err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}
	if err := c.Validate(shareParams); err != nil {
		c.Logger().Errorf("Validation error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	share, err := sh.share_service.CreateShare(c.Request().Context(), *NewShareParamsToCreateShareParams(shareParams, authCtx.User.ID))
	if err != nil {
		c.Logger().Errorf("Could not share object: %v", err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity)
	}
	return c.JSON(http.StatusCreated, resources.ShareFromModel(share, authCtx.User.ID))
}

func (sh *ShareHandler) ShareHandlerGet(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "No auth context")
	}
	if !authCtx.Authenticated {
		return echo.NewHTTPError(http.StatusForbidden, "Unauthorized")
	}
	shareId := c.Param("shareId")
	list, err := sh.share_service.GetShare(c.Request().Context(), shareId, authCtx.User.ID)
	if err != nil {
		c.Logger().Errorf("Failed to get share: %v", err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "Failed to retrieve share")
	}
	return c.JSON(http.StatusOK, resources.ShareFromModel(list, authCtx.User.ID))
}
