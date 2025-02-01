package resources

import (
	"math"
	"time"

	"github.com/stashsphere/backend/models"
)

type PropertyType string

const (
	DateTime PropertyType = "datetime"
	String   PropertyType = "string"
	Float    PropertyType = "float"
)

type PropertyDatetime struct {
	Type  string    `json:"type"`
	Name  string    `json:"name"`
	Value time.Time `json:"value"`
}

type PropertyString struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PropertyFloat struct {
	Type  string  `json:"type"`
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

func PropertyFromModel(property *models.Property) interface{} {
	switch property.Type {
	case models.PropertyTypeDatetime:
		var valueDatetime time.Time
		if property.ValueDatetime.Valid {
			valueDatetime = property.ValueDatetime.Time
		} else {
			valueDatetime = time.Now()
		}
		return &PropertyDatetime{
			Type:  "datetime",
			Name:  property.Name,
			Value: valueDatetime,
		}
	case models.PropertyTypeString:
		var valueString string
		if property.ValueString.Valid {
			valueString = property.ValueString.String
		} else {
			valueString = "unknown"
		}
		return &PropertyString{
			Type:  "string",
			Name:  property.Name,
			Value: valueString,
		}
	case models.PropertyTypeFloat:
		var unit string
		if property.Unit.Valid {
			unit = property.Unit.String
		} else {
			unit = "unknown"
		}
		var valueAsFloat float64
		if property.ValueFloat.Valid {
			valueAsFloat = property.ValueFloat.Float64
		} else {
			valueAsFloat = math.NaN()
		}
		return &PropertyFloat{
			Type:  "float",
			Name:  property.Name,
			Value: valueAsFloat,
			Unit:  unit,
		}
	default:
		return nil
	}
}

func PropertiesFromModelSlice(mProperties models.PropertySlice) []interface{} {
	properties := make([]interface{}, len(mProperties))
	for i, mProperty := range mProperties {
		properties[i] = PropertyFromModel(mProperty)
	}
	return properties
}
