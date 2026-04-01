package auditlog

type Action string

const (
	ActionAuthLoggedIn   Action = "auth.logged_in"
	ActionAuthLoggedOut  Action = "auth.logged_out"
	ActionUserRegistered Action = "user.registered"

	ActionImageCreated           Action = "image.created"
	ActionImageUpdated           Action = "image.updated"
	ActionImageDeleted           Action = "image.deleted"
	ActionImageThumbnailUploaded Action = "image.thumbnail_uploaded"
	ActionImageThumbnailDeleted  Action = "image.thumbnail_deleted"
	ActionImageSnapshotCreated   Action = "image.snapshot_created"

	ActionSandboxCreated    Action = "sandbox.created"
	ActionSandboxUpdated    Action = "sandbox.updated"
	ActionSandboxTTLUpdated Action = "sandbox.ttl_updated"
	ActionSandboxDeleted    Action = "sandbox.deleted"

	ActionUserCreated          Action = "user.created"
	ActionUserUpdated          Action = "user.updated"
	ActionUserDeleted          Action = "user.deleted"
	ActionUserWhitelisted      Action = "user.whitelisted"
	ActionUserWhitelistRemoved Action = "user.whitelist_removed"
)

var allActions = []Action{
	ActionAuthLoggedIn,
	ActionAuthLoggedOut,
	ActionUserRegistered,
	ActionImageCreated,
	ActionImageUpdated,
	ActionImageDeleted,
	ActionImageThumbnailUploaded,
	ActionImageThumbnailDeleted,
	ActionImageSnapshotCreated,
	ActionSandboxCreated,
	ActionSandboxUpdated,
	ActionSandboxTTLUpdated,
	ActionSandboxDeleted,
	ActionUserCreated,
	ActionUserUpdated,
	ActionUserDeleted,
	ActionUserWhitelisted,
	ActionUserWhitelistRemoved,
}

type ResourceType string

const (
	ResourceTypeImage   ResourceType = "image"
	ResourceTypeSandbox ResourceType = "sandbox"
	ResourceTypeUser    ResourceType = "user"
)

func AllActions() []Action {
	return append([]Action(nil), allActions...)
}
