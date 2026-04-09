package domain

import (
	"errors"
	"time"
)

var (
	ErrTaskNotFound  = errors.New("task not found")
	ErrLabelNotFound = errors.New("label not found")
)

type Task struct {
	ID          int64      `json:"id"`
	OwnerID     int64      `json:"ownerId"`
	ProjectID   *int64     `json:"projectId"`
	WorkspaceID *int64     `json:"workspaceId"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	AssigneeID  *int64     `json:"assigneeId"`
	LabelIDs    []byte     `json:"labelIds"`
	DueDate     *time.Time `json:"dueDate"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type Label struct {
	ID        int64     `json:"id"`
	OwnerID   int64     `json:"ownerId"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"createdAt"`
}

type TaskInput struct {
	OwnerID     int64
	ProjectID   *int64
	WorkspaceID *int64
	Title       string
	Description string
	Priority    string
	AssigneeID  *int64
	LabelIDs    []byte
	DueDate     *time.Time
}

type TaskUpdateInput struct {
	ID          int64
	OwnerID     int64
	ProjectID   *int64
	WorkspaceID *int64
	Title       string
	Description string
	Status      string
	Priority    string
	AssigneeID  *int64
	LabelIDs    []byte
	DueDate     *time.Time
}

type LabelInput struct {
	OwnerID int64
	Name    string
	Color   string
}
