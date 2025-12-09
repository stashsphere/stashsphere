package factories

import (
	"github.com/Pallinder/go-randomdata"
	"github.com/bluele/factory-go/factory"
	"github.com/stashsphere/backend/services"
)

var ListFactory = factory.NewFactory(
	&services.CreateListParams{},
).Attr("Name", func(a factory.Args) (interface{}, error) {
	return randomdata.SillyName(), nil
}).Attr("ThingIds", func(a factory.Args) (interface{}, error) {
	return []string{}, nil
}).Attr("SharingState", func(a factory.Args) (interface{}, error) {
	return "private", nil
})
