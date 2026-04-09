//go:build integration

package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/mr-pixel-kg/shopshredder/api/internal/config"
	"github.com/mr-pixel-kg/shopshredder/api/internal/http/dto"
	"github.com/mr-pixel-kg/shopshredder/api/internal/http/handlers"
	authmw "github.com/mr-pixel-kg/shopshredder/api/internal/http/middleware"
	"github.com/mr-pixel-kg/shopshredder/api/internal/repositories"
	"github.com/mr-pixel-kg/shopshredder/api/internal/services"
	"github.com/mr-pixel-kg/shopshredder/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAuthFlow_RegisterLoginMeAndLogout(t *testing.T) {
	db := testutil.OpenIntegrationDB(t)
	testutil.ResetIntegrationDB(t, db)

	s, authService := newIntegrationServer(db)
	authHandler := handlers.AuthHandler{Auth: authService, Audit: newTestAuditService(db)}

	public := fuego.Group(s, "/api")
	fuego.Post(public, "/auth/register", authHandler.Register, option.DefaultStatusCode(http.StatusCreated))
	fuego.Post(public, "/auth/login", authHandler.Login)

	authed := fuego.Group(s, "/api",
		option.Middleware(authmw.Auth(authService)),
	)
	fuego.Post(authed, "/auth/logout", authHandler.Logout, option.DefaultStatusCode(http.StatusNoContent))
	fuego.Get(authed, "/auth/me", authHandler.Me)

	email := "api-auth-flow@example.com"
	password := "Sup3rS3cret!"

	registerBody := dto.RegisterRequest{Email: email, Password: password}
	registerRec := performJSONRequest(t, s, http.MethodPost, "/api/auth/register", registerBody, "")
	require.Equal(t, http.StatusCreated, registerRec.Code, registerRec.Body.String())

	loginBody := dto.LoginRequest{Email: email, Password: password}
	loginRec := performJSONRequest(t, s, http.MethodPost, "/api/auth/login", loginBody, "")
	require.Equal(t, http.StatusOK, loginRec.Code, loginRec.Body.String())

	var loginResp dto.LoginResponse
	require.NoError(t, json.Unmarshal(loginRec.Body.Bytes(), &loginResp))
	require.NotEmpty(t, loginResp.Token)
	assert.Equal(t, email, loginResp.User.Email)

	meRec := performJSONRequest(t, s, http.MethodGet, "/api/auth/me", nil, "Bearer "+loginResp.Token)
	require.Equal(t, http.StatusOK, meRec.Code, meRec.Body.String())

	logoutRec := performJSONRequest(t, s, http.MethodPost, "/api/auth/logout", nil, "Bearer "+loginResp.Token)
	require.Equal(t, http.StatusNoContent, logoutRec.Code, logoutRec.Body.String())

	meAfterLogoutRec := performJSONRequest(t, s, http.MethodGet, "/api/auth/me", nil, "Bearer "+loginResp.Token)
	assert.Equal(t, http.StatusOK, meAfterLogoutRec.Code, meAfterLogoutRec.Body.String())
}

func TestProtectedRouteRejectsMissingAuthorizationHeader(t *testing.T) {
	db := testutil.OpenIntegrationDB(t)
	testutil.ResetIntegrationDB(t, db)

	s, authService := newIntegrationServer(db)
	authHandler := handlers.AuthHandler{Auth: authService, Audit: newTestAuditService(db)}

	authed := fuego.Group(s, "/api",
		option.Middleware(authmw.Auth(authService)),
	)
	fuego.Get(authed, "/auth/me", authHandler.Me)

	rec := performJSONRequest(t, s, http.MethodGet, "/api/auth/me", nil, "")
	assert.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
}

func newTestAuthServices(db *gorm.DB) (*services.AuthService, *services.AuditService) {
	userRepo := repositories.NewUserRepository(db)
	auditRepo := repositories.NewAuditLogRepository(db)

	passwordService := services.NewPasswordService()
	tokenService := services.NewTokenService(config.AuthConfig{
		JWTSecret:     "integration-test-secret",
		JWTTTLMinutes: 60,
	})
	auditService := services.NewAuditService(auditRepo)
	authService := services.NewAuthService(userRepo, passwordService, tokenService, config.RegistrationConfig{
		Mode: config.RegistrationModePublic,
	})

	return authService, auditService
}

func newTestAuditService(db *gorm.DB) *services.AuditService {
	return services.NewAuditService(repositories.NewAuditLogRepository(db))
}

func newIntegrationServer(db *gorm.DB) (*fuego.Server, *services.AuthService) {
	authService, _ := newTestAuthServices(db)
	s := fuego.NewServer(
		fuego.WithAddr("localhost:0"),
		fuego.WithoutAutoGroupTags(),
		fuego.WithoutStartupMessages(),
	)
	return s, authService
}

func performJSONRequest(t *testing.T, s *fuego.Server, method, target string, body any, authorization string) *httptest.ResponseRecorder {
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
	req.Header.Set("Content-Type", "application/json")
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}

	rec := httptest.NewRecorder()
	s.Mux.ServeHTTP(rec, req)
	return rec
}
