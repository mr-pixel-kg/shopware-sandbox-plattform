package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAuthorizationHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		header string
		want   string
		ok     bool
	}{
		{name: "bearer token", header: "Bearer abc.def.ghi", want: "abc.def.ghi", ok: true},
		{name: "lowercase bearer", header: "bearer abc.def.ghi", want: "abc.def.ghi", ok: true},
		{name: "bare token", header: "abc.def.ghi", want: "abc.def.ghi", ok: true},
		{name: "empty", header: "", want: "", ok: false},
		{name: "missing token after bearer", header: "Bearer", want: "", ok: false},
		{name: "wrong scheme", header: "Basic abc", want: "", ok: false},
		{name: "too many parts", header: "Bearer abc def", want: "", ok: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := parseAuthorizationHeader(tt.header)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}
