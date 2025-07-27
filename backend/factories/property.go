package factories

import (
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/bluele/factory-go/factory"
	"github.com/stashsphere/backend/operations"
)

var FloatPropertyFactory = factory.NewFactory(&operations.CreatePropertyFloatParams{}).Attr("Name", func(factory.Args) (interface{}, error) {
	return randomdata.SillyName(), nil
}).Attr("Value", func(a factory.Args) (interface{}, error) {
	return randomdata.Decimal(0, 10000), nil
})

var StringPropertyFactory = factory.NewFactory(&operations.CreatePropertyStringParams{}).Attr("Name", func(factory.Args) (interface{}, error) {
	return randomdata.SillyName(), nil
}).Attr("Value", func(a factory.Args) (interface{}, error) {
	return randomdata.SillyName(), nil
})

var DatetimePropertyFactory = factory.NewFactory(&operations.CreatePropertyDatetimeParams{}).Attr("Name", func(factory.Args) (interface{}, error) {
	return randomdata.SillyName(), nil
}).Attr("Value", func(a factory.Args) (interface{}, error) {
	return time.Unix(int64(randomdata.Number(996154451, 2289994450)), 0), nil
})
