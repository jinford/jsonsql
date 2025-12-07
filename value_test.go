package jsonsql

import (
	"encoding/json"
	"errors"
	"testing"
)

type testProfile struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func TestValue_Scan_Struct(t *testing.T) {
	input := []byte(`{"name":"Alice","email":"alice@example.com"}`)
	var v Value[testProfile]

	if err := v.Scan(input); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if v.V.Name != "Alice" {
		t.Errorf("expected Name=Alice, got %s", v.V.Name)
	}
	if v.V.Email != "alice@example.com" {
		t.Errorf("expected Email=alice@example.com, got %s", v.V.Email)
	}
}

func TestValue_Scan_Map(t *testing.T) {
	input := []byte(`{"key":"value","count":42}`)
	var v Value[map[string]any]

	if err := v.Scan(input); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if v.V["key"] != "value" {
		t.Errorf("expected key=value, got %v", v.V["key"])
	}
	if v.V["count"] != float64(42) {
		t.Errorf("expected count=42, got %v", v.V["count"])
	}
}

func TestValue_Scan_String(t *testing.T) {
	input := `"hello world"`
	var v Value[string]

	if err := v.Scan(input); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if v.V != "hello world" {
		t.Errorf("expected 'hello world', got %s", v.V)
	}
}

func TestValue_Scan_RawMessage(t *testing.T) {
	input := json.RawMessage(`{"test":true}`)
	var v Value[map[string]bool]

	if err := v.Scan(input); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if !v.V["test"] {
		t.Errorf("expected test=true, got %v", v.V["test"])
	}
}

func TestValue_Scan_Nil_ReturnsError(t *testing.T) {
	var v Value[testProfile]

	err := v.Scan(nil)
	if err == nil {
		t.Fatal("expected error for nil input")
	}
	if !errors.Is(err, ErrNullNotAllowed) {
		t.Errorf("expected ErrNullNotAllowed, got %v", err)
	}
}

func TestValue_Scan_JSONNull_ReturnsError(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{"null bytes", []byte("null")},
		{"null string", "null"},
		{"null with leading space", []byte(" null")},
		{"null with trailing space", []byte("null ")},
		{"null with both spaces", []byte(" null ")},
		{"null with newline", []byte("\nnull\n")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v Value[testProfile]

			err := v.Scan(tt.input)
			if err == nil {
				t.Fatalf("expected error for %q", tt.name)
			}
			if !errors.Is(err, ErrNullNotAllowed) {
				t.Errorf("expected ErrNullNotAllowed, got %v", err)
			}
		})
	}
}

func TestValue_Scan_UnsupportedType(t *testing.T) {
	var v Value[testProfile]

	err := v.Scan(123)
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

func TestValue_Scan_InvalidJSON(t *testing.T) {
	var v Value[testProfile]

	err := v.Scan([]byte(`{invalid json}`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestValue_Value_Struct(t *testing.T) {
	v := Value[testProfile]{
		V: testProfile{Name: "Bob", Email: "bob@example.com"},
	}

	result, err := v.Value()
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

func TestValue_Value_Map(t *testing.T) {
	v := Value[map[string]int]{
		V: map[string]int{"a": 1, "b": 2},
	}

	result, err := v.Value()
	if err != nil {
		t.Fatalf("Value failed: %v", err)
	}

	data, ok := result.([]byte)
	if !ok {
		t.Fatalf("expected []byte, got %T", result)
	}

	var parsed map[string]int
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if parsed["a"] != 1 || parsed["b"] != 2 {
		t.Errorf("unexpected result: %+v", parsed)
	}
}

func TestValue_Roundtrip(t *testing.T) {
	original := Value[testProfile]{
		V: testProfile{Name: "Charlie", Email: "charlie@example.com"},
	}

	data, err := original.Value()
	if err != nil {
		t.Fatalf("Value failed: %v", err)
	}

	var restored Value[testProfile]
	if err := restored.Scan(data); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if restored.V != original.V {
		t.Errorf("roundtrip failed: expected %+v, got %+v", original.V, restored.V)
	}
}

// unmarshalableType contains a channel which cannot be marshaled to JSON.
type unmarshalableType struct {
	Ch chan int
}

func TestValue_Value_MarshalError(t *testing.T) {
	v := Value[unmarshalableType]{
		V: unmarshalableType{Ch: make(chan int)},
	}

	_, err := v.Value()
	if err == nil {
		t.Fatal("expected error for unmarshalable type")
	}
}

func TestValue_Scan_Slice(t *testing.T) {
	input := []byte(`[1, 2, 3, 4, 5]`)
	var v Value[[]int]

	if err := v.Scan(input); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	expected := []int{1, 2, 3, 4, 5}
	if len(v.V) != len(expected) {
		t.Fatalf("expected length %d, got %d", len(expected), len(v.V))
	}
	for i, val := range expected {
		if v.V[i] != val {
			t.Errorf("expected v.V[%d]=%d, got %d", i, val, v.V[i])
		}
	}
}

func TestValue_Scan_SliceOfStructs(t *testing.T) {
	input := []byte(`[{"name":"Alice","email":"alice@example.com"},{"name":"Bob","email":"bob@example.com"}]`)
	var v Value[[]testProfile]

	if err := v.Scan(input); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(v.V) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(v.V))
	}
	if v.V[0].Name != "Alice" {
		t.Errorf("expected v.V[0].Name=Alice, got %s", v.V[0].Name)
	}
	if v.V[1].Name != "Bob" {
		t.Errorf("expected v.V[1].Name=Bob, got %s", v.V[1].Name)
	}
}

func TestValue_Scan_EmptySlice(t *testing.T) {
	input := []byte(`[]`)
	var v Value[[]int]

	if err := v.Scan(input); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if v.V == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(v.V) != 0 {
		t.Errorf("expected length 0, got %d", len(v.V))
	}
}

func TestValue_Value_Slice(t *testing.T) {
	v := Value[[]string]{
		V: []string{"a", "b", "c"},
	}

	result, err := v.Value()
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

func TestValue_Roundtrip_Slice(t *testing.T) {
	original := Value[[]testProfile]{
		V: []testProfile{
			{Name: "Alice", Email: "alice@example.com"},
			{Name: "Bob", Email: "bob@example.com"},
		},
	}

	data, err := original.Value()
	if err != nil {
		t.Fatalf("Value failed: %v", err)
	}

	var restored Value[[]testProfile]
	if err := restored.Scan(data); err != nil {
		t.Fatalf("Scan failed: %v", err)
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

func TestValue_Get(t *testing.T) {
	v := NewValue(testProfile{Name: "Alice", Email: "alice@example.com"})

	got := v.Get()

	if got.Name != "Alice" {
		t.Errorf("expected Name=Alice, got %s", got.Name)
	}
	if got.Email != "alice@example.com" {
		t.Errorf("expected Email=alice@example.com, got %s", got.Email)
	}
}
