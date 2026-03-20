package dto

type CreateSandboxRequest struct {
	ImageID    string `json:"imageId" validate:"required" format:"uuid" example:"8ae13ed9-cfb1-4941-a248-bc74b9fb6a24"`
	TTLMinutes *int   `json:"ttlMinutes"`
}

type ExtendTTLRequest struct {
	TTLMinutes int `json:"ttlMinutes"`
}

type CreateSnapshotRequest struct {
	ImagePayload
}
