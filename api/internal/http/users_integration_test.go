//go:build integration

package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/handlers"
	authmw "github.com/manuel/shopware-testenv-platform/api/internal/http/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/manuel/shopware-testenv-platform/api/internal/repositories"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
	"github.com/manuel/shopware-testenv-platform/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAdminUsersCRUD(t *testing.T) {
	db := testutil.OpenIntegrationDB(t)
	testutil.ResetIntegrationDB(t, db)

	router := newIntegrationRouter()
	authService, auditService := newTestAuthServices(db)
	userService := services.NewUserService(repositories.NewUserRepository(db), services.NewPasswordService())
	authHandler := handlers.NewAuthHandler(authService, auditService)
	userHandler := handlers.NewUserHandler(userService, auditService)

	router.POST("/api/auth/login", authHandler.Login)

	private := router.Group("/api")
	private.Use(authmw.Auth(authService))
	admin := private.Group("/admin")
	admin.Use(authmw.RequireAdmin())
	admin.GET("/users", userHandler.List)
	admin.GET("/users/:id", userHandler.Get)
	admin.POST("/users", userHandler.Create)
	admin.PUT("/users/:id", userHandler.Update)
	admin.DELETE("/users/:id", userHandler.Delete)

	adminToken := createAdminToken(t, db, router)
	password := "Sup3rS3cret!"

	createRec := performJSONRequest(t, router, http.MethodPost, "/api/admin/users", map[string]any{
		"email":    fmt.Sprintf("crud-user-%d@example.com", time.Now().UnixNano()),
		"role":     models.RoleUser,
		"password": password,
	}, "Bearer "+adminToken)
	require.Equal(t, http.StatusCreated, createRec.Code, createRec.Body.String())

	var created models.User
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))
	assert.Equal(t, models.RoleUser, created.Role)

	storedCreated, err := repositories.NewUserRepository(db).FindByID(created.ID)
	require.NoError(t, err)
	assert.False(t, storedCreated.IsPending())

	getRec := performJSONRequest(t, router, http.MethodGet, "/api/admin/users/"+created.ID.String(), nil, "Bearer "+adminToken)
	require.Equal(t, http.StatusOK, getRec.Code, getRec.Body.String())

	updateRec := performJSONRequest(t, router, http.MethodPut, "/api/admin/users/"+created.ID.String(), map[string]any{
		"email": created.Email,
		"role":  models.RoleAdmin,
	}, "Bearer "+adminToken)
	require.Equal(t, http.StatusOK, updateRec.Code, updateRec.Body.String())

	listRec := performJSONRequest(t, router, http.MethodGet, "/api/admin/users", nil, "Bearer "+adminToken)
	require.Equal(t, http.StatusOK, listRec.Code, listRec.Body.String())
	assert.Contains(t, listRec.Body.String(), created.Email)

	deleteRec := performJSONRequest(t, router, http.MethodDelete, "/api/admin/users/"+created.ID.String(), nil, "Bearer "+adminToken)
	require.Equal(t, http.StatusNoContent, deleteRec.Code, deleteRec.Body.String())

	getDeletedRec := performJSONRequest(t, router, http.MethodGet, "/api/admin/users/"+created.ID.String(), nil, "Bearer "+adminToken)
	assert.Equal(t, http.StatusNotFound, getDeletedRec.Code, getDeletedRec.Body.String())
}

func createAdminToken(t *testing.T, db *gorm.DB, router *echo.Echo) string {
	t.Helper()

	passwordService := services.NewPasswordService()
	password := "Adm1nSup3rS3cret!"
	passwordHash, err := passwordService.Hash(password)
	require.NoError(t, err)

	admin := &models.User{
		ID:           uuid.New(),
		Email:        fmt.Sprintf("admin-%d@example.com", time.Now().UnixNano()),
		PasswordHash: passwordHash,
		Role:         models.RoleAdmin,
	}
	require.NoError(t, repositories.NewUserRepository(db).Create(admin))

	loginRec := performJSONRequest(t, router, http.MethodPost, "/api/auth/login", map[string]any{
		"email":    admin.Email,
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
