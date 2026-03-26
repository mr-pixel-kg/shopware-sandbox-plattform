package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/middleware"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/responses"
	"github.com/manuel/shopware-testenv-platform/api/internal/logging"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
)

type AuditHandler struct {
	audit *services.AuditService
}

func NewAuditHandler(audit *services.AuditService) *AuditHandler {
	return &AuditHandler{audit: audit}
}

// List godoc
// @Summary      List audit logs
// @Description  Returns recent audit log entries, optionally limited
// @Tags         AuditLogs
// @Security     BearerAuth
// @Produce      json
// @Param        limit query int false "Max entries (1-200, default 50)" minimum(1) maximum(200) example(50)
// @Success      200 {array} dto.AuditLogResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/audit-logs [get]
func (h *AuditHandler) List(c echo.Context) error {
	limit := 50
	if value := c.QueryParam("limit"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	auth := middleware.MustAuth(c)
	logs, err := h.audit.List(limit)
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "AUDIT_LOG_LIST_FAILED", "Could not load audit logs")
	}
	slog.Debug("audit logs listed", logging.RequestFields(c, "component", "audit", "user_id", auth.UserID.String(), "limit", limit, "count", len(logs))...)

	response := make([]dto.AuditLogResponse, 0, len(logs))
	for _, logEntry := range logs {
		response = append(response, dto.AuditLogResponse{
			ID:        logEntry.ID,
			User:      toUserSummary(logEntry.User),
			Action:    logEntry.Action,
			IPAddress: logEntry.IPAddress,
			Details:   logEntry.Details,
			CreatedAt: logEntry.CreatedAt,
		})
	}

	return c.JSON(http.StatusOK, response)
}
