package middleware

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const (
	clientIDHeaderName = "X-Client-Id"
	clientIDCookieName = "client_id"
)

func EnsureClientID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := extractClientID(r)
			ctx := withClientID(r.Context(), id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractClientID(r *http.Request) *uuid.UUID {
	if id := parseClientID(r.Header.Get(clientIDHeaderName)); id != nil {
		return id
	}

	if cookie, err := r.Cookie(clientIDCookieName); err == nil && cookie != nil {
		return parseClientID(cookie.Value)
	}

	return nil
}

func parseClientID(raw string) *uuid.UUID {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	id, err := uuid.Parse(raw)
	if err != nil {
		return nil
	}
	return &id
}
