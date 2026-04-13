package domain

import "time"

type Project struct {
	ID          int64
	WorkspaceID int64
	OwnerID     int64
	Name        string
	Description string
	Icon        string
	Color       string
	Status      string
	StartDate   *time.Time
	DueDate     *time.Time
	CreatedBy   int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProjectMember struct {
	ID        int64
	ProjectID int64
	UserID    int64
	Name      string
	Email     string
	Role      string // manager | member | viewer
	AddedBy   *int64
	AddedAt   time.Time
}

type ProjectInput struct {
	WorkspaceID int64
	OwnerID     int64
	Name        string
	Description string
	Icon        string
	Color       string
	StartDate   *time.Time
	DueDate     *time.Time
	CreatedBy   int64
}

type UpdateProjectInput struct {
	ID          int64
	OwnerID     int64
	Name        string
	Description string
	Icon        string
	Color       string
	Status      string
	StartDate   *time.Time
	DueDate     *time.Time
}

type AddProjectMemberInput struct {
	ProjectID int64
	UserID    int64
	Role      string
	AddedBy   int64
}

type RemoveProjectMemberInput struct {
	ProjectID int64
	UserID    int64
}

type UpdateProjectMemberRoleInput struct {
	ProjectID int64
	UserID    int64
	Role      string
}

type TransferProjectOwnershipInput struct {
	ProjectID  int64
	NewOwnerID int64
	CallerID   int64
}
