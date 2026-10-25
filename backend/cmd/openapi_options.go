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

// multipartFileBody documents a multipart/form-data request body with a single
// required `file` field of the given content type.
// fuego derives request bodies from Go types and cannot express file parts.
func multipartFileBody(description string, fileContentType string) func(*fuego.BaseRoute) {
	return func(r *fuego.BaseRoute) {
		file := openapi3.NewStringSchema().WithFormat("binary")
		file.Description = description
		schema := openapi3.NewObjectSchema().
			WithProperty("file", file).
			WithRequired([]string{"file"})
		mediaType := openapi3.NewMediaType().WithSchema(schema)
		mediaType.Encoding = map[string]*openapi3.Encoding{
			"file": {ContentType: fileContentType},
		}
		r.Operation.RequestBody = &openapi3.RequestBodyRef{
			Value: openapi3.NewRequestBody().
				WithDescription(description).
				WithRequired(true).
				WithContent(openapi3.Content{"multipart/form-data": mediaType}),
		}
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
