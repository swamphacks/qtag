package qtag

import (
	"encoding"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
)

// Decode is a convenience wrapper around [Unmarshal] for standard HTTP handlers.
// It extracts the query parameters directly from the *http.Request URL and parses
// them into the provided struct pointer.
//
// Example:
//
//	func SearchHandler(w http.ResponseWriter, r *http.Request) {
//		var filter SearchFilter
//		if err := qtag.Decode(r, &filter); err != nil {
//			http.Error(w, err.Error(), http.StatusBadRequest)
//			return
//		}
//		// ... use filter ...
//	}
func Decode[T any](r *http.Request, v *T) error {
	return Unmarshal(r.URL.Query(), v)
}

// Unmarshal parses URL query parameters into a struct using reflection.
// It requires a valid pointer to a struct.
//
// # Behavior & Tag Routing
//
//   - Basic Mapping: Define a `qt:"name"` tag to map a query parameter to a struct field.
//   - Default Values: Add `,default=value` to provide a fallback if the parameter is
//     missing from the URL. Example: `qt:"limit,default=20"`.
//   - Ignored Fields: Use `qt:"-"` to explicitly tell the parser to skip a field entirely.
//   - Tagged Nested Structs: If a struct field has a `qt` tag, it acts as a namespace
//     using dot-notation. Example: `qt:"user"` creates keys like `?user.name=Alice`.
//   - Untagged Nested Structs: If an embedded or nested struct has NO `qt` tag
//     (or `qt:""`), its fields are "flattened" and inherit the parent's namespace.
//
// # Supported Types
//
//   - Standard Types: string, bool, int/8/16/32/64, float32/64.
//   - Custom Types: Any type that implements [encoding.TextUnmarshaler] is automatically
//     supported. This includes standard library types like time.Time (which expects RFC 3339).
//
// # Common Pitfalls
//
//   - Passing by Value: You MUST pass a pointer to the struct (e.g., &myStruct),
//     otherwise the function will return an error.
//   - Unexported Fields: Fields starting with a lowercase letter cannot be set via
//     reflection and are silently skipped.
//   - Struct Initialization: If you pre-populate your struct with values before calling
//     [Unmarshal], and a query parameter is missing, qtag will leave your pre-populated
//     value alone unless you defined a default tag, which will overwrite it.
//
// Example:
//
//	type Pagination struct {
//		Limit int `qt:"limit,default=20"`
//		Page  int `qt:"page,default=1"`
//	}
//
//	type SearchQuery struct {
//		Pagination               // Untagged: inherits keys flatly (?limit=50)
//		User       UserFilter    `qt:"user"` // Tagged: uses dot-notation (?user.name=Alice)
//		InternalID string        `qt:"-"`
//	}
//
//	values, _ := url.ParseQuery("limit=50&user.name=Alice")
//	var q SearchQuery
//	err := qtag.Unmarshal(values, &q)
func Unmarshal[T any](data url.Values, v *T) error {
	if v == nil {
		return errors.New("qtag: cannot unmarshal into nil pointer")
	}

	elm := reflect.ValueOf(v).Elem()

	if elm.Kind() != reflect.Struct {
		return errors.New("qtag: cannot unmarshal into a non-struct")
	}

	// Kick off the recursive decoder with an empty root prefix
	return decodeStruct(data, elm, "")
}

// decodeStruct is the internal recursive engine that handles the reflection tree.
// It builds dot-notation keys based on the prefix and delegates type conversion.
func decodeStruct(values url.Values, v reflect.Value, prefix string) error {
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("qtag: expected struct, got %s", v.Kind())
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fieldType := t.Field(i)
		fieldVal := v.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		rawTag := fieldType.Tag.Get("qt")
		parsedTags := parseTags(rawTag)

		if parsedTags.Ignore {
			continue
		}

		queryKey := parsedTags.Key
		if prefix != "" && queryKey != "" {
			queryKey = prefix + "." + queryKey
		} else if queryKey == "" {
			queryKey = prefix
		}

		strVal := values.Get(queryKey)
		if strVal == "" && parsedTags.Default != nil {
			strVal = *parsedTags.Default
		}

		if strVal != "" && fieldVal.CanAddr() {
			if u, ok := fieldVal.Addr().Interface().(encoding.TextUnmarshaler); ok {
				err := u.UnmarshalText([]byte(strVal))
				if err != nil {
					return fmt.Errorf("qtag: failed to unmarshal %s: %w", queryKey, err)
				}
				continue
			}
		}

		if fieldVal.Kind() == reflect.Struct {
			err := decodeStruct(values, fieldVal, queryKey)
			if err != nil {
				return err
			}
			continue
		}

		if strVal == "" || parsedTags.Key == "" {
			continue
		}

		switch fieldVal.Kind() {
		case reflect.String:
			fieldVal.SetString(strVal)

		case reflect.Bool:
			parsedBool, err := strconv.ParseBool(strVal)
			if err != nil {
				return fmt.Errorf("qtag: cannot parse %s='%s' into a boolean", queryKey, strVal)
			}
			fieldVal.SetBool(parsedBool)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			parsedInt, err := strconv.ParseInt(strVal, 10, 64)
			if err != nil {
				return fmt.Errorf("qtag: cannot parse %s='%s' into an integer", queryKey, strVal)
			}
			fieldVal.SetInt(parsedInt)

		case reflect.Float32, reflect.Float64:
			bitSize := 64
			if fieldVal.Kind() == reflect.Float32 {
				bitSize = 32
			}
			parsedFloat, err := strconv.ParseFloat(strVal, bitSize)
			if err != nil {
				return fmt.Errorf("qtag: cannot parse %s='%s' into a float", queryKey, strVal)
			}
			fieldVal.SetFloat(parsedFloat)

		default:
			return fmt.Errorf("qtag: unexpected type %s in tag %s", fieldVal.Kind(), queryKey)
		}
	}

	return nil
}
