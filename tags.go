package qtag

import "strings"

type TagOptions struct {
	Key     string  // "key" - Note: This will take the first key field in the tags array
	Default *string // "default=..."
	Ignore  bool    // "-"
}

// parseTags parses the "qt" struct tag string into a TagOptions struct.
// It extracts the primary query key, any default values, and ignore directives.
//
// Behavior:
//   - Empty tag (""): Returns an empty Key with Ignore set to false. This signals
//     the decoder to flatten the field if it's a struct, or skip it if it's a base type.
//   - Ignore tag ("-"): Returns Ignore set to true. The decoder will skip this field.
//   - Standard tag ("name,default=10"): Extracts "name" as the Key and "10" as the Default.
//   - Options without a key (",default=10"): Leaves Key empty but still parses the options.
func parseTags(tags string) TagOptions {
	opts := TagOptions{
		Key:     "",
		Default: nil,
		Ignore:  false,
	}

	// No tags or qt is empty
	if tags == "" {
		return opts
	}

	parts := strings.Split(tags, ",")
	if len(parts) == 0 {
		opts.Ignore = true
		return opts
	}

	for _, p := range parts {
		if kv := strings.Split(p, "="); len(kv) == 2 && kv[0] == "default" {
			opts.Default = &kv[1]
		} else if p == "-" {
			opts.Ignore = true
		} else if opts.Key == "" && p != "" {
			opts.Key = p
		}
	}

	// If ignored, reset state
	if opts.Ignore {
		opts.Key = ""
		opts.Default = nil
	}

	return opts
}
