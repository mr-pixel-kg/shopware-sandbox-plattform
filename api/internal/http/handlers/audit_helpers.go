package handlers

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	auditcontracts "github.com/manuel/shopware-testenv-platform/api/internal/auditlog"
	mw "github.com/manuel/shopware-testenv-platform/api/internal/http/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
)

func newAuditLogInput(
	r *http.Request,
	userID *uuid.UUID,
	action auditcontracts.Action,
	resourceType *auditcontracts.ResourceType,
	resourceID *uuid.UUID,
	details map[string]any,
) services.AuditLogInput {
	return services.AuditLogInput{
		Actor: services.AuditActor{
			UserID:    userID,
			IPAddress: optionalString(strings.TrimSpace(r.RemoteAddr)),
			UserAgent: optionalString(strings.TrimSpace(r.UserAgent())),
			ClientID:  mw.ClientIDFromContext(r),
		},
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
	}
}

func newAuditActor(r *http.Request, userID *uuid.UUID) services.AuditActor {
	return newAuditLogInput(r, userID, "", nil, nil, nil).Actor
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
