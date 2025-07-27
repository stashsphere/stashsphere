package factories

import (
	"github.com/Pallinder/go-randomdata"
	"github.com/bluele/factory-go/factory"
	"github.com/stashsphere/backend/operations"
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
}).Attr("Properties", func(a factory.Args) (interface{}, error) {
	properties := []operations.CreatePropertyParams{}
	amount := randomdata.Number(1, 10)
	for i := 0; i < amount; i++ {
		selection := randomdata.Number(0, 3)
		if selection == 0 {
			property, err := FloatPropertyFactory.Create()
			if err != nil {
				return nil, err
			}
			properties = append(properties, property.(*operations.CreatePropertyFloatParams))
		} else if selection == 1 {
			property, err := StringPropertyFactory.Create()
			if err != nil {
				return nil, err
			}
			properties = append(properties, property.(*operations.CreatePropertyStringParams))
		} else if selection == 2 {
			property, err := DatetimePropertyFactory.Create()
			if err != nil {
				return nil, err
			}
			properties = append(properties, property.(*operations.CreatePropertyDatetimeParams))
		}
	}
	return properties, nil
})
