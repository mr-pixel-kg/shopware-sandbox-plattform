package registry

import (
	"fmt"
	"log/slog"
	"strings"
)

func ValuesFromJSONMap(m map[string]any) map[string]string {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		switch t := v.(type) {
		case nil:
			continue
		case string:
			out[k] = t
		case bool,
			float64, float32,
			int, int32, int64,
			uint, uint32, uint64:
			out[k] = fmt.Sprint(t)
		default:
			slog.Debug("metadata value dropped: unsupported type",
				"component", "registry", "key", k, "type", fmt.Sprintf("%T", v))
		}
	}
	return out
}

func ValuesToJSONMap(m map[string]string) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func HostnameFromURL(u string) string {
	if u == "" {
		return ""
	}
	s := u
	if i := strings.Index(s, "://"); i >= 0 {
		s = s[i+3:]
	}
	if i := strings.IndexAny(s, "/?#"); i >= 0 {
		s = s[:i]
	}
	return s
}
