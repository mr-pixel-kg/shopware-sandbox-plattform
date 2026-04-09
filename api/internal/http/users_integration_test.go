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
	"github.com/mr-pixel-kg/shopshredder/api/internal/http/dto"
	"github.com/mr-pixel-kg/shopshredder/api/internal/http/handlers"
	authmw "github.com/mr-pixel-kg/shopshredder/api/internal/http/middleware"
	"github.com/mr-pixel-kg/shopshredder/api/internal/models"
	"github.com/mr-pixel-kg/shopshredder/api/internal/repositories"
	"github.com/mr-pixel-kg/shopshredder/api/internal/services"
	"github.com/mr-pixel-kg/shopshredder/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminUsersCRUD(t *testing.T) {
	db := testutil.OpenIntegrationDB(t)
	testutil.ResetIntegrationDB(t, db)

	s, authService := newIntegrationServer(db)
	auditService := newTestAuditService(db)
	userService := services.NewUserService(repositories.NewUserRepository(db), services.NewPasswordService())
	authHandler := handlers.AuthHandler{Auth: authService, Audit: auditService}
	userHandler := handlers.UserHandler{Users: userService, Audit: auditService}

	public := fuego.Group(s, "/api")
	fuego.Post(public, "/auth/register", authHandler.Register, option.DefaultStatusCode(http.StatusCreated))
	fuego.Post(public, "/auth/login", authHandler.Login)

	admin := fuego.Group(s, "/api",
		option.Middleware(authmw.Auth(authService)),
		option.Middleware(authmw.RequireAdmin()),
	)
	fuego.Get(admin, "/users", userHandler.List)
	fuego.Get(admin, "/users/{id}", userHandler.Get)
	fuego.Post(admin, "/users", userHandler.Create, option.DefaultStatusCode(http.StatusCreated))
	fuego.Patch(admin, "/users/{id}", userHandler.Update)
	fuego.Delete(admin, "/users/{id}", userHandler.Delete, option.DefaultStatusCode(http.StatusNoContent))

	adminToken := createAdminToken(t, db, s)
	password := "Sup3rS3cret!"

	createRec := performJSONRequest(t, s, http.MethodPost, "/api/users", map[string]any{
		"email":    fmt.Sprintf("crud-user-%d@example.com", time.Now().UnixNano()),
		"role":     models.RoleUser,
		"password": password,
	}, "Bearer "+adminToken)
	require.Equal(t, http.StatusCreated, createRec.Code, createRec.Body.String())

	var created dto.UserResponse
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))
	assert.Equal(t, models.RoleUser, created.Role)

	storedCreated, err := repositories.NewUserRepository(db).FindByID(created.ID)
	require.NoError(t, err)
	assert.False(t, storedCreated.IsPending())

	getRec := performJSONRequest(t, s, http.MethodGet, "/api/users/"+created.ID.String(), nil, "Bearer "+adminToken)
	require.Equal(t, http.StatusOK, getRec.Code, getRec.Body.String())

	updateRec := performJSONRequest(t, s, http.MethodPatch, "/api/users/"+created.ID.String(), map[string]any{
		"email": created.Email,
		"role":  models.RoleAdmin,
	}, "Bearer "+adminToken)
	require.Equal(t, http.StatusOK, updateRec.Code, updateRec.Body.String())

	listRec := performJSONRequest(t, s, http.MethodGet, "/api/users", nil, "Bearer "+adminToken)
	require.Equal(t, http.StatusOK, listRec.Code, listRec.Body.String())
	assert.Contains(t, listRec.Body.String(), created.Email)

	deleteRec := performJSONRequest(t, s, http.MethodDelete, "/api/users/"+created.ID.String(), nil, "Bearer "+adminToken)
	require.Equal(t, http.StatusNoContent, deleteRec.Code, deleteRec.Body.String())

	getDeletedRec := performJSONRequest(t, s, http.MethodGet, "/api/users/"+created.ID.String(), nil, "Bearer "+adminToken)
	assert.Equal(t, http.StatusNotFound, getDeletedRec.Code, getDeletedRec.Body.String())
}
