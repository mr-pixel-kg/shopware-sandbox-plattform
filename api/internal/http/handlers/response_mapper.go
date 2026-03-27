package handlers

import (
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
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
