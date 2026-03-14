package propertyschemas_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	propertyschemas "github.com/stashsphere/backend/property_schemas"
	"github.com/stretchr/testify/assert"
)

func TestValidatorSetup(t *testing.T) {
	validator, err := propertyschemas.NewRootSchemaValidator()
	assert.NoError(t, err)
	assert.NotNil(t, validator)
}

func TestCollectionSimple(t *testing.T) {
	validator, err := propertyschemas.NewRootSchemaValidator()
	assert.NoError(t, err)
	collection, err := validator.BuildCollection()
	assert.NoError(t, err)
	assert.NotNil(t, collection)
	assert.NotEmpty(t, collection.Schemas)
	assert.NotEmpty(t, collection.Aliases)
}

func copyFileToTestDir(fileName string, target string) error {
	valid1, err := propertyschemas.TEST_DATA_SCHEMAS.Open(fileName)
	if err != nil {
		return err
	}
	defer valid1.Close()

	dest, err := os.Create(target)
	if err != nil {
		return err
	}
	defer dest.Close()
	_, err = io.Copy(dest, valid1)
	if err != nil {
		return err
	}
	return nil
}

func TestBuildCollectionWithValidSchema(t *testing.T) {
	valids := []string{"valid_boolean.json", "valid_string.json"}

	for _, valid := range valids {
		testFile := filepath.Join("test_data", valid)

		validator, err := propertyschemas.NewRootSchemaValidator()
		assert.NoError(t, err)

		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test_name_schema.json")

		err = copyFileToTestDir(testFile, tmpFile)
		assert.NoError(t, err)

		collection, err := validator.BuildCollection(testFile)
		assert.NoError(t, err)
		assert.NotNil(t, collection)
		assert.Contains(t, collection.Schemas, "custom_name")
		assert.Contains(t, collection.Aliases, "cn")
		assert.Equal(t, "custom_name", collection.Aliases["cn"])
	}
}

func TestBuildCollectionWithInvalidSchema(t *testing.T) {
	invalids := []string{"invalid_1.json", "invalid_2.json"}

	for _, invalid := range invalids {
		testFile := filepath.Join("test_data", invalid)

		validator, err := propertyschemas.NewRootSchemaValidator()
		assert.NoError(t, err)

		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "invalid_schema.json")

		err = copyFileToTestDir(testFile, tmpFile)
		assert.NoError(t, err)

		collection, err := validator.BuildCollection(tmpFile)
		assert.Error(t, err)
		assert.Nil(t, collection)
		assert.Contains(t, err.Error(), "did not match any root schemas")

	}
}
