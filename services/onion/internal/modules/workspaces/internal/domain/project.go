package domain

import "time"

// Project visibility values. Private (the default) limits the project to its
// creator and members; public additionally exposes this struct's metadata on
// the owner's profile, never task access.
const (
	VisibilityPrivate = "private"
	VisibilityPublic  = "public"
)

func ValidVisibility(v string) bool {
	return v == VisibilityPrivate || v == VisibilityPublic
}

type Project struct {
	ID          int64
	WorkspaceID int64
	OwnerID     int64
	BusinessID  int64
	Name        string
	Description string
	Icon        string
	Color       string
	Status      string
	Visibility  string
	IsPersonal  bool
	StartDate   *time.Time
	DueDate     *time.Time
	CreatedBy   int64
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// Recycle-bin timestamps. ArchivedAt set means the project sits in the bin;
	// PurgeAt is the permanent-deletion deadline stamped from the business's
	// retention setting at archive time.
	ArchivedAt *time.Time
	PurgeAt    *time.Time

	// VisibleViaReport is a PER-CALLER projection, not a stored property: true
	// when this caller can see the project only because they manage someone
	// involved in it. The same project is false for its own members and true for
	// their manager, so it is computed per response and never written to the
	// database. The repository leaves it zero; the application layer sets it.
	VisibleViaReport bool

	// Display projections populated by list/detail reads.
	TaskCount      int32
	CompletedCount int32
	MemberCount    int32
	Code           string // e.g. "FLW" — used to build Task.Identifier ("FLW-42")
}

type ProjectTaskCounts struct {
	Total int
	Done  int
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
	BusinessID  int64
	Name        string
	Description string
	Icon        string
	Color       string
	Visibility  string // empty means VisibilityPrivate; validated by the service
	IsPersonal  bool   // true only for the auto-provisioned My Tasks bucket
	StartDate   *time.Time
	DueDate     *time.Time
	CreatedBy   int64
}

type UpdateProjectInput struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
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

type MoveProjectInput struct {
	ProjectID   int64
	WorkspaceID int64
	CallerID    int64
}

// FlowProfile is a business member's in-app profile: identity plus the projects
// they chose to make discoverable.
//
// Bio has no column behind it yet, so it is always empty. It stays on the
// struct because the contract declares it and dropping it would make the
// mapping silently lossy the day storage lands.
type FlowProfile struct {
	UserID         int64
	Name           string
	Email          string
	Initials       string
	RoleTitle      string
	Bio            string
	PublicProjects []ProfileProject
}

// ProfileProject is the metadata card for one public project. It carries no
// task-level fields on purpose: public means discoverable, not openable.
type ProfileProject struct {
	ID             int64
	Name           string
	Icon           string
	Color          string
	Description    string
	Status         string
	TaskCount      int32
	CompletedCount int32
}
