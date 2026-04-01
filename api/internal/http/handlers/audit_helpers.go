package handlers

import (
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	auditcontracts "github.com/manuel/shopware-testenv-platform/api/internal/auditlog"
	mw "github.com/manuel/shopware-testenv-platform/api/internal/http/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
)

func newAuditLogInput(
	c echo.Context,
	userID *uuid.UUID,
	action auditcontracts.Action,
	resourceType *auditcontracts.ResourceType,
	resourceID *uuid.UUID,
	details map[string]any,
) services.AuditLogInput {
	return services.AuditLogInput{
		Actor: services.AuditActor{
			UserID:      userID,
			IPAddress:   optionalString(strings.TrimSpace(c.RealIP())),
			UserAgent:   optionalString(strings.TrimSpace(c.Request().UserAgent())),
			ClientToken: mw.ClientToken(c),
		},
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
	}
}

func newAuditActor(c echo.Context, userID *uuid.UUID) services.AuditActor {
	return newAuditLogInput(c, userID, "", nil, nil, nil).Actor
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
