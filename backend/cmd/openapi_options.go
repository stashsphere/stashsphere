package cmd

import (
	"strconv"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-fuego/fuego"
)

// noContentResponse documents a response without a body.
// fuego's option.AddResponse always attaches a body schema.
func noContentResponse(code int, description string) func(*fuego.BaseRoute) {
	return func(r *fuego.BaseRoute) {
		if r.Operation.Responses == nil {
			r.Operation.Responses = openapi3.NewResponses()
		}
		response := openapi3.NewResponse().WithDescription(description)
		// non-nil but empty: fuego only adds a body to responses without content
		response.Content = openapi3.Content{}
		r.Operation.Responses.Set(strconv.Itoa(code), &openapi3.ResponseRef{Value: response})
	}
}

// redirectResponse documents a bodiless redirect response and makes it the
// route's default, so fuego does not invent a 200 with an untyped body.
func redirectResponse(code int, description string) func(*fuego.BaseRoute) {
	return func(r *fuego.BaseRoute) {
		r.DefaultStatusCode = code
		noContentResponse(code, description)(r)
		r.Operation.Responses.Value(strconv.Itoa(code)).Value.Headers = openapi3.Headers{
			"Location": &openapi3.HeaderRef{
				Value: &openapi3.Header{
					Parameter: openapi3.Parameter{
						Description: "Redirect target",
						Schema:      openapi3.NewStringSchema().NewRef(),
					},
				},
			},
		}
	}
}
