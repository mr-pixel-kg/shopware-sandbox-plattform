package middleware

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

func RequestLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := &responseWriter{ResponseWriter: w}
			next.ServeHTTP(ww, r)

			duration := time.Since(start)
			fields := []any{
				"component", "http",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.status,
				"bytes", ww.bytes,
				"latency_ms", float64(duration.Nanoseconds()) / 1e6,
			}

			switch {
			case ww.status >= http.StatusInternalServerError:
				slog.Error("http request completed", fields...)
			case ww.status >= http.StatusBadRequest:
				slog.Warn("http request completed", fields...)
			default:
				slog.Info("http request completed", fields...)
			}
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, fmt.Errorf("underlying ResponseWriter does not implement http.Hijacker")
}

func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}
