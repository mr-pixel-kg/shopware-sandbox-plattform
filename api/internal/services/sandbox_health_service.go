package services

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/manuel/shopware-testenv-platform/api/internal/repositories"
)

type SandboxHealthEvent struct {
	SandboxID     uuid.UUID `json:"sandboxId"`
	Status        string    `json:"status"`
	Ready         bool      `json:"ready"`
	URL           string    `json:"url"`
	HTTPStatus    int       `json:"httpStatus,omitempty"`
	LatencyMs     int64     `json:"latencyMs,omitempty"`
	FailureReason string    `json:"failureReason,omitempty"`
	Message       string    `json:"message,omitempty"`
	CheckedAt     time.Time `json:"checkedAt"`
}

type sandboxHealthState struct {
	subscribers map[chan SandboxHealthEvent]struct{}
	cancel      context.CancelFunc
	last        *SandboxHealthEvent
}

type SandboxHealthService struct {
	repo     *repositories.SandboxRepository
	interval time.Duration
	client   *http.Client

	mu     sync.Mutex
	active map[uuid.UUID]*sandboxHealthState
}

func NewSandboxHealthService(repo *repositories.SandboxRepository) *SandboxHealthService {
	return &SandboxHealthService{
		repo:     repo,
		interval: 5 * time.Second,
		client: &http.Client{
			Timeout: 3 * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		active: make(map[uuid.UUID]*sandboxHealthState),
	}
}

func (s *SandboxHealthService) Watch(sandbox *models.Sandbox) (<-chan SandboxHealthEvent, func()) {
	ch := make(chan SandboxHealthEvent, 8)

	if sandbox.Status != models.SandboxStatusStarting {
		ch <- sandboxHealthEventFromStatus(sandbox)
		close(ch)
		return ch, func() {}
	}

	s.mu.Lock()
	state := s.active[sandbox.ID]
	if state == nil {
		ctx, cancel := context.WithCancel(context.Background())
		state = &sandboxHealthState{
			subscribers: map[chan SandboxHealthEvent]struct{}{ch: {}},
			cancel:      cancel,
		}
		s.active[sandbox.ID] = state
		go s.monitor(ctx, sandbox.ID)
	} else {
		state.subscribers[ch] = struct{}{}
		if state.last != nil {
			ch <- *state.last
		}
	}
	s.mu.Unlock()

	return ch, func() {
		s.removeSubscriber(sandbox.ID, ch)
	}
}

func (s *SandboxHealthService) monitor(ctx context.Context, sandboxID uuid.UUID) {
	defer s.stopMonitor(sandboxID)

	s.runProbe(sandboxID)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if s.runProbe(sandboxID) {
				return
			}
		}
	}
}

func (s *SandboxHealthService) runProbe(sandboxID uuid.UUID) bool {
	sandbox, err := s.repo.FindByID(sandboxID)
	if err != nil {
		s.broadcastFinal(sandboxID, SandboxHealthEvent{
			SandboxID: sandboxID,
			Status:    "not_found",
			Ready:     false,
			Message:   "Sandbox not found",
			CheckedAt: time.Now().UTC(),
		})
		return true
	}

	if sandbox.Status != models.SandboxStatusStarting {
		s.broadcastFinal(sandboxID, sandboxHealthEventFromStatus(sandbox))
		return true
	}

	event := s.probeSandbox(sandbox)
	if event.Ready {
		sandbox.Status = models.SandboxStatusRunning
		if err := s.repo.Update(sandbox); err != nil {
			event.Ready = false
			event.Status = "error"
			event.FailureReason = "status_update_failed"
			event.Message = fmt.Sprintf("Could not persist running state: %v", err)
			s.broadcast(sandboxID, event)
			return false
		}
		event.Status = "ready"
		s.broadcastFinal(sandboxID, event)
		return true
	}

	s.broadcast(sandboxID, event)
	return false
}

func (s *SandboxHealthService) probeSandbox(sandbox *models.Sandbox) SandboxHealthEvent {
	event := SandboxHealthEvent{
		SandboxID: sandbox.ID,
		Status:    "probing",
		Ready:     false,
		URL:       sandbox.URL,
		CheckedAt: time.Now().UTC(),
	}

	if sandbox.URL == "" {
		event.FailureReason = "missing_url"
		event.Message = "Sandbox URL is empty"
		return event
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, sandbox.URL, nil)
	if err != nil {
		event.FailureReason = "invalid_url"
		event.Message = err.Error()
		return event
	}

	start := time.Now()
	resp, err := s.client.Do(req)
	event.LatencyMs = time.Since(start).Milliseconds()
	if err != nil {
		event.FailureReason = classifyProbeError(err)
		event.Message = err.Error()
		return event
	}
	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
	event.HTTPStatus = resp.StatusCode

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusBadRequest {
		event.Ready = true
		event.Message = "Sandbox URL is reachable"
		return event
	}

	event.FailureReason = "http_status"
	event.Message = fmt.Sprintf("Sandbox URL returned HTTP %d", resp.StatusCode)
	return event
}

func (s *SandboxHealthService) broadcast(sandboxID uuid.UUID, event SandboxHealthEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := s.active[sandboxID]
	if state == nil {
		return
	}

	state.last = &event
	for ch := range state.subscribers {
		select {
		case ch <- event:
		default:
		}
	}
}

func (s *SandboxHealthService) broadcastFinal(sandboxID uuid.UUID, event SandboxHealthEvent) {
	s.mu.Lock()
	state := s.active[sandboxID]
	if state == nil {
		s.mu.Unlock()
		return
	}

	delete(s.active, sandboxID)
	state.last = &event
	subscribers := make([]chan SandboxHealthEvent, 0, len(state.subscribers))
	for ch := range state.subscribers {
		subscribers = append(subscribers, ch)
	}
	s.mu.Unlock()

	for _, ch := range subscribers {
		select {
		case ch <- event:
		default:
		}
		close(ch)
	}
}

func (s *SandboxHealthService) removeSubscriber(sandboxID uuid.UUID, ch chan SandboxHealthEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := s.active[sandboxID]
	if state == nil {
		return
	}

	delete(state.subscribers, ch)
	if len(state.subscribers) == 0 {
		delete(s.active, sandboxID)
		state.cancel()
	}
}

func (s *SandboxHealthService) stopMonitor(sandboxID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := s.active[sandboxID]
	if state == nil {
		return
	}

	if len(state.subscribers) == 0 {
		delete(s.active, sandboxID)
	}
}

func sandboxHealthEventFromStatus(sandbox *models.Sandbox) SandboxHealthEvent {
	event := SandboxHealthEvent{
		SandboxID: sandbox.ID,
		Status:    string(sandbox.Status),
		Ready:     sandbox.Status == models.SandboxStatusRunning,
		URL:       sandbox.URL,
		CheckedAt: time.Now().UTC(),
	}

	switch sandbox.Status {
	case models.SandboxStatusRunning:
		event.Message = "Sandbox URL is reachable"
	case models.SandboxStatusStarting:
		event.Status = "probing"
		event.Message = "Sandbox is still starting"
	default:
		event.Message = fmt.Sprintf("Sandbox is %s", sandbox.Status)
	}

	return event
}

func classifyProbeError(err error) string {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		err = urlErr.Err
	}

	var hostnameErr *net.DNSError
	if errors.As(err, &hostnameErr) {
		return "dns_failed"
	}

	var unknownAuthorityErr x509.UnknownAuthorityError
	if errors.As(err, &unknownAuthorityErr) {
		return "tls_untrusted_certificate"
	}

	var certInvalidErr x509.CertificateInvalidError
	if errors.As(err, &certInvalidErr) {
		return "tls_invalid_certificate"
	}

	var recordHeaderErr tls.RecordHeaderError
	if errors.As(err, &recordHeaderErr) {
		return "tls_handshake_failed"
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		if opErr.Timeout() {
			return "timeout"
		}
		if opErr.Op == "dial" {
			return "connect_failed"
		}
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}

	if errors.Is(err, syscall.ECONNREFUSED) {
		return "connection_refused"
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "timeout"):
		return "timeout"
	case strings.Contains(msg, "tls"):
		return "tls_handshake_failed"
	case strings.Contains(msg, "connection refused"):
		return "connection_refused"
	default:
		return "request_failed"
	}
}
