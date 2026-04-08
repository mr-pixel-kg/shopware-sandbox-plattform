//go:build integration

package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/google/uuid"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/handlers"
	authmw "github.com/manuel/shopware-testenv-platform/api/internal/http/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/manuel/shopware-testenv-platform/api/internal/repositories"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
	"github.com/manuel/shopware-testenv-platform/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func TestAuditLogsAdminCanListWithPaginationAndFilters(t *testing.T) {
	db := testutil.OpenIntegrationDB(t)
	testutil.ResetIntegrationDB(t, db)

	s, authService := newIntegrationServer(db)
	auditService := newTestAuditService(db)
	authHandler := handlers.AuthHandler{Auth: authService, Audit: auditService}
	auditHandler := handlers.AuditHandler{Audit: auditService}

	public := fuego.Group(s, "/api")
	authHandler.MountPublicRoutes(public)

	admin := fuego.Group(s, "/api",
		option.Middleware(authmw.Auth(authService)),
		option.Middleware(authmw.RequireAdmin()),
	)
	auditHandler.MountRoutes(admin)

	adminToken := createAdminToken(t, db, s)
	user := createAuditHTTPUser(t, db, "audit-http-user")
	clientID := uuid.New()
	resourceID := uuid.New()
	now := time.Now().UTC()
	resourceType := "sandbox"

	repo := repositories.NewAuditLogRepository(db)
	createAuditHTTPLog(t, repo, models.AuditLog{
		ID:           uuid.New(),
		UserID:       &user.ID,
		Action:       "sandbox.created",
		UserAgent:    strPtrHTTP("Mozilla/5.0"),
		ClientID:     &clientID,
		ResourceType: &resourceType,
		ResourceID:   &resourceID,
		Details:      datatypes.JSON([]byte(`{"step":"created"}`)),
		Timestamp:    now.Add(-2 * time.Minute),
	})
	expected := createAuditHTTPLog(t, repo, models.AuditLog{
		ID:           uuid.New(),
		UserID:       &user.ID,
		Action:       "sandbox.created",
		UserAgent:    strPtrHTTP("Shopware-Test/1.0"),
		ClientID:     &clientID,
		ResourceType: &resourceType,
		ResourceID:   &resourceID,
		Details:      datatypes.JSON([]byte(`{"step":"deleted"}`)),
		Timestamp:    now.Add(-1 * time.Minute),
	})
	createAuditHTTPLog(t, repo, models.AuditLog{
		ID:           uuid.New(),
		Action:       "image.created",
		ResourceType: strPtrHTTP("image"),
		ResourceID:   &resourceID,
		Details:      datatypes.JSON([]byte(`{}`)),
		Timestamp:    now,
	})

	target := fmt.Sprintf(
		"/api/audit-logs?limit=1&offset=0&action=sandbox.created&resourceType=sandbox&resourceId=%s&clientId=%s",
		resourceID.String(),
		clientID.String(),
	)
	rec := performJSONRequest(t, s, http.MethodGet, target, nil, "Bearer "+adminToken)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var response dto.AuditLogListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Len(t, response.Data, 1)
	assert.Equal(t, expected.ID, response.Data[0].ID)
	assert.Equal(t, "sandbox.created", response.Data[0].Action)
	assert.NotNil(t, response.Data[0].User)
	assert.Equal(t, user.Email, response.Data[0].User.Email)
	assert.Equal(t, 1, response.Meta.Pagination.Limit)
	assert.Equal(t, 0, response.Meta.Pagination.Offset)
	assert.Equal(t, 1, response.Meta.Pagination.Count)
	assert.Equal(t, int64(2), response.Meta.Pagination.Total)
	assert.True(t, response.Meta.Pagination.HasMore)
	require.NotNil(t, response.Meta.Filters.Action)
	assert.Equal(t, "sandbox.created", *response.Meta.Filters.Action)
	require.NotNil(t, response.Meta.Filters.ResourceType)
	assert.Equal(t, "sandbox", *response.Meta.Filters.ResourceType)
	require.NotNil(t, response.Meta.Filters.ResourceID)
	assert.Equal(t, resourceID, *response.Meta.Filters.ResourceID)
	require.NotNil(t, response.Meta.Filters.ClientID)
	assert.Equal(t, clientID, *response.Meta.Filters.ClientID)
}

func TestAuditLogsRequireAdmin(t *testing.T) {
	db := testutil.OpenIntegrationDB(t)
	testutil.ResetIntegrationDB(t, db)

	s, authService := newIntegrationServer(db)
	auditService := newTestAuditService(db)
	authHandler := handlers.AuthHandler{Auth: authService, Audit: auditService}
	auditHandler := handlers.AuditHandler{Audit: auditService}

	public := fuego.Group(s, "/api")
	authHandler.MountPublicRoutes(public)

	admin := fuego.Group(s, "/api",
		option.Middleware(authmw.Auth(authService)),
		option.Middleware(authmw.RequireAdmin()),
	)
	auditHandler.MountRoutes(admin)

	userToken := createUserToken(t, db, s)
	rec := performJSONRequest(t, s, http.MethodGet, "/api/audit-logs", nil, "Bearer "+userToken)

	require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
	assert.Contains(t, rec.Body.String(), "Admin access required")
}

func TestAuditLogsRejectInvalidQueryParameters(t *testing.T) {
	db := testutil.OpenIntegrationDB(t)
	testutil.ResetIntegrationDB(t, db)

	s, authService := newIntegrationServer(db)
	auditService := newTestAuditService(db)
	authHandler := handlers.AuthHandler{Auth: authService, Audit: auditService}
	auditHandler := handlers.AuditHandler{Audit: auditService}

	public := fuego.Group(s, "/api")
	authHandler.MountPublicRoutes(public)

	admin := fuego.Group(s, "/api",
		option.Middleware(authmw.Auth(authService)),
		option.Middleware(authmw.RequireAdmin()),
	)
	auditHandler.MountRoutes(admin)

	adminToken := createAdminToken(t, db, s)
	rec := performJSONRequest(t, s, http.MethodGet, "/api/audit-logs?from=invalid", nil, "Bearer "+adminToken)

	require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
	assert.Contains(t, rec.Body.String(), "Invalid from timestamp")
}

func TestAuditLogFacetsReturnStableUsersAndActions(t *testing.T) {
	db := testutil.OpenIntegrationDB(t)
	testutil.ResetIntegrationDB(t, db)

	s, authService := newIntegrationServer(db)
	auditService := newTestAuditService(db)
	authHandler := handlers.AuthHandler{Auth: authService, Audit: auditService}
	auditHandler := handlers.AuditHandler{Audit: auditService}

	public := fuego.Group(s, "/api")
	authHandler.MountPublicRoutes(public)

	admin := fuego.Group(s, "/api",
		option.Middleware(authmw.Auth(authService)),
		option.Middleware(authmw.RequireAdmin()),
	)
	auditHandler.MountRoutes(admin)

	adminToken := createAdminToken(t, db, s)
	user := createAuditHTTPUser(t, db, "audit-facet-user")
	repo := repositories.NewAuditLogRepository(db)
	now := time.Now().UTC()

	createAuditHTTPLog(t, repo, models.AuditLog{
		ID:        uuid.New(),
		UserID:    &user.ID,
		Action:    "sandbox.created",
		Details:   datatypes.JSON([]byte(`{}`)),
		Timestamp: now,
	})

	rec := performJSONRequest(t, s, http.MethodGet, "/api/audit-logs/facets?from="+now.Add(-time.Hour).Format(time.RFC3339), nil, "Bearer "+adminToken)

	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var response dto.AuditLogFacetsResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.Contains(t, response.Actions, "sandbox.created")
	assert.Contains(t, response.Actions, "image.created")
	assert.Contains(t, response.Users, dto.UserSummary{
		ID:        user.ID,
		Email:     user.Email,
		AvatarURL: dto.GravatarURL(user.Email, 80),
	})
}

func createAuditHTTPLog(t *testing.T, repo *repositories.AuditLogRepository, entry models.AuditLog) models.AuditLog {
	t.Helper()
	require.NoError(t, repo.Create(&entry))
	return entry
}

func createUserToken(t *testing.T, db *gorm.DB, s *fuego.Server) string {
	t.Helper()
	return createRoleToken(t, db, s, models.RoleUser)
}

func createAuditHTTPUser(t *testing.T, db *gorm.DB, prefix string) *models.User {
	t.Helper()

	user := &models.User{
		ID:           uuid.New(),
		Email:        fmt.Sprintf("%s-%d@example.com", prefix, time.Now().UnixNano()),
		PasswordHash: "hashed-password",
		Role:         models.RoleUser,
	}
	require.NoError(t, repositories.NewUserRepository(db).Create(user))
	return user
}

func strPtrHTTP(value string) *string {
	return &value
}

func createRoleToken(t *testing.T, db *gorm.DB, s *fuego.Server, role string) string {
	t.Helper()

	passwordService := services.NewPasswordService()
	password := "RoleSup3rS3cret!"
	passwordHash, err := passwordService.Hash(password)
	require.NoError(t, err)

	user := &models.User{
		ID:           uuid.New(),
		Email:        fmt.Sprintf("%s-%d@example.com", role, time.Now().UnixNano()),
		PasswordHash: passwordHash,
		Role:         role,
	}
	require.NoError(t, repositories.NewUserRepository(db).Create(user))

	loginRec := performJSONRequest(t, s, http.MethodPost, "/api/auth/login", map[string]any{
		"email":    user.Email,
		"password": password,
	}, "")
	require.Equal(t, http.StatusOK, loginRec.Code, loginRec.Body.String())

	var loginResp struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.Unmarshal(loginRec.Body.Bytes(), &loginResp))
	require.NotEmpty(t, loginResp.Token)

	return loginResp.Token
}

func createAdminToken(t *testing.T, db *gorm.DB, s *fuego.Server) string {
	t.Helper()
	return createRoleToken(t, db, s, models.RoleAdmin)
}
