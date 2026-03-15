package operations

import (
	"context"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/rs/zerolog/log"
	"github.com/stashsphere/backend/models"
)

type CreatePropertyFloatParams struct {
	Name  string
	Value float64
	Unit  *string
}

type CreatePropertyStringParams struct {
	Name  string
	Value string
}

type CreatePropertyDatetimeParams struct {
	Name  string
	Value time.Time
}

type CreatePropertyBooleanParams struct {
	Name  string
	Value bool
}

type CreatePropertyParams interface {
	Data() any
	Type() string
}

func (p CreatePropertyFloatParams) Data() any    { return p }
func (p CreatePropertyDatetimeParams) Data() any { return p }
func (p CreatePropertyStringParams) Data() any   { return p }
func (p CreatePropertyBooleanParams) Data() any  { return p }

func (p CreatePropertyFloatParams) Type() string    { return "float" }
func (p CreatePropertyDatetimeParams) Type() string { return "datetime" }
func (p CreatePropertyStringParams) Type() string   { return "string" }
func (p CreatePropertyBooleanParams) Type() string  { return "boolean" }

func CreateProperty(ctx context.Context, exec boil.ContextExecutor, thingId string, params CreatePropertyParams) (*models.Property, error) {
	propertyID, err := gonanoid.New()
	if err != nil {
		return nil, err
	}

	property := models.Property{
		ID: propertyID,
	}

	switch params.Type() {
	case "string":
		data := params.Data().(CreatePropertyStringParams)
		property.Type = models.PropertyTypeString
		property.ThingID = thingId
		property.Name = data.Name
		property.ValueString = null.NewString(data.Value, true)
	case "float":
		data := params.Data().(CreatePropertyFloatParams)
		property.Type = models.PropertyTypeFloat
		property.ThingID = thingId
		property.Name = data.Name
		property.ValueFloat = null.NewFloat64(data.Value, true)
		if data.Unit != nil {
			property.Unit = null.NewString(*data.Unit, true)
		}
	case "datetime":
		data := params.Data().(CreatePropertyDatetimeParams)
		property.Type = models.PropertyTypeDatetime
		property.ThingID = thingId
		property.Name = data.Name
		property.ValueDatetime = null.NewTime(data.Value, true)
	case "boolean":
		data := params.Data().(CreatePropertyBooleanParams)
		property.Type = models.PropertyTypeBoolean
		property.ThingID = thingId
		property.Name = data.Name
		property.ValueBoolean = null.NewBool(data.Value, true)
	}
	err = property.Insert(ctx, exec, boil.Infer())
	if err != nil {
		log.Error().Msgf("Failed to insert property: %v", err)
		return nil, err
	}
	return &property, nil
}
