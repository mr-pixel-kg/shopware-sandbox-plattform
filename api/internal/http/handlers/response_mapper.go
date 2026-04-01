package handlers

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"text/template"

	"github.com/manuel/shopware-testenv-platform/api/internal/config"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/manuel/shopware-testenv-platform/api/internal/registry"
)

func toUserSummary(user *models.User) *dto.UserSummary {
	if user == nil {
		return nil
	}

	return &dto.UserSummary{
		ID:    user.ID,
		Email: user.Email,
	}
}

func toImageResponse(image *models.Image) dto.ImageResponse {
	return dto.ImageResponse{
		ID:           image.ID,
		Name:         image.Name,
		Tag:          image.Tag,
		Title:        image.Title,
		Description:  image.Description,
		ThumbnailURL: image.ThumbnailURL,
		IsPublic:     image.IsPublic,
		Status:       image.Status,
		Error:        image.Error,
		Metadata:     image.Metadata,
		RegistryRef:  image.RegistryRef,
		Owner:        toUserSummary(image.Owner),
		CreatedAt:    image.CreatedAt,
		UpdatedAt:    image.UpdatedAt,
		DeletedAt:    image.DeletedAt,
	}
}

func toImageResponses(images []models.Image) []dto.ImageResponse {
	out := make([]dto.ImageResponse, len(images))
	for i := range images {
		out[i] = toImageResponse(&images[i])
	}
	return out
}

func toSandboxResponse(sandbox *models.Sandbox) dto.SandboxResponse {
	return dto.SandboxResponse{
		ID:             sandbox.ID,
		ImageID:        sandbox.ImageID,
		Owner:          toUserSummary(sandbox.Owner),
		GuestSessionID: sandbox.GuestSessionID,
		DisplayName:    sandbox.DisplayName,
		Status:         sandbox.Status,
		StateReason:    sandbox.StateReason,
		ContainerID:    sandbox.ContainerID,
		ContainerName:  sandbox.ContainerName,
		URL:            sandbox.URL,
		Port:           sandbox.Port,
		ClientIP:       sandbox.ClientIP,
		Metadata:       sandbox.Metadata,
		ExpiresAt:      sandbox.ExpiresAt,
		LastSeenAt:     sandbox.LastSeenAt,
		CreatedAt:      sandbox.CreatedAt,
		UpdatedAt:      sandbox.UpdatedAt,
		DeletedAt:      sandbox.DeletedAt,
	}
}

func toSandboxResponses(sandboxes []models.Sandbox) []dto.SandboxResponse {
	out := make([]dto.SandboxResponse, len(sandboxes))
	for i := range sandboxes {
		out[i] = toSandboxResponse(&sandboxes[i])
	}
	return out
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
