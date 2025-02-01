package factories

import (
	"github.com/Pallinder/go-randomdata"
	"github.com/bluele/factory-go/factory"
	"github.com/stashsphere/backend/services"
)

var UserFactory = factory.NewFactory(
	&services.CreateUserParams{},
).Attr("Name", func(a factory.Args) (interface{}, error) {
	return randomdata.FullName(randomdata.RandomGender), nil
}).Attr("Password", func(a factory.Args) (interface{}, error) {
	return randomdata.City(), nil
}).Attr("Email", func(a factory.Args) (interface{}, error) {
	return randomdata.Email(), nil
})
