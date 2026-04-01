package handlers

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/manuel/shopware-testenv-platform/api/internal/apperror"
)

func bindAndValidate(c echo.Context, target any) error {
	if err := c.Bind(target); err != nil {
		return apperror.BadRequest("VALIDATION_ERROR", "Invalid request body")
	}

	if err := c.Validate(target); err != nil {
		return apperror.BadRequest("VALIDATION_ERROR", "Request validation failed").WithDetails([]string{err.Error()})
	}

	return nil
}

func parseUUIDParam(c echo.Context, name string, code string, message string) (uuid.UUID, error) {
	value := c.Param(name)
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, apperror.BadRequest(code, message)
	}

	return id, nil
}

func validationError(message string) error {
	return apperror.BadRequest("VALIDATION_ERROR", message)
}
