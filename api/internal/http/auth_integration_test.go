//go:build integration

package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/manuel/shopware-testenv-platform/api/internal/config"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/handlers"
	authmw "github.com/manuel/shopware-testenv-platform/api/internal/http/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/repositories"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
	"github.com/manuel/shopware-testenv-platform/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAuthFlow_RegisterLoginMeAndLogout(t *testing.T) {
	db := testutil.OpenIntegrationDB(t)
	testutil.ResetIntegrationDB(t, db)

	router := echo.New()
	authService, auditService := newTestAuthServices(db)
	authHandler := handlers.NewAuthHandler(authService, auditService)

	router.POST("/api/auth/register", authHandler.Register)
	router.POST("/api/auth/login", authHandler.Login)

	private := router.Group("/api")
	private.Use(authmw.Auth(authService))
	private.GET("/me", authHandler.Me)
	private.POST("/auth/logout", authHandler.Logout)

	email := "api-auth-flow@example.com"
	password := "Sup3rS3cret!"

	registerBody := dto.RegisterRequest{Email: email, Password: password}
	registerRec := performJSONRequest(t, router, http.MethodPost, "/api/auth/register", registerBody, "")
	require.Equal(t, http.StatusCreated, registerRec.Code, registerRec.Body.String())

	loginBody := dto.LoginRequest{Email: email, Password: password}
	loginRec := performJSONRequest(t, router, http.MethodPost, "/api/auth/login", loginBody, "")
	require.Equal(t, http.StatusOK, loginRec.Code, loginRec.Body.String())

	var loginResp dto.AuthLoginResponse
	require.NoError(t, json.Unmarshal(loginRec.Body.Bytes(), &loginResp))
	require.NotEmpty(t, loginResp.Token)
	assert.Equal(t, email, loginResp.User.Email)

	meRec := performJSONRequest(t, router, http.MethodGet, "/api/me", nil, loginResp.Token)
	require.Equal(t, http.StatusOK, meRec.Code, meRec.Body.String())

	logoutRec := performJSONRequest(t, router, http.MethodPost, "/api/auth/logout", nil, "Bearer "+loginResp.Token)
	require.Equal(t, http.StatusNoContent, logoutRec.Code, logoutRec.Body.String())

	meAfterLogoutRec := performJSONRequest(t, router, http.MethodGet, "/api/me", nil, "Bearer "+loginResp.Token)
	assert.Equal(t, http.StatusUnauthorized, meAfterLogoutRec.Code, meAfterLogoutRec.Body.String())
}

func TestProtectedRouteRejectsMissingAuthorizationHeader(t *testing.T) {
	db := testutil.OpenIntegrationDB(t)
	testutil.ResetIntegrationDB(t, db)

	router := echo.New()
	authService, auditService := newTestAuthServices(db)
	authHandler := handlers.NewAuthHandler(authService, auditService)

	private := router.Group("/api")
	private.Use(authmw.Auth(authService))
	private.GET("/me", authHandler.Me)

	rec := performJSONRequest(t, router, http.MethodGet, "/api/me", nil, "")
	assert.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
}

func newTestAuthServices(db *gorm.DB) (*services.AuthService, *services.AuditService) {
	userRepo := repositories.NewUserRepository(db)
	sessionRepo := repositories.NewSessionRepository(db)
	auditRepo := repositories.NewAuditLogRepository(db)

	passwordService := services.NewPasswordService()
	tokenService := services.NewTokenService(config.AuthConfig{
		JWTSecret:          "integration-test-secret",
		JWTTTLMinutes:      60,
		GuestJWTTTLMinutes: 60,
		GuestCookieName:    "test_guest",
	})
	auditService := services.NewAuditService(auditRepo)
	authService := services.NewAuthService(userRepo, sessionRepo, passwordService, tokenService)

	return authService, auditService
}

func performJSONRequest(t *testing.T, e *echo.Echo, method, target string, body any, authorization string) *httptest.ResponseRecorder {
	t.Helper()

	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
	}

	req := httptest.NewRequest(method, target, bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if authorization != "" {
		req.Header.Set(echo.HeaderAuthorization, authorization)
	}

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}
