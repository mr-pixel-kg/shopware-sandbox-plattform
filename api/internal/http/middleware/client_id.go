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
			if id == nil {
				generated := uuid.New()
				id = &generated
			}

			if _, err := r.Cookie(clientIDCookieName); err != nil {
				http.SetCookie(w, &http.Cookie{
					Name:     clientIDCookieName,
					Value:    id.String(),
					Path:     "/",
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
				})
			}

			ctx := withClientID(r.Context(), id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractClientID(r *http.Request) *uuid.UUID {
	if cookie, err := r.Cookie(clientIDCookieName); err == nil && cookie != nil {
		if id := parseClientID(cookie.Value); id != nil {
			return id
		}
	}

	return parseClientID(r.Header.Get(clientIDHeaderName))
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
