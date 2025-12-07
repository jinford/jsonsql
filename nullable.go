package jsonsql

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Compile-time interface satisfaction checks
var (
	_ sql.Scanner   = (*Nullable[struct{}])(nil)
	_ driver.Valuer = Nullable[struct{}]{}
)

// Nullable[T] is a generic type for NULL-able JSON columns.
// Valid indicates whether V holds a valid value.
// When Valid is false, the value represents NULL.
type Nullable[T any] struct {
	V     T
	Valid bool
}

// NewNullable creates a new Nullable[T] with the given value and valid flag.
// If valid is false, V is set to the zero value of T.
func NewNullable[T any](v T, valid bool) Nullable[T] {
	if !valid {
		return Null[T]()
	}
	return NullableFrom(v)
}

// NullableFrom creates a new Nullable[T] with Valid=true and the given value.
func NullableFrom[T any](v T) Nullable[T] {
	return Nullable[T]{V: v, Valid: true}
}

// Null creates a new Nullable[T] with Valid=false (represents NULL).
func Null[T any]() Nullable[T] {
	return Nullable[T]{Valid: false}
}

// NullableFromPtr creates a Nullable[T] from a pointer.
// Returns Null[T]() if ptr is nil.
func NullableFromPtr[T any](ptr *T) Nullable[T] {
	if ptr == nil {
		return Null[T]()
	}
	return NullableFrom(*ptr)
}

// ToPtr returns a pointer to the value if Valid is true, otherwise nil.
func (n Nullable[T]) ToPtr() *T {
	if !n.Valid {
		return nil
	}
	return &n.V
}

// Get returns the value and a boolean indicating whether it is valid.
func (n Nullable[T]) Get() (T, bool) {
	return n.V, n.Valid
}

// Scan implements sql.Scanner interface.
// It unmarshals JSON data from the database into V.
// Sets Valid=false for nil, empty []byte, empty string, or JSON literal "null".
func (n *Nullable[T]) Scan(src any) error {
	if src == nil {
		n.Valid = false
		var zero T
		n.V = zero
		return nil
	}

	var data []byte
	switch s := src.(type) {
	case []byte:
		if len(s) == 0 {
			n.Valid = false
			var zero T
			n.V = zero
			return nil
		}
		data = s
	case string:
		if len(s) == 0 {
			n.Valid = false
			var zero T
			n.V = zero
			return nil
		}
		data = []byte(s)
	case json.RawMessage:
		if len(s) == 0 {
			n.Valid = false
			var zero T
			n.V = zero
			return nil
		}
		data = s
	default:
		return fmt.Errorf("jsonsql.Nullable.Scan: unsupported type %T", src)
	}

	// JSON literal null (with optional whitespace) should be treated as NULL (Valid=false)
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		n.Valid = false
		var zero T
		n.V = zero
		return nil
	}

	if err := json.Unmarshal(data, &n.V); err != nil {
		return fmt.Errorf("jsonsql.Nullable.Scan: %w", err)
	}
	n.Valid = true
	return nil
}

// Value implements driver.Valuer interface.
// Returns nil (NULL) when Valid is false.
// Otherwise marshals V to JSON bytes.
func (n Nullable[T]) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	data, err := json.Marshal(n.V)
	if err != nil {
		return nil, fmt.Errorf("jsonsql.Nullable.Value: %w", err)
	}
	return data, nil
}
