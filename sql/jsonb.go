package sql

import (
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
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("failed to unmarshal JSONB: %w", err)
	}

	*j = result
	return nil
}

// String returns the JSON string representation
func (j JSONB) String() string {
	if j == nil {
		return "null"
	}
	data, err := json.Marshal(j)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// Set sets a value in the JSONB
func (j JSONB) Set(key string, value any) {
	if j == nil {
		return
	}
	j[key] = value
}

// Get retrieves a value from the JSONB
func (j JSONB) Get(key string) (any, bool) {
	if j == nil {
		return nil, false
	}
	value, exists := j[key]
	return value, exists
}

// GetString retrieves a string value from the JSONB
func (j JSONB) GetString(key string) (string, bool) {
	value, exists := j.Get(key)
	if !exists {
		return "", false
	}
	str, ok := value.(string)
	return str, ok
}

// GetInt retrieves an int value from the JSONB
func (j JSONB) GetInt(key string) (int, bool) {
	value, exists := j.Get(key)
	if !exists {
		return 0, false
	}

	switch v := value.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return int(i), true
		}
	}
	return 0, false
}

// GetBool retrieves a bool value from the JSONB
func (j JSONB) GetBool(key string) (bool, bool) {
	value, exists := j.Get(key)
	if !exists {
		return false, false
	}
	boolean, ok := value.(bool)
	return boolean, ok
}

// Delete removes a key from the JSONB
func (j JSONB) Delete(key string) {
	if j == nil {
		return
	}
	delete(j, key)
}

// Has checks if a key exists in the JSONB
func (j JSONB) Has(key string) bool {
	if j == nil {
		return false
	}
	_, exists := j[key]
	return exists
}

// Keys returns all keys in the JSONB
func (j JSONB) Keys() []string {
	if j == nil {
		return nil
	}
	keys := make([]string, 0, len(j))
	for k := range j {
		keys = append(keys, k)
	}
	return keys
}

// IsEmpty checks if the JSONB is empty
func (j JSONB) IsEmpty() bool {
	return len(j) == 0
}

// Clone creates a deep copy of the JSONB
func (j JSONB) Clone() JSONB {
	if j == nil {
		return nil
	}

	clone := make(JSONB, len(j))
	for k, v := range j {
		clone[k] = v
	}
	return clone
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
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("failed to unmarshal JSONB array: %w", err)
	}

	*j = result
	return nil
}
