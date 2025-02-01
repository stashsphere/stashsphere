package factories

import (
	"github.com/Pallinder/go-randomdata"
	"github.com/bluele/factory-go/factory"
	"github.com/stashsphere/backend/services"
)

var ThingFactory = factory.NewFactory(
	&services.CreateThingParams{},
).Attr("Name", func(a factory.Args) (interface{}, error) {
	return randomdata.SillyName(), nil
}).Attr("Description", func(a factory.Args) (interface{}, error) {
	return randomdata.Paragraph(), nil
}).Attr("PrivateNote", func(a factory.Args) (interface{}, error) {
	return "Private!" + randomdata.Paragraph(), nil
})
