package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/resources"
	"github.com/stashsphere/backend/services"
	"github.com/stashsphere/backend/utils"
)

type ListHandler struct {
	list_service *services.ListService
}

func NewListHandler(list_service *services.ListService) *ListHandler {
	return &ListHandler{
		list_service,
	}
}

type NewListParams struct {
	Name     string   `json:"name" validate:"gt=0"`
	ThingIds []string `json:"thing_ids" validate:"required"`
}

func NewListParamsToCreateListParams(param NewListParams, ownerId string) services.CreateListParams {
	return services.CreateListParams{
		Name:     param.Name,
		ThingIds: param.ThingIds,
		OwnerId:  ownerId,
	}
}

func (lh *ListHandler) ListHandlerPost(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	params := NewListParams{}
	if err := c.Bind(&params); err != nil {
		return &utils.ParameterError{Err: err}
	}
	if err := c.Validate(params); err != nil {
		return &utils.ParameterError{Err: err}
	}
	list, err := lh.list_service.CreateList(c.Request().Context(), NewListParamsToCreateListParams(params, authCtx.User.ID))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, resources.ReducedListFromModel(list, authCtx.User.ID))
}

func (lh *ListHandler) ListHandlerShow(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	listId := c.Param("listId")
	list, err := lh.list_service.GetList(c.Request().Context(), listId, authCtx.User.ID)
	if err != nil {
		return err
	}
	sharedListIds, err := lh.list_service.GetSharedListIdsForUser(c.Request().Context(), authCtx.User.ID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resources.ListFromModel(list, authCtx.User.ID, sharedListIds))
}

type ListsParams struct {
	Page    uint64 `query:"page"`
	PerPage uint64 `query:"perPage"`
}

func (lh *ListHandler) ListHandlerIndex(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	var params ListsParams
	if err := c.Bind(&params); err != nil {
		return &utils.ParameterError{Err: err}
	}
	if params.PerPage == 0 {
		params.PerPage = 50
	}
	totalCount, totalPageCount, lists, err := lh.list_service.GetListsForUser(c.Request().Context(),
		services.GetListsForUserParams{
			UserId:   authCtx.User.ID,
			PerPage:  params.PerPage,
			Page:     params.Page,
			Paginate: true,
		},
	)
	if err != nil {
		return err
	}
	sharedListIds, err := lh.list_service.GetSharedListIdsForUser(c.Request().Context(), authCtx.User.ID)
	if err != nil {
		return err
	}
	paginated := resources.PaginatedLists{
		Things:         resources.ListsFromModelSlice(lists, authCtx.User.ID, sharedListIds),
		PerPage:        uint64(params.PerPage),
		Page:           uint64(params.Page),
		TotalPageCount: totalPageCount,
		TotalCount:     totalCount,
	}
	return c.JSON(http.StatusOK, paginated)
}

type UpdateListParams struct {
	Name     string   `json:"name" validate:"gt=0"`
	ThingIds []string `json:"thing_ids" validate:"required"`
}

func UpdateListParamsToUpdateListParams(p UpdateListParams) services.UpdateListParams {
	return services.UpdateListParams{
		Name:     p.Name,
		ThingIds: p.ThingIds,
	}
}

func (lh *ListHandler) ListHandlerPatch(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "No auth context")
	}
	if !authCtx.Authenticated {
		return c.Redirect(http.StatusSeeOther, "/user/login")
	}
	listId := c.Param("listId")
	listParams := UpdateListParams{}
	if err := c.Bind(&listParams); err != nil {
		return &utils.ParameterError{Err: err}
	}
	if err := c.Validate(listParams); err != nil {
		return &utils.ParameterError{Err: err}
	}
	list, err := lh.list_service.UpdateList(c.Request().Context(), listId, authCtx.User.ID, UpdateListParamsToUpdateListParams(listParams))
	if err != nil {
		return err
	}
	c.Logger().Infof("List edited: %v", list.ID)
	list, err = lh.list_service.GetList(c.Request().Context(), listId, authCtx.User.ID)
	if err != nil {
		return err
	}
	sharedListIds, err := lh.list_service.GetSharedListIdsForUser(c.Request().Context(), authCtx.User.ID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resources.ListFromModel(list, authCtx.User.ID, sharedListIds))
}
