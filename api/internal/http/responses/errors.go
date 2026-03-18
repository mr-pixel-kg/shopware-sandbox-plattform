package responses

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/manuel/shopware-testenv-platform/api/internal/apperror"
)

func Error(c echo.Context, status int, code, message string) error {
	return FromAppError(c, apperror.New(status, code, message))
}

func FromAppError(c echo.Context, err *apperror.AppError) error {
	// Keep the external payload intentionally small and stable even if the
	// internal error object carries more debugging context.
	body := map[string]any{
		"error": map[string]any{
			"code":    err.Code,
			"message": err.Message,
		},
	}

	if err.Details != nil {
		body["error"].(map[string]any)["details"] = err.Details
	}

	return c.JSON(err.StatusCode, body)
}

func FromError(c echo.Context, err error) error {
	var appErr *apperror.AppError
	if ok := AsAppError(err, &appErr); ok {
		return FromAppError(c, appErr)
	}

	// Unknown errors are normalized into the shared response format so handlers
	// do not accidentally leak arbitrary error strings or stack details.
	return FromAppError(c, apperror.Internal("INTERNAL_ERROR", http.StatusText(http.StatusInternalServerError)).WithCause(err))
}

func AsAppError(err error, target **apperror.AppError) bool {
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
