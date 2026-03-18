package apperror

import "net/http"

type AppError struct {
	// StatusCode is the HTTP status that should be exposed to API clients.
	StatusCode int
	// Code is a stable machine-readable identifier for frontend branching.
	Code string
	// Message is the user-facing error message returned by the API.
	Message string
	// Details can carry optional structured context for clients and debugging.
	Details any
	// Cause keeps the original internal error without forcing it into the response.
	Cause error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func New(statusCode int, code, message string) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}

func (e *AppError) WithDetails(details any) *AppError {
	e.Details = details
	return e
}

func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

func BadRequest(code, message string) *AppError {
	return New(http.StatusBadRequest, code, message)
}

func Unauthorized(message string) *AppError {
	return New(http.StatusUnauthorized, "UNAUTHORIZED", message)
}

func NotFound(code, message string) *AppError {
	return New(http.StatusNotFound, code, message)
}

func Conflict(code, message string) *AppError {
	return New(http.StatusConflict, code, message)
}

func Internal(code, message string) *AppError {
	return New(http.StatusInternalServerError, code, message)
}
