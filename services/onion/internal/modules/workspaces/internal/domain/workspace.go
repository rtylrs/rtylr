package domain

import (
	"errors"
	"time"
)

var (
	ErrWorkspaceNotFound = errors.New("workspace not found")
	ErrProjectNotFound   = errors.New("project not found")
	ErrTaskNotFound      = errors.New("task not found")
	ErrCommentNotFound   = errors.New("comment not found")
	ErrLabelNotFound     = errors.New("label not found")
	ErrUnauthorized      = errors.New("unauthorized")
)

type Workspace struct {
	ID          int64
	OwnerID     int64
	Name        string
	Description string
	Icon        string
	Color       string
	IsDefault   bool
	CreatedBy   int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type WorkspaceMember struct {
	ID          int64
	WorkspaceID int64
	UserID      int64
	Role        string
	JoinedAt    time.Time
}

type WorkspaceInput struct {
	OwnerID     int64
	Name        string
	Description string
	Icon        string
	Color       string
	CreatedBy   int64
}

type UpdateWorkspaceInput struct {
	ID          int64
	OwnerID     int64
	Name        string
	Description string
	Icon        string
	Color       string
}

type AddWorkspaceMemberInput struct {
	WorkspaceID int64
	UserID      int64
	Role        string
}

type RemoveWorkspaceMemberInput struct {
	WorkspaceID int64
	UserID      int64
}
