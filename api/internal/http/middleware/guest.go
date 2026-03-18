package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
	"github.com/manuel/shopware-testenv-platform/api/internal/types"
)

const guestContextKey = "guest"

func EnsureGuestSession(guestService *services.GuestSessionService, cookieName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Public routes auto-provision a guest identity so anonymous users can
			// still list "their" sandbox sessions later on.
			cookie, _ := c.Cookie(cookieName)
			tokenValue := ""
			if cookie != nil {
				tokenValue = cookie.Value
			}

			token, sessionID, err := guestService.Ensure(tokenValue)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "could not create guest session")
			}

			if token != tokenValue {
				// The cookie is http-only because the frontend only needs the backend
				// to correlate the guest session, not direct cookie access from JS.
				c.SetCookie(&http.Cookie{
					Name:     cookieName,
					Value:    token,
					Path:     "/",
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
				})
			}

			c.Set(guestContextKey, types.GuestContext{SessionID: sessionID})
			return next(c)
		}
	}
}

func MustGuest(c echo.Context) types.GuestContext {
	return c.Get(guestContextKey).(types.GuestContext)
}
