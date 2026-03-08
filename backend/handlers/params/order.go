package params

import (
	"bytes"
	"fmt"
	"strings"
)

type Order struct {
	FieldName   string `validate:"oneof=created_at access_reason"`
	FieldSortBy string `validate:"oneof=asc desc"`
}

// encoding.TextUnmarshaler
func (o *Order) UnmarshalText(text []byte) error {
	s := bytes.TrimSpace(text)
	if len(text) == 0 {
		return nil
	}
	parts := strings.SplitN(string(s), ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid order format %s", s)
	}

	o.FieldName = parts[0]   // Trimmed key e.g. created_at
	o.FieldSortBy = parts[1] // Trimmed value e.g. desc

	return nil
}
