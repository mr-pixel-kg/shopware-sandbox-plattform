package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"text/template"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/mr-pixel-kg/shopshredder/api/internal/config"
	"github.com/mr-pixel-kg/shopshredder/api/internal/http/dto"
	"github.com/mr-pixel-kg/shopshredder/api/internal/http/errs"
	mw "github.com/mr-pixel-kg/shopshredder/api/internal/http/middleware"
	"github.com/mr-pixel-kg/shopshredder/api/internal/models"
	"github.com/mr-pixel-kg/shopshredder/api/internal/registry"
	"github.com/mr-pixel-kg/shopshredder/api/internal/services"
)

type SandboxHandler struct {
	Sandboxes *services.SandboxService
	Health    *services.SandboxHealthService
	Auth      *services.AuthService
}

func (h SandboxHandler) List(c fuego.ContextNoBody) (dto.SandboxListResponse, error) {
	r := c.Request()
	auth := mw.AuthFromContext(r)
	user := mw.UserFromContext(r)

	limit, offset, err := parsePaginationParams(r)
	if err != nil {
		return dto.SandboxListResponse{}, err
	}

	input := services.SandboxListInput{Limit: limit, Offset: offset}

	if auth == nil {
		clientID := mw.ClientIDFromContext(r)
		if clientID == nil {
			return dto.SandboxListResponse{}, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "Client identification required"}
		}
		input.ClientID = clientID
	} else {
		isAdminAll := user != nil && user.IsAdmin() && r.URL.Query().Get("scope") == "all"
		if !isAdminAll {
			input.UserID = &auth.UserID
			input.ClientID = mw.ClientIDFromContext(r)
		}
	}

	result, err := h.Sandboxes.ListPaginated(input)
	if err != nil {
		return dto.SandboxListResponse{}, fuego.HTTPError{Status: http.StatusInternalServerError, Detail: "Could not load sandboxes"}
	}

	sshCfg := h.Sandboxes.SSHConfig()
	out := make([]dto.SandboxResponse, len(result.Sandboxes))
	for i, sb := range result.Sandboxes {
		out[i] = sandboxToResponse(&result.Sandboxes[i], sshCfg, h.Sandboxes.ResolveSSHEntry(sb.ImageID))
	}
	return dto.SandboxListResponse{
		Data: out,
		Meta: dto.PaginatedMeta{
			Pagination: buildPaginationMeta(len(out), result.Limit, result.Offset, result.Total),
		},
	}, nil
}

func (h SandboxHandler) Get(c fuego.ContextNoBody) (dto.SandboxResponse, error) {
	id, err := parsePathUUID(c, "id")
	if err != nil {
		return dto.SandboxResponse{}, err
	}

	sandbox, err := h.Sandboxes.FindByID(id)
	if err != nil {
		return dto.SandboxResponse{}, fuego.HTTPError{Status: http.StatusNotFound, Detail: "Sandbox not found"}
	}

	sshCfg := h.Sandboxes.SSHConfig()
	return sandboxToResponse(sandbox, sshCfg, h.Sandboxes.ResolveSSHEntry(sandbox.ImageID)), nil
}

func (h SandboxHandler) Create(c fuego.ContextWithBody[dto.CreateSandboxRequest]) (dto.SandboxResponse, error) {
	body, err := c.Body()
	if err != nil {
		return dto.SandboxResponse{}, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "Invalid request body"}
	}

	imageID, err := uuid.Parse(body.ImageID)
	if err != nil {
		return dto.SandboxResponse{}, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "Invalid image id"}
	}

	r := c.Request()
	auth := mw.AuthFromContext(r)
	clientID := mw.ClientIDFromContext(r)

	input := services.CreateSandboxInput{
		ImageID:     imageID,
		ClientID:    clientID,
		ClientIP:    extractIP(r),
		TTLMinutes:  body.TTLMinutes,
		DisplayName: body.DisplayName,
		Metadata:    body.Metadata,
	}

	if auth != nil {
		input.UserID = &auth.UserID
		input.AuditActor = newAuditActor(r, &auth.UserID)
		slog.Debug("sandbox creation requested", "component", "sandbox", "user_id", auth.UserID, "image_id", imageID, "ttl_minutes", body.TTLMinutes)
	} else {
		if clientID == nil {
			return dto.SandboxResponse{}, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "Client identification required for guests"}
		}
		input.AuditActor = newAuditActor(r, nil)
		slog.Debug("guest sandbox creation requested", "component", "sandbox", "client_id", clientID, "image_id", imageID)
	}

	sandbox, err := h.Sandboxes.Create(r.Context(), input)
	if err != nil {
		return dto.SandboxResponse{}, mapSandboxError(err)
	}
	h.Health.StartMonitoring(sandbox.ID)

	slog.Info("sandbox created", "component", "sandbox", "sandbox_id", sandbox.ID, "image_id", sandbox.ImageID, "expires_at", sandbox.ExpiresAt)
	sshCfg := h.Sandboxes.SSHConfig()
	return sandboxToResponse(sandbox, sshCfg, h.Sandboxes.ResolveSSHEntry(sandbox.ImageID)), nil
}

func (h SandboxHandler) Update(c fuego.ContextWithBody[dto.UpdateSandboxRequest]) (dto.SandboxResponse, error) {
	id, err := parsePathUUID(c, "id")
	if err != nil {
		return dto.SandboxResponse{}, err
	}

	body, err := c.Body()
	if err != nil {
		return dto.SandboxResponse{}, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "Invalid request body"}
	}

	r := c.Request()
	auth := mw.MustAuth(r)
	sandbox, err := h.Sandboxes.UpdateSandbox(services.UpdateSandboxInput{
		SandboxID:   id,
		UserID:      &auth.UserID,
		DisplayName: body.DisplayName,
		TTLMinutes:  body.TTLMinutes,
		ClientIP:    extractIP(r),
		AuditActor:  newAuditActor(r, &auth.UserID),
	})
	if err != nil {
		return dto.SandboxResponse{}, mapSandboxError(err)
	}

	slog.Info("sandbox updated", "component", "sandbox", "user_id", auth.UserID, "sandbox_id", id)
	sshCfg := h.Sandboxes.SSHConfig()
	return sandboxToResponse(sandbox, sshCfg, h.Sandboxes.ResolveSSHEntry(sandbox.ImageID)), nil
}

func (h SandboxHandler) Delete(c fuego.ContextNoBody) (any, error) {
	id, err := parsePathUUID(c, "id")
	if err != nil {
		return nil, err
	}

	r := c.Request()
	auth := mw.AuthFromContext(r)

	if auth == nil {
		clientID := mw.ClientIDFromContext(r)
		if clientID == nil {
			return nil, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "Client identification required for guests"}
		}
		if err := h.Sandboxes.DeleteForGuest(r.Context(), id, *clientID, newAuditActor(r, nil)); err != nil {
			return nil, mapSandboxError(err)
		}
		slog.Info("guest sandbox deleted", "component", "sandbox", "client_id", clientID, "sandbox_id", id)
		return nil, nil
	}

	user := mw.UserFromContext(r)
	sandbox, err := h.Sandboxes.FindByID(id)
	if err != nil {
		return nil, mapSandboxError(err)
	}

	if user == nil || !user.IsAdmin() {
		ownsViaUser := sandbox.OwnerID != nil && *sandbox.OwnerID == auth.UserID
		clientID := mw.ClientIDFromContext(r)
		ownsViaClient := sandbox.ClientID != nil && clientID != nil && *sandbox.ClientID == *clientID
		if !ownsViaUser && !ownsViaClient {
			return nil, mapSandboxError(services.ErrSandboxAccessDenied)
		}
	}

	slog.Debug("sandbox deletion requested", "component", "sandbox", "user_id", auth.UserID, "sandbox_id", id)
	if err := h.Sandboxes.Delete(r.Context(), id, newAuditActor(r, &auth.UserID)); err != nil {
		return nil, mapSandboxError(err)
	}

	slog.Info("sandbox deleted", "component", "sandbox", "user_id", auth.UserID, "sandbox_id", id)
	return nil, nil
}

func (h SandboxHandler) Snapshot(c fuego.ContextWithBody[dto.CreateSnapshotRequest]) (dto.ImageResponse, error) {
	id, err := parsePathUUID(c, "id")
	if err != nil {
		return dto.ImageResponse{}, err
	}

	body, err := c.Body()
	if err != nil {
		return dto.ImageResponse{}, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "Invalid request body"}
	}

	r := c.Request()
	auth := mw.MustAuth(r)
	slog.Debug("sandbox snapshot requested", "component", "sandbox", "user_id", auth.UserID, "sandbox_id", id, "name", body.Name, "tag", body.Tag)

	metadataJSON, _ := json.Marshal(body.Metadata)
	image, err := h.Sandboxes.CreateSnapshot(r.Context(), services.CreateSnapshotInput{
		SandboxID:   id,
		Name:        body.Name,
		Tag:         body.Tag,
		Title:       body.Title,
		Description: body.Description,
		IsPublic:    body.IsPublic,
		ClientIP:    extractIP(r),
		UserID:      &auth.UserID,
		Metadata:    metadataJSON,
		AuditActor:  newAuditActor(r, &auth.UserID),
	})
	if err != nil {
		return dto.ImageResponse{}, mapSandboxError(err)
	}

	slog.Info("sandbox snapshot created", "component", "sandbox", "user_id", auth.UserID, "sandbox_id", id, "image_id", image.ID, "image", image.FullName())
	return imageToResponse(image), nil
}

func (h SandboxHandler) HealthSSE(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		errs.Write(w, http.StatusBadRequest, "Invalid sandbox id")
		return
	}

	sandbox, err := h.Sandboxes.FindByID(id)
	if err != nil {
		errs.Write(w, http.StatusNotFound, "Sandbox not found")
		return
	}

	if err := h.authorizeHealthAccess(w, r, sandbox); err != nil {
		return
	}

	writeSSEHeaders(w)
	ch, cancel := h.Health.Watch(sandbox)
	defer cancel()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			sendSSEEvent(w, dto.SandboxHealthEvent{
				SandboxID:     event.SandboxID,
				Status:        event.Status,
				Ready:         event.Ready,
				URL:           event.URL,
				HTTPStatus:    event.HTTPStatus,
				LatencyMs:     event.LatencyMs,
				FailureReason: event.FailureReason,
				Message:       event.Message,
				CheckedAt:     event.CheckedAt,
			})
		}
	}
}

func (h SandboxHandler) StreamSSE(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		errs.Write(w, http.StatusBadRequest, "Invalid sandbox id")
		return
	}

	sandbox, err := h.Sandboxes.FindByID(id)
	if err != nil {
		errs.Write(w, http.StatusNotFound, "Sandbox not found")
		return
	}

	if err := h.authorizeHealthAccess(w, r, sandbox); err != nil {
		return
	}

	writeSSEHeaders(w)
	ctx := r.Context()
	ch := h.Health.WatchStream(ctx, sandbox)
	for event := range ch {
		sendSSEEvent(w, dto.SandboxStreamEvent{
			ID:          event.SandboxID,
			Status:      event.Status,
			StateReason: event.StateReason,
		})
	}
}

func (h SandboxHandler) authorizeHealthAccess(w http.ResponseWriter, r *http.Request, sandbox *models.Sandbox) error {
	userToken := r.URL.Query().Get("access_token")
	if userToken == "" {
		if token, ok := mw.ParseAuthorizationHeader(r.Header.Get("Authorization")); ok {
			userToken = token
		}
	}

	if userToken != "" {
		user, err := h.Auth.Authenticate(userToken)
		if err != nil {
			errs.Write(w, http.StatusUnauthorized, "Invalid or expired token")
			return err
		}
		if user.IsAdmin() {
			return nil
		}
		if sandbox.OwnerID != nil && *sandbox.OwnerID == user.ID {
			return nil
		}
		errs.Write(w, http.StatusForbidden, "Sandbox access denied")
		return fmt.Errorf("forbidden")
	}

	if sandbox.ClientID == nil {
		errs.Write(w, http.StatusUnauthorized, "Missing bearer token")
		return fmt.Errorf("unauthorized")
	}

	clientID := mw.ClientIDFromContext(r)
	if clientID == nil || *sandbox.ClientID != *clientID {
		errs.Write(w, http.StatusForbidden, "Sandbox access denied")
		return fmt.Errorf("forbidden")
	}

	return nil
}

func mapSandboxError(err error) error {
	switch {
	case errors.Is(err, services.ErrSandboxLimitReached):
		return fuego.HTTPError{Status: http.StatusConflict, Detail: "Maximum number of sandboxes reached"}
	case errors.Is(err, services.ErrSandboxNotFound):
		return fuego.HTTPError{Status: http.StatusNotFound, Detail: "Sandbox not found"}
	case errors.Is(err, services.ErrSandboxAccessDenied):
		return fuego.HTTPError{Status: http.StatusForbidden, Detail: "Sandbox does not belong to the current user"}
	default:
		return fuego.HTTPError{Status: http.StatusInternalServerError, Detail: "Sandbox operation failed"}
	}
}

func sandboxToResponse(sb *models.Sandbox, sshCfg config.SSHConfig, sshEntry *registry.SSHEntry) dto.SandboxResponse {
	var owner *dto.UserSummary
	if sb.Owner != nil {
		owner = &dto.UserSummary{ID: sb.Owner.ID, Email: sb.Owner.Email, AvatarURL: dto.GravatarURL(sb.Owner.Email, 80)}
	}
	var ssh *dto.SSHConnectionInfo
	if sshCfg.Enabled {
		ssh = buildSSHInfo(sb, sshCfg, sshEntry)
	}
	return dto.SandboxResponse{
		ID: sb.ID, ImageID: sb.ImageID, Owner: owner,
		ClientID: sb.ClientID, DisplayName: sb.DisplayName,
		Status: sb.Status, StateReason: sb.StateReason,
		ContainerID: sb.ContainerID, ContainerName: sb.ContainerName,
		URL: sb.URL, Port: sb.Port, SSH: ssh, ClientIP: sb.ClientIP,
		Metadata:  sb.Metadata,
		ExpiresAt: sb.ExpiresAt, LastSeenAt: sb.LastSeenAt,
		CreatedAt: sb.CreatedAt, UpdatedAt: sb.UpdatedAt,
	}
}

func buildSSHInfo(sandbox *models.Sandbox, sshCfg config.SSHConfig, sshEntry *registry.SSHEntry) *dto.SSHConnectionInfo {
	if !sshCfg.Enabled || sshEntry == nil || !sandbox.Status.IsActive() {
		return nil
	}
	host := resolveSSHHost(sshCfg.Host, sandbox)
	username := sshEntry.Username + "+" + sandbox.ID.String()
	return &dto.SSHConnectionInfo{
		Host:     host,
		Port:     sshCfg.Port,
		Username: username,
		Password: sshEntry.Password,
		Command:  fmt.Sprintf("ssh %s@%s -p %d", username, host, sshCfg.Port),
	}
}

func resolveSSHHost(hostTemplate string, sandbox *models.Sandbox) string {
	if hostTemplate == "" {
		return extractHostname(sandbox.URL)
	}
	if !strings.Contains(hostTemplate, "{{") {
		return hostTemplate
	}
	tmpl, err := template.New("ssh_host").Parse(hostTemplate)
	if err != nil {
		return extractHostname(sandbox.URL)
	}
	shortID := sandbox.ContainerID
	if len(shortID) > 12 {
		shortID = shortID[:12]
	}
	data := struct {
		ContainerName    string
		ContainerID      string
		ContainerShortID string
		SandboxID        string
	}{
		ContainerName:    sandbox.ContainerName,
		ContainerID:      sandbox.ContainerID,
		ContainerShortID: shortID,
		SandboxID:        sandbox.ID.String(),
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return extractHostname(sandbox.URL)
	}
	return buf.String()
}

func extractHostname(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "localhost"
	}
	return u.Hostname()
}
