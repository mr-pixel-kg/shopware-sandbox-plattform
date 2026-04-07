package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/manuel/shopware-testenv-platform/api/internal/apperror"
	auditcontracts "github.com/manuel/shopware-testenv-platform/api/internal/auditlog"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
	mw "github.com/manuel/shopware-testenv-platform/api/internal/http/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
)

type WhitelistHandler struct {
	Users *services.UserService
	Audit *services.AuditService
}

func (h WhitelistHandler) MountRoutes(s *fuego.Server) {
	wl := fuego.Group(s, "/whitelist")
	fuego.Get(wl, "", h.list,
		option.Summary("List whitelisted emails"),
		option.Description("Return all pending (whitelisted but not yet registered) users"),
		option.Tags("Whitelist"),
	)
	fuego.Post(wl, "", h.add,
		option.Summary("Add email to whitelist"),
		option.Description("Create a pending user row so the email can register in whitelist mode"),
		option.Tags("Whitelist"),
		option.DefaultStatusCode(http.StatusCreated),
	)
	fuego.Delete(wl, "/{id}", h.remove,
		option.Summary("Remove email from whitelist"),
		option.Description("Delete a pending user row (only works for users that have not yet registered)"),
		option.Tags("Whitelist"),
		option.DefaultStatusCode(http.StatusNoContent),
	)
}

func (h WhitelistHandler) list(c fuego.ContextNoBody) ([]dto.UserResponse, error) {
	users, err := h.Users.ListPending()
	if err != nil {
		return nil, fuego.HTTPError{Status: http.StatusInternalServerError, Detail: "Could not list whitelist"}
	}

	out := make([]dto.UserResponse, len(users))
	for i, u := range users {
		out[i] = dto.UserResponse{
			ID: u.ID, Email: u.Email, Role: u.Role,
			IsPending: u.IsPending(), CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
		}
	}
	return out, nil
}

func (h WhitelistHandler) add(c fuego.ContextWithBody[dto.AddWhitelistRequest]) (dto.UserResponse, error) {
	body, err := c.Body()
	if err != nil {
		return dto.UserResponse{}, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "Invalid request body"}
	}

	auth := mw.MustAuth(c.Request())
	user, err := h.Users.AddWhitelist(body.Email, body.Role)
	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			return dto.UserResponse{}, fuego.HTTPError{Status: appErr.StatusCode, Detail: appErr.Message}
		}
		return dto.UserResponse{}, fuego.HTTPError{Status: http.StatusInternalServerError, Detail: "Could not add to whitelist"}
	}

	slog.Info("email whitelisted", "component", "admin", "whitelisted_email", body.Email)
	resourceType := auditcontracts.ResourceTypeUser
	_ = h.Audit.Log(newAuditLogInput(c.Request(), &auth.UserID, auditcontracts.ActionUserWhitelisted, &resourceType, &user.ID, map[string]any{
		"email": user.Email, "role": user.Role,
	}))

	return dto.UserResponse{
		ID: user.ID, Email: user.Email, Role: user.Role,
		IsPending: user.IsPending(), CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
	}, nil
}

func (h WhitelistHandler) remove(c fuego.ContextNoBody) (any, error) {
	id, err := parsePathUUID(c, "id")
	if err != nil {
		return nil, err
	}

	user, err := h.Users.Get(id)
	if err != nil {
		return nil, fuego.HTTPError{Status: http.StatusNotFound, Detail: "Whitelist entry not found"}
	}

	if err := h.Users.RemoveWhitelist(id); err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			return nil, fuego.HTTPError{Status: appErr.StatusCode, Detail: appErr.Message}
		}
		return nil, fuego.HTTPError{Status: http.StatusInternalServerError, Detail: "Could not remove whitelist entry"}
	}

	auth := mw.MustAuth(c.Request())
	slog.Info("email removed from whitelist", "component", "admin", "removed_email", user.Email)
	resourceType := auditcontracts.ResourceTypeUser
	_ = h.Audit.Log(newAuditLogInput(c.Request(), &auth.UserID, auditcontracts.ActionUserWhitelistRemoved, &resourceType, &user.ID, map[string]any{
		"email": user.Email,
	}))

	return nil, nil
}
