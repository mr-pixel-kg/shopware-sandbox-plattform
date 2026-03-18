package logging

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/manuel/shopware-testenv-platform/api/internal/apperror"
)

func RequestFields(c echo.Context, extra ...any) []any {
	fields := []any{
		"request_id", c.Response().Header().Get(echo.HeaderXRequestID),
		"method", c.Request().Method,
		"path", c.Request().URL.Path,
		"route", c.Path(),
		"remote_ip", c.RealIP(),
		"user_agent", c.Request().UserAgent(),
	}

	return append(fields, extra...)
}

func ErrorFields(err error, extra ...any) []any {
	fields := []any{"error", err.Error()}

	var appErr *apperror.AppError
	if ok := apperrorAs(err, &appErr); ok {
		fields = append(fields,
			"status", appErr.StatusCode,
			"error_code", appErr.Code,
		)
		if appErr.Cause != nil {
			fields = append(fields, "cause", appErr.Cause.Error())
		}
	}

	return append(fields, extra...)
}

func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return redact(email, 2)
	}

	return fmt.Sprintf("%s@%s", redact(parts[0], 1), parts[1])
}

func redact(value string, visible int) string {
	if value == "" {
		return ""
	}
	if len(value) <= visible {
		return strings.Repeat("*", len(value))
	}

	return value[:visible] + strings.Repeat("*", len(value)-visible)
}

func apperrorAs(err error, target **apperror.AppError) bool {
	if err == nil {
		return false
	}

	appErr, ok := err.(*apperror.AppError)
	if !ok {
		return false
	}

	*target = appErr
	return true
}

func LogRequestError(c echo.Context, message string, err error, extra ...any) {
	slog.Error(message, append(RequestFields(c), ErrorFields(err, extra...)...)...)
}
