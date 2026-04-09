package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-fuego/fuego"
	auditcontracts "github.com/mr-pixel-kg/shopshredder/api/internal/auditlog"
	"github.com/mr-pixel-kg/shopshredder/api/internal/http/dto"
	mw "github.com/mr-pixel-kg/shopshredder/api/internal/http/middleware"
	"github.com/mr-pixel-kg/shopshredder/api/internal/services"
)

type WhitelistHandler struct {
	Users *services.UserService
	Audit *services.AuditService
}

func (h WhitelistHandler) List(c fuego.ContextNoBody) (dto.UserListResponse, error) {
	users, err := h.Users.ListPending()
	if err != nil {
		return dto.UserListResponse{}, fuego.HTTPError{Status: http.StatusInternalServerError, Detail: "Could not list whitelist"}
	}

	out := make([]dto.UserResponse, len(users))
	for i, u := range users {
		out[i] = dto.UserResponse{
			ID: u.ID, Email: u.Email, AvatarURL: dto.GravatarURL(u.Email, 80), Role: u.Role,
			IsPending: u.IsPending(), CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
		}
	}
	return dto.UserListResponse{
		Data: out,
		Meta: dto.PaginatedMeta{
			Pagination: buildPaginationMeta(len(out), len(out), 0, int64(len(out))),
		},
	}, nil
}

func (h WhitelistHandler) Add(c fuego.ContextWithBody[dto.AddWhitelistRequest]) (dto.UserResponse, error) {
	body, err := c.Body()
	if err != nil {
		return dto.UserResponse{}, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "Invalid request body"}
	}

	auth := mw.MustAuth(c.Request())
	user, err := h.Users.AddWhitelist(body.Email, body.Role)
	if err != nil {
		return dto.UserResponse{}, mapUserError(err)
	}

	slog.Info("email whitelisted", "component", "admin", "whitelisted_email", body.Email)
	resourceType := auditcontracts.ResourceTypeUser
	_ = h.Audit.Log(newAuditLogInput(c.Request(), &auth.UserID, auditcontracts.ActionUserWhitelisted, &resourceType, &user.ID, map[string]any{
		"email": user.Email, "role": user.Role,
	}))

	return dto.UserResponse{
		ID: user.ID, Email: user.Email, AvatarURL: dto.GravatarURL(user.Email, 80), Role: user.Role,
		IsPending: user.IsPending(), CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
	}, nil
}

func (h WhitelistHandler) Remove(c fuego.ContextNoBody) (any, error) {
	id, err := parsePathUUID(c, "id")
	if err != nil {
		return nil, err
	}

	user, err := h.Users.Get(id)
	if err != nil {
		return nil, fuego.HTTPError{Status: http.StatusNotFound, Detail: "Whitelist entry not found"}
	}

	if err := h.Users.RemoveWhitelist(id); err != nil {
		return nil, mapUserError(err)
	}

	auth := mw.MustAuth(c.Request())
	slog.Info("email removed from whitelist", "component", "admin", "removed_email", user.Email)
	resourceType := auditcontracts.ResourceTypeUser
	_ = h.Audit.Log(newAuditLogInput(c.Request(), &auth.UserID, auditcontracts.ActionUserWhitelistRemoved, &resourceType, &user.ID, map[string]any{
		"email": user.Email,
	}))

	return nil, nil
}
