package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"gorm.io/datatypes"
)

type ErrorDetail struct {
	Code    string `json:"code" example:"VALIDATION_ERROR"`
	Message string `json:"message" example:"Invalid request body"`
	Details any    `json:"details,omitempty" swaggertype:"object"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type AuthLoginResponse struct {
	Token string      `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
	User  models.User `json:"user"`
}

type UserSummary struct {
	ID    uuid.UUID `json:"id" format:"uuid" example:"5cc66f6f-5c71-4be4-9f2d-639dc4b8c8c2"`
	Email string    `json:"email" example:"jane.doe@example.com"`
}

type AuditLogResponse struct {
	ID        uuid.UUID      `json:"id" format:"uuid" example:"4d0dbf0d-1034-42ef-8b6d-7eb3ceef99cf"`
	User      *UserSummary   `json:"user"`
	Action    string         `json:"action" example:"sandbox.created"`
	IPAddress string         `json:"ipAddress" example:"203.0.113.25"`
	Details   datatypes.JSON `json:"details" swaggertype:"object"`
	CreatedAt time.Time      `json:"createdAt" example:"2026-03-20T10:15:00Z"`
}

type ImageResponse struct {
	ID           uuid.UUID      `json:"id" format:"uuid" example:"8ae13ed9-cfb1-4941-a248-bc74b9fb6a24"`
	Name         string         `json:"name" example:"dockware/dev"`
	Tag          string         `json:"tag" example:"6.6.9.0"`
	Title        *string        `json:"title,omitempty" example:"Shopware Demo Image"`
	Description  *string        `json:"description,omitempty" example:"Prepared image for sales demos and internal QA."`
	ThumbnailURL *string        `json:"thumbnailUrl,omitempty" example:"https://cdn.example.com/images/shopware-demo.png"`
	IsPublic     bool           `json:"isPublic" example:"true"`
	Status       string         `json:"status" example:"ready"`
	Error        *string        `json:"error,omitempty" example:"pull access denied"`
	Metadata     datatypes.JSON `json:"metadata,omitempty" swaggertype:"string"`
	RegistryRef  *string        `json:"registryRef,omitempty" example:"dockware/dev"`
	Owner        *UserSummary   `json:"owner,omitempty"`
	CreatedAt    time.Time      `json:"createdAt" example:"2026-03-20T10:15:00Z"`
	UpdatedAt    time.Time      `json:"updatedAt" example:"2026-03-20T10:20:00Z"`
	DeletedAt    *time.Time     `json:"deletedAt,omitempty"`
}

type SandboxResponse struct {
	ID             uuid.UUID            `json:"id" format:"uuid" example:"0b443c82-d8a3-49a7-b59a-26ce327c7341"`
	ImageID        uuid.UUID            `json:"imageId" format:"uuid" example:"8ae13ed9-cfb1-4941-a248-bc74b9fb6a24"`
	Owner          *UserSummary         `json:"owner,omitempty"`
	GuestSessionID *uuid.UUID           `json:"guestSessionId,omitempty" format:"uuid" example:"db7fcb92-c2ff-4c20-9ac2-5a2504ab6326"`
	DisplayName    string               `json:"displayName" example:"My Test Shop"`
	Status         models.SandboxStatus `json:"status" enums:"starting,running,stopped,expired,deleted,failed" example:"running"`
	ContainerID    string               `json:"containerId" example:"1a2b3c4d5e6f7g8h9i0j"`
	ContainerName  string               `json:"containerName" example:"sandbox-0b443c82"`
	URL            string               `json:"url" example:"https://sandbox-0b443c82.demo.shopshredder.de"`
	Port           *int                 `json:"port,omitempty" example:"8080"`
	ClientIP       string               `json:"clientIp" example:"203.0.113.25"`
	Metadata       datatypes.JSON       `json:"metadata,omitempty" swaggertype:"string"`
	ExpiresAt      *time.Time           `json:"expiresAt,omitempty" example:"2026-03-20T12:00:00Z"`
	LastSeenAt     *time.Time           `json:"lastSeenAt,omitempty" example:"2026-03-20T10:45:00Z"`
	CreatedAt      time.Time            `json:"createdAt" example:"2026-03-20T10:15:00Z"`
	UpdatedAt      time.Time            `json:"updatedAt" example:"2026-03-20T10:20:00Z"`
	DeletedAt      *time.Time           `json:"deletedAt,omitempty"`
}

type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

type ImagePullProgressEvent struct {
	Percent int    `json:"percent" example:"75"`
	Status  string `json:"status" example:"pulling"`
}

type PendingPullResponse struct {
	ID      string  `json:"id" example:"8ae13ed9-cfb1-4941-a248-bc74b9fb6a24"`
	Name    string  `json:"name" example:"dockware/shopware"`
	Tag     string  `json:"tag" example:"6.6.9.0"`
	Title   *string `json:"title,omitempty" example:"Shopware Demo Image"`
	Percent int     `json:"percent" example:"42"`
	Status  string  `json:"status" example:"pulling"`
}

type SandboxHealthEvent struct {
	SandboxID     string `json:"sandboxId" example:"0b443c82-d8a3-49a7-b59a-26ce327c7341"`
	Status        string `json:"status" example:"probing"`
	Ready         bool   `json:"ready" example:"false"`
	URL           string `json:"url" example:"https://sandbox-0b443c82.demo.shopshredder.de"`
	HTTPStatus    int    `json:"httpStatus,omitempty" example:"200"`
	LatencyMs     int64  `json:"latencyMs,omitempty" example:"412"`
	FailureReason string `json:"failureReason,omitempty" example:"tls_handshake_failed"`
	Message       string `json:"message,omitempty" example:"Sandbox URL is reachable"`
	CheckedAt     string `json:"checkedAt" example:"2026-03-23T10:15:07Z"`
}
