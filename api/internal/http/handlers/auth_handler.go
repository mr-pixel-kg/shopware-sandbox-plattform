package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-fuego/fuego"
	auditcontracts "github.com/mr-pixel-kg/shopshredder/api/internal/auditlog"
	"github.com/mr-pixel-kg/shopshredder/api/internal/http/dto"
	mw "github.com/mr-pixel-kg/shopshredder/api/internal/http/middleware"
	"github.com/mr-pixel-kg/shopshredder/api/internal/services"
)

type AuthHandler struct {
	Auth  *services.AuthService
	Audit *services.AuditService
}

func (h AuthHandler) Register(c fuego.ContextWithBody[dto.RegisterRequest]) (dto.UserResponse, error) {
	body, err := c.Body()
	if err != nil {
		return dto.UserResponse{}, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "Invalid request body"}
	}

	slog.Debug("register request received", "component", "auth", "email", body.Email)
	user, err := h.Auth.Register(body.Email, body.Password)
	if err != nil {
		if errors.Is(err, services.ErrEmailNotWhitelisted) {
			return dto.UserResponse{}, fuego.HTTPError{Status: http.StatusForbidden, Detail: "Email not whitelisted for registration"}
		}
		return dto.UserResponse{}, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "Could not register user"}
	}

	slog.Info("user registered", "component", "auth", "user_id", user.ID, "email", user.Email)
	resourceType := auditcontracts.ResourceTypeUser
	_ = h.Audit.Log(newAuditLogInput(c.Request(), &user.ID, auditcontracts.ActionUserRegistered, &resourceType, &user.ID, map[string]any{"email": user.Email}))

	return dto.UserResponse{
		ID: user.ID, Email: user.Email, AvatarURL: dto.GravatarURL(user.Email, 80), Role: user.Role,
		IsPending: user.IsPending(), CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
	}, nil
}

func (h AuthHandler) Login(c fuego.ContextWithBody[dto.LoginRequest]) (dto.LoginResponse, error) {
	body, err := c.Body()
	if err != nil {
		return dto.LoginResponse{}, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "Invalid request body"}
	}

	slog.Debug("login request received", "component", "auth", "email", body.Email)
	token, user, err := h.Auth.Login(body.Email, body.Password)
	if err != nil {
		return dto.LoginResponse{}, fuego.HTTPError{Status: http.StatusUnauthorized, Detail: "Email or password is invalid"}
	}

	slog.Info("user logged in", "component", "auth", "user_id", user.ID, "email", user.Email)
	_ = h.Audit.Log(newAuditLogInput(c.Request(), &user.ID, auditcontracts.ActionAuthLoggedIn, nil, nil, map[string]any{}))

	return dto.LoginResponse{
		Token: token,
		User: dto.UserResponse{
			ID: user.ID, Email: user.Email, AvatarURL: dto.GravatarURL(user.Email, 80), Role: user.Role,
			IsPending: user.IsPending(), CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

func (h AuthHandler) Logout(c fuego.ContextNoBody) (any, error) {
	auth := mw.MustAuth(c.Request())
	slog.Info("user logged out", "component", "auth", "user_id", auth.UserID)
	_ = h.Audit.Log(newAuditLogInput(c.Request(), &auth.UserID, auditcontracts.ActionAuthLoggedOut, nil, nil, map[string]any{}))
	return nil, nil
}

func (h AuthHandler) Me(c fuego.ContextNoBody) (dto.UserResponse, error) {
	user := mw.UserFromContext(c.Request())
	return dto.UserResponse{
		ID: user.ID, Email: user.Email, AvatarURL: dto.GravatarURL(user.Email, 80), Role: user.Role,
		IsPending: user.IsPending(), CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
	}, nil
}
