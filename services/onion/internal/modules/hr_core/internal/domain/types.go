package domain

import "time"

type EmployeeRecord struct {
	ID                    int64     `json:"id"`
	OwnerID               int64     `json:"owner_id"`
	BusinessID            int64     `json:"business_id"`
	UserID                int64     `json:"user_id"`
	DepartmentID          int64     `json:"department_id"`
	LocationID            int64     `json:"location_id"`
	WorkScheduleID        int64     `json:"work_schedule_id"`
	FirstName             string    `json:"first_name"`
	LastName              string    `json:"last_name"`
	Email                 string    `json:"email"`
	Phone                 string    `json:"phone"`
	EmploymentType        string    `json:"employment_type"`
	Status                string    `json:"status"`
	SalaryCents           int64     `json:"salary_cents"`
	HourlyRateCents       int64     `json:"hourly_rate_cents"`
	OvertimeRateCents     int64     `json:"overtime_rate_cents"`
	StartDate             time.Time `json:"start_date"`
	EndDate               time.Time `json:"end_date"`
	Address               string    `json:"address"`
	EmergencyContactName  string    `json:"emergency_contact_name"`
	EmergencyContactPhone string    `json:"emergency_contact_phone"`
	Notes                 string    `json:"notes"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type DepartmentRecord struct {
	ID                    int64     `json:"id"`
	OwnerID               int64     `json:"owner_id"`
	BusinessID            int64     `json:"business_id"`
	LocationID            int64     `json:"location_id"`
	ParentDepartmentID    int64     `json:"parent_department_id"`
	Name                  string    `json:"name"`
	DefaultWorkScheduleID int64     `json:"default_work_schedule_id"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type LocationRecord struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	BusinessID int64     `json:"business_id"`
	Name       string    `json:"name"`
	Address    string    `json:"address"`
	City       string    `json:"city"`
	Country    string    `json:"country"`
	Timezone   string    `json:"timezone"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	IsDefault  bool      `json:"is_default"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type WorkScheduleRecord struct {
	ID           int64     `json:"id"`
	OwnerID      int64     `json:"owner_id"`
	BusinessID   int64     `json:"business_id"`
	Name         string    `json:"name"`
	ScheduleData string    `json:"schedule_data"`
	IsDefault    bool      `json:"is_default"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RoleRecord struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	BusinessID  int64     `json:"business_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Permissions string    `json:"permissions"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PermissionRecord struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

type EmployeeRoleRecord struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	BusinessID int64     `json:"business_id"`
	EmployeeID int64     `json:"employee_id"`
	RoleID     int64     `json:"role_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type SetupStepRecord struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	BusinessID  int64     `json:"business_id"`
	StepKey     string    `json:"step_key"`
	StepLabel   string    `json:"step_label"`
	IsCompleted bool      `json:"is_completed"`
	CompletedAt time.Time `json:"completed_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type HRCompanyConfigRecord struct {
	ID                    int64     `json:"id"`
	OwnerID               int64     `json:"owner_id"`
	BusinessID            int64     `json:"business_id"`
	CompanyName           string    `json:"company_name"`
	FiscalYearStartMonth  int32     `json:"fiscal_year_start_month"`
	DefaultCurrency       string    `json:"default_currency"`
	DefaultTimezone       string    `json:"default_timezone"`
	DefaultWorkScheduleID int64     `json:"default_work_schedule_id"`
	OvertimeEnabled       bool      `json:"overtime_enabled"`
	OvertimeMultiplier    float64   `json:"overtime_multiplier"`
	DirectoryEnabled      bool      `json:"directory_enabled"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type ApprovalChainRecord struct {
	ID          int64                      `json:"id"`
	OwnerID     int64                      `json:"owner_id"`
	BusinessID  int64                      `json:"business_id"`
	RequestType string                     `json:"request_type"`
	Name        string                     `json:"name"`
	Levels      []ApprovalChainLevelRecord `json:"levels"`
	CreatedAt   time.Time                  `json:"created_at"`
	UpdatedAt   time.Time                  `json:"updated_at"`
}

type ApprovalChainLevelRecord struct {
	ID                 int64     `json:"id"`
	OwnerID            int64     `json:"owner_id"`
	BusinessID         int64     `json:"business_id"`
	ApprovalChainID    int64     `json:"approval_chain_id"`
	LevelOrder         int32     `json:"level_order"`
	ApproverType       string    `json:"approver_type"`
	ApproverRoleID     int64     `json:"approver_role_id"`
	ApproverEmployeeID int64     `json:"approver_employee_id"`
	CreatedAt          time.Time `json:"created_at"`
}

type InviteRecord struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	BusinessID int64     `json:"business_id"`
	EmployeeID int64     `json:"employee_id"`
	Token      string    `json:"token"`
	ExpiresAt  time.Time `json:"expires_at"`
	ConsumedAt time.Time `json:"consumed_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type ListEmployeesParams struct {
	Limit        int32
	Offset       int32
	DepartmentID *int64
	LocationID   *int64
	Status       *string
}
