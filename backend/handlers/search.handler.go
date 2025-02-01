package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/resources"
	"github.com/stashsphere/backend/services"
)

type SearchHandler struct {
	search_service *services.SearchService
	list_service   *services.ListService
}

func NewSearchHandler(search_service *services.SearchService, list_service *services.ListService) *SearchHandler {
	return &SearchHandler{
		search_service,
		list_service,
	}
}

type SearchParams struct {
	Query string `query:"query"`
}

func (sh *SearchHandler) SearchHandlerGet(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "No auth context")
	}
	if !authCtx.Authenticated {
		return c.Redirect(http.StatusSeeOther, "/user/login")
	}
	searchParams := SearchParams{}
	if err := c.Bind(&searchParams); err != nil {
		fmt.Printf("Validation Failed %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not parse params")
	}
	results, err := sh.search_service.Search(c.Request().Context(), authCtx.User.ID, &services.SearchParams{Query: searchParams.Query})
	if err != nil {
		c.Logger().Error("Could not search: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not search")
	}
	sharedListIds, err := sh.list_service.GetSharedListIdsForUser(c.Request().Context(), authCtx.User.ID)
	if err != nil {
		c.Logger().Warn("Could not get shared lists: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, resources.SearchResultsFromModel(results, authCtx.User.ID, sharedListIds))
}
