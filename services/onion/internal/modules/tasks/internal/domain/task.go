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
	OwnerID     int64      `json:"owner_id"`
	ProjectID   *int64     `json:"project_id"`
	WorkspaceID *int64     `json:"workspace_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	AssigneeID  *int64     `json:"assignee_id"`
	LabelIDs    []byte     `json:"label_ids"`
	DueDate     *time.Time `json:"due_date"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Label struct {
	ID        int64     `json:"id"`
	OwnerID   int64     `json:"owner_id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
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
