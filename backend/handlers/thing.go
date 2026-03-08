package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/benjajaja/jtug"
	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/handlers/params"
	"github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/resources"
	"github.com/stashsphere/backend/services"
	"github.com/stashsphere/backend/utils"
)

type ThingHandler struct {
	thingService *services.ThingService
	listService  *services.ListService
}

func NewThingHandler(thingService *services.ThingService, listService *services.ListService) *ThingHandler {
	return &ThingHandler{thingService, listService}
}

type ThingsParams struct {
	Page           uint64         `query:"page"`
	PerPage        uint64         `query:"perPage"`
	FilterOwnerIds []string       `query:"filterOwnerId"`
	SearchTerm     string         `query:"searchTerm"`
	Paginate       *bool          `query:"paginate"`
	Order          []params.Order `query:"order" validate:"max=2,dive"`
}

func ParamsOrderToThingServiceOrder(order []params.Order) ([]services.ThingOrder, error) {
	res := make([]services.ThingOrder, len(order))
	for i, o := range order {
		so := services.ThingOrder{}
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

func (th *ThingHandler) ThingHandlerIndex(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	var params ThingsParams
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
	if params.Paginate != nil && *params.Paginate == false {
		paginate = false
	}

	order, err := ParamsOrderToThingServiceOrder(params.Order)
	if err != nil {
		return err
	}
	result, err := th.thingService.GetThingsForUser(c.Request().Context(),
		services.GetThingsForUserParams{
			UserId:         authCtx.User.UserId,
			PerPage:        params.PerPage,
			Page:           params.Page,
			Paginate:       paginate,
			FilterOwnerIds: params.FilterOwnerIds,
			SearchTerm:     params.SearchTerm,
			Order:          order,
		},
	)
	if err != nil {
		return err
	}
	sharedListIds, err := th.listService.GetSharedListIdsForUser(c.Request().Context(), authCtx.User.UserId)
	if err != nil {
		return err
	}

	reasonMap := make(map[string]*resources.AccessReason)
	for thingId, reason := range result.ThingReasonMap {
		reasonMap[thingId] = resources.AccessReasonFromOperations(reason)
	}

	paginated := resources.PaginatedThings{
		Things:         resources.ThingsFromModelSlice(result.Things, authCtx.User.UserId, sharedListIds, reasonMap),
		PerPage:        uint64(params.PerPage),
		Page:           uint64(params.Page),
		TotalPageCount: result.TotalPages,
		TotalCount:     result.TotalCount,
	}
	return c.JSON(http.StatusOK, paginated)
}

func (th *ThingHandler) ThingHandlerSummary(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	summary, err := th.thingService.GetSummaryForUser(c.Request().Context(), authCtx.User.UserId)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, summary)
}

type PropertyTypeTag string

const (
	PropertyTypeString   = PropertyTypeTag("string")
	PropertyTypeFloat    = PropertyTypeTag("float")
	PropertyTypeDatetime = PropertyTypeTag("datetime")
)

type PropertyStringParam struct {
	Name  string `json:"name" validate:"gt=0"`
	Value string `json:"value" validate:"gt=0"`
}

type PropertyFloatParam struct {
	Name  string  `json:"name" validate:"gt=0"`
	Value float64 `json:"value"`
	Unit  *string `json:"unit"`
}

type PropertyDatetimeParam struct {
	Name  string    `json:"name" validate:"gt=0"`
	Value time.Time `json:"value"`
}

type PropertyUnion = jtug.Union[PropertyTypeTag]
type PropertyList = jtug.UnionList[PropertyTypeTag, PropertyMapper]
type PropertyMapper struct{}

func (PropertyMapper) Unmarshal(b []byte, t PropertyTypeTag) (jtug.Union[PropertyTypeTag], error) {
	switch t {
	case PropertyTypeString:
		var value PropertyStringParam
		return value, json.Unmarshal(b, &value)
	case PropertyTypeFloat:
		var value PropertyFloatParam
		return value, json.Unmarshal(b, &value)
	case PropertyTypeDatetime:
		var value PropertyDatetimeParam
		return value, json.Unmarshal(b, &value)
	default:
		return nil, fmt.Errorf("unknown property type: %v", t)
	}
}

type NewThingParams struct {
	Name         string       `json:"name" validate:"gt=3"`
	PrivateNote  string       `json:"privateNote"`
	Description  string       `json:"description"`
	ImagesIds    []string     `json:"imagesIds"`
	Properties   PropertyList `json:"properties"`
	Quantity     uint64       `json:"quantity"`
	QuantityUnit string       `json:"quantityUnit"`
	SharingState string       `json:"sharingState" validate:"oneof=private friends friends-of-friends"`
}

func NewThingParamsToCreateThingParams(param NewThingParams, ownerId string) services.CreateThingParams {
	properties := []operations.CreatePropertyParams{}
	for i := range param.Properties {
		switch t := param.Properties[i].(type) {
		case PropertyStringParam:
			properties = append(properties, operations.CreatePropertyStringParams{
				Name:  t.Name,
				Value: t.Value,
			})
		case PropertyFloatParam:
			properties = append(properties, operations.CreatePropertyFloatParams{
				Name:  t.Name,
				Value: t.Value,
				Unit:  t.Unit,
			})
		case PropertyDatetimeParam:
			properties = append(properties, operations.CreatePropertyDatetimeParams{
				Name:  t.Name,
				Value: t.Value,
			})
		}
	}
	return services.CreateThingParams{
		Name:         param.Name,
		OwnerId:      ownerId,
		Properties:   properties,
		ImagesIds:    param.ImagesIds,
		Description:  param.Description,
		PrivateNote:  param.PrivateNote,
		Quantity:     param.Quantity,
		QuantityUnit: param.QuantityUnit,
		SharingState: param.SharingState,
	}
}

func (th *ThingHandler) ThingHandlerPost(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	thingParams := NewThingParams{}
	if err := c.Bind(&thingParams); err != nil {
		return &utils.ParameterError{Err: err}
	}
	if err := c.Validate(thingParams); err != nil {
		return &utils.ParameterError{Err: err}
	}
	thing, err := th.thingService.CreateThing(c.Request().Context(), NewThingParamsToCreateThingParams(thingParams, authCtx.User.UserId))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, resources.ReducedThingFromModel(thing, authCtx.User.UserId))
}

type UpdateThingParams = NewThingParams

func UpdateThingParamsToUpdateThingParams(param UpdateThingParams) services.UpdateThingParams {
	properties := []operations.CreatePropertyParams{}
	for i := range param.Properties {
		switch t := param.Properties[i].(type) {
		case PropertyStringParam:
			properties = append(properties, operations.CreatePropertyStringParams{
				Name:  t.Name,
				Value: t.Value,
			})
		case PropertyFloatParam:
			properties = append(properties, operations.CreatePropertyFloatParams{
				Name:  t.Name,
				Value: t.Value,
				Unit:  t.Unit,
			})
		case PropertyDatetimeParam:
			properties = append(properties, operations.CreatePropertyDatetimeParams{
				Name:  t.Name,
				Value: t.Value,
			})
		}
	}
	return services.UpdateThingParams{
		Name:         param.Name,
		Properties:   properties,
		ImagesIds:    param.ImagesIds,
		Description:  param.Description,
		PrivateNote:  param.PrivateNote,
		Quantity:     param.Quantity,
		QuantityUnit: param.QuantityUnit,
		SharingState: param.SharingState,
	}
}

func (th *ThingHandler) ThingHandlerPatch(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	thingId := c.Param("thingId")
	thingParams := UpdateThingParams{}
	if err := c.Bind(&thingParams); err != nil {
		return &utils.ParameterError{Err: err}
	}
	if err := c.Validate(thingParams); err != nil {
		return &utils.ParameterError{Err: err}
	}
	updated_thing, err := th.thingService.EditThing(c.Request().Context(), thingId, authCtx.User.UserId, UpdateThingParamsToUpdateThingParams(thingParams))
	if err != nil {
		return err
	}
	sharedListIds, err := th.listService.GetSharedListIdsForUser(c.Request().Context(), authCtx.User.UserId)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resources.ThingFromModel(updated_thing, authCtx.User.UserId, sharedListIds, nil))
}

func (th *ThingHandler) ThingHandlerShow(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	thingId := c.Param("thingId")
	thingWithReason, err := th.thingService.GetThing(c.Request().Context(), thingId, authCtx.User.UserId)
	if err != nil {
		return err
	}
	sharedListIds, err := th.listService.GetSharedListIdsForUser(c.Request().Context(), authCtx.User.UserId)
	if err != nil {
		return err
	}

	reasonMap := make(map[string]*resources.AccessReason)
	reasonMap[thingWithReason.Thing.ID] = resources.AccessReasonFromOperations(thingWithReason.Reason)

	return c.JSON(http.StatusOK, resources.ThingFromModel(&thingWithReason.Thing, authCtx.User.UserId, sharedListIds, reasonMap))
}

func (th *ThingHandler) ThingHandlerDelete(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return utils.NoAuthContextError{}
	}
	if !authCtx.Authenticated {
		return utils.NotAuthenticatedError{}
	}
	thingId := c.Param("thingId")
	err := th.thingService.DeleteThing(c.Request().Context(), thingId, authCtx.User.UserId)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
