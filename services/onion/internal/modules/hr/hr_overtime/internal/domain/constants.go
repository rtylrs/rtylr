package domain

// Overtime request status values.
const (
	OvertimeStatusPending  = "pending"
	OvertimeStatusApproved = "approved"
	OvertimeStatusRejected = "rejected"

	// PermissionManageEmployees allows an HR administrator to create a
	// request for another employee; self-service requests are allowed for the
	// caller's own employee record.
	PermissionManageEmployees = "manage_employees"
	// PermissionManageHRConfig gates tenant-wide overtime policy mutations.
	PermissionManageHRConfig = "manage_hr_config"
)
