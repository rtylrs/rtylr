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
	Role      string
	JoinedAt  time.Time
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
