package qtag

type TagOptions struct {
	Key     string  // "key" - Note: This will take the first key field in the tags array
	Default *string // "default=..."
	Ignore  bool    // "-"
}

func parseTags(tags string) TagOptions {
	opts := TagOptions{
		Key:     "",
		Default: nil,
		Ignore:  false,
	}

	parts := split(tags, ',')
	if len(parts) == 0 {
		opts.Ignore = true
		return opts
	}

	for _, p := range parts {
		if kv := split(p, '='); len(kv) == 2 && kv[0] == "default" {
			opts.Default = &kv[1]
		} else if p == "-" {
			opts.Ignore = true
		} else if opts.Key == "" && p != "" {
			opts.Key = p
		}
	}

	// Handle empty key values etc
	if opts.Key == "" {
		opts.Ignore = true
	}

	return opts
}

func split(s string, sep byte) []string {
	out := make([]string, 0)
	last := 0

	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			if last < i {
				out = append(out, s[last:i])
			}
			last = i + 1
		}
	}

	return out
}
