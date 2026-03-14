package propertyschemas

import "embed"

//go:embed root/*.json
var ROOT_SCHEMAS embed.FS

//go:embed data/*.json
var DATA_SCHEMAS embed.FS

//go:embed test_data/*.json
var TEST_DATA_SCHEMAS embed.FS
