package qtag

import (
	"net/url"
	"testing"
	"time"
)

type BasicParams struct {
	Limit   int64   `qt:"limit"`
	Page    int     `qt:"page"`
	Active  bool    `qt:"active"`
	Score   float64 `qt:"score"`
	Ignored string  `qt:"-"`
}

type DefaultParams struct {
	Limit int    `qt:"limit,default=50"`
	Sort  string `qt:"sort,default=desc"`
}

type NestedParams struct {
	User struct {
		Name string `qt:"name"`
		Age  int    `qt:"age"`
	} `qt:"user"`
}

type FlatParams struct {
	Pagination struct {
		Limit int `qt:"limit"`
		Page  int `qt:"page"`
	} // No qt tag, should flatten
	Search string `qt:"search"`
}

type CustomTypeParams struct {
	Timestamp time.Time `qt:"timestamp"` // Implements TextUnmarshaler
}

type EdgeCaseParams struct {
	unexported string `qt:"unexported"` // Should be safely ignored
	Public     string `qt:"public"`
}

// --- The Tests ---

func TestUnmarshal(t *testing.T) {

	t.Run("Basic Types and Ignored Fields", func(t *testing.T) {
		var p BasicParams
		values := url.Values{
			"limit":   {"200"},
			"page":    {"3"},
			"active":  {"true"},
			"score":   {"99.5"},
			"Ignored": {"should_not_parse"},
		}

		err := Unmarshal(values, &p)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if p.Limit != 200 || p.Page != 3 || p.Active != true || p.Score != 99.5 {
			t.Errorf("Basic parsing failed. Got: %+v", p)
		}
		if p.Ignored != "" {
			t.Errorf("Ignored field was parsed: %s", p.Ignored)
		}
	})

	t.Run("Default Values", func(t *testing.T) {
		var p DefaultParams
		values := url.Values{}

		err := Unmarshal(values, &p)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if p.Limit != 50 {
			t.Errorf("Expected default limit 50, got %d", p.Limit)
		}
		if p.Sort != "desc" {
			t.Errorf("Expected default sort 'desc', got %s", p.Sort)
		}
	})

	t.Run("Nested Structs (Dot Notation)", func(t *testing.T) {
		var p NestedParams
		values := url.Values{
			"user.name": {"Alice"},
			"user.age":  {"30"},
		}

		err := Unmarshal(values, &p)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if p.User.Name != "Alice" || p.User.Age != 30 {
			t.Errorf("Nested parsing failed. Got: %+v", p)
		}
	})

	t.Run("Flat / Embedded Structs (Untagged)", func(t *testing.T) {
		var p FlatParams
		values := url.Values{
			"limit":  {"10"}, // Notice there is no "pagination." prefix
			"page":   {"2"},
			"search": {"golang"},
		}

		err := Unmarshal(values, &p)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if p.Pagination.Limit != 10 || p.Pagination.Page != 2 || p.Search != "golang" {
			t.Errorf("Flat struct parsing failed. Got: %+v", p)
		}
	})

	t.Run("TextUnmarshaler (time.Time)", func(t *testing.T) {
		var p CustomTypeParams
		values := url.Values{
			"timestamp": {"2024-01-01T15:04:05Z"}, // RFC3339 format, so should parse.
		}

		err := Unmarshal(values, &p)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expectedTime, _ := time.Parse(time.RFC3339, "2024-01-01T15:04:05Z")
		if !p.Timestamp.Equal(expectedTime) {
			t.Errorf("TextUnmarshaler failed. Expected %v, got %v", expectedTime, p.Timestamp)
		}
	})

	t.Run("Unexported Fields Ignored", func(t *testing.T) {
		var p EdgeCaseParams
		values := url.Values{
			"unexported": {"secret_value"},
			"public":     {"hello"},
		}

		err := Unmarshal(values, &p)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if p.unexported != "" {
			t.Errorf("Unexported field was illegally modified!")
		}
		if p.Public != "hello" {
			t.Errorf("Public field failed to parse. Got: %s", p.Public)
		}
	})

	// --- Error Handling Tests ---

	t.Run("Error: Type Conversion Failure", func(t *testing.T) {
		var p BasicParams
		values := url.Values{
			"limit": {"not_a_number"}, // Should fail to parse into int64
		}

		err := Unmarshal(values, &p)
		if err == nil {
			t.Fatal("Expected an error for invalid integer conversion, but got nil")
		}
	})

	t.Run("Error: Nil Pointer Input", func(t *testing.T) {
		values := url.Values{}
		var p *BasicParams = nil

		err := Unmarshal(values, p)
		if err == nil {
			t.Fatal("Expected an error for nil pointer, but got nil")
		}
	})

	t.Run("Error: Non-Pointer Input", func(t *testing.T) {
		values := url.Values{}

		// Passing by value, not by pointer
		// NOTE: In Go, if you define the parameter as `v *T`, the compiler actually
		// stops you from passing a value. This test is conceptually what happens if
		// someone bypassed the generic signature. Our internal `elm.Kind() != reflect.Struct`
		// catch handles the rest.

		// To properly test the reflection safety net, we test passing a pointer to a non-struct:
		nonStruct := 5
		err := Unmarshal(values, &nonStruct)
		if err == nil {
			t.Fatal("Expected an error for non-struct pointer, but got nil")
		}
	})
}
