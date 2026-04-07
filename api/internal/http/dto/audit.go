package dto

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type AuditLogResponse struct {
	ID           uuid.UUID      `json:"id" format:"uuid" example:"4d0dbf0d-1034-42ef-8b6d-7eb3ceef99cf"`
	User         *UserSummary   `json:"user"`
	Action       string         `json:"action" example:"sandbox.created"`
	IPAddress    *string        `json:"ipAddress,omitempty" example:"203.0.113.25"`
	UserAgent    *string        `json:"userAgent,omitempty" example:"Mozilla/5.0"`
	ClientID     *uuid.UUID     `json:"clientId,omitempty" format:"uuid" example:"4d0dbf0d-1034-42ef-8b6d-7eb3ceef99cf"`
	ResourceType *string        `json:"resourceType,omitempty" example:"sandbox"`
	ResourceID   *uuid.UUID     `json:"resourceId,omitempty" format:"uuid" example:"5cc66f6f-5c71-4be4-9f2d-639dc4b8c8c2"`
	Details      datatypes.JSON `json:"details" swaggertype:"object"`
	Timestamp    time.Time      `json:"timestamp" example:"2026-03-20T10:15:00Z"`
}

type AuditLogListFilters struct {
	UserID       *uuid.UUID `json:"userId,omitempty" format:"uuid"`
	Action       *string    `json:"action,omitempty" example:"sandbox.created"`
	ResourceType *string    `json:"resourceType,omitempty" example:"sandbox"`
	ResourceID   *uuid.UUID `json:"resourceId,omitempty" format:"uuid"`
	ClientID     *uuid.UUID `json:"clientId,omitempty" format:"uuid"`
	From         *time.Time `json:"from,omitempty" example:"2026-04-01T00:00:00Z"`
	To           *time.Time `json:"to,omitempty" example:"2026-04-02T00:00:00Z"`
}

type AuditLogListMeta struct {
	Pagination PaginationMeta      `json:"pagination"`
	Filters    AuditLogListFilters `json:"filters"`
}

type AuditLogListResponse struct {
	Data []AuditLogResponse `json:"data"`
	Meta AuditLogListMeta   `json:"meta"`
}

type AuditLogFacetsResponse struct {
	Users   []UserSummary `json:"users"`
	Actions []string      `json:"actions"`
}
