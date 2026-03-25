package model

import "time"

// Permission defines the access level of an API key.
type Permission string

const (
	PermissionRead  Permission = "read"
	PermissionWrite Permission = "write"
)

// APIKey represents a stored API key (key hash is never returned in responses).
type APIKey struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	KeyHash    string     `json:"-"`
	Permission Permission `json:"permission"`
	CreatedAt  time.Time  `json:"created_at"`
}

// CreateAPIKeyRequest is the payload for creating an API key.
type CreateAPIKeyRequest struct {
	Name       string     `json:"name" validate:"required,min=2,max=100"`
	Permission Permission `json:"permission" validate:"required,oneof=read write"`
}

// CreateAPIKeyResponse is returned only on key creation; Key is shown once.
type CreateAPIKeyResponse struct {
	APIKey
	Key string `json:"key"`
}
