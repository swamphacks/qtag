package qtag

import (
	"net/url"
	"testing"
)

type Params1 struct {
	Limit int64 `qt:"limit"`
	Page  int   `qt:"page"`
}

func TestTesting(t *testing.T) {
	t.Run("Basic Struct", func(t *testing.T) {
		var params Params1
		values := url.Values{}
		values.Set("limit", "200")
		values.Set("page", "3")
		err := Unmarshal(values, &params)
		if err != nil {
			t.Error(err)
		}

		if params.Limit != 200 {
			t.Errorf("Expected 200 for `limit` but got %d", params.Limit)
		}

		if params.Page != 3 {
			t.Errorf("Expected 3 for `page` but got %d", params.Page)
		}
	})

	t.Run("Field unmarshal error", func(t *testing.T) {
		var params Params1
		values := url.Values{}
		values.Set("limit", "hello")
		values.Set("page", "3")
		err := Unmarshal(values, &params)
		if err == nil {
			t.Error(err)
		}

	})
}
