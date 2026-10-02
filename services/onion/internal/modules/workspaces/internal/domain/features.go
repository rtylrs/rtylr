package domain

import "time"

// ─────────────────────────────────────────────
// Cycles
// ─────────────────────────────────────────────

type Cycle struct {
	ID          int64
	WorkspaceID int64
	ProjectID   *int64
	Name        string
	Description string
	Status      string // upcoming | active | completed
	StartDate   time.Time
	EndDate     time.Time
	Scope       int // total tasks in cycle
	Completed   int // done tasks in cycle
	CreatedBy   int64
	CreatedAt   time.Time
}

type CycleInput struct {
	BusinessID  int64
	WorkspaceID int64
	ProjectID   *int64
	Name        string
	Description string
	StartDate   time.Time
	EndDate     time.Time
	CreatedBy   int64
	Status      string
}

type UpdateCycleInput struct {
	ID         int64
	ActorID    int64
	BusinessID int64
	Name       string
	Status     string
	EndDate    time.Time
}

// ─────────────────────────────────────────────
// Milestones
// ─────────────────────────────────────────────

type Milestone struct {
	ID          int64
	ProjectID   int64
	WorkspaceID int64
	Name        string
	Description string
	Status      string // planning | active | at_risk | completed
	DueDate     time.Time
	CompletedAt *time.Time
	TaskCount   int
	DoneCount   int
	Progress    int // 0–100
	Project     *MilestoneProject
	CreatedBy   int64
	CreatedAt   time.Time
}

type MilestoneProject struct {
	Name  string
	Color string
}

type MilestoneInput struct {
	ProjectID   int64
	WorkspaceID int64
	BusinessID  int64
	Name        string
	Description string
	DueDate     time.Time
	CreatedBy   int64
}

type UpdateMilestoneInput struct {
	ID          int64
	ActorID     int64
	BusinessID  int64
	Name        string
	Description string
	Status      string
	DueDate     time.Time
}

// ─────────────────────────────────────────────
// Views
// ─────────────────────────────────────────────

type View struct {
	ID          int64
	WorkspaceID int64
	Name        string
	Icon        string
	Filters     map[string]interface{}
	GroupBy     string
	SortBy      string
	SortDir     string
	Visibility  string
	CreatedBy   int64
	CreatedAt   time.Time
}

type ViewInput struct {
	BusinessID  int64
	WorkspaceID int64
	Name        string
	Icon        string
	Filters     map[string]interface{}
	GroupBy     string
	SortBy      string
	SortDir     string
	Visibility  string
	CreatedBy   int64
}

type UpdateViewInput struct {
	ID         int64
	ActorID    int64
	BusinessID int64
	Name       string
	Icon       string
	Filters    map[string]interface{}
	GroupBy    string
	SortBy     string
	SortDir    string
	Visibility string
}

// ─────────────────────────────────────────────
// Notifications
// ─────────────────────────────────────────────

type Notification struct {
	ID            int64
	UserID        int64
	WorkspaceID   *int64
	Type          string // assigned | mention | comment | status_change | due_soon
	Title         string
	Description   string
	TaskID        *int64
	TaskTitle     string
	ActorID       *int64
	ActorName     string
	ActorInitials string
	Project       string
	ProjectColor  string
	IsRead        bool
	CreatedAt     time.Time
}

type CreateNotificationInput struct {
	UserID        int64
	WorkspaceID   *int64
	Type          string
	Title         string
	Description   string
	TaskID        *int64
	TaskTitle     string
	ActorID       *int64
	ActorName     string
	ActorInitials string
	Project       string
	ProjectColor  string
}

// ─────────────────────────────────────────────
// Insights
// ─────────────────────────────────────────────

type InsightsVelocityPoint struct {
	Cycle          string
	Completed      int
	Total          int
	CompletionRate float64
}

type InsightsBurndownPoint struct {
	Day    int
	Ideal  float64
	Actual float64
}

type InsightsMemberTime struct {
	ID    string
	Name  string
	Hours float64
}

type InsightsThroughput struct {
	Week      string
	Completed int
}

type Insights struct {
	Velocity     []InsightsVelocityPoint
	Burndown     []InsightsBurndownPoint
	CycleTime    float64
	TimeByMember []InsightsMemberTime
	Throughput   []InsightsThroughput
	TotalTasks   int
	DoneTasks    int
}

// ─────────────────────────────────────────────
// Team Workload
// ─────────────────────────────────────────────

type TeamMemberWorkload struct {
	ID            string
	Name          string
	Initials      string
	Capacity      int
	AssignedHours float64
	Load          float64
	Tasks         []Task
}

// ─────────────────────────────────────────────
// User Settings
// ─────────────────────────────────────────────

type UserSettings struct {
	UserID                 int64
	TaskAssigned           bool
	TaskComments           bool
	Mentions               bool
	StatusChanges          bool
	WeeklyDigest           bool
	EmailNotifications     bool
	SidebarPreferencesJSON string
}
