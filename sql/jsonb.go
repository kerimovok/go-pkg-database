package sql

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONB represents a PostgreSQL JSONB column
type JSONB map[string]any

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	data, err := json.Marshal(j)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSONB: %w", err)
	}
	return data, nil
}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into JSONB", value)
	}

	if len(data) == 0 {
		*j = make(map[string]any)
		return nil
	}

	var result map[string]any
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&result); err != nil {
		return fmt.Errorf("failed to unmarshal JSONB: %w", err)
	}

	*j = result
	return nil
}

// JSONBArray represents a PostgreSQL JSONB array
type JSONBArray []any

// Value implements the driver.Valuer interface
func (j JSONBArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	data, err := json.Marshal(j)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSONB array: %w", err)
	}
	return data, nil
}

// Scan implements the sql.Scanner interface
func (j *JSONBArray) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into JSONB array", value)
	}

	if len(data) == 0 {
		*j = make([]any, 0)
		return nil
	}

	var result []any
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&result); err != nil {
		return fmt.Errorf("failed to unmarshal JSONB array: %w", err)
	}

	*j = result
	return nil
}
