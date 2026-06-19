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

// ─────────────────────────────────────────────
// Workspace
// ─────────────────────────────────────────────

type Workspace struct {
	ID          int64
	BusinessID  int64
	OwnerID     int64
	Name        string
	Description string
	Icon        string
	Color       string
	IsDefault   bool
	IsPersonal  bool
	CreatedBy   int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type WorkspaceMember struct {
	ID          int64
	WorkspaceID int64
	UserID      int64
	Name        string
	Email       string
	Role        string // manager | member | viewer
	InvitedBy   *int64
	JoinedAt    time.Time
}

type WorkspaceInput struct {
	BusinessID  int64
	OwnerID     int64
	Name        string
	Description string
	Icon        string
	Color       string
	IsPersonal  bool
	CreatedBy   int64
}

type UpdateWorkspaceInput struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	Name        string
	Description string
	Icon        string
	Color       string
}

type AddWorkspaceMemberInput struct {
	WorkspaceID int64
	UserID      int64
	Role        string // manager | member | viewer
	InvitedBy   int64
}

type RemoveWorkspaceMemberInput struct {
	WorkspaceID int64
	UserID      int64
}

type UpdateWorkspaceMemberRoleInput struct {
	WorkspaceID int64
	UserID      int64
	Role        string
}
