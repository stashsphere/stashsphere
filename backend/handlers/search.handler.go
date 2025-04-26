package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/resources"
	"github.com/stashsphere/backend/services"
	"github.com/stashsphere/backend/utils"
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
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	searchParams := SearchParams{}
	if err := c.Bind(&searchParams); err != nil {
		return &utils.ParameterError{Err: err}
	}
	results, err := sh.search_service.Search(c.Request().Context(), authCtx.User.ID, &services.SearchParams{Query: searchParams.Query})
	if err != nil {
		return err
	}
	sharedListIds, err := sh.list_service.GetSharedListIdsForUser(c.Request().Context(), authCtx.User.ID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resources.SearchResultsFromModel(results, authCtx.User.ID, sharedListIds))
}
