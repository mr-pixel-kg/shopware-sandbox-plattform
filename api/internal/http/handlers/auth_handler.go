package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	auditcontracts "github.com/manuel/shopware-testenv-platform/api/internal/auditlog"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
	mw "github.com/manuel/shopware-testenv-platform/api/internal/http/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
)

type AuthHandler struct {
	Auth  *services.AuthService
	Audit *services.AuditService
}

func (h AuthHandler) MountPublicRoutes(s *fuego.Server) {
	auth := fuego.Group(s, "/auth")
	fuego.Post(auth, "/register", h.register,
		option.Summary("Register a new user"),
		option.Description("Create a new user account with email and password"),
		option.Tags("Auth"),
		option.DefaultStatusCode(http.StatusCreated),
	)
	fuego.Post(auth, "/login", h.login,
		option.Summary("Log in"),
		option.Description("Authenticate with email and password, receive a JWT token"),
		option.Tags("Auth"),
	)
}

func (h AuthHandler) MountAuthedRoutes(s *fuego.Server) {
	auth := fuego.Group(s, "/auth")
	fuego.Post(auth, "/logout", h.logout,
		option.Summary("Log out"),
		option.Description("Invalidate the current session token"),
		option.Tags("Auth"),
		option.DefaultStatusCode(http.StatusNoContent),
	)
	fuego.Get(auth, "/me", h.me,
		option.Summary("Get current user"),
		option.Description("Return the authenticated user's profile"),
		option.Tags("Auth"),
	)
}

func (h AuthHandler) register(c fuego.ContextWithBody[dto.RegisterRequest]) (dto.UserResponse, error) {
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

func (h AuthHandler) login(c fuego.ContextWithBody[dto.LoginRequest]) (dto.LoginResponse, error) {
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

func (h AuthHandler) logout(c fuego.ContextNoBody) (any, error) {
	auth := mw.MustAuth(c.Request())
	slog.Info("user logged out", "component", "auth", "user_id", auth.UserID)
	_ = h.Audit.Log(newAuditLogInput(c.Request(), &auth.UserID, auditcontracts.ActionAuthLoggedOut, nil, nil, map[string]any{}))
	return nil, nil
}

func (h AuthHandler) me(c fuego.ContextNoBody) (dto.UserResponse, error) {
	user := mw.UserFromContext(c.Request())
	return dto.UserResponse{
		ID: user.ID, Email: user.Email, AvatarURL: dto.GravatarURL(user.Email, 80), Role: user.Role,
		IsPending: user.IsPending(), CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
	}, nil
}
