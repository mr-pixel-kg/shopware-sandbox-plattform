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

type PendingPullResponse struct {
	ID      string  `json:"id" example:"8ae13ed9-cfb1-4941-a248-bc74b9fb6a24"`
	Name    string  `json:"name" example:"dockware/shopware"`
	Tag     string  `json:"tag" example:"6.6.9.0"`
	Title   *string `json:"title,omitempty" example:"Shopware Demo Image"`
	Percent int     `json:"percent" example:"42"`
	Status  string  `json:"status" example:"pulling"`
}
