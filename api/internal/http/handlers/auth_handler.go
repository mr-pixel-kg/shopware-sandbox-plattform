package handlers

import (
	"errors"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/manuel/shopware-testenv-platform/api/internal/apperror"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
	mw "github.com/manuel/shopware-testenv-platform/api/internal/http/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/responses"
	"github.com/manuel/shopware-testenv-platform/api/internal/logging"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
)

type AuthHandler struct {
	auth  *services.AuthService
	audit *services.AuditService
}

func NewAuthHandler(auth *services.AuthService, audit *services.AuditService) *AuthHandler {
	return &AuthHandler{auth: auth, audit: audit}
}

// Register godoc
// @Summary      Register a new user
// @Description  Create a new user account with email and password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body dto.RegisterRequest true "Registration credentials"
// @Success      201 {object} models.User
// @Failure      400 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse
// @Router       /api/auth/register [post]
func (h *AuthHandler) Register(c echo.Context) error {
	var input dto.RegisterRequest
	if err := c.Bind(&input); err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid request body"))
	}

	slog.Debug("register request received", logging.RequestFields(c, "component", "auth", "email", logging.MaskEmail(input.Email))...)
	user, err := h.auth.Register(input.Email, input.Password)
	if err != nil {
		if errors.Is(err, services.ErrEmailNotWhitelisted) {
			return responses.FromAppError(c, apperror.Forbidden("Email not whitelisted for registration"))
		}
		return responses.FromAppError(c, apperror.BadRequest("REGISTER_FAILED", "Could not register user").WithCause(err))
	}

	slog.Info("user registered", logging.RequestFields(c,
		"component", "auth",
		"user_id", user.ID.String(),
		"email", logging.MaskEmail(user.Email),
	)...)
	_ = h.audit.Log(&user.ID, "auth.registered", c.RealIP(), map[string]any{"email": user.Email})
	return c.JSON(201, user)
}

// Login godoc
// @Summary      Log in
// @Description  Authenticate with email and password, receive a JWT token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body dto.LoginRequest true "Login credentials"
// @Success      200 {object} dto.AuthLoginResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Router       /api/auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var input dto.LoginRequest
	if err := c.Bind(&input); err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid request body"))
	}

	slog.Debug("login request received", logging.RequestFields(c, "component", "auth", "email", logging.MaskEmail(input.Email))...)
	token, user, err := h.auth.Login(input.Email, input.Password)
	if err != nil {
		return responses.FromAppError(c, apperror.Unauthorized("Email or password is invalid").WithCause(err))
	}

	slog.Info("user logged in", logging.RequestFields(c,
		"component", "auth",
		"user_id", user.ID.String(),
		"email", logging.MaskEmail(user.Email),
	)...)
	_ = h.audit.Log(&user.ID, "auth.logged_in", c.RealIP(), map[string]any{})
	return c.JSON(200, dto.AuthLoginResponse{
		Token: token,
		User:  *user,
	})
}

// Logout godoc
// @Summary      Log out
// @Description  Invalidate the current session token
// @Tags         Auth
// @Security     BearerAuth
// @Produce      json
// @Success      204
// @Failure      401 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/auth/logout [post]
func (h *AuthHandler) Logout(c echo.Context) error {
	auth := mw.MustAuth(c)
	slog.Debug("logout request received", logging.RequestFields(c, "component", "auth", "user_id", auth.UserID.String(), "token_id", auth.TokenID)...)
	if err := h.auth.Logout(auth.TokenID); err != nil {
		return responses.FromAppError(c, apperror.Internal("LOGOUT_FAILED", "Could not log out").WithCause(err))
	}

	slog.Info("user logged out", logging.RequestFields(c, "component", "auth", "user_id", auth.UserID.String(), "token_id", auth.TokenID)...)
	_ = h.audit.Log(&auth.UserID, "auth.logged_out", c.RealIP(), map[string]any{})
	return c.NoContent(204)
}

// Me godoc
// @Summary      Get current user
// @Description  Return the authenticated user's profile
// @Tags         Auth
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} models.User
// @Failure      401 {object} dto.ErrorResponse
// @Router       /api/me [get]
func (h *AuthHandler) Me(c echo.Context) error {
	auth := mw.MustAuth(c)
	slog.Debug("profile requested", logging.RequestFields(c, "component", "auth", "user_id", auth.UserID.String())...)
	return c.JSON(200, c.Get("user"))
}
