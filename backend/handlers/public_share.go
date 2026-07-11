package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/resources"
	"github.com/stashsphere/backend/services"
	"github.com/stashsphere/backend/utils"
)

type PublicShareHandler struct {
	publicShareService *services.PublicShareService
}

func NewPublicShareHandler(publicShareService *services.PublicShareService) *PublicShareHandler {
	return &PublicShareHandler{
		publicShareService,
	}
}

type NewPublicShareParams struct {
	ObjectId string `json:"objectId"`
}

func (ph *PublicShareHandler) PublicShareHandlerPost(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	shareParams := NewPublicShareParams{}
	if err := c.Bind(&shareParams); err != nil {
		return &utils.ParameterError{Err: err}
	}
	if err := c.Validate(shareParams); err != nil {
		return &utils.ParameterError{Err: err}
	}
	share, err := ph.publicShareService.CreatePublicShare(c.Request().Context(), services.CreatePublicShareParams{
		ObjectId: shareParams.ObjectId,
		OwnerId:  authCtx.User.UserId,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, resources.PublicShareInfoFromModel(share))
}

func (ph *PublicShareHandler) PublicShareHandlerIndex(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	shares, err := ph.publicShareService.GetPublicSharesForUser(c.Request().Context(), authCtx.User.UserId)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resources.PublicShareIndexEntriesFromModelSlice(shares))
}

// PublicShareHandlerGet is intentionally accessible without authentication:
// knowledge of the share token grants read access.
func (ph *PublicShareHandler) PublicShareHandlerGet(c echo.Context) error {
	token := c.Param("token")
	share, err := ph.publicShareService.GetPublicShare(c.Request().Context(), token)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resources.PublicShareFromModel(share))
}

func (ph *PublicShareHandler) PublicShareHandlerDelete(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	token := c.Param("token")
	err := ph.publicShareService.DeletePublicShare(c.Request().Context(), token, authCtx.User.UserId)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}
