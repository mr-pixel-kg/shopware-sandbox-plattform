package dto

import "github.com/manuel/shopware-testenv-platform/api/internal/models"

type ErrorDetail struct {
	Code    string `json:"code" example:"VALIDATION_ERROR"`
	Message string `json:"message" example:"Invalid request body"`
	Details any    `json:"details,omitempty" swaggertype:"object"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type AuthLoginResponse struct {
	Token string      `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
	User  models.User `json:"user"`
}

type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

type ImagePullProgressEvent struct {
	Percent int    `json:"percent" example:"75"`
	Status  string `json:"status" example:"pulling"`
}
