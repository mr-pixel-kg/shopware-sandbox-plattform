package handlers

import (
	"bufio"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/mr-pixel-kg/shopshredder/api/internal/http/dto"
	"github.com/mr-pixel-kg/shopshredder/api/internal/http/errs"
	mw "github.com/mr-pixel-kg/shopshredder/api/internal/http/middleware"
	"github.com/mr-pixel-kg/shopshredder/api/internal/services"
)

type LogHandler struct {
	Logs *services.LogService
}

func (h LogHandler) ListSources(c fuego.ContextNoBody) ([]dto.LogSourceResponse, error) {
	id, err := parsePathUUID(c, "id")
	if err != nil {
		return nil, err
	}

	r := c.Request()
	auth := mw.MustAuth(r)
	user := mw.UserFromContext(r)

	sandbox, err := h.Logs.ValidateAccess(services.ValidateLogAccessInput{
		SandboxID: id,
		UserID:    auth.UserID,
		IsAdmin:   user.IsAdmin(),
	})
	if err != nil {
		return nil, mapLogError(err)
	}

	sources := h.Logs.GetLogSources(sandbox)
	out := make([]dto.LogSourceResponse, len(sources))
	for i, src := range sources {
		out[i] = dto.LogSourceResponse{
			Key:   src.Key,
			Label: src.Label,
			Type:  string(src.Type),
		}
	}
	return out, nil
}

func (h LogHandler) StreamLog(w http.ResponseWriter, r *http.Request) {
	sandboxID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		errs.Write(w, http.StatusBadRequest, "Invalid sandbox id")
		return
	}

	key := r.PathValue("key")
	if key == "" {
		errs.Write(w, http.StatusBadRequest, "Log source key is required")
		return
	}

	auth := mw.MustAuth(r)
	user := mw.UserFromContext(r)

	sandbox, err := h.Logs.ValidateAccess(services.ValidateLogAccessInput{
		SandboxID: sandboxID,
		UserID:    auth.UserID,
		IsAdmin:   user.IsAdmin(),
	})
	if err != nil {
		writeLogError(w, err)
		return
	}

	source, err := h.Logs.FindLogSource(sandbox, key)
	if err != nil {
		writeLogError(w, err)
		return
	}

	logStream, err := h.Logs.StreamLog(r.Context(), sandbox.ContainerID, *source)
	if err != nil {
		errs.Write(w, http.StatusInternalServerError, "Failed to open log stream")
		return
	}
	defer logStream.Reader.Close()

	writeSSEHeaders(w)

	var reader io.Reader
	if logStream.TTY {
		reader = logStream.Reader
	} else {
		pr, pw := io.Pipe()
		defer func() { _ = pr.Close() }()
		go func() {
			defer func() { _ = pw.Close() }()
			_, _ = stdcopy.StdCopy(pw, pw, logStream.Reader)
		}()
		reader = pr
	}

	ctx := r.Context()
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 256*1024), 256*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}
		sendSSEEvent(w, dto.LogEvent{Line: scanner.Text()})
	}
	if err := scanner.Err(); err != nil {
		slog.Warn("log stream scanner error", "error", err, "sandbox_id", sandboxID, "key", key)
	}
}

func writeLogError(w http.ResponseWriter, err error) {
	httpErr := mapLogError(err).(fuego.HTTPError)
	errs.Write(w, httpErr.Status, httpErr.Detail)
}

func mapLogError(err error) error {
	switch {
	case errors.Is(err, services.ErrSandboxNotFound):
		return fuego.HTTPError{Status: http.StatusNotFound, Detail: "Sandbox not found"}
	case errors.Is(err, services.ErrLogNotRunning):
		return fuego.HTTPError{Status: http.StatusConflict, Detail: "Sandbox is not running"}
	case errors.Is(err, services.ErrLogAccessDenied):
		return fuego.HTTPError{Status: http.StatusForbidden, Detail: "Log access denied"}
	case errors.Is(err, services.ErrLogSourceNotFound):
		return fuego.HTTPError{Status: http.StatusNotFound, Detail: "Log source not found"}
	default:
		return fuego.HTTPError{Status: http.StatusInternalServerError, Detail: "Log operation failed"}
	}
}
