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

// Takes in url.Values and a pointer to the struct you want to destructure the query parameters into. It will only accept a valid struct pointer.
//
// Usage:
//
//	Add a field tag `qt:""` in the struct you pass in.
//	The tag will be the name of the query parameter that will be unmarshaled into that field.
//
//	For example:
//
//	Url: https://example.com/some/endpoint?limit=200&page=200
//
//	type QueryParams struct {
//	     Limit int64 `qt:"limit"`
//	     Page int64 `qt:"page"`
//	}
func Unmarshal[T any](data url.Values, v *T) error {
	if v == nil {
		return errors.New("Cannot unmarshal into null pointer.")
	}

	elm := reflect.ValueOf(v).Elem()
	elmType := elm.Type()

	if elm.Kind() != reflect.Struct {
		return errors.New("Cannot unmarshal into a non-struct.")
	}

	for i := range elm.NumField() {
		t, f := elmType.Field(i), elm.Field(i)

		if !f.CanSet() {
			continue
		}

		tags := t.Tag.Get("qt")
		if tags == "" || tags == "-" {
			continue
		}

		value := data.Get(tags)
		if value == "" {
			continue
		}

		// Check for TextUnmarshal
		if f.CanAddr() {
			if u, ok := f.Addr().Interface().(encoding.TextUnmarshaler); ok {
				err := u.UnmarshalText([]byte(value))
				if err == nil {
					continue
				}
			}
		}

		switch f.Kind() {
		case reflect.String:
			f.SetString(value)
		case reflect.Bool:
			b, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("Cannot parse %s with value %s into a boolean.", tags, value)
			}
			f.SetBool(b)
		case reflect.Int, reflect.Int64:
			int, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("Cannot parse %s with value %s into an integer.", tags, value)
			}
			f.SetInt(int)
		case reflect.Float64:
			float, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("Cannot parse %s with value %s into a float64.", tags, value)
			}
			f.SetFloat(float)
		case reflect.Float32:
			float, err := strconv.ParseFloat(value, 32)
			if err != nil {
				return fmt.Errorf("Cannot parse %s with value %s into a float32.", tags, value)
			}
			f.SetFloat(float)
		default:
			return fmt.Errorf("Unexpected type in tag %s", tags)
		}
	}

	return nil
}

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

		tags := fieldType.Tag.Get("qt")
		if tags == "-" {
			continue
		}
	}

	return nil
}

func Decode[T any](r *http.Request, v *T) error {
	return Unmarshal(r.URL.Query(), v)
}
