package handlers

import (
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/manuel/shopware-testenv-platform/api/internal/apperror"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
	mw "github.com/manuel/shopware-testenv-platform/api/internal/http/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/responses"
	"github.com/manuel/shopware-testenv-platform/api/internal/logging"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
)

type SandboxHandler struct {
	sandboxes       *services.SandboxService
	images          *services.ImageService
	resolver        RegistryResolver
	health          *services.SandboxHealthService
	auth            *services.AuthService
	guest           *services.GuestSessionService
	guestCookieName string
}

func NewSandboxHandler(
	sandboxes *services.SandboxService,
	images *services.ImageService,
	resolver RegistryResolver,
	health *services.SandboxHealthService,
	auth *services.AuthService,
	guest *services.GuestSessionService,
	guestCookieName string,
) *SandboxHandler {
	return &SandboxHandler{
		sandboxes:       sandboxes,
		images:          images,
		resolver:        resolver,
		health:          health,
		auth:            auth,
		guest:           guest,
		guestCookieName: guestCookieName,
	}
}

// List godoc
// @Summary      List all active sandboxes
// @Description  Returns all sandboxes that are currently active (admin view)
// @Tags         Sandboxes
// @Security     BearerAuth
// @Produce      json
// @Success      200 {array} dto.SandboxResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/sandboxes [get]
func (h *SandboxHandler) List(c echo.Context) error {
	sandboxes, err := h.sandboxes.ListActive()
	if err != nil {
		return responses.FromAppError(c, apperror.Internal("SANDBOX_LIST_FAILED", "Could not load sandboxes").WithCause(err))
	}
	auth := mw.MustAuth(c)
	h.enrichSandboxMetadata(sandboxes)
	slog.Debug("listed all sandboxes", logging.RequestFields(c, "component", "sandbox", "user_id", auth.UserID.String(), "count", len(sandboxes))...)
	return c.JSON(200, toSandboxResponses(sandboxes))
}

// ListMine godoc
// @Summary      List my sandboxes
// @Description  Returns sandboxes owned by the authenticated user
// @Tags         Sandboxes
// @Security     BearerAuth
// @Produce      json
// @Success      200 {array} dto.SandboxResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/me/sandboxes [get]
func (h *SandboxHandler) ListMine(c echo.Context) error {
	auth := mw.MustAuth(c)
	sandboxes, err := h.sandboxes.ListByUser(auth.UserID)
	if err != nil {
		return responses.FromAppError(c, apperror.Internal("SANDBOX_LIST_FAILED", "Could not load own sandboxes").WithCause(err))
	}
	h.enrichSandboxMetadata(sandboxes)
	slog.Debug("listed user sandboxes", logging.RequestFields(c, "component", "sandbox", "user_id", auth.UserID.String(), "count", len(sandboxes))...)
	return c.JSON(200, toSandboxResponses(sandboxes))
}

// ListGuest godoc
// @Summary      List guest sandboxes
// @Description  Returns sandboxes for the current guest session (cookie-based)
// @Tags         Public
// @Produce      json
// @Success      200 {array} dto.SandboxResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/public/sandboxes [get]
func (h *SandboxHandler) ListGuest(c echo.Context) error {
	guest := mw.MustGuest(c)
	sandboxes, err := h.sandboxes.ListByGuestSession(guest.SessionID)
	if err != nil {
		return responses.FromAppError(c, apperror.Internal("SANDBOX_LIST_FAILED", "Could not load guest sandboxes").WithCause(err))
	}
	h.enrichSandboxMetadata(sandboxes)
	slog.Debug("listed guest sandboxes", logging.RequestFields(c, "component", "sandbox", "guest_session_id", guest.SessionID.String(), "count", len(sandboxes))...)
	return c.JSON(200, toSandboxResponses(sandboxes))
}

// Get godoc
// @Summary      Get sandbox by ID
// @Description  Returns a single sandbox by its UUID
// @Tags         Sandboxes
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Sandbox ID" format(uuid)
// @Success      200 {object} dto.SandboxResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/sandboxes/{id} [get]
func (h *SandboxHandler) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid sandbox id"))
	}

	sandbox, err := h.sandboxes.FindByID(id)
	if err != nil {
		return responses.FromAppError(c, apperror.NotFound("SANDBOX_NOT_FOUND", "Sandbox not found").WithCause(err))
	}
	h.enrichSandbox(sandbox)
	slog.Debug("sandbox loaded", logging.RequestFields(c, "component", "sandbox", "sandbox_id", sandbox.ID.String(), "status", sandbox.Status)...)
	return c.JSON(200, toSandboxResponse(sandbox))
}

// Health godoc
// @Summary      Stream sandbox health
// @Description  SSE endpoint streaming sandbox readiness for active subscribers.
// @Tags         Sandboxes
// @Produce      text/event-stream
// @Param        id path string true "Sandbox ID" format(uuid)
// @Param        access_token query string false "Bearer token fallback for EventSource"
// @Success      200 {object} dto.SandboxHealthEvent "Last emitted SSE event payload"
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/sandboxes/{id}/health [get]
func (h *SandboxHandler) Health(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid sandbox id"))
	}

	sandbox, err := h.sandboxes.FindByID(id)
	if err != nil {
		return responses.FromAppError(c, apperror.NotFound("SANDBOX_NOT_FOUND", "Sandbox not found").WithCause(err))
	}

	if err := h.authorizeHealthAccess(c, sandbox); err != nil {
		return err
	}

	writeSSEHeaders(c)
	ch, cancel := h.health.Watch(sandbox)
	defer cancel()

	ctx := c.Request().Context()
	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-ch:
			if !ok {
				return nil
			}
			sendSSEEvent(c, event)
		}
	}
}

// CreatePublicDemo godoc
// @Summary      Create a public demo sandbox
// @Description  Spin up a new sandbox for a guest visitor
// @Tags         Public
// @Accept       json
// @Produce      json
// @Param        body body dto.CreateSandboxRequest true "Sandbox configuration"
// @Success      201 {object} models.Sandbox
// @Failure      400 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/public/demos [post]
func (h *SandboxHandler) CreatePublicDemo(c echo.Context) error {
	var input dto.CreateSandboxRequest
	if err := c.Bind(&input); err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid request body"))
	}

	imageID, err := uuid.Parse(input.ImageID)
	if err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid image id"))
	}

	guest := mw.MustGuest(c)
	slog.Debug("public demo creation requested", logging.RequestFields(c,
		"component", "sandbox",
		"guest_session_id", guest.SessionID.String(),
		"image_id", imageID.String(),
	)...)
	sandbox, err := h.sandboxes.Create(c.Request().Context(), services.CreateSandboxInput{
		ImageID:        imageID,
		GuestSessionID: &guest.SessionID,
		ClientIP:       c.RealIP(),
		Metadata:       input.Metadata,
	})
	if err != nil {
		return mapSandboxError(c, err)
	}
	h.health.StartMonitoring(sandbox.ID)

	slog.Info("public demo created", logging.RequestFields(c,
		"component", "sandbox",
		"guest_session_id", guest.SessionID.String(),
		"sandbox_id", sandbox.ID.String(),
		"image_id", sandbox.ImageID.String(),
		"expires_at", sandbox.ExpiresAt,
	)...)
	h.enrichSandbox(sandbox)
	return c.JSON(201, sandbox)
}

// CreatePrivateSandbox godoc
// @Summary      Create a private sandbox
// @Description  Spin up a new sandbox for an authenticated user
// @Tags         Sandboxes
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.CreateSandboxRequest true "Sandbox configuration"
// @Success      201 {object} models.Sandbox
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/sandboxes [post]
func (h *SandboxHandler) CreatePrivateSandbox(c echo.Context) error {
	var input dto.CreateSandboxRequest
	if err := c.Bind(&input); err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid request body"))
	}

	imageID, err := uuid.Parse(input.ImageID)
	if err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid image id"))
	}

	auth := mw.MustAuth(c)
	slog.Debug("private sandbox creation requested", logging.RequestFields(c,
		"component", "sandbox",
		"user_id", auth.UserID.String(),
		"image_id", imageID.String(),
		"ttl_minutes", input.TTLMinutes,
	)...)
	sandbox, err := h.sandboxes.Create(c.Request().Context(), services.CreateSandboxInput{
		ImageID:     imageID,
		UserID:      &auth.UserID,
		ClientIP:    c.RealIP(),
		TTLMinutes:  input.TTLMinutes,
		DisplayName: input.DisplayName,
		Metadata:    input.Metadata,
	})
	if err != nil {
		return mapSandboxError(c, err)
	}
	h.health.StartMonitoring(sandbox.ID)

	slog.Info("private sandbox created", logging.RequestFields(c,
		"component", "sandbox",
		"user_id", auth.UserID.String(),
		"sandbox_id", sandbox.ID.String(),
		"image_id", sandbox.ImageID.String(),
		"expires_at", sandbox.ExpiresAt,
	)...)
	h.enrichSandbox(sandbox)
	return c.JSON(201, sandbox)
}

// Update godoc
// @Summary      Update sandbox details
// @Description  Update display name of a sandbox owned by the authenticated user
// @Tags         Sandboxes
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Sandbox ID" format(uuid)
// @Param        body body dto.UpdateSandboxRequest true "Update payload"
// @Success      200 {object} models.Sandbox
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/sandboxes/{id} [patch]
func (h *SandboxHandler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid sandbox id"))
	}

	var input dto.UpdateSandboxRequest
	if err := c.Bind(&input); err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid request body"))
	}

	auth := mw.MustAuth(c)
	sandbox, err := h.sandboxes.UpdateSandbox(services.UpdateSandboxInput{
		SandboxID:   id,
		UserID:      &auth.UserID,
		DisplayName: input.DisplayName,
		ClientIP:    c.RealIP(),
	})
	if err != nil {
		return mapSandboxError(c, err)
	}

	slog.Info("sandbox updated", logging.RequestFields(c,
		"component", "sandbox",
		"user_id", auth.UserID.String(),
		"sandbox_id", id.String(),
	)...)
	h.enrichSandbox(sandbox)
	return c.JSON(200, sandbox)
}

// ExtendTTL godoc
// @Summary      Extend sandbox TTL
// @Description  Add additional time to a sandbox's expiration
// @Tags         Sandboxes
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Sandbox ID" format(uuid)
// @Param        body body dto.ExtendTTLRequest true "TTL extension"
// @Success      200 {object} models.Sandbox
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/sandboxes/{id}/ttl [patch]
func (h *SandboxHandler) ExtendTTL(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid sandbox id"))
	}

	var input dto.ExtendTTLRequest
	if err := c.Bind(&input); err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid request body"))
	}
	if input.TTLMinutes != nil && *input.TTLMinutes < 0 {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "ttlMinutes must be 0 (unlimited) or greater"))
	}

	auth := mw.MustAuth(c)
	slog.Debug("sandbox TTL extension requested", logging.RequestFields(c,
		"component", "sandbox",
		"user_id", auth.UserID.String(),
		"sandbox_id", id.String(),
		"ttl_minutes", input.TTLMinutes,
	)...)
	sandbox, err := h.sandboxes.ExtendTTL(id, input.TTLMinutes, c.RealIP(), &auth.UserID)
	if err != nil {
		return mapSandboxError(c, err)
	}

	slog.Info("sandbox TTL extended", logging.RequestFields(c,
		"component", "sandbox",
		"user_id", auth.UserID.String(),
		"sandbox_id", id.String(),
		"new_expires_at", sandbox.ExpiresAt,
	)...)
	return c.JSON(200, sandbox)
}

// Delete godoc
// @Summary      Delete a sandbox
// @Description  Stop and remove an authenticated user's sandbox
// @Tags         Sandboxes
// @Security     BearerAuth
// @Param        id path string true "Sandbox ID" format(uuid)
// @Success      204
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/sandboxes/{id} [delete]
func (h *SandboxHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid sandbox id"))
	}

	auth := mw.MustAuth(c)
	slog.Debug("sandbox deletion requested", logging.RequestFields(c, "component", "sandbox", "user_id", auth.UserID.String(), "sandbox_id", id.String())...)
	if err := h.sandboxes.Delete(c.Request().Context(), id, c.RealIP(), &auth.UserID); err != nil {
		return mapSandboxError(c, err)
	}

	slog.Info("sandbox deleted", logging.RequestFields(c, "component", "sandbox", "user_id", auth.UserID.String(), "sandbox_id", id.String())...)
	return c.NoContent(204)
}

// DeleteGuest godoc
// @Summary      Delete a guest sandbox
// @Description  Stop and remove a guest session's sandbox
// @Tags         Public
// @Param        id path string true "Sandbox ID" format(uuid)
// @Success      204
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/public/sandboxes/{id} [delete]
func (h *SandboxHandler) DeleteGuest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid sandbox id"))
	}

	guest := mw.MustGuest(c)
	slog.Debug("guest sandbox deletion requested", logging.RequestFields(c, "component", "sandbox", "guest_session_id", guest.SessionID.String(), "sandbox_id", id.String())...)
	if err := h.sandboxes.DeleteForGuest(c.Request().Context(), id, guest.SessionID, c.RealIP()); err != nil {
		return mapSandboxError(c, err)
	}

	slog.Info("guest sandbox deleted", logging.RequestFields(c, "component", "sandbox", "guest_session_id", guest.SessionID.String(), "sandbox_id", id.String())...)
	return c.NoContent(204)
}

// Snapshot godoc
// @Summary      Create a snapshot image from a sandbox
// @Description  Commit the current state of a running sandbox as a new Docker image
// @Tags         Sandboxes
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Sandbox ID" format(uuid)
// @Param        body body dto.CreateSnapshotRequest true "Snapshot metadata"
// @Success      201 {object} models.Image
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/sandboxes/{id}/snapshot [post]
func (h *SandboxHandler) Snapshot(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid sandbox id"))
	}

	var input dto.CreateSnapshotRequest
	if err := c.Bind(&input); err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid request body"))
	}

	auth := mw.MustAuth(c)
	slog.Debug("sandbox snapshot requested", logging.RequestFields(c,
		"component", "sandbox",
		"user_id", auth.UserID.String(),
		"sandbox_id", id.String(),
		"name", input.Name,
		"tag", input.Tag,
		"is_public", input.IsPublic,
	)...)
	metadataJSON, _ := json.Marshal(input.Metadata)
	image, err := h.sandboxes.CreateSnapshot(c.Request().Context(), services.CreateSnapshotInput{
		SandboxID:   id,
		Name:        input.Name,
		Tag:         input.Tag,
		Title:       input.Title,
		Description: input.Description,
		IsPublic:    input.IsPublic,
		ClientIP:    c.RealIP(),
		UserID:      &auth.UserID,
		Metadata:    metadataJSON,
	})
	if err != nil {
		return mapSandboxError(c, err)
	}

	slog.Info("sandbox snapshot created", logging.RequestFields(c,
		"component", "sandbox",
		"user_id", auth.UserID.String(),
		"sandbox_id", id.String(),
		"image_id", image.ID.String(),
		"image", image.FullName(),
	)...)
	return c.JSON(201, image)
}

func mapSandboxError(c echo.Context, err error) error {
	switch err {
	case services.ErrSandboxLimitReached:
		return responses.FromAppError(c, apperror.Conflict("SANDBOX_LIMIT_REACHED", "Maximum number of sandboxes reached"))
	case services.ErrSandboxNotFound:
		return responses.FromAppError(c, apperror.NotFound("SANDBOX_NOT_FOUND", "Sandbox not found"))
	case services.ErrSandboxAccessDenied:
		return responses.FromAppError(c, apperror.New(403, "SANDBOX_ACCESS_DENIED", "Sandbox does not belong to the current user"))
	default:
		return responses.FromAppError(c, apperror.Internal("SANDBOX_ERROR", "Sandbox operation failed").WithCause(err))
	}
}

func (h *SandboxHandler) authorizeHealthAccess(c echo.Context, sandbox *models.Sandbox) error {
	userToken := c.QueryParam("access_token")
	if userToken == "" {
		authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
		if token, ok := parseAuthorizationHeader(authHeader); ok {
			userToken = token
		}
	}

	if userToken != "" {
		user, _, err := h.auth.Authenticate(userToken)
		if err != nil {
			return responses.FromAppError(c, apperror.Unauthorized("Invalid or expired token"))
		}

		if sandbox.OwnerID != nil && *sandbox.OwnerID == user.ID {
			return nil
		}

		return responses.FromAppError(c, apperror.New(403, "SANDBOX_ACCESS_DENIED", "Sandbox access denied"))
	}

	if sandbox.GuestSessionID == nil {
		return responses.FromAppError(c, apperror.Unauthorized("Missing bearer token"))
	}

	cookie, err := c.Cookie(h.guestCookieName)
	if err != nil || cookie == nil || cookie.Value == "" {
		return responses.FromAppError(c, apperror.Unauthorized("Missing guest session"))
	}

	sessionID, _, err := h.guest.Validate(cookie.Value)
	if err != nil {
		return responses.FromAppError(c, apperror.Unauthorized("Invalid guest session"))
	}

	if *sandbox.GuestSessionID != sessionID {
		return responses.FromAppError(c, apperror.New(403, "SANDBOX_ACCESS_DENIED", "Sandbox access denied"))
	}

	return nil
}

func parseAuthorizationHeader(authHeader string) (string, bool) {
	parts := strings.Fields(authHeader)
	switch len(parts) {
	case 1:
		if parts[0] == "" || strings.EqualFold(parts[0], "Bearer") {
			return "", false
		}
		return parts[0], true
	case 2:
		if !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
			return "", false
		}
		return parts[1], true
	default:
		return "", false
	}
}
