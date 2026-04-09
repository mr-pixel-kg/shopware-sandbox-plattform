package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/mr-pixel-kg/shopshredder/api/internal/http/dto"
	"github.com/mr-pixel-kg/shopshredder/api/internal/services"
)

type AuditHandler struct {
	Audit *services.AuditService
}

func (h AuditHandler) List(c fuego.ContextNoBody) (dto.AuditLogListResponse, error) {
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
			userSummary = &dto.UserSummary{ID: logEntry.User.ID, Email: logEntry.User.Email, AvatarURL: dto.GravatarURL(logEntry.User.Email, 80)}
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
			Pagination: buildPaginationMeta(len(response), result.Limit, result.Offset, result.Total),
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

func (h AuditHandler) Facets(c fuego.ContextNoBody) (dto.AuditLogFacetsResponse, error) {
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
		users = append(users, dto.UserSummary{ID: user.ID, Email: user.Email, AvatarURL: dto.GravatarURL(user.Email, 80)})
	}

	return dto.AuditLogFacetsResponse{Users: users, Actions: result.Actions}, nil
}

func parseAuditLogListInput(r *http.Request) (services.AuditLogListInput, error) {
	q := r.URL.Query()

	limit, offset, err := parsePaginationParams(r)
	if err != nil {
		return services.AuditLogListInput{}, err
	}
	input := services.AuditLogListInput{Limit: limit, Offset: offset}

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
