package propertyschemas

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

type SchemaIcons struct {
	Schema map[string]any `json:"schema"`
	Icons  []string       `json:"icons"`
}

type SchemaCollection struct {
	Schemas map[string]SchemaIcons `json:"schemas"`
	Aliases map[string]string      `json:"aliases"`
}

type SchemaValidator struct {
	schemas []*jsonschema.Schema
}

func NewSchemaCollection() *SchemaCollection {
	return &SchemaCollection{
		Schemas: make(map[string]SchemaIcons),
		Aliases: make(map[string]string),
	}
}

func (sc *SchemaCollection) Hash() (string, error) {
	as_json, err := json.Marshal(sc)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(as_json)
	return fmt.Sprintf("%x", hash), nil
}

func NewRootSchemaValidator() (*SchemaValidator, error) {
	validator := &SchemaValidator{}

	err := fs.WalkDir(ROOT_SCHEMAS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".json") {
			f, err := ROOT_SCHEMAS.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			doc, err := jsonschema.UnmarshalJSON(f)
			if err != nil {
				return err
			}

			c := jsonschema.NewCompiler()
			c.AssertFormat()

			url := "schema://root/" + path
			if err := c.AddResource(url, doc); err != nil {
				return err
			}
			schema, err := c.Compile(url)
			if err != nil {
				return err
			}
			validator.schemas = append(validator.schemas, schema)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return validator, err
}

func readDataFS(fs embed.FS, path string) ([]byte, error) {
	f, err := fs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open schema file %w", err)
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("Could not read schema %w", err)
	}
	return data, nil
}

func readDataLocal(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open schema file %w", err)
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("Could not read schema %w", err)
	}
	return data, nil
}

// reduced version of a schema, easier to unmarshall then going through
// jsonschema when it's not used for validation afterwards
type CommonRootSchema struct {
	Properties struct {
		Name struct {
			Const string `json:"const"`
		} `json:"name"`
	} `json:"properties"`
	Aliases []string `json:"x-aliases"`
	Icons   []string `json:"x-icons"`
}

func (v *SchemaValidator) BuildCollection(localPaths ...string) (*SchemaCollection, error) {
	type DataFileName struct {
		Data []byte
		Path string
	}
	var schemaData []DataFileName
	// upstream schemas
	err := fs.WalkDir(DATA_SCHEMAS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".json") {
			data, err := readDataFS(DATA_SCHEMAS, path)
			if err != nil {
				return err
			}
			schemaData = append(schemaData, DataFileName{
				Data: data,
				Path: path,
			})
		}
		return nil
	})
	// instance schemas
	for _, localPath := range localPaths {
		if strings.HasSuffix(localPath, ".json") {
			data, err := readDataLocal(localPath)
			if err != nil {
				return nil, err
			}
			schemaData = append(schemaData, DataFileName{
				Data: data,
				Path: localPath,
			})
		}
	}
	collection := NewSchemaCollection()

	// schemas are loaded, let's verify them
	for _, el := range schemaData {
		reader := bytes.NewReader(el.Data)

		doc, err := jsonschema.UnmarshalJSON(reader)
		if err != nil {
			return nil, fmt.Errorf("Could not unmarshal JSON for schema %s: %w", el.Path, err)
		}
		var lastErr error
		for _, rootSchema := range v.schemas {
			err := rootSchema.Validate(doc)
			if err == nil {
				var parsed CommonRootSchema
				var raw map[string]any
				// it matches the root schema, extract information
				err = json.Unmarshal(el.Data, &parsed)
				if err != nil {
					return nil, fmt.Errorf("Could not unmarshal JSON for schema %s: %w", el.Path, err)
				}
				err = json.Unmarshal(el.Data, &raw)
				if err != nil {
					return nil, fmt.Errorf("Could not unmarshal JSON for schema %s: %w", el.Path, err)
				}
				name := parsed.Properties.Name.Const
				collection.Schemas[name] = SchemaIcons{
					Schema: raw,
					Icons:  parsed.Icons,
				}
				for _, alias := range parsed.Aliases {
					existing, ok := collection.Aliases[alias]
					if ok {
						return nil, fmt.Errorf("Alias %s already exists for %s. schema: %s", alias, existing, el.Path)
					}
					collection.Aliases[alias] = name
				}
				// we found a matching schema
				lastErr = nil
				break
			} else {
				lastErr = err
			}
		}
		if lastErr != nil {
			// we loop all root schemas and did not found a matching one
			return nil, fmt.Errorf("Schema %s did not match any root schemas. Last error: %w", el.Path, lastErr)
		}
	}

	return collection, err
}
