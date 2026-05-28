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
