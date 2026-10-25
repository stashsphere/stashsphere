package cmd

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// convertNullableTo31 rewrites the OpenAPI 3.0 `nullable: true` keyword that
// kin-openapi generates into the OpenAPI 3.1 form `type: [<type>, "null"]`.
// Approach from https://github.com/go-fuego/fuego/pull/838#issuecomment-5142427264
// Only schemas that carry `nullable` themselves are rewritten, so referenced
// component schemas do not become nullable for all of their users (the
// regression discussed in that PR).
func convertNullableTo31(doc *openapi3.T) {
	walkSchemas(doc, func(schema *openapi3.Schema) {
		if !schema.Nullable || schema.Type == nil || len(*schema.Type) == 0 {
			return
		}
		if !schema.Type.Includes(openapi3.TypeNull) {
			*schema.Type = append(*schema.Type, openapi3.TypeNull)
		}
		schema.Nullable = false
	})
}

// walkSchemas calls fn once for every schema reachable from the document.
func walkSchemas(doc *openapi3.T, fn func(*openapi3.Schema)) {
	seen := map[*openapi3.Schema]bool{}
	walkSchemaRefs(doc, func(ref *openapi3.SchemaRef) {
		if !seen[ref.Value] {
			seen[ref.Value] = true
			fn(ref.Value)
		}
	})
}

// walkSchemaRefs calls fn for every schema reference reachable from the
// document before descending into the schema it points to, so fn may replace
// the reference. Each schema is descended into once.
func walkSchemaRefs(doc *openapi3.T, fn func(*openapi3.SchemaRef)) {
	seen := map[*openapi3.Schema]bool{}
	var walk func(ref *openapi3.SchemaRef)
	walk = func(ref *openapi3.SchemaRef) {
		if ref == nil || ref.Value == nil {
			return
		}
		fn(ref)
		schema := ref.Value
		if seen[schema] {
			return
		}
		seen[schema] = true

		for _, property := range schema.Properties {
			walk(property)
		}
		walk(schema.Items)
		walk(schema.AdditionalProperties.Schema)
		walk(schema.Not)
		for _, list := range []openapi3.SchemaRefs{schema.AllOf, schema.AnyOf, schema.OneOf} {
			for _, item := range list {
				walk(item)
			}
		}
	}
	walkContent := func(content openapi3.Content) {
		for _, mediaType := range content {
			walk(mediaType.Schema)
		}
	}
	walkParameters := func(parameters openapi3.Parameters) {
		for _, parameter := range parameters {
			if parameter.Value != nil {
				walk(parameter.Value.Schema)
				walkContent(parameter.Value.Content)
			}
		}
	}

	if doc.Components != nil {
		for _, schema := range doc.Components.Schemas {
			walk(schema)
		}
	}
	if doc.Paths == nil {
		return
	}
	for _, pathItem := range doc.Paths.Map() {
		walkParameters(pathItem.Parameters)
		for _, operation := range pathItem.Operations() {
			walkParameters(operation.Parameters)
			if operation.RequestBody != nil && operation.RequestBody.Value != nil {
				walkContent(operation.RequestBody.Value.Content)
			}
			if operation.Responses == nil {
				continue
			}
			for _, response := range operation.Responses.Map() {
				if response.Value == nil {
					continue
				}
				walkContent(response.Value.Content)
				for _, header := range response.Value.Headers {
					if header.Value != nil {
						walk(header.Value.Schema)
					}
				}
			}
		}
	}
}
