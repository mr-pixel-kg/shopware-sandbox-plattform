package services

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/manuel/shopware-testenv-platform/api/internal/models"
	"github.com/manuel/shopware-testenv-platform/api/internal/registry"
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
	StateReason   string    `json:"stateReason,omitempty"`
	CheckedAt     time.Time `json:"checkedAt"`
}

type sandboxHealthState struct {
	subscribers map[chan SandboxHealthEvent]struct{}
	cancel      context.CancelFunc
	last        *SandboxHealthEvent
}

type HealthCheckResolver interface {
	ResolveEntry(imageName string) *registry.ImageEntry
}

type SandboxHealthService struct {
	repo     *repositories.SandboxRepository
	imgRepo  *repositories.ImageRepository
	resolver HealthCheckResolver
	interval time.Duration
	client   *http.Client

	mu          sync.Mutex
	active      map[uuid.UUID]*sandboxHealthState
	healthCache map[uuid.UUID]*registry.HealthCheckConfig
}

func NewSandboxHealthService(
	repo *repositories.SandboxRepository,
	imgRepo *repositories.ImageRepository,
	resolver HealthCheckResolver,
) *SandboxHealthService {
	return &SandboxHealthService{
		repo:     repo,
		imgRepo:  imgRepo,
		resolver: resolver,
		interval: 5 * time.Second,
		client: &http.Client{
			Timeout: 5 * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		active:      make(map[uuid.UUID]*sandboxHealthState),
		healthCache: make(map[uuid.UUID]*registry.HealthCheckConfig),
	}
}

func (s *SandboxHealthService) StartMonitoring(sandboxID uuid.UUID) {
	s.mu.Lock()
	if _, exists := s.active[sandboxID]; exists {
		s.mu.Unlock()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.active[sandboxID] = &sandboxHealthState{
		subscribers: make(map[chan SandboxHealthEvent]struct{}),
		cancel:      cancel,
	}
	s.mu.Unlock()

	go s.monitor(ctx, sandboxID)
}

func (s *SandboxHealthService) StartMonitoringActive() {
	sandboxes, err := s.repo.ListAllActive()
	if err != nil {
		slog.Error("load active sandboxes for health monitoring failed", "component", "sandbox_health", "error", err.Error())
		return
	}

	for _, sandbox := range sandboxes {
		switch sandbox.Status {
		case models.SandboxStatusStarting, models.SandboxStatusRunning:
			s.StartMonitoring(sandbox.ID)
		}
	}
}

func (s *SandboxHealthService) Watch(sandbox *models.Sandbox) (<-chan SandboxHealthEvent, func()) {
	ch := make(chan SandboxHealthEvent, 8)

	switch sandbox.Status {
	case models.SandboxStatusStarting, models.SandboxStatusRunning:
		s.StartMonitoring(sandbox.ID)
	default:
		ch <- sandboxHealthEventFromStatus(sandbox)
		close(ch)
		return ch, func() {}
	}

	s.mu.Lock()
	state := s.active[sandbox.ID]
	state.subscribers[ch] = struct{}{}
	if state.last != nil {
		ch <- *state.last
	}
	s.mu.Unlock()

	return ch, func() {
		s.removeSubscriber(sandbox.ID, ch)
	}
}

func (s *SandboxHealthService) monitor(ctx context.Context, sandboxID uuid.UUID) {
	defer s.stopMonitor(sandboxID)

	s.runProbe(sandboxID)

	interval := s.interval
	sandbox, err := s.repo.FindByID(sandboxID)
	if err == nil {
		if hc := s.getHealthConfig(sandbox); hc != nil && hc.Interval.Duration > 0 {
			interval = hc.Interval.Duration
		}
	}

	ticker := time.NewTicker(interval)
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
			Status:    models.HealthStatusNotFound,
			Ready:     false,
			Message:   "Sandbox not found",
			CheckedAt: time.Now().UTC(),
		})
		return true
	}

	switch sandbox.Status {
	case models.SandboxStatusStarting:
		event := s.probeSandbox(sandbox, false)
		if event.Ready {
			sandbox.Status = models.SandboxStatusRunning
			sandbox.StateReason = nil
			if err := s.repo.Update(sandbox); err != nil {
				event.Ready = false
				event.Status = "error"
				event.Message = fmt.Sprintf("Could not persist running state: %v", err)
				s.broadcast(sandboxID, event)
				return false
			}
			event.Status = models.HealthStatusReady
			event.StateReason = ""
		} else {
			event.StateReason = stateReasonFromProbe(event)
		}
		s.broadcast(sandboxID, event)
		return false

	case models.SandboxStatusRunning:
		event := s.probeSandbox(sandbox, true)
		if !event.Ready {
			event.Status = models.HealthStatusOffline
			event.StateReason = stateReasonFromProbe(event)
		}
		s.broadcast(sandboxID, event)
		return false

	default:
		s.broadcastFinal(sandboxID, sandboxHealthEventFromStatus(sandbox))
		return true
	}
}

func (s *SandboxHealthService) getHealthConfig(sandbox *models.Sandbox) *registry.HealthCheckConfig {
	s.mu.Lock()
	hc, ok := s.healthCache[sandbox.ImageID]
	s.mu.Unlock()
	if ok {
		return hc
	}

	if s.imgRepo == nil {
		return nil
	}

	img, err := s.imgRepo.FindByID(sandbox.ImageID)
	if err == nil && s.resolver != nil {
		if entry := s.resolver.ResolveEntry(img.RegistryName()); entry != nil {
			hc = entry.HealthCheck
		}
	}

	s.mu.Lock()
	s.healthCache[sandbox.ImageID] = hc
	s.mu.Unlock()
	return hc
}

func (s *SandboxHealthService) probeSandbox(sandbox *models.Sandbox, wasReady bool) SandboxHealthEvent {
	status := models.HealthStatusProbing
	if wasReady || sandbox.Status == models.SandboxStatusRunning {
		status = models.HealthStatusOffline
	}

	probeURL := sandbox.URL
	hc := s.getHealthConfig(sandbox)
	if hc != nil && hc.Path != "" {
		probeURL = strings.TrimRight(sandbox.URL, "/") + hc.Path
	}

	event := SandboxHealthEvent{
		SandboxID: sandbox.ID,
		Status:    status,
		Ready:     false,
		URL:       sandbox.URL,
		CheckedAt: time.Now().UTC(),
	}

	if probeURL == "" {
		event.FailureReason = "missing_url"
		event.Message = "Sandbox URL is empty"
		return event
	}

	timeout := s.client.Timeout
	if hc != nil && hc.Timeout.Duration > 0 {
		timeout = hc.Timeout.Duration
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, probeURL, nil)
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
		event.Status = models.HealthStatusReady
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
	state.cancel()
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

	if sandbox.StateReason != nil {
		event.StateReason = *sandbox.StateReason
	}

	switch sandbox.Status {
	case models.SandboxStatusRunning:
		event.Status = models.HealthStatusReady
		event.Message = "Sandbox URL is reachable"
	case models.SandboxStatusStarting:
		event.Status = models.HealthStatusProbing
		event.Message = "Sandbox is still starting"
	default:
		event.Message = fmt.Sprintf("Sandbox is %s", sandbox.Status)
	}

	return event
}

type SandboxStreamEvent struct {
	SandboxID   uuid.UUID
	Status      string
	StateReason string
}

func (s *SandboxHealthService) WatchStream(ctx context.Context, sandbox *models.Sandbox) <-chan SandboxStreamEvent {
	out := make(chan SandboxStreamEvent, 8)

	switch sandbox.Status {
	case models.SandboxStatusStarting, models.SandboxStatusRunning:
		go s.streamFromHealth(ctx, sandbox, out)
	default:
		go s.streamFromDB(ctx, sandbox.ID, out)
	}

	return out
}

func (s *SandboxHealthService) streamFromHealth(ctx context.Context, sandbox *models.Sandbox, out chan<- SandboxStreamEvent) {
	defer close(out)

	ch, cancel := s.Watch(sandbox)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			status := event.Status
			if event.Ready {
				status = string(models.SandboxStatusRunning)
			}
			select {
			case out <- SandboxStreamEvent{
				SandboxID:   event.SandboxID,
				Status:      status,
				StateReason: event.StateReason,
			}:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (s *SandboxHealthService) streamFromDB(ctx context.Context, id uuid.UUID, out chan<- SandboxStreamEvent) {
	defer close(out)

	send := func() bool {
		sandbox, err := s.repo.FindByID(id)
		if err != nil {
			select {
			case out <- SandboxStreamEvent{SandboxID: id, Status: string(models.SandboxStatusFailed)}:
			case <-ctx.Done():
			}
			return true
		}
		reason := ""
		if sandbox.StateReason != nil {
			reason = *sandbox.StateReason
		}
		select {
		case out <- SandboxStreamEvent{SandboxID: sandbox.ID, Status: string(sandbox.Status), StateReason: reason}:
		case <-ctx.Done():
			return true
		}
		return !sandbox.Status.IsTransitional()
	}

	if send() {
		return
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if send() {
				return
			}
		}
	}
}

func stateReasonFromProbe(event SandboxHealthEvent) string {
	if event.FailureReason != "" {
		return "Nicht erreichbar"
	}
	return "Warte auf HTTP-Bereitschaft"
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
