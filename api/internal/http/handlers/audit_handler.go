package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/google/uuid"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
)

type AuditHandler struct {
	Audit *services.AuditService
}

func (h AuditHandler) MountRoutes(s *fuego.Server) {
	logs := fuego.Group(s, "/audit-logs")
	fuego.Get(logs, "", h.list,
		option.Summary("List audit logs"),
		option.Description("Returns recent audit log entries with pagination and filtering"),
		option.Tags("AuditLogs"),
		option.QueryInt("limit", "Max entries (1-500, default 50)"),
		option.QueryInt("offset", "Offset for pagination"),
		option.Query("userId", "Filter by user ID"),
		option.Query("action", "Filter by action"),
		option.Query("resourceType", "Filter by resource type"),
		option.Query("resourceId", "Filter by resource ID"),
		option.Query("clientId", "Filter by client ID"),
		option.Query("from", "Filter from timestamp (inclusive, RFC3339)"),
		option.Query("to", "Filter to timestamp (inclusive, RFC3339)"),
	)
	fuego.Get(logs, "/facets", h.facets,
		option.Summary("List audit log facets"),
		option.Description("Returns available audit filter values for the current query window"),
		option.Tags("AuditLogs"),
	)
}

func (h AuditHandler) list(c fuego.ContextNoBody) (dto.AuditLogListResponse, error) {
	r := c.Request()
	filters, err := parseAuditLogListInput(r)
	if err != nil {
		return dto.AuditLogListResponse{}, err
	}

	result, err := h.Audit.List(filters)
	if err != nil {
		return dto.AuditLogListResponse{}, fuego.HTTPError{Status: http.StatusInternalServerError, Detail: "Could not load audit logs"}
	}

	response := make([]dto.AuditLogResponse, 0, len(result.Logs))
	for _, logEntry := range result.Logs {
		var userSummary *dto.UserSummary
		if logEntry.User != nil {
			userSummary = &dto.UserSummary{ID: logEntry.User.ID, Email: logEntry.User.Email}
		}
		response = append(response, dto.AuditLogResponse{
			ID:           logEntry.ID,
			User:         userSummary,
			Action:       logEntry.Action,
			IPAddress:    logEntry.IPAddress,
			UserAgent:    logEntry.UserAgent,
			ClientID:     logEntry.ClientID,
			ResourceType: logEntry.ResourceType,
			ResourceID:   logEntry.ResourceID,
			Details:      logEntry.Details,
			Timestamp:    logEntry.Timestamp,
		})
	}

	return dto.AuditLogListResponse{
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
				ClientID:     filters.ClientID,
				From:         filters.From,
				To:           filters.To,
			},
		},
	}, nil
}

func (h AuditHandler) facets(c fuego.ContextNoBody) (dto.AuditLogFacetsResponse, error) {
	r := c.Request()
	input, err := parseAuditLogFacetInput(r)
	if err != nil {
		return dto.AuditLogFacetsResponse{}, err
	}

	result, err := h.Audit.ListFacets(input)
	if err != nil {
		return dto.AuditLogFacetsResponse{}, fuego.HTTPError{Status: http.StatusInternalServerError, Detail: "Could not load audit log facets"}
	}

	users := make([]dto.UserSummary, 0, len(result.Users))
	for _, user := range result.Users {
		users = append(users, dto.UserSummary{ID: user.ID, Email: user.Email})
	}

	return dto.AuditLogFacetsResponse{Users: users, Actions: result.Actions}, nil
}

func parseAuditLogListInput(r *http.Request) (services.AuditLogListInput, error) {
	q := r.URL.Query()
	input := services.AuditLogListInput{Limit: 50, Offset: 0}

	if v := strings.TrimSpace(q.Get("limit")); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil || parsed <= 0 || parsed > 500 {
			return input, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "limit must be between 1 and 500"}
		}
		input.Limit = parsed
	}
	if v := strings.TrimSpace(q.Get("offset")); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil || parsed < 0 {
			return input, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "offset must be 0 or greater"}
		}
		input.Offset = parsed
	}

	var err error
	input.UserID, err = parseOptionalUUIDQuery(q.Get("userId"), "Invalid userId")
	if err != nil {
		return input, err
	}
	input.ResourceID, err = parseOptionalUUIDQuery(q.Get("resourceId"), "Invalid resourceId")
	if err != nil {
		return input, err
	}
	input.ClientID, err = parseOptionalUUIDQuery(q.Get("clientId"), "Invalid clientId")
	if err != nil {
		return input, err
	}
	input.From, err = parseOptionalTimeQuery(q.Get("from"), "Invalid from timestamp")
	if err != nil {
		return input, err
	}
	input.To, err = parseOptionalTimeQuery(q.Get("to"), "Invalid to timestamp")
	if err != nil {
		return input, err
	}
	if input.From != nil && input.To != nil && input.From.After(*input.To) {
		return input, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "from must be before or equal to to"}
	}

	if v := strings.TrimSpace(q.Get("action")); v != "" {
		input.Action = &v
	}
	if v := strings.TrimSpace(q.Get("resourceType")); v != "" {
		input.ResourceType = &v
	}

	return input, nil
}

func parseAuditLogFacetInput(r *http.Request) (services.AuditLogFacetInput, error) {
	q := r.URL.Query()

	resourceID, err := parseOptionalUUIDQuery(q.Get("resourceId"), "Invalid resourceId")
	if err != nil {
		return services.AuditLogFacetInput{}, err
	}
	clientID, err := parseOptionalUUIDQuery(q.Get("clientId"), "Invalid clientId")
	if err != nil {
		return services.AuditLogFacetInput{}, err
	}
	from, err := parseOptionalTimeQuery(q.Get("from"), "Invalid from timestamp")
	if err != nil {
		return services.AuditLogFacetInput{}, err
	}
	to, err := parseOptionalTimeQuery(q.Get("to"), "Invalid to timestamp")
	if err != nil {
		return services.AuditLogFacetInput{}, err
	}
	if from != nil && to != nil && from.After(*to) {
		return services.AuditLogFacetInput{}, fuego.HTTPError{Status: http.StatusBadRequest, Detail: "from must be before or equal to to"}
	}

	input := services.AuditLogFacetInput{
		ResourceID: resourceID,
		ClientID:   clientID,
		From:       from,
		To:         to,
	}
	if v := strings.TrimSpace(q.Get("action")); v != "" {
		input.Action = &v
	}
	if v := strings.TrimSpace(q.Get("resourceType")); v != "" {
		input.ResourceType = &v
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
		return nil, fuego.HTTPError{Status: http.StatusBadRequest, Detail: message}
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
		return nil, fuego.HTTPError{Status: http.StatusBadRequest, Detail: message}
	}
	return &value, nil
}
