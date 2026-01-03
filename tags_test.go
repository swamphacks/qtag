package qtag

import "testing"

func TestTags(t *testing.T) {
	t.Run("Basic tag", func(t *testing.T) {
		out := parseTags("limit")
		if out.Key != "limit" {
			t.Errorf("Key was %s, not 'limit'", out.Key)
		}

		if out.Default != nil {
			t.Errorf("Default was not nil like expected.")
		}
	})

	t.Run("Basic default", func(t *testing.T) {
		out := parseTags("limit,default=10")
		if out.Key != "limit" {
			t.Errorf("Key was %s, not 'limit'", out.Key)
		}

		if *out.Default != "10" {
			t.Errorf("Default was %s, not 10 like expected.", *out.Default)
		}
	})
}
