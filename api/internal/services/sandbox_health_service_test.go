package services

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSandboxHealthServiceProbeSandboxReady(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	service := NewSandboxHealthService(nil, nil, nil)
	event := service.probeSandbox(&models.Sandbox{
		ID:     uuid.New(),
		Status: models.SandboxStatusStarting,
		URL:    server.URL,
	}, false)

	assert.True(t, event.Ready)
	assert.Equal(t, "ready", event.Status)
	assert.Equal(t, http.StatusOK, event.HTTPStatus)
	assert.Equal(t, "Sandbox URL is reachable", event.Message)
}

func TestSandboxHealthServiceProbeSandboxPending(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	service := NewSandboxHealthService(nil, nil, nil)
	event := service.probeSandbox(&models.Sandbox{
		ID:     uuid.New(),
		Status: models.SandboxStatusStarting,
		URL:    server.URL,
	}, false)

	assert.False(t, event.Ready)
	assert.Equal(t, "probing", event.Status)
	assert.Equal(t, http.StatusServiceUnavailable, event.HTTPStatus)
	assert.Equal(t, "http_status", event.FailureReason)
	assert.Equal(t, "Sandbox URL returned HTTP 503", event.Message)
}

func TestSandboxHealthEventFromStatusRunning(t *testing.T) {
	sandbox := &models.Sandbox{
		ID:     uuid.New(),
		Status: models.SandboxStatusRunning,
		URL:    "https://example.invalid",
	}

	event := sandboxHealthEventFromStatus(sandbox)

	require.True(t, event.Ready)
	assert.Equal(t, "ready", event.Status)
	assert.Equal(t, "Sandbox URL is reachable", event.Message)
}
