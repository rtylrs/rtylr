package domain

import (
	"errors"
	"time"
)

var ErrProjectNotFound = errors.New("project not found")

type Project struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	WorkspaceID *int64    `json:"workspace_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Color       string    `json:"color"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
