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
	// component is the name the union is published under in
	// components/schemas; componentResponse can use it as a response body
	component string
}

// registerOpenAPITypes installs fuego's schema customizer, which can only be
// set once. Types in `overrides` are documented with the given schema.
// Unions document interface-typed fields as `oneOf` the variant schemas with a
// discriminator. Without this, fuego renders both as untyped schemas.
// Must be called before routes are registered. The returned function must be
// called on the generated document to reference the union components, see
// useUnionComponents.
func registerOpenAPITypes(openAPI *fuego.OpenAPI, overrides map[reflect.Type]*openapi3.Schema, unions ...openAPIUnion) func(*openapi3.T) {
	schemas := map[reflect.Type]*openapi3.Schema{}
	for t, schema := range overrides {
		schemas[t] = schema
	}
	components := map[reflect.Type]string{}
	// inline copies of union schemas generated for fields, by component name
	unionCopies := map[*openapi3.Schema]string{}
	for _, union := range unions {
		schema := &openapi3.Schema{
			Discriminator: &openapi3.Discriminator{
				PropertyName: union.discriminator,
				Mapping:      map[string]openapi3.MappingRef{},
			},
		}
		for tag, variant := range union.variants {
			variantTag := fuego.SchemaTagFromType(openAPI, variant)
			ref := variantTag.Ref
			pinDiscriminator(openAPI.Description().Components.Schemas, ref, union.discriminator, tag)
			schema.OneOf = append(schema.OneOf, openapi3.NewSchemaRef(ref, variantTag.Value))
			schema.Discriminator.Mapping[tag] = openapi3.MappingRef{Ref: ref}
		}
		sort.Slice(schema.OneOf, func(i, j int) bool { return schema.OneOf[i].Ref < schema.OneOf[j].Ref })
		names := make([]string, len(schema.OneOf))
		for i, ref := range schema.OneOf {
			names[i] = strings.TrimPrefix(ref.Ref, "#/components/schemas/")
		}
		schema.Description = "One of " + strings.Join(names, ", ") + ", discriminated by `" + union.discriminator + "`"
		openAPI.Description().Components.Schemas[union.component] = schema.NewRef()
		schemas[union.iface] = schema
		components[union.iface] = union.component
	}

	openAPI.SetGeneratorSchemaCustomizer(func(_ string, t reflect.Type, _ reflect.StructTag, schema *openapi3.Schema) error {
		if override, ok := schemas[t]; ok {
			nullable := schema.Nullable
			*schema = *override
			schema.Nullable = nullable
		}
		if component, ok := components[t]; ok {
			unionCopies[schema] = component
		}
		if t.Kind() == reflect.Struct {
			nullableRefFields(t, schema)
		}
		return nil
	})

	return func(doc *openapi3.T) {
		useUnionComponents(doc, unionCopies)
	}
}

// useUnionComponents replaces the inline union schemas that the generator
// creates for every interface-typed field with a reference to the union
// component. The generator only exports structs as components, so without
// this every field would get its own copy of the union, which client
// generators turn into one distinct type per field.
func useUnionComponents(doc *openapi3.T, unionCopies map[*openapi3.Schema]string) {
	walkSchemaRefs(doc, func(ref *openapi3.SchemaRef) {
		component, ok := unionCopies[ref.Value]
		if !ok || ref.Ref != "" {
			return
		}
		componentRef := openapi3.NewSchemaRef("#/components/schemas/"+component, doc.Components.Schemas[component].Value)
		if !ref.Value.Nullable {
			*ref = *componentRef
			return
		}
		*ref = *openapi3.NewSchemaRef("", &openapi3.Schema{
			AnyOf: openapi3.SchemaRefs{
				componentRef,
				openapi3.NewSchemaRef("", &openapi3.Schema{Type: &openapi3.Types{openapi3.TypeNull}}),
			},
		})
	})
}

// nullableRefFields marks properties that reference a component schema and
// can be JSON null (a pointer field without omitempty) as `anyOf` the reference
// and null. The generator drops nullability at $ref sites so the shared
// component itself does not become nullable for all of its users.
func nullableRefFields(t reflect.Type, schema *openapi3.Schema) {
	for i := range t.NumField() {
		field := t.Field(i)
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			nullableRefFields(field.Type, schema)
			continue
		}
		if !field.IsExported() || field.Type.Kind() != reflect.Pointer {
			continue
		}
		name, options, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "-" || strings.Contains(options, "omitempty") {
			continue
		}
		if name == "" {
			name = field.Name
		}
		property, ok := schema.Properties[name]
		if !ok || property.Ref == "" {
			continue
		}
		schema.Properties[name] = openapi3.NewSchemaRef("", &openapi3.Schema{
			AnyOf: openapi3.SchemaRefs{
				property,
				openapi3.NewSchemaRef("", &openapi3.Schema{Type: &openapi3.Types{openapi3.TypeNull}}),
			},
		})
	}
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
		// attach the schema so the spec validator can resolve the ref
		var componentSchema *openapi3.Schema
		if schemaRef, ok := r.OpenAPI.Description().Components.Schemas[component]; ok {
			componentSchema = schemaRef.Value
		}
		content := openapi3.NewContentWithJSONSchemaRef(openapi3.NewSchemaRef("#/components/schemas/"+component, componentSchema))
		response := openapi3.NewResponse().WithDescription(description).WithContent(content)
		r.Operation.Responses.Set(strconv.Itoa(code), &openapi3.ResponseRef{Value: response})
	}
}
