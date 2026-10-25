package cmd

import (
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-fuego/fuego"
)

// openAPIUnion describes a Go interface type whose values are one of several
// structs, tagged in JSON by the `discriminator` field.
type openAPIUnion struct {
	iface         reflect.Type
	discriminator string
	// variants maps the discriminator value to a zero value of the struct
	variants map[string]any
	// component, if set, also publishes the union as components/schemas/<component>
	// so it can be used as a response with componentResponse
	component string
}

// registerOpenAPITypes installs fuego's schema customizer, which can only be
// set once. Types in `overrides` are documented with the given schema.
// Unions document interface-typed fields as `oneOf` the variant schemas with a
// discriminator. Without this, fuego renders both as untyped schemas.
// Must be called before routes are registered.
func registerOpenAPITypes(openAPI *fuego.OpenAPI, overrides map[reflect.Type]*openapi3.Schema, unions ...openAPIUnion) {
	schemas := map[reflect.Type]*openapi3.Schema{}
	for t, schema := range overrides {
		schemas[t] = schema
	}
	for _, union := range unions {
		schema := &openapi3.Schema{
			Discriminator: &openapi3.Discriminator{
				PropertyName: union.discriminator,
				Mapping:      openapi3.StringMap{},
			},
		}
		for tag, variant := range union.variants {
			ref := fuego.SchemaTagFromType(openAPI, variant).Ref
			pinDiscriminator(openAPI.Description().Components.Schemas, ref, union.discriminator, tag)
			schema.OneOf = append(schema.OneOf, openapi3.NewSchemaRef(ref, nil))
			schema.Discriminator.Mapping[tag] = ref
		}
		sort.Slice(schema.OneOf, func(i, j int) bool { return schema.OneOf[i].Ref < schema.OneOf[j].Ref })
		if union.component != "" {
			names := make([]string, len(schema.OneOf))
			for i, ref := range schema.OneOf {
				names[i] = strings.TrimPrefix(ref.Ref, "#/components/schemas/")
			}
			schema.Description = "One of " + strings.Join(names, ", ") + ", discriminated by `" + union.discriminator + "`"
			openAPI.Description().Components.Schemas[union.component] = schema.NewRef()
		}
		schemas[union.iface] = schema
	}

	openAPI.SetGeneratorSchemaCustomizer(func(_ string, t reflect.Type, _ reflect.StructTag, schema *openapi3.Schema) error {
		if override, ok := schemas[t]; ok {
			*schema = *override
		}
		return nil
	})
}

// pinDiscriminator makes the discriminator a required property of the variant
// schema that only allows the variant's tag.
func pinDiscriminator(components openapi3.Schemas, ref string, property string, tag string) {
	schemaRef, ok := components[strings.TrimPrefix(ref, "#/components/schemas/")]
	if !ok || schemaRef.Value == nil {
		return
	}
	variant := schemaRef.Value
	if variant.Properties == nil {
		variant.Properties = openapi3.Schemas{}
	}
	variant.Properties[property] = openapi3.NewStringSchema().WithEnum(tag).NewRef()
	for _, required := range variant.Required {
		if required == property {
			return
		}
	}
	variant.Required = append(variant.Required, property)
}

// componentResponse documents a JSON response whose body is the named
// component schema, e.g. a union registered with registerOpenAPITypes.
func componentResponse(code int, description string, component string) func(*fuego.BaseRoute) {
	return func(r *fuego.BaseRoute) {
		if r.Operation.Responses == nil {
			r.Operation.Responses = openapi3.NewResponses()
		}
		content := openapi3.NewContentWithJSONSchemaRef(openapi3.NewSchemaRef("#/components/schemas/"+component, nil))
		response := openapi3.NewResponse().WithDescription(description).WithContent(content)
		r.Operation.Responses.Set(strconv.Itoa(code), &openapi3.ResponseRef{Value: response})
	}
}
