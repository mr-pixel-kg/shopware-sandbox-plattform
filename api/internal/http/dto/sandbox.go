package dto

type CreateSandboxRequest struct {
	ImageID     string            `json:"imageId" validate:"required" format:"uuid" example:"8ae13ed9-cfb1-4941-a248-bc74b9fb6a24"`
	TTLMinutes  *int              `json:"ttlMinutes" example:"120"`
	DisplayName *string           `json:"displayName,omitempty" example:"My Test Shop"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type UpdateSandboxRequest struct {
	DisplayName *string `json:"displayName" example:"My Test Shop"`
}

type ExtendTTLRequest struct {
	TTLMinutes *int `json:"ttlMinutes" example:"60"`
}

type CreateSnapshotRequest struct {
	ImagePayload
}
