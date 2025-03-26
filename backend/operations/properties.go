package operations

import (
	"context"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/rs/zerolog/log"
	"github.com/stashsphere/backend/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
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

type CreatePropertyParams interface {
	Data() interface{}
	Type() string
}

func (p CreatePropertyFloatParams) Data() interface{}    { return p }
func (p CreatePropertyDatetimeParams) Data() interface{} { return p }
func (p CreatePropertyStringParams) Data() interface{}   { return p }

func (p CreatePropertyFloatParams) Type() string    { return "float" }
func (p CreatePropertyDatetimeParams) Type() string { return "datetime" }
func (p CreatePropertyStringParams) Type() string   { return "string" }

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
	}
	err = property.Insert(ctx, exec, boil.Infer())
	if err != nil {
		log.Error().Msgf("Failed to insert property: %v", err)
		return nil, err
	}
	return &property, nil
}
