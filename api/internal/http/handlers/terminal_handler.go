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
	"github.com/mr-pixel-kg/shopshredder/api/internal/docker"
	"github.com/mr-pixel-kg/shopshredder/api/internal/http/dto"
	"github.com/mr-pixel-kg/shopshredder/api/internal/http/errs"
	mw "github.com/mr-pixel-kg/shopshredder/api/internal/http/middleware"
	"github.com/mr-pixel-kg/shopshredder/api/internal/services"
)

const (
	wsReadLimit  = 64 * 1024
	wsPingPeriod = 30 * time.Second
	wsPongWait   = 10 * time.Second
	execBufSize  = 4096
)

type TerminalHandler struct {
	Terminals      *services.TerminalService
	Auth           *services.AuthService
	AllowedOrigins []string
}

func (h TerminalHandler) upgrader() *websocket.Upgrader {
	return &websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true
			}
			for _, allowed := range h.AllowedOrigins {
				if allowed == "*" || allowed == origin {
					return true
				}
			}
			return false
		},
	}
}

func (h TerminalHandler) Connect(w http.ResponseWriter, r *http.Request) {
	sandboxID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		errs.Write(w, http.StatusBadRequest, "Invalid sandbox id")
		return
	}

	token := r.URL.Query().Get("access_token")
	if token == "" {
		if t, ok := mw.ParseAuthorizationHeader(r.Header.Get("Authorization")); ok {
			token = t
		}
	}
	if token == "" {
		errs.Write(w, http.StatusUnauthorized, "Missing access token")
		return
	}

	user, err := h.Auth.Authenticate(token)
	if err != nil {
		errs.Write(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	sandbox, err := h.Terminals.ValidateAccess(services.ValidateTerminalAccessInput{
		SandboxID: sandboxID,
		UserID:    user.ID,
		IsAdmin:   user.IsAdmin(),
	})
	if err != nil {
		writeTerminalError(w, err)
		return
	}

	cols, rows := parseTerminalSize(r)

	execSession, err := h.Terminals.OpenSession(r.Context(), sandbox, cols, rows)
	if err != nil {
		writeTerminalError(w, err)
		return
	}

	ws, err := h.upgrader().Upgrade(w, r, nil)
	if err != nil {
		_ = execSession.Close()
		h.Terminals.CloseSession(sandboxID)
		slog.Error("websocket upgrade failed", "error", err, "sandbox_id", sandboxID)
		return
	}

	slog.Info("terminal session started", "sandbox_id", sandboxID, "user_id", user.ID)

	h.bridgeConnection(ws, execSession, sandboxID)

	slog.Info("terminal session ended", "sandbox_id", sandboxID, "user_id", user.ID)
}

func writeTerminalError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrSandboxNotFound):
		errs.Write(w, http.StatusNotFound, "Sandbox not found")
	case errors.Is(err, services.ErrTerminalNotRunning):
		errs.Write(w, http.StatusConflict, "Sandbox is not running")
	case errors.Is(err, services.ErrTerminalAccessDenied):
		errs.Write(w, http.StatusForbidden, "Terminal access denied")
	case errors.Is(err, services.ErrTerminalSessionLimit):
		errs.Write(w, http.StatusConflict, "Too many active terminal sessions")
	default:
		errs.Write(w, http.StatusInternalServerError, "Terminal error")
	}
}

func (h TerminalHandler) bridgeConnection(ws *websocket.Conn, exec *docker.ExecSession, sandboxID uuid.UUID) {
	cfg := h.Terminals.Config()
	idleTimeout := time.Duration(cfg.IdleTimeoutMinutes) * time.Minute
	maxDuration := time.Duration(cfg.MaxDurationMinutes) * time.Minute
	deadline := time.Now().Add(maxDuration)

	defer func() {
		_ = ws.Close()
		_ = exec.Close()
		h.Terminals.CloseSession(sandboxID)
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

func parseTerminalSize(r *http.Request) (cols, rows uint) {
	cols = 80
	rows = 24
	if v, err := strconv.ParseUint(r.URL.Query().Get("cols"), 10, 32); err == nil && v > 0 {
		cols = uint(v)
	}
	if v, err := strconv.ParseUint(r.URL.Query().Get("rows"), 10, 32); err == nil && v > 0 {
		rows = uint(v)
	}
	return
}
