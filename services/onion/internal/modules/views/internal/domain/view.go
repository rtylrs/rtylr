package domain

import (
	"errors"
	"time"
)

var ErrViewNotFound = errors.New("view not found")

type View struct {
	ID          int64
	OwnerID     int64
	WorkspaceID int64
	Name        string
	Filters     []byte
	SortBy      string
	GroupBy     string
	IsDefault   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ViewInput struct {
	OwnerID     int64
	WorkspaceID int64
	Name        string
	Filters     []byte
	SortBy      string
	GroupBy     string
	IsDefault   bool
}

type ViewUpdateInput struct {
	ID        int64
	OwnerID   int64
	Name      string
	Filters   []byte
	SortBy    string
	GroupBy   string
	IsDefault bool
}
