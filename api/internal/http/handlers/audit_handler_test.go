package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAuditLogListInputDefaultsAndTrimsValues(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	resourceID := uuid.New()
	clientToken := uuid.New()
	req := httptest.NewRequest("GET", "/api/audit-logs?limit=25&offset=10&userId="+userID.String()+
		"&action=%20sandbox.created%20&resourceType=%20sandbox%20&resourceId="+resourceID.String()+
		"&clientToken="+clientToken.String()+"&from=2026-04-01T10:00:00Z&to=2026-04-01T12:00:00Z", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	input, err := parseAuditLogListInput(c)

	require.NoError(t, err)
	assert.Equal(t, 25, input.Limit)
	assert.Equal(t, 10, input.Offset)
	require.NotNil(t, input.UserID)
	assert.Equal(t, userID, *input.UserID)
	require.NotNil(t, input.Action)
	assert.Equal(t, "sandbox.created", *input.Action)
	require.NotNil(t, input.ResourceType)
	assert.Equal(t, "sandbox", *input.ResourceType)
	require.NotNil(t, input.ResourceID)
	assert.Equal(t, resourceID, *input.ResourceID)
	require.NotNil(t, input.ClientToken)
	assert.Equal(t, clientToken, *input.ClientToken)
	require.NotNil(t, input.From)
	assert.Equal(t, "2026-04-01T10:00:00Z", input.From.Format("2006-01-02T15:04:05Z07:00"))
	require.NotNil(t, input.To)
	assert.Equal(t, "2026-04-01T12:00:00Z", input.To.Format("2006-01-02T15:04:05Z07:00"))
}

func TestParseAuditLogListInputUsesDefaultPagination(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/api/audit-logs", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	input, err := parseAuditLogListInput(c)

	require.NoError(t, err)
	assert.Equal(t, 50, input.Limit)
	assert.Equal(t, 0, input.Offset)
	assert.Nil(t, input.Action)
	assert.Nil(t, input.ResourceType)
}

func TestParseAuditLogListInputRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		query  string
		detail string
	}{
		{name: "invalid limit", query: "limit=0", detail: "limit must be between 1 and 500"},
		{name: "invalid offset", query: "offset=-1", detail: "offset must be 0 or greater"},
		{name: "invalid user id", query: "userId=not-a-uuid", detail: "Invalid userId"},
		{name: "invalid resource id", query: "resourceId=not-a-uuid", detail: "Invalid resourceId"},
		{name: "invalid client token", query: "clientToken=not-a-uuid", detail: "Invalid clientToken"},
		{name: "invalid from", query: "from=nope", detail: "Invalid from timestamp"},
		{name: "invalid to", query: "to=nope", detail: "Invalid to timestamp"},
		{name: "from after to", query: "from=2026-04-01T12:00:00Z&to=2026-04-01T10:00:00Z", detail: "from must be before or equal to to"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest("GET", "/api/audit-logs?"+tt.query, nil)
			rec := httptest.NewRecorder()
			c := echo.New().NewContext(req, rec)

			_, err := parseAuditLogListInput(c)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.detail)
		})
	}
}

func TestParseAuditLogFacetInputDefaultsAndTrimsValues(t *testing.T) {
	t.Parallel()

	resourceID := uuid.New()
	clientToken := uuid.New()
	req := httptest.NewRequest("GET", "/api/audit-logs/facets?action=%20sandbox.created%20&resourceType=%20sandbox%20&resourceId="+resourceID.String()+
		"&clientToken="+clientToken.String()+"&from=2026-04-01T10:00:00Z&to=2026-04-01T12:00:00Z", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	input, err := parseAuditLogFacetInput(c)

	require.NoError(t, err)
	require.NotNil(t, input.Action)
	assert.Equal(t, "sandbox.created", *input.Action)
	require.NotNil(t, input.ResourceType)
	assert.Equal(t, "sandbox", *input.ResourceType)
	require.NotNil(t, input.ResourceID)
	assert.Equal(t, resourceID, *input.ResourceID)
	require.NotNil(t, input.ClientToken)
	assert.Equal(t, clientToken, *input.ClientToken)
}
