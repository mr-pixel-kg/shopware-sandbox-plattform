package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/manuel/shopware-testenv-platform/api/internal/apperror"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/responses"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
	"github.com/manuel/shopware-testenv-platform/api/internal/types"
)

const authContextKey = "auth"

func Auth(authService *services.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Employee routes rely on a bearer token, while guest routes use a cookie
			// handled by the dedicated guest middleware.
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return responses.FromAppError(c, apperror.Unauthorized("Missing bearer token"))
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				return responses.FromAppError(c, apperror.Unauthorized("Invalid authorization header"))
			}

			user, tokenID, err := authService.Authenticate(token)
			if err != nil {
				return responses.FromAppError(c, apperror.Unauthorized("Invalid or expired token"))
			}

			// Store the authenticated user on the request context so handlers do not
			// need to parse or validate the token again.
			c.Set(authContextKey, types.AuthContext{UserID: user.ID, TokenID: tokenID, SessionType: "user"})
			c.Set("user", user)
			return next(c)
		}
	}
}

func MustAuth(c echo.Context) types.AuthContext {
	return c.Get(authContextKey).(types.AuthContext)
}
