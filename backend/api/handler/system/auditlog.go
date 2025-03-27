package system

import (
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
)

// Auth godoc
// @Summary Endpoint to the last 300 audit log entries.
// @Description Get the last 300 audit log entries.
// @Tags System Management
// @Accept */*
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Security BasicAuth
// @Router /api/auditlog [get]
func (h *SystemHandler) SystemGetLastAuditLog(c echo.Context) error {

	res, err := h.AuditLogService.AuditLogRepository.GetLast(300)
	if err != nil {
		slog.Error("Failed to retrieve audit log entries", err)
		return c.JSON(http.StatusInternalServerError, "Failed to retrieve audit log entries")
	}
	return c.JSON(http.StatusOK, res)

}
