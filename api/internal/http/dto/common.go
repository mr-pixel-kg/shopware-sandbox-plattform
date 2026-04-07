package dto

type PaginationMeta struct {
	Limit   int   `json:"limit" example:"50"`
	Offset  int   `json:"offset" example:"0"`
	Count   int   `json:"count" example:"50"`
	Total   int64 `json:"total" example:"137"`
	HasMore bool  `json:"hasMore" example:"true"`
}

type PaginatedMeta struct {
	Pagination PaginationMeta `json:"pagination"`
}

type ErrorDetail struct {
	Code    string `json:"code" example:"VALIDATION_ERROR"`
	Message string `json:"message" example:"Invalid request body"`
	Details any    `json:"details,omitempty" swaggertype:"object"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}
