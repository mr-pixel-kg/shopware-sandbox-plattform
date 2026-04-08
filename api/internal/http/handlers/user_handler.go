package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/google/uuid"
	"github.com/manuel/shopware-testenv-platform/api/internal/apperror"
	auditcontracts "github.com/manuel/shopware-testenv-platform/api/internal/auditlog"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
	mw "github.com/manuel/shopware-testenv-platform/api/internal/http/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
	"gorm.io/gorm"
)

type UserHandler struct {
	Users *services.UserService
	Audit *services.AuditService
}

func (h UserHandler) MountRoutes(s *fuego.Server) {
	users := fuego.Group(s, "/users")
	fuego.Get(users, "", h.list,
		option.Summary("List users"),
		option.Description("Return all users, including pending invited users"),
		option.Tags("Users"),
		option.QueryInt("limit", "Max entries per page (1-500, default 50)"),
		option.QueryInt("offset", "Offset for pagination (default 0)"),
	)
	fuego.Get(users, "/{id}", h.get,
		option.Summary("Get user"),
		option.Description("Return a single user by ID"),
		option.Tags("Users"),
	)
	fuego.Post(users, "", h.create,
		option.Summary("Create user"),
		option.Description("Create an active user or invite a pending user when no password is provided"),
		option.Tags("Users"),
		option.DefaultStatusCode(http.StatusCreated),
	)
	fuego.Patch(users, "/{id}", h.update,
		option.Summary("Update user"),
		option.Description("Update a user's email, role, and optionally password"),
		option.Tags("Users"),
	)
	fuego.Delete(users, "/{id}", h.delete,
		option.Summary("Delete user"),
		option.Description("Delete a user by ID"),
		option.Tags("Users"),
		option.DefaultStatusCode(http.StatusNoContent),
	)
}

func (h UserHandler) list(c fuego.ContextNoBody) (dto.UserListResponse, error) {
	limit, offset, err := parsePaginationParams(c.Request())
	if err != nil {
		return dto.UserListResponse{}, err
	}

	result, err := h.Users.ListPaginated(services.UserListInput{Limit: limit, Offset: offset})
	if err != nil {
		return dto.UserListResponse{}, fuego.HTTPError{Status: http.StatusInternalServerError, Detail: "Could not list users"}
	}

	out := make([]dto.UserResponse, len(result.Users))
	for i, u := range result.Users {
		out[i] = dto.UserResponse{
			ID: u.ID, Email: u.Email, AvatarURL: dto.GravatarURL(u.Email, 80), Role: u.Role,
			IsPending: u.IsPending(), CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
		}
	}
	return dto.UserListResponse{
		Data: out,
		Meta: dto.PaginatedMeta{
			Pagination: buildPaginationMeta(len(out), result.Limit, result.Offset, result.Total),
		},
	}, nil
}

func (h UserHandler) get(c fuego.ContextNoBody) (dto.UserResponse, error) {
	id, err := parsePathUUID(c, "id")
	if err != nil {
		return dto.UserResponse{}, err
	}

	user, err := h.Users.Get(id)
	if err != nil {
		return dto.UserResponse{}, mapUserError(err)
	}

	return dto.UserResponse{
		ID: user.ID, Email: user.Email, AvatarURL: dto.GravatarURL(user.Email, 80), Role: user.Role,
		IsPending: user.IsPending(), CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
	}, nil
}

func (h UserHandler) create(c fuego.ContextWithBody[dto.CreateUserRequest]) (dto.UserResponse, error) {
	body, err := c.Body()
	if err != nil {
		return dto.UserResponse{}, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "Invalid request body"}
	}

	user, err := h.Users.Create(body.Email, body.Role, body.Password)
	if err != nil {
		return dto.UserResponse{}, mapUserError(err)
	}

	auth := mw.MustAuth(c.Request())
	slog.Info("user created", "component", "admin", "user_id", user.ID, "email", user.Email, "pending", user.IsPending())
	resourceType := auditcontracts.ResourceTypeUser
	_ = h.Audit.Log(newAuditLogInput(c.Request(), &auth.UserID, auditcontracts.ActionUserCreated, &resourceType, &user.ID, map[string]any{
		"email": user.Email, "role": user.Role, "pending": user.IsPending(),
	}))

	return dto.UserResponse{
		ID: user.ID, Email: user.Email, AvatarURL: dto.GravatarURL(user.Email, 80), Role: user.Role,
		IsPending: user.IsPending(), CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
	}, nil
}

func (h UserHandler) update(c fuego.ContextWithBody[dto.UpdateUserRequest]) (dto.UserResponse, error) {
	id, err := parsePathUUID(c, "id")
	if err != nil {
		return dto.UserResponse{}, err
	}

	body, err := c.Body()
	if err != nil {
		return dto.UserResponse{}, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "Invalid request body"}
	}

	user, err := h.Users.Update(id, body.Email, body.Role, body.Password)
	if err != nil {
		return dto.UserResponse{}, mapUserError(err)
	}

	auth := mw.MustAuth(c.Request())
	slog.Info("user updated", "component", "admin", "user_id", user.ID, "email", user.Email)
	resourceType := auditcontracts.ResourceTypeUser
	_ = h.Audit.Log(newAuditLogInput(c.Request(), &auth.UserID, auditcontracts.ActionUserUpdated, &resourceType, &user.ID, map[string]any{
		"email": user.Email, "role": user.Role,
	}))

	return dto.UserResponse{
		ID: user.ID, Email: user.Email, AvatarURL: dto.GravatarURL(user.Email, 80), Role: user.Role,
		IsPending: user.IsPending(), CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
	}, nil
}

func (h UserHandler) delete(c fuego.ContextNoBody) (any, error) {
	id, err := parsePathUUID(c, "id")
	if err != nil {
		return nil, err
	}

	auth := mw.MustAuth(c.Request())
	if auth.UserID == id {
		return nil, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "You cannot delete your own user"}
	}

	user, err := h.Users.Get(id)
	if err != nil {
		return nil, mapUserError(err)
	}

	if err := h.Users.Delete(id); err != nil {
		return nil, mapUserError(err)
	}

	slog.Info("user deleted", "component", "admin", "user_id", user.ID, "email", user.Email)
	resourceType := auditcontracts.ResourceTypeUser
	_ = h.Audit.Log(newAuditLogInput(c.Request(), &auth.UserID, auditcontracts.ActionUserDeleted, &resourceType, &user.ID, map[string]any{
		"email": user.Email,
	}))

	return nil, nil
}

func mapUserError(err error) error {
	if err == nil {
		return nil
	}
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		return fuego.HTTPError{Status: appErr.StatusCode, Detail: appErr.Message}
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fuego.HTTPError{Status: http.StatusNotFound, Detail: "User not found"}
	}
	return fuego.HTTPError{Status: http.StatusInternalServerError, Detail: "User operation failed"}
}

func parsePathUUID(c interface{ PathParam(string) string }, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.PathParam(name))
	if err != nil {
		return uuid.Nil, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "Invalid " + name}
	}
	return id, nil
}
