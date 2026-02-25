package qtag

import (
	"testing"
)

func TestParseTags(t *testing.T) {
	// Helper function to safely compare string pointers
	checkDefault := func(expected, actual *string) bool {
		if expected == nil && actual == nil {
			return true
		}
		if expected != nil && actual != nil {
			return *expected == *actual
		}
		return false
	}

	tests := []struct {
		name     string
		tagInput string
		wantKey  string
		wantDef  *string
		wantIgn  bool
	}{
		{
			name:     "Basic tag",
			tagInput: "limit",
			wantKey:  "limit",
			wantDef:  nil,
			wantIgn:  false,
		},
		{
			name:     "Basic default",
			tagInput: "limit,default=10",
			wantKey:  "limit",
			wantDef:  func() *string { s := "10"; return &s }(),
			wantIgn:  false,
		},
		{
			name:     "Empty tag (Untagged struct)",
			tagInput: "",
			wantKey:  "",
			wantDef:  nil,
			wantIgn:  false,
		},
		{
			name:     "Explicit ignore",
			tagInput: "-",
			wantKey:  "",
			wantDef:  nil,
			wantIgn:  true,
		},
		{
			name:     "Ignore with other options (should still ignore)",
			tagInput: "-,default=10",
			wantKey:  "",
			wantDef:  nil,
			wantIgn:  true,
		},
		{
			name:     "Empty key with default (Weird, but valid)",
			tagInput: ",default=20",
			wantKey:  "",
			wantDef:  func() *string { s := "20"; return &s }(),
			wantIgn:  false,
		},
		{
			name:     "Empty default value",
			tagInput: "search,default=",
			wantKey:  "search",
			wantDef:  func() *string { s := ""; return &s }(),
			wantIgn:  false,
		},
		{
			name:     "Unknown options are ignored safely",
			tagInput: "limit,unknown=foo,default=50",
			wantKey:  "limit",
			wantDef:  func() *string { s := "50"; return &s }(),
			wantIgn:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTags(tt.tagInput)

			if got.Key != tt.wantKey {
				t.Errorf("parseTags(%q) Key = %q, want %q", tt.tagInput, got.Key, tt.wantKey)
			}

			if !checkDefault(tt.wantDef, got.Default) {
				gotDefStr := "<nil>"
				if got.Default != nil {
					gotDefStr = *got.Default
				}
				wantErrStr := "<nil>"
				if tt.wantDef != nil {
					wantErrStr = *tt.wantDef
				}
				t.Errorf("parseTags(%q) Default = %s, want %s", tt.tagInput, gotDefStr, wantErrStr)
			}

			if got.Ignore != tt.wantIgn {
				t.Errorf("parseTags(%q) Ignore = %v, want %v", tt.tagInput, got.Ignore, tt.wantIgn)
			}
		})
	}
}
