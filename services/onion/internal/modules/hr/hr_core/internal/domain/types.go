package domain

import "time"

type EmployeeRecord struct {
	ID                    int64
	OwnerID               int64
	BusinessID            int64
	UserID                int64
	DepartmentID          int64
	LocationID            int64
	WorkScheduleID        int64
	FirstName             string
	LastName              string
	Email                 string
	Phone                 string
	EmploymentType        string
	Status                string
	SalaryCents           int64
	HourlyRateCents       int64
	OvertimeRateCents     int64
	StartDate             time.Time
	EndDate               time.Time
	Address               string
	EmergencyContactName  string
	EmergencyContactPhone string
	Notes                 string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type TerminatedEmployee struct {
	BusinessID int64
	UserID     int64
}

type DepartmentRecord struct {
	ID                    int64
	OwnerID               int64
	BusinessID            int64
	LocationID            int64
	ParentDepartmentID    int64
	Name                  string
	DefaultWorkScheduleID int64
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type LocationRecord struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	Name       string
	Address    string
	City       string
	Country    string
	Timezone   string
	Latitude   float64
	Longitude  float64
	IsDefault  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type PublicLocationRecord struct {
	ID       int64
	Name     string
	Slug     string
	Address  string
	Timezone string
	IsActive bool
}

type WorkScheduleRecord struct {
	ID           int64
	OwnerID      int64
	BusinessID   int64
	Name         string
	ScheduleData string
	IsDefault    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RoleRecord struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	Name        string
	Description string
	Permissions string
	IsSystem    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PermissionRecord struct {
	Key         string
	Label       string
	Description string
	Category    string
}
type SetupStepRecord struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	StepKey     string
	StepLabel   string
	IsCompleted bool
	CompletedAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type HRCompanyConfigRecord struct {
	ID                    int64
	OwnerID               int64
	BusinessID            int64
	CompanyName           string
	FiscalYearStartMonth  int32
	DefaultCurrency       string
	DefaultTimezone       string
	DefaultWorkScheduleID int64
	OvertimeEnabled       bool
	OvertimeMultiplier    float64
	DirectoryEnabled      bool
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type ApprovalChainRecord struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	RequestType string
	Name        string
	Levels      []ApprovalChainLevelRecord
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ApprovalChainLevelRecord struct {
	ID                 int64
	OwnerID            int64
	BusinessID         int64
	ApprovalChainID    int64
	LevelOrder         int32
	ApproverType       string
	ApproverRoleID     int64
	ApproverEmployeeID int64
	CreatedAt          time.Time
}

type InviteRecord struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	EmployeeID int64
	Token      string
	ExpiresAt  time.Time
	ConsumedAt time.Time
	CreatedAt  time.Time
}

type ListEmployeesParams struct {
	Limit        int32
	Offset       int32
	DepartmentID *int64
	LocationID   *int64
	Status       *string
}
