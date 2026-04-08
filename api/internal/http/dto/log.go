package dto

type LogSourceResponse struct {
	Key   string `json:"key" example:"container"`
	Label string `json:"label" example:"Container Output"`
	Type  string `json:"type" enums:"docker,file" example:"docker"`
}

type LogEvent struct {
	Line string `json:"line"`
}
