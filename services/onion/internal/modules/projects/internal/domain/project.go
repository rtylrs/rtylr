package domain

import (
	"errors"
	"time"
)

var ErrProjectNotFound = errors.New("project not found")

type Project struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"ownerId"`
	WorkspaceID *int64    `json:"workspaceId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Color       string    `json:"color"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ProjectInput struct {
	OwnerID     int64
	WorkspaceID *int64
	Name        string
	Description string
	Color       string
}

type ProjectUpdateInput struct {
	ID          int64
	OwnerID     int64
	WorkspaceID *int64
	Name        string
	Description string
	Status      string
	Color       string
}
