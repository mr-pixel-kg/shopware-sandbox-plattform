package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/manuel/shopware-testenv-platform/api/internal/http/errs"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
	"github.com/manuel/shopware-testenv-platform/api/internal/types"
)

func Auth(authService *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			if authHeader == "" {
				slog.Warn("missing authorization header",
					"component", "auth",
					"method", r.Method,
					"path", r.URL.Path,
				)
				errs.Write(w, http.StatusUnauthorized, "Missing bearer token")
				return
			}

			token, ok := ParseAuthorizationHeader(authHeader)
			if !ok {
				slog.Warn("invalid authorization header format",
					"component", "auth",
					"method", r.Method,
					"path", r.URL.Path,
				)
				errs.Write(w, http.StatusUnauthorized, "Invalid authorization header")
				return
			}

			user, err := authService.Authenticate(token)
			if err != nil {
				slog.Warn("token authentication failed",
					"component", "auth",
					"method", r.Method,
					"path", r.URL.Path,
					"error", err.Error(),
				)
				errs.Write(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			ctx := withAuth(r.Context(), types.AuthContext{UserID: user.ID})
			ctx = withUser(ctx, user)
			slog.Debug("request authenticated",
				"component", "auth",
				"method", r.Method,
				"path", r.URL.Path,
				"user_id", user.ID.String(),
			)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAdmin() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := UserFromContext(r)
			if user == nil || !user.IsAdmin() {
				errs.Write(w, http.StatusForbidden, "Admin access required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func ParseAuthorizationHeader(authHeader string) (string, bool) {
	parts := strings.Fields(authHeader)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}
