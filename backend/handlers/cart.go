package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/resources"
	"github.com/stashsphere/backend/services"
	"github.com/stashsphere/backend/utils"
)

type CartHandler struct {
	cart_service *services.CartService
}

func NewCartHandler(cart_service *services.CartService) *CartHandler {
	return &CartHandler{
		cart_service,
	}
}

func (ch *CartHandler) Index(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}

	entries, err := ch.cart_service.GetCart(c.Request().Context(), authCtx.User.UserId)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resources.CartFromModelSlice(entries))
}

type CartParams struct {
	ThingIds []string `json:"thingIds"`
}

func (ch *CartHandler) Put(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	params := CartParams{}
	if err := c.Bind(&params); err != nil {
		return &utils.ParameterError{Err: err}
	}
	entries, err := ch.cart_service.UpdateCart(c.Request().Context(), services.UpdateCartParams{
		UserId:   authCtx.User.UserId,
		ThingIds: params.ThingIds,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resources.CartFromModelSlice(entries))
}
