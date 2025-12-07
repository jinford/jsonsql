package jsonsql

import (
	"encoding/json"
	"testing"
)

func TestNullable_Scan_Struct(t *testing.T) {
	input := []byte(`{"name":"Alice","email":"alice@example.com"}`)
	var n Nullable[testProfile]

	if err := n.Scan(input); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if !n.Valid {
		t.Error("expected Valid=true")
	}
	if n.V.Name != "Alice" {
		t.Errorf("expected Name=Alice, got %s", n.V.Name)
	}
	if n.V.Email != "alice@example.com" {
		t.Errorf("expected Email=alice@example.com, got %s", n.V.Email)
	}
}

func TestNullable_Scan_String(t *testing.T) {
	input := `"hello"`
	var n Nullable[string]

	if err := n.Scan(input); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if !n.Valid {
		t.Error("expected Valid=true")
	}
	if n.V != "hello" {
		t.Errorf("expected 'hello', got %s", n.V)
	}
}

func TestNullable_Scan_RawMessage(t *testing.T) {
	input := json.RawMessage(`{"test":123}`)
	var n Nullable[map[string]int]

	if err := n.Scan(input); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if !n.Valid {
		t.Error("expected Valid=true")
	}
	if n.V["test"] != 123 {
		t.Errorf("expected test=123, got %v", n.V["test"])
	}
}

func TestNullable_Scan_Nil(t *testing.T) {
	n := Nullable[testProfile]{
		V:     testProfile{Name: "Previous"},
		Valid: true,
	}

	if err := n.Scan(nil); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if n.Valid {
		t.Error("expected Valid=false for nil")
	}
	if n.V.Name != "" {
		t.Errorf("expected zero value, got %+v", n.V)
	}
}

func TestNullable_Scan_EmptyBytes(t *testing.T) {
	var n Nullable[testProfile]

	if err := n.Scan([]byte{}); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if n.Valid {
		t.Error("expected Valid=false for empty bytes")
	}
}

func TestNullable_Scan_EmptyString(t *testing.T) {
	var n Nullable[testProfile]

	if err := n.Scan(""); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if n.Valid {
		t.Error("expected Valid=false for empty string")
	}
}

func TestNullable_Scan_EmptyRawMessage(t *testing.T) {
	var n Nullable[testProfile]

	if err := n.Scan(json.RawMessage{}); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if n.Valid {
		t.Error("expected Valid=false for empty RawMessage")
	}
}

func TestNullable_Scan_JSONNull_Bytes(t *testing.T) {
	n := Nullable[testProfile]{
		V:     testProfile{Name: "Previous"},
		Valid: true,
	}

	if err := n.Scan([]byte("null")); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if n.Valid {
		t.Error("expected Valid=false for JSON null")
	}
	if n.V.Name != "" {
		t.Errorf("expected zero value, got %+v", n.V)
	}
}

func TestNullable_Scan_JSONNull_String(t *testing.T) {
	n := Nullable[testProfile]{
		V:     testProfile{Name: "Previous"},
		Valid: true,
	}

	if err := n.Scan("null"); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if n.Valid {
		t.Error("expected Valid=false for JSON null string")
	}
	if n.V.Name != "" {
		t.Errorf("expected zero value, got %+v", n.V)
	}
}

func TestNullable_Scan_JSONNull_RawMessage(t *testing.T) {
	n := Nullable[testProfile]{
		V:     testProfile{Name: "Previous"},
		Valid: true,
	}

	if err := n.Scan(json.RawMessage("null")); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if n.Valid {
		t.Error("expected Valid=false for JSON null RawMessage")
	}
	if n.V.Name != "" {
		t.Errorf("expected zero value, got %+v", n.V)
	}
}

func TestNullable_Scan_JSONNull_WithWhitespace(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{"leading space", []byte(" null")},
		{"trailing space", []byte("null ")},
		{"both spaces", []byte(" null ")},
		{"newline", []byte("\nnull\n")},
		{"tabs", []byte("\tnull\t")},
		{"mixed whitespace", []byte(" \t\nnull \n")},
		{"string with space", " null "},
		{"RawMessage with space", json.RawMessage(" null ")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := Nullable[testProfile]{
				V:     testProfile{Name: "Previous"},
				Valid: true,
			}

			if err := n.Scan(tt.input); err != nil {
				t.Fatalf("Scan failed: %v", err)
			}

			if n.Valid {
				t.Errorf("expected Valid=false for %q", tt.name)
			}
			if n.V.Name != "" {
				t.Errorf("expected zero value, got %+v", n.V)
			}
		})
	}
}

func TestNullable_Scan_UnsupportedType(t *testing.T) {
	var n Nullable[testProfile]

	err := n.Scan(123)
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

func TestNullable_Scan_InvalidJSON(t *testing.T) {
	var n Nullable[testProfile]

	err := n.Scan([]byte(`{invalid}`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestNullable_Value_Valid(t *testing.T) {
	n := Nullable[testProfile]{
		V:     testProfile{Name: "Bob", Email: "bob@example.com"},
		Valid: true,
	}

	result, err := n.Value()
	if err != nil {
		t.Fatalf("Value failed: %v", err)
	}

	data, ok := result.([]byte)
	if !ok {
		t.Fatalf("expected []byte, got %T", result)
	}

	var parsed testProfile
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if parsed.Name != "Bob" || parsed.Email != "bob@example.com" {
		t.Errorf("unexpected result: %+v", parsed)
	}
}

func TestNullable_Value_Invalid(t *testing.T) {
	n := Nullable[testProfile]{
		V:     testProfile{Name: "Bob"},
		Valid: false,
	}

	result, err := n.Value()
	if err != nil {
		t.Fatalf("Value failed: %v", err)
	}

	if result != nil {
		t.Errorf("expected nil for invalid, got %v", result)
	}
}

func TestNullable_ZeroValue_IsNull(t *testing.T) {
	var n Nullable[testProfile]

	result, err := n.Value()
	if err != nil {
		t.Fatalf("Value failed: %v", err)
	}

	if result != nil {
		t.Errorf("expected nil for zero value, got %v", result)
	}
}

func TestNullable_Roundtrip(t *testing.T) {
	original := Nullable[testProfile]{
		V:     testProfile{Name: "Charlie", Email: "charlie@example.com"},
		Valid: true,
	}

	data, err := original.Value()
	if err != nil {
		t.Fatalf("Value failed: %v", err)
	}

	var restored Nullable[testProfile]
	if err := restored.Scan(data); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if !restored.Valid {
		t.Error("expected Valid=true after roundtrip")
	}
	if restored.V != original.V {
		t.Errorf("roundtrip failed: expected %+v, got %+v", original.V, restored.V)
	}
}

func TestNullable_Roundtrip_Null(t *testing.T) {
	original := Nullable[testProfile]{Valid: false}

	data, err := original.Value()
	if err != nil {
		t.Fatalf("Value failed: %v", err)
	}

	var restored Nullable[testProfile]
	if err := restored.Scan(data); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if restored.Valid {
		t.Error("expected Valid=false after null roundtrip")
	}
}

func TestNullable_Map(t *testing.T) {
	n := Nullable[map[string]any]{
		V:     map[string]any{"key": "value", "count": 42},
		Valid: true,
	}

	data, err := n.Value()
	if err != nil {
		t.Fatalf("Value failed: %v", err)
	}

	var restored Nullable[map[string]any]
	if err := restored.Scan(data); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if !restored.Valid {
		t.Error("expected Valid=true")
	}
	if restored.V["key"] != "value" {
		t.Errorf("expected key=value, got %v", restored.V["key"])
	}
}

func TestNullable_Value_MarshalError(t *testing.T) {
	n := Nullable[unmarshalableType]{
		V:     unmarshalableType{Ch: make(chan int)},
		Valid: true,
	}

	_, err := n.Value()
	if err == nil {
		t.Fatal("expected error for unmarshalable type")
	}
}

func TestNullable_Scan_Slice(t *testing.T) {
	input := []byte(`[1, 2, 3, 4, 5]`)
	var n Nullable[[]int]

	if err := n.Scan(input); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if !n.Valid {
		t.Error("expected Valid=true")
	}

	expected := []int{1, 2, 3, 4, 5}
	if len(n.V) != len(expected) {
		t.Fatalf("expected length %d, got %d", len(expected), len(n.V))
	}
	for i, val := range expected {
		if n.V[i] != val {
			t.Errorf("expected n.V[%d]=%d, got %d", i, val, n.V[i])
		}
	}
}

func TestNullable_Scan_SliceOfStructs(t *testing.T) {
	input := []byte(`[{"name":"Alice","email":"alice@example.com"},{"name":"Bob","email":"bob@example.com"}]`)
	var n Nullable[[]testProfile]

	if err := n.Scan(input); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if !n.Valid {
		t.Error("expected Valid=true")
	}
	if len(n.V) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(n.V))
	}
	if n.V[0].Name != "Alice" {
		t.Errorf("expected n.V[0].Name=Alice, got %s", n.V[0].Name)
	}
	if n.V[1].Name != "Bob" {
		t.Errorf("expected n.V[1].Name=Bob, got %s", n.V[1].Name)
	}
}

func TestNullable_Scan_EmptySlice(t *testing.T) {
	input := []byte(`[]`)
	var n Nullable[[]int]

	if err := n.Scan(input); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if !n.Valid {
		t.Error("expected Valid=true for empty slice")
	}
	if n.V == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(n.V) != 0 {
		t.Errorf("expected length 0, got %d", len(n.V))
	}
}

func TestNullable_Value_Slice(t *testing.T) {
	n := Nullable[[]string]{
		V:     []string{"a", "b", "c"},
		Valid: true,
	}

	result, err := n.Value()
	if err != nil {
		t.Fatalf("Value failed: %v", err)
	}

	data, ok := result.([]byte)
	if !ok {
		t.Fatalf("expected []byte, got %T", result)
	}

	var parsed []string
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if len(parsed) != 3 || parsed[0] != "a" || parsed[1] != "b" || parsed[2] != "c" {
		t.Errorf("unexpected result: %+v", parsed)
	}
}

func TestNullable_Roundtrip_Slice(t *testing.T) {
	original := Nullable[[]testProfile]{
		V: []testProfile{
			{Name: "Alice", Email: "alice@example.com"},
			{Name: "Bob", Email: "bob@example.com"},
		},
		Valid: true,
	}

	data, err := original.Value()
	if err != nil {
		t.Fatalf("Value failed: %v", err)
	}

	var restored Nullable[[]testProfile]
	if err := restored.Scan(data); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if !restored.Valid {
		t.Error("expected Valid=true after roundtrip")
	}
	if len(restored.V) != len(original.V) {
		t.Fatalf("roundtrip failed: expected %d elements, got %d", len(original.V), len(restored.V))
	}
	for i := range original.V {
		if restored.V[i] != original.V[i] {
			t.Errorf("roundtrip failed at index %d: expected %+v, got %+v", i, original.V[i], restored.V[i])
		}
	}
}

func TestNullable_Slice_Null(t *testing.T) {
	n := Nullable[[]testProfile]{
		V: []testProfile{
			{Name: "Previous"},
		},
		Valid: true,
	}

	if err := n.Scan(nil); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if n.Valid {
		t.Error("expected Valid=false for nil")
	}
	if n.V != nil {
		t.Errorf("expected nil slice, got %+v", n.V)
	}
}

func TestNullable_ToPtr_Valid(t *testing.T) {
	n := NullableFrom(testProfile{Name: "Alice", Email: "alice@example.com"})

	ptr := n.ToPtr()

	if ptr == nil {
		t.Fatal("expected non-nil pointer")
	}
	if ptr.Name != "Alice" {
		t.Errorf("expected Name=Alice, got %s", ptr.Name)
	}
	if ptr.Email != "alice@example.com" {
		t.Errorf("expected Email=alice@example.com, got %s", ptr.Email)
	}
}

func TestNullable_ToPtr_Invalid(t *testing.T) {
	n := Null[testProfile]()

	ptr := n.ToPtr()

	if ptr != nil {
		t.Errorf("expected nil pointer, got %+v", ptr)
	}
}

func TestNullableFromPtr_NonNil(t *testing.T) {
	profile := testProfile{Name: "Bob", Email: "bob@example.com"}
	n := NullableFromPtr(&profile)

	if !n.Valid {
		t.Error("expected Valid=true")
	}
	if n.V.Name != "Bob" {
		t.Errorf("expected Name=Bob, got %s", n.V.Name)
	}
	if n.V.Email != "bob@example.com" {
		t.Errorf("expected Email=bob@example.com, got %s", n.V.Email)
	}
}

func TestNullableFromPtr_Nil(t *testing.T) {
	n := NullableFromPtr[testProfile](nil)

	if n.Valid {
		t.Error("expected Valid=false")
	}
	if n.V.Name != "" || n.V.Email != "" {
		t.Errorf("expected zero value, got %+v", n.V)
	}
}

func TestNullable_ToPtr_Roundtrip(t *testing.T) {
	original := testProfile{Name: "Charlie", Email: "charlie@example.com"}
	n := NullableFromPtr(&original)
	ptr := n.ToPtr()

	if ptr == nil {
		t.Fatal("expected non-nil pointer")
	}
	if *ptr != original {
		t.Errorf("roundtrip failed: expected %+v, got %+v", original, *ptr)
	}
}

func TestNullable_ToPtr_Roundtrip_Nil(t *testing.T) {
	n := NullableFromPtr[testProfile](nil)
	ptr := n.ToPtr()

	if ptr != nil {
		t.Errorf("expected nil pointer, got %+v", ptr)
	}
}

func TestNullable_Get_Valid(t *testing.T) {
	n := NullableFrom(testProfile{Name: "Alice", Email: "alice@example.com"})

	v, ok := n.Get()

	if !ok {
		t.Error("expected ok=true")
	}
	if v.Name != "Alice" {
		t.Errorf("expected Name=Alice, got %s", v.Name)
	}
	if v.Email != "alice@example.com" {
		t.Errorf("expected Email=alice@example.com, got %s", v.Email)
	}
}

func TestNullable_Get_Invalid(t *testing.T) {
	n := Null[testProfile]()

	v, ok := n.Get()

	if ok {
		t.Error("expected ok=false")
	}
	if v.Name != "" || v.Email != "" {
		t.Errorf("expected zero value, got %+v", v)
	}
}
