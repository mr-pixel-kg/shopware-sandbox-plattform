package middleware

import (
	"log/slog"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/manuel/shopware-testenv-platform/api/internal/apperror"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/responses"
	"github.com/manuel/shopware-testenv-platform/api/internal/logging"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
	"github.com/manuel/shopware-testenv-platform/api/internal/types"
)

const authContextKey = "auth"

func Auth(authService *services.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Employee routes rely on a bearer token, while guest routes use a cookie
			// handled by the dedicated guest middleware.
			authHeader := strings.TrimSpace(c.Request().Header.Get(echo.HeaderAuthorization))
			if authHeader == "" {
				slog.Warn("missing authorization header", logging.RequestFields(c, "component", "auth")...)
				return responses.FromAppError(c, apperror.Unauthorized("Missing bearer token"))
			}

			token, ok := parseAuthorizationHeader(authHeader)
			if !ok {
				slog.Warn("invalid authorization header format", logging.RequestFields(c, "component", "auth")...)
				return responses.FromAppError(c, apperror.Unauthorized("Invalid authorization header"))
			}

			user, tokenID, err := authService.Authenticate(token)
			if err != nil {
				slog.Warn("token authentication failed", append(logging.RequestFields(c, "component", "auth"), "error", err.Error())...)
				return responses.FromAppError(c, apperror.Unauthorized("Invalid or expired token"))
			}

			// Store the authenticated user on the request context so handlers do not
			// need to parse or validate the token again.
			c.Set(authContextKey, types.AuthContext{UserID: user.ID, TokenID: tokenID, SessionType: "user"})
			c.Set("user", user)
			slog.Debug("request authenticated", logging.RequestFields(c, "component", "auth", "user_id", user.ID.String(), "token_id", tokenID)...)
			return next(c)
		}
	}
}

func MustAuth(c echo.Context) types.AuthContext {
	return c.Get(authContextKey).(types.AuthContext)
}

func parseAuthorizationHeader(authHeader string) (string, bool) {
	parts := strings.Fields(authHeader)
	switch len(parts) {
	case 1:
		// Swagger UI users often paste only the JWT value into the auth dialog.
		if parts[0] == "" || strings.EqualFold(parts[0], "Bearer") {
			return "", false
		}
		return parts[0], true
	case 2:
		if !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			return "", false
		}
		return parts[1], true
	default:
		return "", false
	}
}
