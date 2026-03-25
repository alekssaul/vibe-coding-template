package model

import "time"

// Item represents a generic CRUD resource.
type Item struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateItemRequest is the payload for creating an item.
type CreateItemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateItemRequest is the payload for updating an item.
type UpdateItemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
