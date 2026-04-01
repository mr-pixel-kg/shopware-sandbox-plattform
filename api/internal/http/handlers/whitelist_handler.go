package handlers

import (
	"log/slog"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/manuel/shopware-testenv-platform/api/internal/apperror"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
	mw "github.com/manuel/shopware-testenv-platform/api/internal/http/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/responses"
	"github.com/manuel/shopware-testenv-platform/api/internal/logging"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/manuel/shopware-testenv-platform/api/internal/repositories"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
)

type WhitelistHandler struct {
	users *repositories.UserRepository
	audit *services.AuditService
}

func NewWhitelistHandler(users *repositories.UserRepository, audit *services.AuditService) *WhitelistHandler {
	return &WhitelistHandler{users: users, audit: audit}
}

// List godoc
// @Summary      List whitelisted emails
// @Description  Return all pending (whitelisted but not yet registered) users
// @Tags         Admin
// @Security     BearerAuth
// @Produce      json
// @Success      200 {array} models.User
// @Failure      403 {object} dto.ErrorResponse
// @Router       /api/admin/whitelist [get]
func (h *WhitelistHandler) List(c echo.Context) error {
	users, err := h.users.ListPending()
	if err != nil {
		return responses.FromAppError(c, apperror.Internal("WHITELIST_LIST_FAILED", "Could not list whitelisted emails").WithCause(err))
	}
	return c.JSON(200, users)
}

// Add godoc
// @Summary      Add email to whitelist
// @Description  Create a pending user row so the email can register in whitelist mode
// @Tags         Admin
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.AddWhitelistRequest true "Email to whitelist"
// @Success      201 {object} models.User
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse
// @Router       /api/admin/whitelist [post]
func (h *WhitelistHandler) Add(c echo.Context) error {
	var input dto.AddWhitelistRequest
	if err := bindAndValidate(c, &input); err != nil {
		return responses.FromError(c, err)
	}

	auth := mw.MustAuth(c)
	user := &models.User{
		ID:    uuid.New(),
		Email: input.Email,
		Role:  input.Role,
	}

	if err := h.users.Create(user); err != nil {
		return responses.FromAppError(c, apperror.Conflict("EMAIL_EXISTS", "Email already exists").WithCause(err))
	}

	slog.Info("email whitelisted", logging.RequestFields(c,
		"component", "admin",
		"whitelisted_email", logging.MaskEmail(input.Email),
	)...)
	_ = h.audit.Log(&auth.UserID, "admin.whitelist_added", c.RealIP(), map[string]any{"email": input.Email})
	return c.JSON(201, user)
}

// Remove godoc
// @Summary      Remove email from whitelist
// @Description  Delete a pending user row (only works for users that have not yet registered)
// @Tags         Admin
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Pending user ID" format(uuid)
// @Success      204
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/admin/whitelist/{id} [delete]
func (h *WhitelistHandler) Remove(c echo.Context) error {
	id, err := parseUUIDParam(c, "id", "INVALID_ID", "Invalid UUID")
	if err != nil {
		return responses.FromError(c, err)
	}

	user, err := h.users.FindByID(id)
	if err != nil {
		return responses.FromAppError(c, apperror.NotFound("NOT_FOUND", "Whitelisted email not found").WithCause(err))
	}

	if !user.IsPending() {
		return responses.FromAppError(c, apperror.BadRequest("NOT_PENDING", "User has already registered"))
	}

	if err := h.users.DeletePending(id); err != nil {
		return responses.FromAppError(c, apperror.Internal("WHITELIST_DELETE_FAILED", "Could not remove whitelisted email").WithCause(err))
	}

	auth := mw.MustAuth(c)
	slog.Info("email removed from whitelist", logging.RequestFields(c,
		"component", "admin",
		"removed_email", logging.MaskEmail(user.Email),
	)...)
	_ = h.audit.Log(&auth.UserID, "admin.whitelist_removed", c.RealIP(), map[string]any{"email": user.Email})
	return c.NoContent(204)
}
