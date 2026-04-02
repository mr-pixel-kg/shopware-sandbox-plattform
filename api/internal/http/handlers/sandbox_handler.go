package handlers

import (
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/manuel/shopware-testenv-platform/api/internal/apperror"
	"github.com/manuel/shopware-testenv-platform/api/internal/config"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
	mw "github.com/manuel/shopware-testenv-platform/api/internal/http/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/responses"
	"github.com/manuel/shopware-testenv-platform/api/internal/logging"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
)

type SandboxHandler struct {
	sandboxes *services.SandboxService
	images    *services.ImageService
	resolver  RegistryResolver
	health    *services.SandboxHealthService
	auth      *services.AuthService
	sshCfg    config.SSHConfig
}

func NewSandboxHandler(
	sandboxes *services.SandboxService,
	images *services.ImageService,
	resolver RegistryResolver,
	health *services.SandboxHealthService,
	auth *services.AuthService,
	sshCfg config.SSHConfig,
) *SandboxHandler {
	return &SandboxHandler{
		sandboxes: sandboxes,
		images:    images,
		resolver:  resolver,
		health:    health,
		auth:      auth,
		sshCfg:    sshCfg,
	}
}

// List godoc
// @Summary      List sandboxes
// @Description  Admins see all sandboxes. Regular users see their own. Use ?owner=self for own, ?clientId=<uuid> for guest sandboxes.
// @Tags         Sandboxes
// @Security     BearerAuth
// @Produce      json
// @Param        owner query string false "Filter: 'self' for own sandboxes"
// @Param        clientId query string false "Filter by client ID" format(uuid)
// @Success      200 {array} dto.SandboxResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/sandboxes [get]
func (h *SandboxHandler) List(c echo.Context) error {
	auth := mw.MustAuth(c)
	user := c.Get("user").(*models.User)

	var (
		sandboxes []models.Sandbox
		err       error
	)

	if clientIDStr := c.QueryParam("clientId"); clientIDStr != "" {
		parsed, parseErr := uuid.Parse(clientIDStr)
		if parseErr != nil {
			return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid clientId"))
		}
		sandboxes, err = h.sandboxes.ListByClientID(parsed)
	} else if user.IsAdmin() && c.QueryParam("owner") != "self" {
		sandboxes, err = h.sandboxes.ListAll()
	} else {
		sandboxes, err = h.sandboxes.ListByUser(auth.UserID)
	}

	if err != nil {
		return responses.FromAppError(c, apperror.Internal("SANDBOX_LIST_FAILED", "Could not load sandboxes").WithCause(err))
	}
	slog.Debug("listed sandboxes", logging.RequestFields(c, "component", "sandbox", "user_id", auth.UserID.String(), "count", len(sandboxes))...)
	resp := h.enrichSandboxResponses(sandboxes)
	return c.JSON(200, resp)
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
	id, err := parseUUIDParam(c, "id", "VALIDATION_ERROR", "Invalid sandbox id")
	if err != nil {
		return responses.FromError(c, err)
	}

	sandbox, err := h.sandboxes.FindByID(id)
	if err != nil {
		return responses.FromAppError(c, apperror.NotFound("SANDBOX_NOT_FOUND", "Sandbox not found").WithCause(err))
	}
	slog.Debug("sandbox loaded", logging.RequestFields(c, "component", "sandbox", "sandbox_id", sandbox.ID.String(), "status", sandbox.Status)...)
	resp := h.enrichSandboxResponse(sandbox)
	return c.JSON(200, resp)
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
	id, err := parseUUIDParam(c, "id", "VALIDATION_ERROR", "Invalid sandbox id")
	if err != nil {
		return responses.FromError(c, err)
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

// Create godoc
// @Summary      Create a sandbox
// @Description  Spin up a new sandbox. Always requires auth. Stores X-Client-Id header on sandbox automatically.
// @Tags         Sandboxes
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.CreateSandboxRequest true "Sandbox configuration"
// @Success      201 {object} dto.SandboxResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/sandboxes [post]
func (h *SandboxHandler) Create(c echo.Context) error {
	var input dto.CreateSandboxRequest
	if err := bindAndValidate(c, &input); err != nil {
		return responses.FromError(c, err)
	}

	imageID, err := uuid.Parse(input.ImageID)
	if err != nil {
		return responses.FromError(c, validationError("Invalid image id"))
	}

	auth := mw.MustAuth(c)
	clientID := mw.ClientID(c)
	slog.Debug("sandbox creation requested", logging.RequestFields(c,
		"component", "sandbox",
		"user_id", auth.UserID.String(),
		"image_id", imageID.String(),
		"ttl_minutes", input.TTLMinutes,
	)...)
	sandbox, err := h.sandboxes.Create(c.Request().Context(), services.CreateSandboxInput{
		ImageID:     imageID,
		UserID:      &auth.UserID,
		ClientID:    clientID,
		ClientIP:    c.RealIP(),
		TTLMinutes:  input.TTLMinutes,
		DisplayName: input.DisplayName,
		Metadata:    input.Metadata,
		AuditActor:  newAuditActor(c, &auth.UserID),
	})
	if err != nil {
		return mapSandboxError(c, err)
	}
	h.health.StartMonitoring(sandbox.ID)

	slog.Info("sandbox created", logging.RequestFields(c,
		"component", "sandbox",
		"user_id", auth.UserID.String(),
		"sandbox_id", sandbox.ID.String(),
		"image_id", sandbox.ImageID.String(),
		"expires_at", sandbox.ExpiresAt,
	)...)
	resp := h.enrichSandboxResponse(sandbox)
	return c.JSON(201, resp)
}

// Update godoc
// @Summary      Update sandbox
// @Description  Update display name and/or extend TTL of a sandbox owned by the authenticated user
// @Tags         Sandboxes
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Sandbox ID" format(uuid)
// @Param        body body dto.UpdateSandboxRequest true "Update payload"
// @Success      200 {object} dto.SandboxResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/sandboxes/{id} [patch]
func (h *SandboxHandler) Update(c echo.Context) error {
	id, err := parseUUIDParam(c, "id", "VALIDATION_ERROR", "Invalid sandbox id")
	if err != nil {
		return responses.FromError(c, err)
	}

	var input dto.UpdateSandboxRequest
	if err := bindAndValidate(c, &input); err != nil {
		return responses.FromError(c, err)
	}

	auth := mw.MustAuth(c)
	sandbox, err := h.sandboxes.UpdateSandbox(services.UpdateSandboxInput{
		SandboxID:   id,
		UserID:      &auth.UserID,
		DisplayName: input.DisplayName,
		TTLMinutes:  input.TTLMinutes,
		ClientIP:    c.RealIP(),
		AuditActor:  newAuditActor(c, &auth.UserID),
	})
	if err != nil {
		return mapSandboxError(c, err)
	}

	slog.Info("sandbox updated", logging.RequestFields(c,
		"component", "sandbox",
		"user_id", auth.UserID.String(),
		"sandbox_id", id.String(),
	)...)
	resp := h.enrichSandboxResponse(sandbox)
	return c.JSON(200, resp)
}

// Delete godoc
// @Summary      Delete a sandbox
// @Description  Stop and remove a sandbox. Checks ownership by user ID or X-Client-Id header.
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
	id, err := parseUUIDParam(c, "id", "VALIDATION_ERROR", "Invalid sandbox id")
	if err != nil {
		return responses.FromError(c, err)
	}

	auth := mw.MustAuth(c)
	user := c.Get("user").(*models.User)

	sandbox, err := h.sandboxes.FindByID(id)
	if err != nil {
		return mapSandboxError(c, err)
	}

	if !user.IsAdmin() {
		ownsViaUser := sandbox.OwnerID != nil && *sandbox.OwnerID == auth.UserID
		clientID := mw.ClientID(c)
		ownsViaClient := sandbox.ClientID != nil && clientID != nil && *sandbox.ClientID == *clientID
		if !ownsViaUser && !ownsViaClient {
			return mapSandboxError(c, services.ErrSandboxAccessDenied)
		}
	}

	slog.Debug("sandbox deletion requested", logging.RequestFields(c, "component", "sandbox", "user_id", auth.UserID.String(), "sandbox_id", id.String())...)
	if err := h.sandboxes.Delete(c.Request().Context(), id, newAuditActor(c, &auth.UserID)); err != nil {
		return mapSandboxError(c, err)
	}

	slog.Info("sandbox deleted", logging.RequestFields(c, "component", "sandbox", "user_id", auth.UserID.String(), "sandbox_id", id.String())...)
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
// @Router       /api/sandboxes/{id}/snapshots [post]
func (h *SandboxHandler) Snapshot(c echo.Context) error {
	id, err := parseUUIDParam(c, "id", "VALIDATION_ERROR", "Invalid sandbox id")
	if err != nil {
		return responses.FromError(c, err)
	}

	var input dto.CreateSnapshotRequest
	if err := bindAndValidate(c, &input); err != nil {
		return responses.FromError(c, err)
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
		AuditActor:  newAuditActor(c, &auth.UserID),
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

// Stream godoc
// @Summary      Stream sandbox state
// @Description  SSE endpoint streaming real-time state updates for a single sandbox
// @Tags         Sandboxes
// @Security     BearerAuth
// @Produce      text/event-stream
// @Param        id path string true "Sandbox ID" format(uuid)
// @Success      200 {object} dto.SandboxStreamEvent
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Router       /api/sandboxes/{id}/stream [get]
func (h *SandboxHandler) Stream(c echo.Context) error {
	id, err := parseUUIDParam(c, "id", "VALIDATION_ERROR", "Invalid sandbox id")
	if err != nil {
		return responses.FromError(c, err)
	}

	sandbox, err := h.sandboxes.FindByID(id)
	if err != nil {
		return responses.FromAppError(c, apperror.NotFound("SANDBOX_NOT_FOUND", "Sandbox not found").WithCause(err))
	}

	if err := h.authorizeHealthAccess(c, sandbox); err != nil {
		return err
	}

	writeSSEHeaders(c)

	ctx := c.Request().Context()
	ch := h.health.WatchStream(ctx, sandbox)
	for event := range ch {
		sendSSEEvent(c, dto.SandboxStreamEvent{
			ID:          event.SandboxID.String(),
			Status:      event.Status,
			StateReason: event.StateReason,
		})
	}
	return nil
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
		if token, ok := mw.ParseAuthorizationHeader(authHeader); ok {
			userToken = token
		}
	}

	if userToken != "" {
		user, err := h.auth.Authenticate(userToken)
		if err != nil {
			return responses.FromAppError(c, apperror.Unauthorized("Invalid or expired token"))
		}

		if sandbox.OwnerID != nil && *sandbox.OwnerID == user.ID {
			return nil
		}

		return responses.FromAppError(c, apperror.New(403, "SANDBOX_ACCESS_DENIED", "Sandbox access denied"))
	}

	if sandbox.ClientID == nil {
		return responses.FromAppError(c, apperror.Unauthorized("Missing bearer token"))
	}

	clientID := mw.ClientID(c)
	if clientID == nil || *sandbox.ClientID != *clientID {
		return responses.FromAppError(c, apperror.New(403, "SANDBOX_ACCESS_DENIED", "Sandbox access denied"))
	}

	return nil
}

// CreateDemo godoc
// @Summary      Create a guest demo sandbox
// @Description  Create a sandbox for a guest visitor. Identified by X-Client-Id header. No auth required. Server applies default TTL.
// @Tags         Demos
// @Accept       json
// @Produce      json
// @Param        body body dto.CreateDemoRequest true "Demo sandbox configuration"
// @Success      201 {object} dto.SandboxResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/demos [post]
func (h *SandboxHandler) CreateDemo(c echo.Context) error {
	var input dto.CreateDemoRequest
	if err := bindAndValidate(c, &input); err != nil {
		return responses.FromError(c, err)
	}

	imageID, err := uuid.Parse(input.ImageID)
	if err != nil {
		return responses.FromError(c, validationError("Invalid image id"))
	}

	clientID := mw.ClientID(c)
	slog.Debug("demo creation requested", logging.RequestFields(c, "component", "sandbox", "image_id", imageID.String())...)
	sandbox, err := h.sandboxes.Create(c.Request().Context(), services.CreateSandboxInput{
		ImageID:    imageID,
		ClientID:   clientID,
		ClientIP:   c.RealIP(),
		AuditActor: newAuditActor(c, nil),
	})
	if err != nil {
		return mapSandboxError(c, err)
	}
	h.health.StartMonitoring(sandbox.ID)

	slog.Info("demo created", logging.RequestFields(c,
		"component", "sandbox",
		"sandbox_id", sandbox.ID.String(),
		"image_id", sandbox.ImageID.String(),
	)...)
	resp := h.enrichSandboxResponse(sandbox)
	return c.JSON(201, resp)
}

// ListDemos godoc
// @Summary      List guest demo sandboxes
// @Description  Returns sandboxes belonging to the given client ID. No auth required.
// @Tags         Demos
// @Produce      json
// @Param        clientId query string true "Client ID" format(uuid)
// @Success      200 {array} dto.SandboxResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/demos [get]
func (h *SandboxHandler) ListDemos(c echo.Context) error {
	clientIDStr := c.QueryParam("clientId")
	if clientIDStr == "" {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "clientId query parameter is required"))
	}
	parsed, err := uuid.Parse(clientIDStr)
	if err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid clientId"))
	}

	sandboxes, err := h.sandboxes.ListByClientID(parsed)
	if err != nil {
		return responses.FromAppError(c, apperror.Internal("SANDBOX_LIST_FAILED", "Could not load demo sandboxes").WithCause(err))
	}
	slog.Debug("listed demo sandboxes", logging.RequestFields(c, "component", "sandbox", "client_id", parsed.String(), "count", len(sandboxes))...)
	resp := h.enrichSandboxResponses(sandboxes)
	return c.JSON(200, resp)
}

// DeleteDemo godoc
// @Summary      Delete a guest demo sandbox
// @Description  Delete a sandbox owned by the X-Client-Id. No auth required.
// @Tags         Demos
// @Param        id path string true "Sandbox ID" format(uuid)
// @Success      204
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/demos/{id} [delete]
func (h *SandboxHandler) DeleteDemo(c echo.Context) error {
	id, err := parseUUIDParam(c, "id", "VALIDATION_ERROR", "Invalid sandbox id")
	if err != nil {
		return responses.FromError(c, err)
	}

	clientID := mw.ClientID(c)
	if clientID == nil {
		return responses.FromAppError(c, apperror.BadRequest("MISSING_CLIENT_ID", "X-Client-Id header is required"))
	}

	if err := h.sandboxes.DeleteForGuest(c.Request().Context(), id, *clientID, newAuditActor(c, nil)); err != nil {
		return mapSandboxError(c, err)
	}

	slog.Info("demo deleted", logging.RequestFields(c, "component", "sandbox", "client_id", clientID.String(), "sandbox_id", id.String())...)
	return c.NoContent(204)
}
