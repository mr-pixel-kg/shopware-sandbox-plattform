package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/manuel/shopware-testenv-platform/api/internal/apperror"
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
// @Param        offset query int false "Offset for pagination" minimum(0) example(0)
// @Param        userId query string false "Filter by user ID" format(uuid)
// @Param        action query string false "Filter by action" example("sandbox.created")
// @Param        resourceType query string false "Filter by resource type" example("sandbox")
// @Param        resourceId query string false "Filter by resource ID" format(uuid)
// @Param        clientToken query string false "Filter by client token" format(uuid)
// @Param        from query string false "Filter from timestamp (inclusive)" format(date-time)
// @Param        to query string false "Filter to timestamp (inclusive)" format(date-time)
// @Success      200 {object} dto.AuditLogListResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/audit-logs [get]
func (h *AuditHandler) List(c echo.Context) error {
	filters, err := parseAuditLogListInput(c)
	if err != nil {
		return responses.FromError(c, err)
	}

	auth := middleware.MustAuth(c)
	result, err := h.audit.List(filters)
	if err != nil {
		return responses.Error(c, http.StatusInternalServerError, "AUDIT_LOG_LIST_FAILED", "Could not load audit logs")
	}
	slog.Debug("audit logs listed", logging.RequestFields(c, "component", "audit", "user_id", auth.UserID.String(), "limit", filters.Limit, "offset", filters.Offset, "count", len(result.Logs), "total", result.Total)...)

	response := make([]dto.AuditLogResponse, 0, len(result.Logs))
	for _, logEntry := range result.Logs {
		response = append(response, dto.AuditLogResponse{
			ID:           logEntry.ID,
			User:         toUserSummary(logEntry.User),
			Action:       logEntry.Action,
			IPAddress:    logEntry.IPAddress,
			UserAgent:    logEntry.UserAgent,
			ClientToken:  logEntry.ClientToken,
			ResourceType: logEntry.ResourceType,
			ResourceID:   logEntry.ResourceID,
			Details:      logEntry.Details,
			Timestamp:    logEntry.Timestamp,
		})
	}

	return c.JSON(http.StatusOK, dto.AuditLogListResponse{
		Data: response,
		Meta: dto.AuditLogListMeta{
			Pagination: dto.PaginationMeta{
				Limit:   result.Limit,
				Offset:  result.Offset,
				Count:   len(response),
				Total:   result.Total,
				HasMore: int64(result.Offset+len(response)) < result.Total,
			},
			Filters: dto.AuditLogListFilters{
				UserID:       filters.UserID,
				Action:       filters.Action,
				ResourceType: filters.ResourceType,
				ResourceID:   filters.ResourceID,
				ClientToken:  filters.ClientToken,
				From:         filters.From,
				To:           filters.To,
			},
		},
	})
}

func parseAuditLogListInput(c echo.Context) (services.AuditLogListInput, error) {
	input := services.AuditLogListInput{
		Limit:  50,
		Offset: 0,
	}

	if value := strings.TrimSpace(c.QueryParam("limit")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 || parsed > 500 {
			return input, apperror.BadRequest("VALIDATION_ERROR", "limit must be between 1 and 500")
		}
		input.Limit = parsed
	}

	if value := strings.TrimSpace(c.QueryParam("offset")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			return input, apperror.BadRequest("VALIDATION_ERROR", "offset must be 0 or greater")
		}
		input.Offset = parsed
	}

	userID, err := parseOptionalUUIDQuery(c.QueryParam("userId"), "Invalid userId")
	if err != nil {
		return input, err
	}
	resourceID, err := parseOptionalUUIDQuery(c.QueryParam("resourceId"), "Invalid resourceId")
	if err != nil {
		return input, err
	}
	clientToken, err := parseOptionalUUIDQuery(c.QueryParam("clientToken"), "Invalid clientToken")
	if err != nil {
		return input, err
	}
	from, err := parseOptionalTimeQuery(c.QueryParam("from"), "Invalid from timestamp")
	if err != nil {
		return input, err
	}
	to, err := parseOptionalTimeQuery(c.QueryParam("to"), "Invalid to timestamp")
	if err != nil {
		return input, err
	}
	if from != nil && to != nil && from.After(*to) {
		return input, apperror.BadRequest("VALIDATION_ERROR", "from must be before or equal to to")
	}

	input.UserID = userID
	input.ResourceID = resourceID
	input.ClientToken = clientToken
	input.From = from
	input.To = to

	if value := strings.TrimSpace(c.QueryParam("action")); value != "" {
		input.Action = &value
	}
	if value := strings.TrimSpace(c.QueryParam("resourceType")); value != "" {
		input.ResourceType = &value
	}

	return input, nil
}

func parseOptionalUUIDQuery(raw string, message string) (*uuid.UUID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	value, err := uuid.Parse(raw)
	if err != nil {
		return nil, apperror.BadRequest("VALIDATION_ERROR", message)
	}
	return &value, nil
}

func parseOptionalTimeQuery(raw string, message string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	value, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, apperror.BadRequest("VALIDATION_ERROR", message)
	}
	return &value, nil
}
