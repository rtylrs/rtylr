package domain

// Employment status values used across the HR modules.
const (
	EmploymentStatusActive   = "active"
	EmploymentStatusInactive = "inactive"
	EmploymentStatusInvited  = "invited"
)

// Employment type values.
const (
	EmploymentTypeFullTime  = "full_time"
	EmploymentTypePartTime  = "part_time"
	EmploymentTypeContract  = "contract"
	EmploymentTypeFreelance = "freelance"
)

// Approval request types used in approval chains.
const (
	ApprovalRequestTypeLeave     = "leave"
	ApprovalRequestTypeOvertime  = "overtime"
	ApprovalRequestTypeEquipment = "equipment"
	ApprovalRequestTypePayroll   = "payroll"
)

// Fine-grained HR permission keys. A role's Permissions JSON (hr_roles.permissions)
// grants a subset of these; HasPermission checks a key from this set. Keep this
// list and AllPermissions() (permissions_catalog.go) in sync — SeedDefaultRoles
// (application/service.go) only ever grants keys from here, so the two can never
// drift apart as long as both reference these constants instead of string literals.
const (
	PermissionViewOwnProfile       = "view_own_profile"
	PermissionViewOwnSchedule      = "view_own_schedule"
	PermissionRequestTimeOff       = "request_time_off"
	PermissionViewTeam             = "view_team"
	PermissionApproveTimeOff       = "approve_time_off"
	PermissionApproveOvertime      = "approve_overtime"
	PermissionManageSchedules      = "manage_schedules"
	PermissionManageEmployees      = "manage_employees"
	PermissionManageRoles          = "manage_roles"
	PermissionManageDepartments    = "manage_departments"
	PermissionManageLocations      = "manage_locations"
	PermissionManageHRConfig       = "manage_hr_config"
	PermissionViewReports          = "view_reports"
	PermissionManagePayroll        = "manage_payroll"
	PermissionManageBenefits       = "manage_benefits"
	PermissionManageApprovalChains = "manage_approval_chains"
	PermissionViewDirectory        = "view_directory"
)
