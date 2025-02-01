package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/benjajaja/jtug"
	"github.com/labstack/echo/v4"
	"github.com/stashsphere/backend/middleware"
	"github.com/stashsphere/backend/operations"
	"github.com/stashsphere/backend/resources"
	"github.com/stashsphere/backend/services"
)

type ThingHandler struct {
	thing_service    *services.ThingService
	list_service     *services.ListService
	property_service *services.PropertyService
}

func NewThingHandler(thing_service *services.ThingService, list_service *services.ListService, property_service *services.PropertyService) *ThingHandler {
	return &ThingHandler{thing_service, list_service, property_service}
}

type ThingsParams struct {
	Page    uint64 `query:"page"`
	PerPage uint64 `query:"perPage"`
}

func (th *ThingHandler) ThingHandlerIndex(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return c.String(http.StatusInternalServerError, "No auth context")
	}
	if !authCtx.Authenticated {
		return c.String(http.StatusUnauthorized, "Not authorized")
	}
	var params ThingsParams
	err := c.Bind(&params)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid parameters")
	}
	if params.PerPage == 0 {
		params.PerPage = 50
	}

	totalCount, totalPageCount, things, err := th.thing_service.GetThingsForUser(c.Request().Context(),
		services.GetThingsForUserParams{
			UserId:   authCtx.User.ID,
			PerPage:  params.PerPage,
			Page:     params.Page,
			Paginate: true,
		},
	)
	if err != nil {
		c.Logger().Error("Could not query owned things for User %d. %v", authCtx.User.ID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "")
	}
	sharedListIds, err := th.list_service.GetSharedListIdsForUser(c.Request().Context(), authCtx.User.ID)
	if err != nil {
		c.Logger().Warn("Could not get shared lists: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	paginated := resources.PaginatedThings{
		Things:         resources.ThingsFromModelSlice(things, authCtx.User.ID, sharedListIds),
		PerPage:        uint64(params.PerPage),
		Page:           uint64(params.Page),
		TotalPageCount: totalPageCount,
		TotalCount:     totalCount,
	}
	return c.JSON(http.StatusOK, paginated)
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
	}
}

func (th *ThingHandler) ThingHandlerPost(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "No auth context")
	}
	if !authCtx.Authenticated {
		return echo.NewHTTPError(http.StatusForbidden, "Unauthorized")
	}
	thingParams := NewThingParams{}
	if err := c.Bind(&thingParams); err != nil {
		c.Logger().Errorf("Bind error: %v", err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}
	if err := c.Validate(thingParams); err != nil {
		c.Logger().Errorf("Validation error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	thing, err := th.thing_service.CreateThing(c.Request().Context(), NewThingParamsToCreateThingParams(thingParams, authCtx.User.ID))
	if err != nil {
		c.Logger().Errorf("Could not create thing: %v", err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity)
	}
	return c.JSON(http.StatusCreated, resources.ReducedThingFromModel(thing, authCtx.User.ID))
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
	}
}

func (th *ThingHandler) ThingHandlerPatch(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "No auth context")
	}
	if !authCtx.Authenticated {
		return c.Redirect(http.StatusSeeOther, "/user/login")
	}
	thingId := c.Param("thingId")
	thingParams := UpdateThingParams{}
	if err := c.Bind(&thingParams); err != nil {
		c.Logger().Errorf("Bind failed: %v", err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}
	if err := c.Validate(thingParams); err != nil {
		c.Logger().Errorf("Validation failed: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "TODO")
	}
	thing, err := th.thing_service.EditThing(c.Request().Context(), thingId, authCtx.User.ID, UpdateThingParamsToUpdateThingParams(thingParams))
	if err != nil {
		c.Logger().Errorf("Failed to edit thing: %v", err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "Failed to edit thing")
	}
	c.Logger().Infof("Thing edited: %v", thing.ID)
	updated_thing, err := th.thing_service.GetThing(c.Request().Context(), thingId, authCtx.User.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "Failed to retrieve updated thing")
	}
	sharedListIds, err := th.list_service.GetSharedListIdsForUser(c.Request().Context(), authCtx.User.ID)
	if err != nil {
		c.Logger().Warn("Could not get shared lists: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, resources.ThingFromModel(updated_thing, authCtx.User.ID, sharedListIds))
}

func (th *ThingHandler) ThingHandlerShow(c echo.Context) error {
	authCtx, ok := c.Get("auth").(*middleware.AuthContext)
	if !ok {
		c.Logger().Errorf("No auth context")
		return echo.NewHTTPError(http.StatusInternalServerError, "No auth context")
	}
	if !authCtx.Authenticated {
		c.Logger().Errorf("User is not authenticated")
		return echo.NewHTTPError(http.StatusForbidden, "Unauthorized")
	}
	thingId := c.Param("thingId")
	thing, err := th.thing_service.GetThing(c.Request().Context(), thingId, authCtx.User.ID)
	if err != nil {
		c.Logger().Errorf("Failed to get thing: %v", err)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "Failed to retrieve thing")
	}
	sharedListIds, err := th.list_service.GetSharedListIdsForUser(c.Request().Context(), authCtx.User.ID)
	if err != nil {
		c.Logger().Warn("Could not get shared lists: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, resources.ThingFromModel(thing, authCtx.User.ID, sharedListIds))
}
