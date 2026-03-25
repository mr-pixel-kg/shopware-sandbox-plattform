package handlers

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/manuel/shopware-testenv-platform/api/internal/apperror"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
	mw "github.com/manuel/shopware-testenv-platform/api/internal/http/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/responses"
	"github.com/manuel/shopware-testenv-platform/api/internal/logging"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
)

type UserHandler struct {
	users *services.UserService
	audit *services.AuditService
}

func NewUserHandler(users *services.UserService, audit *services.AuditService) *UserHandler {
	return &UserHandler{users: users, audit: audit}
}

// List godoc
// @Summary      List users
// @Description  Return all users, including pending invited users
// @Tags         Admin
// @Security     BearerAuth
// @Produce      json
// @Success      200 {array} models.User
// @Failure      403 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/admin/users [get]
func (h *UserHandler) List(c echo.Context) error {
	users, err := h.users.List()
	if err != nil {
		return responses.FromAppError(c, apperror.Internal("USER_LIST_FAILED", "Could not list users").WithCause(err))
	}

	return c.JSON(http.StatusOK, users)
}

// Get godoc
// @Summary      Get user
// @Description  Return a single user by ID
// @Tags         Admin
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "User ID" format(uuid)
// @Success      200 {object} models.User
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/admin/users/{id} [get]
func (h *UserHandler) Get(c echo.Context) error {
	id, err := parseUserID(c)
	if err != nil {
		return err
	}

	user, getErr := h.users.Get(id)
	if getErr != nil {
		return responses.FromError(c, getErr)
	}

	return c.JSON(http.StatusOK, user)
}

// Create godoc
// @Summary      Create user
// @Description  Create an active user or invite a pending user when no password is provided
// @Tags         Admin
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.CreateUserRequest true "User payload"
// @Success      201 {object} models.User
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/admin/users [post]
func (h *UserHandler) Create(c echo.Context) error {
	var input dto.CreateUserRequest
	if err := c.Bind(&input); err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid request body"))
	}

	user, err := h.users.Create(input.Email, input.Role, input.Password)
	if err != nil {
		return responses.FromError(c, err)
	}

	auth := mw.MustAuth(c)
	slog.Info("user created", logging.RequestFields(c,
		"component", "admin",
		"user_id", user.ID.String(),
		"email", logging.MaskEmail(user.Email),
		"pending", user.IsPending(),
	)...)
	_ = h.audit.Log(&auth.UserID, "admin.user_created", c.RealIP(), map[string]any{
		"userId":  user.ID.String(),
		"email":   user.Email,
		"role":    user.Role,
		"pending": user.IsPending(),
	})

	return c.JSON(http.StatusCreated, user)
}

// Update godoc
// @Summary      Update user
// @Description  Update a user's email, role, and optionally password
// @Tags         Admin
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID" format(uuid)
// @Param        body body dto.UpdateUserRequest true "User payload"
// @Success      200 {object} models.User
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/admin/users/{id} [put]
func (h *UserHandler) Update(c echo.Context) error {
	id, err := parseUserID(c)
	if err != nil {
		return err
	}

	var input dto.UpdateUserRequest
	if bindErr := c.Bind(&input); bindErr != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid request body"))
	}

	user, updateErr := h.users.Update(id, input.Email, input.Role, input.Password)
	if updateErr != nil {
		return responses.FromError(c, updateErr)
	}

	auth := mw.MustAuth(c)
	slog.Info("user updated", logging.RequestFields(c,
		"component", "admin",
		"user_id", user.ID.String(),
		"email", logging.MaskEmail(user.Email),
	)...)
	_ = h.audit.Log(&auth.UserID, "admin.user_updated", c.RealIP(), map[string]any{
		"userId": user.ID.String(),
		"email":  user.Email,
		"role":   user.Role,
	})

	return c.JSON(http.StatusOK, user)
}

// Delete godoc
// @Summary      Delete user
// @Description  Delete a user by ID
// @Tags         Admin
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "User ID" format(uuid)
// @Success      204
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/admin/users/{id} [delete]
func (h *UserHandler) Delete(c echo.Context) error {
	id, err := parseUserID(c)
	if err != nil {
		return err
	}

	auth := mw.MustAuth(c)
	if auth.UserID == id {
		return responses.FromAppError(c, apperror.BadRequest("CANNOT_DELETE_SELF", "You cannot delete your own user"))
	}

	user, getErr := h.users.Get(id)
	if getErr != nil {
		return responses.FromError(c, getErr)
	}

	if deleteErr := h.users.Delete(id); deleteErr != nil {
		return responses.FromError(c, deleteErr)
	}

	slog.Info("user deleted", logging.RequestFields(c,
		"component", "admin",
		"user_id", user.ID.String(),
		"email", logging.MaskEmail(user.Email),
	)...)
	_ = h.audit.Log(&auth.UserID, "admin.user_deleted", c.RealIP(), map[string]any{
		"userId": user.ID.String(),
		"email":  user.Email,
	})

	return c.NoContent(http.StatusNoContent)
}

func parseUserID(c echo.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return uuid.Nil, responses.FromAppError(c, apperror.BadRequest("INVALID_ID", "Invalid UUID"))
	}

	return id, nil
}
