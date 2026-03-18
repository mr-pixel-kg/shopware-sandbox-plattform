package logging

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

func EchoRequestLogger() echo.MiddlewareFunc {
	return echomw.RequestLoggerWithConfig(echomw.RequestLoggerConfig{
		LogLatency:       true,
		LogRemoteIP:      true,
		LogHost:          true,
		LogMethod:        true,
		LogURI:           true,
		LogRoutePath:     true,
		LogRequestID:     true,
		LogUserAgent:     true,
		LogStatus:        true,
		LogError:         true,
		LogContentLength: true,
		LogResponseSize:  true,
		LogValuesFunc: func(c echo.Context, v echomw.RequestLoggerValues) error {
			fields := []any{
				"component", "http_request",
				"request_id", v.RequestID,
				"method", v.Method,
				"uri", v.URI,
				"route", v.RoutePath,
				"remote_ip", v.RemoteIP,
				"host", v.Host,
				"user_agent", v.UserAgent,
				"status", v.Status,
				"latency_ms", float64(v.Latency.Nanoseconds()) / 1e6,
				"bytes_in", v.ContentLength,
				"bytes_out", v.ResponseSize,
			}
			if v.Error != nil {
				fields = append(fields, "error", v.Error.Error())
			}

			switch {
			case v.Status >= http.StatusInternalServerError:
				slog.Error("http request completed", fields...)
			case v.Status >= http.StatusBadRequest:
				slog.Warn("http request completed", fields...)
			default:
				slog.Info("http request completed", fields...)
			}

			return nil
		},
	})
}
