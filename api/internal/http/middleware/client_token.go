package middleware

import (
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const clientContextKey = "client"

const (
	clientTokenHeaderName = "X-Client-Token"
	clientTokenCookieName = "client_token"
)

func EnsureClientToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(clientContextKey, extractClientToken(c))
			return next(c)
		}
	}
}

func ClientToken(c echo.Context) *uuid.UUID {
	clientToken, ok := c.Get(clientContextKey).(*uuid.UUID)
	if !ok {
		return nil
	}
	return clientToken
}

func extractClientToken(c echo.Context) *uuid.UUID {
	if tokenID := parseClientToken(c.Request().Header.Get(clientTokenHeaderName)); tokenID != nil {
		return tokenID
	}

	if cookie, err := c.Cookie(clientTokenCookieName); err == nil && cookie != nil {
		return parseClientToken(cookie.Value)
	}

	// Future option:
	// If we later standardize on cookie issuance by the API, the SetCookie logic
	// can live here. For now we only consume an externally managed client token.
	return nil
}

func parseClientToken(raw string) *uuid.UUID {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	tokenID, err := uuid.Parse(raw)
	if err != nil {
		return nil
	}
	return &tokenID
}
