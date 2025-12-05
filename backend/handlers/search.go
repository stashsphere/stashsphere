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
	searchService   *services.SearchService
	listService     *services.ListService
	propertyService *services.PropertyService
}

func NewSearchHandler(searchService *services.SearchService, listService *services.ListService, propertyService *services.PropertyService) *SearchHandler {
	return &SearchHandler{
		searchService,
		listService,
		propertyService,
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
	results, err := sh.searchService.Search(c.Request().Context(), authCtx.User.UserId, &services.SearchParams{Query: searchParams.Query})
	if err != nil {
		return err
	}
	sharedListIds, err := sh.listService.GetSharedListIdsForUser(c.Request().Context(), authCtx.User.UserId)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resources.SearchResultsFromModel(results, authCtx.User.UserId, sharedListIds))
}

type AutoCompleteParams struct {
	Name  string `query:"name"`
	Value string `query:"value"`
}

func PropertyAutoCompleteParamsFromParams(v AutoCompleteParams, userId string) services.PropertyAutoCompleteParams {
	if v.Value == "" {
		return services.PropertyAutoCompleteParams{
			Name:   v.Name,
			Value:  nil,
			UserId: userId,
		}
	} else {
		return services.PropertyAutoCompleteParams{
			Name:   v.Name,
			Value:  &v.Value,
			UserId: userId,
		}
	}
}

func (sh *SearchHandler) AutocompleteGet(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	autocompleteParams := AutoCompleteParams{}
	if err := c.Bind(&autocompleteParams); err != nil {
		return &utils.ParameterError{Err: err}
	}
	result, err := sh.propertyService.AutoComplete(c.Request().Context(), PropertyAutoCompleteParamsFromParams(autocompleteParams, authCtx.User.UserId))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, result)
}
