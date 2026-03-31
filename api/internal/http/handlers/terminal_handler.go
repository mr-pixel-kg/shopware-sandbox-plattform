package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/manuel/shopware-testenv-platform/api/internal/apperror"
	"github.com/manuel/shopware-testenv-platform/api/internal/docker"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/dto"
	"github.com/manuel/shopware-testenv-platform/api/internal/http/responses"
	"github.com/manuel/shopware-testenv-platform/api/internal/services"
)

const (
	wsReadLimit  = 64 * 1024
	wsPingPeriod = 30 * time.Second
	wsPongWait   = 10 * time.Second
	execBufSize  = 4096
)

type TerminalHandler struct {
	terminals      *services.TerminalService
	auth           *services.AuthService
	upgrader       websocket.Upgrader
	allowedOrigins []string
}

func NewTerminalHandler(
	terminals *services.TerminalService,
	auth *services.AuthService,
	allowedOrigins []string,
) *TerminalHandler {
	h := &TerminalHandler{
		terminals:      terminals,
		auth:           auth,
		allowedOrigins: allowedOrigins,
	}
	h.upgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin:     h.checkOrigin,
	}
	return h
}

func (h *TerminalHandler) checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	for _, allowed := range h.allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

// Connect godoc
// @Summary      Open interactive terminal session
// @Description  Interactive shell (docker exec) into the sandbox container
// @Tags         Sandboxes
// @Param        id path string true "Sandbox ID" format(uuid)
// @Param        access_token query string true "Bearer token"
// @Param        cols query int false "Initial terminal columns" default(80)
// @Param        rows query int false "Initial terminal rows" default(24)
// @Success      101 "Switching Protocols – WebSocket connection established"
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse "Sandbox not running or session limit reached"
// @Failure      500 {object} dto.ErrorResponse
// @Router       /api/sandboxes/{id}/terminal [get]
func (h *TerminalHandler) Connect(c echo.Context) error {
	sandboxID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return responses.FromAppError(c, apperror.BadRequest("VALIDATION_ERROR", "Invalid sandbox id"))
	}

	token := c.QueryParam("access_token")
	if token == "" {
		if t, ok := parseAuthorizationHeader(c.Request().Header.Get(echo.HeaderAuthorization)); ok {
			token = t
		}
	}
	if token == "" {
		return responses.FromAppError(c, apperror.Unauthorized("Missing access token"))
	}

	user, _, err := h.auth.Authenticate(token)
	if err != nil {
		return responses.FromAppError(c, apperror.Unauthorized("Invalid or expired token"))
	}

	sandbox, err := h.terminals.ValidateAccess(sandboxID, user)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrSandboxNotFound):
			return responses.FromAppError(c, apperror.NotFound("SANDBOX_NOT_FOUND", "Sandbox not found"))
		case errors.Is(err, services.ErrTerminalNotRunning):
			return responses.FromAppError(c, apperror.Conflict("SANDBOX_NOT_RUNNING", "Sandbox is not running"))
		case errors.Is(err, services.ErrTerminalAccessDenied):
			return responses.FromAppError(c, apperror.New(http.StatusForbidden, "TERMINAL_ACCESS_DENIED", "Terminal access denied"))
		default:
			return responses.FromAppError(c, apperror.Internal("TERMINAL_ERROR", "Terminal error").WithCause(err))
		}
	}

	cols, rows := parseTerminalSize(c)

	execSession, err := h.terminals.OpenSession(c.Request().Context(), sandbox, cols, rows)
	if err != nil {
		if errors.Is(err, services.ErrTerminalSessionLimit) {
			return responses.FromAppError(c, apperror.Conflict("TERMINAL_SESSION_LIMIT", "Too many active terminal sessions"))
		}
		return responses.FromAppError(c, apperror.Internal("TERMINAL_EXEC_FAILED", "Failed to create terminal session").WithCause(err))
	}

	ws, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		_ = execSession.Close()
		h.terminals.CloseSession(sandboxID)
		slog.Error("websocket upgrade failed", "error", err, "sandbox_id", sandboxID)
		return nil
	}

	slog.Info("terminal session started", "sandbox_id", sandboxID, "user_id", user.ID)

	h.bridgeConnection(ws, execSession, sandboxID)

	slog.Info("terminal session ended", "sandbox_id", sandboxID, "user_id", user.ID)
	return nil
}

func (h *TerminalHandler) bridgeConnection(ws *websocket.Conn, exec *docker.ExecSession, sandboxID uuid.UUID) {
	cfg := h.terminals.Config()
	idleTimeout := time.Duration(cfg.IdleTimeoutMinutes) * time.Minute
	maxDuration := time.Duration(cfg.MaxDurationMinutes) * time.Minute
	deadline := time.Now().Add(maxDuration)

	defer func() {
		_ = ws.Close()
		_ = exec.Close()
		h.terminals.CloseSession(sandboxID)
	}()

	done := make(chan struct{})

	go func() {
		defer func() {
			select {
			case <-done:
			default:
				close(done)
			}
		}()
		buf := make([]byte, execBufSize)
		for {
			n, err := exec.Read(buf)
			if n > 0 {
				if writeErr := ws.WriteMessage(websocket.BinaryMessage, buf[:n]); writeErr != nil {
					return
				}
			}
			if err != nil {
				if err != io.EOF {
					sendControlMessage(ws, dto.TerminalMsgError, "Container process ended unexpectedly", 0)
				} else {
					sendControlMessage(ws, dto.TerminalMsgExit, "", 0)
				}
				_ = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				return
			}
		}
	}()

	go func() {
		defer func() {
			select {
			case <-done:
			default:
				close(done)
			}
		}()

		ws.SetReadLimit(wsReadLimit)
		_ = ws.SetReadDeadline(time.Now().Add(idleTimeout))

		ws.SetPongHandler(func(string) error {
			return ws.SetReadDeadline(time.Now().Add(idleTimeout))
		})

		for {
			msgType, data, err := ws.ReadMessage()
			if err != nil {
				return
			}

			nextDeadline := time.Now().Add(idleTimeout)
			if nextDeadline.After(deadline) {
				nextDeadline = deadline
			}
			_ = ws.SetReadDeadline(nextDeadline)

			switch msgType {
			case websocket.BinaryMessage:
				if _, err := exec.Write(data); err != nil {
					return
				}
			case websocket.TextMessage:
				var msg dto.TerminalControlMessage
				if err := json.Unmarshal(data, &msg); err != nil {
					continue
				}
				if msg.Type == dto.TerminalMsgResize && msg.Cols > 0 && msg.Rows > 0 {
					_ = exec.Resize(context.Background(), msg.Cols, msg.Rows)
				}
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(wsPingPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := ws.WriteControl(websocket.PingMessage, nil, time.Now().Add(wsPongWait)); err != nil {
					return
				}
			case <-done:
				return
			}
		}
	}()

	<-done
}

func sendControlMessage(ws *websocket.Conn, msgType, message string, code int) {
	msg := dto.TerminalControlMessage{
		Type:    msgType,
		Message: message,
		Code:    code,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	_ = ws.WriteMessage(websocket.TextMessage, data)
}

func parseTerminalSize(c echo.Context) (cols, rows uint) {
	cols = 80
	rows = 24
	if v, err := strconv.ParseUint(c.QueryParam("cols"), 10, 32); err == nil && v > 0 {
		cols = uint(v)
	}
	if v, err := strconv.ParseUint(c.QueryParam("rows"), 10, 32); err == nil && v > 0 {
		rows = uint(v)
	}
	return
}
