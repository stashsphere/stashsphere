package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/handlers/params"
	"github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/resources"
	"github.com/stashsphere/backend/services"
	"github.com/stashsphere/backend/utils"
)

type ListHandler struct {
	listService *services.ListService
}

func NewListHandler(listService *services.ListService) *ListHandler {
	return &ListHandler{
		listService,
	}
}

type NewListParams struct {
	Name         string   `json:"name" validate:"gt=0"`
	ThingIds     []string `json:"thingIds" validate:"required"`
	SharingState string   `json:"sharingState" validate:"oneof=private friends friends-of-friends"`
}

func NewListParamsToCreateListParams(param NewListParams, ownerId string) services.CreateListParams {
	return services.CreateListParams{
		Name:         param.Name,
		ThingIds:     param.ThingIds,
		OwnerId:      ownerId,
		SharingState: param.SharingState,
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
	list, err := lh.listService.CreateList(c.Request().Context(), NewListParamsToCreateListParams(params, authCtx.User.UserId))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, resources.ReducedListFromModel(list, authCtx.User.UserId))
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
	listWithReason, err := lh.listService.GetList(c.Request().Context(), listId, authCtx.User.UserId)
	if err != nil {
		return err
	}
	sharedListIds, err := lh.listService.GetSharedListIdsForUser(c.Request().Context(), authCtx.User.UserId)
	if err != nil {
		return err
	}

	listReason := resources.AccessReasonFromOperations(listWithReason.Reason)

	return c.JSON(http.StatusOK, resources.ListFromModel(&listWithReason.List, authCtx.User.UserId, sharedListIds, listReason, nil))
}

type ListsParams struct {
	Page           uint64         `query:"page"`
	PerPage        uint64         `query:"perPage"`
	FilterOwnerIds []string       `query:"filterOwnerId"`
	Paginate       *bool          `query:"paginate"`
	Order          []params.Order `query:"order" validate:"max=2,dive"`
}

func ParamsOrderToListServiceOrder(order []params.Order) ([]services.ListOrder, error) {
	res := make([]services.ListOrder, len(order))
	for i, o := range order {
		so := services.ListOrder{}
		switch o.FieldName {
		case "access_reason":
			so.Field = services.OrderFieldAccessReason
			break
		case "created_at":
			so.Field = services.OrderFieldCreatedAt
			break
		default:
			return nil, utils.ParameterError{Err: fmt.Errorf("Invalid field.")}
		}
		switch o.FieldSortBy {
		case "desc":
			so.Direction = services.OrderDescending
			break
		case "asc":
			so.Direction = services.OrderAscending
			break
		default:
			return nil, utils.ParameterError{Err: fmt.Errorf("Invalid field.")}
		}
		res[i] = so
	}
	return res, nil
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
	if err := c.Validate(&params); err != nil {
		return &utils.ParameterError{Err: err}
	}
	if params.PerPage == 0 {
		params.PerPage = 50
	}
	paginate := true
	if params.Paginate != nil && *params.Paginate == true {
		paginate = true
	}

	order, err := ParamsOrderToListServiceOrder(params.Order)
	if err != nil {
		return err
	}
	result, err := lh.listService.GetListsForUser(c.Request().Context(),
		services.GetListsForUserParams{
			UserId:         authCtx.User.UserId,
			PerPage:        params.PerPage,
			Page:           params.Page,
			FilterOwnerIds: params.FilterOwnerIds,
			Paginate:       paginate,
			Order:          order,
		},
	)
	if err != nil {
		return err
	}
	sharedListIds, err := lh.listService.GetSharedListIdsForUser(c.Request().Context(), authCtx.User.UserId)
	if err != nil {
		return err
	}

	listReasonMap := make(map[string]*resources.AccessReason)
	for listId, reason := range result.ListReasonMap {
		listReasonMap[listId] = resources.AccessReasonFromOperations(reason)
	}

	thingReasonMap := make(map[string]*resources.AccessReason)
	for thingId, reason := range result.ThingReasonMap {
		thingReasonMap[thingId] = resources.AccessReasonFromOperations(reason)
	}

	paginated := resources.PaginatedLists{
		Things:         resources.ListsFromModelSlice(result.Lists, authCtx.User.UserId, sharedListIds, listReasonMap, thingReasonMap),
		PerPage:        uint64(params.PerPage),
		Page:           uint64(params.Page),
		TotalPageCount: result.TotalPages,
		TotalCount:     result.TotalCount,
	}
	return c.JSON(http.StatusOK, paginated)
}

type UpdateListParams struct {
	Name         string   `json:"name" validate:"gt=0"`
	ThingIds     []string `json:"thingIds" validate:"required"`
	SharingState string   `json:"sharingState" validate:"oneof=private friends friends-of-friends"`
}

func UpdateListParamsToUpdateListParams(p UpdateListParams) services.UpdateListParams {
	return services.UpdateListParams{
		Name:         p.Name,
		ThingIds:     p.ThingIds,
		SharingState: p.SharingState,
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
	list, err := lh.listService.UpdateList(c.Request().Context(), listId, authCtx.User.UserId, UpdateListParamsToUpdateListParams(listParams))
	if err != nil {
		return err
	}
	c.Logger().Infof("List edited: %v", list.ID)
	listWithReason, err := lh.listService.GetList(c.Request().Context(), listId, authCtx.User.UserId)
	if err != nil {
		return err
	}
	sharedListIds, err := lh.listService.GetSharedListIdsForUser(c.Request().Context(), authCtx.User.UserId)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resources.ListFromModel(&listWithReason.List, authCtx.User.UserId, sharedListIds, nil, nil))
}

func (lh *ListHandler) ListHandlerDelete(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	listId := c.Param("listId")
	err := lh.listService.DeleteList(c.Request().Context(), listId, authCtx.User.UserId)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
