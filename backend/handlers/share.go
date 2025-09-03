package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/resources"
	"github.com/stashsphere/backend/services"
	"github.com/stashsphere/backend/utils"
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
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	shareParams := NewShareParams{}
	if err := c.Bind(&shareParams); err != nil {
		return &utils.ParameterError{Err: err}
	}
	if err := c.Validate(shareParams); err != nil {
		return &utils.ParameterError{Err: err}
	}
	share, err := sh.share_service.CreateShare(c.Request().Context(), *NewShareParamsToCreateShareParams(shareParams, authCtx.User.UserId))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, resources.ShareFromModel(share, authCtx.User.UserId))
}

func (sh *ShareHandler) ShareHandlerGet(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	shareId := c.Param("shareId")
	share, err := sh.share_service.GetShare(c.Request().Context(), shareId, authCtx.User.UserId)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resources.ShareFromModel(share, authCtx.User.UserId))
}

func (sh *ShareHandler) ShareHandlerDelete(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	shareId := c.Param("shareId")
	err := sh.share_service.DeleteShare(c.Request().Context(), shareId, authCtx.User.UserId)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}
