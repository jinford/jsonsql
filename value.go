package jsonsql

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

// Compile-time interface satisfaction checks
var (
	_ sql.Scanner   = (*Value[struct{}])(nil)
	_ driver.Valuer = Value[struct{}]{}
)

// ErrNullNotAllowed is returned when Scan receives nil for Value[T] (NOT NULL).
var ErrNullNotAllowed = errors.New("jsonsql: null value not allowed for NOT NULL field")

// Value[T] is a generic type for NOT NULL JSON columns.
// It wraps any type T and provides Scan/Value methods for database/sql compatibility.
type Value[T any] struct {
	V T
}

// NewValue creates a new Value[T] with the given value.
func NewValue[T any](v T) Value[T] {
	return Value[T]{V: v}
}

// Get returns the value.
func (v Value[T]) Get() T {
	return v.V
}

// Scan implements sql.Scanner interface.
// It unmarshals JSON data from the database into V.
// Returns ErrNullNotAllowed if src is nil or JSON literal "null" (NOT NULL constraint violation).
func (v *Value[T]) Scan(src any) error {
	if src == nil {
		return ErrNullNotAllowed
	}

	var data []byte
	switch s := src.(type) {
	case []byte:
		data = s
	case string:
		data = []byte(s)
	case json.RawMessage:
		data = s
	default:
		return fmt.Errorf("jsonsql.Value.Scan: unsupported type %T", src)
	}

	// JSON literal null (with optional whitespace) is not allowed for NOT NULL field
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return ErrNullNotAllowed
	}

	if err := json.Unmarshal(data, &v.V); err != nil {
		return fmt.Errorf("jsonsql.Value.Scan: %w", err)
	}
	return nil
}

// Value implements driver.Valuer interface.
// It marshals V to JSON bytes for database storage.
func (v Value[T]) Value() (driver.Value, error) {
	data, err := json.Marshal(v.V)
	if err != nil {
		return nil, fmt.Errorf("jsonsql.Value.Value: %w", err)
	}
	return data, nil
}
