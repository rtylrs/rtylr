package domain

// AllPermissions returns the full catalog of fine-grained HR permission keys
// a role's Permissions JSON may grant. This is the single source of truth for
// what a role can be given — SeedDefaultRoles (application/service.go) builds
// its hardcoded default role sets from these same constants, so the catalog
// and what actually gets seeded never drift apart. Enforcement call sites in
// other modules (e.g. hr_payroll) cannot import these constants across Go's
// internal-package boundary and still use bare string literals — this catalog
// does not (yet) guarantee those stay in sync.
func AllPermissions() []PermissionRecord {
	return []PermissionRecord{
		{Key: PermissionViewOwnProfile, Label: "View own profile", Description: "See your own employee profile", Category: "general"},
		{Key: PermissionViewOwnSchedule, Label: "View own schedule", Description: "See your own work schedule", Category: "general"},
		{Key: PermissionRequestTimeOff, Label: "Request time off", Description: "Submit time-off requests", Category: "time_off"},
		{Key: PermissionViewTeam, Label: "View team", Description: "See other employees on the team", Category: "employees"},
		{Key: PermissionApproveTimeOff, Label: "Approve time off", Description: "Approve or reject time-off requests", Category: "time_off"},
		{Key: PermissionApproveOvertime, Label: "Approve overtime", Description: "Approve or reject overtime requests", Category: "payroll"},
		{Key: PermissionManageSchedules, Label: "Manage schedules", Description: "Create and edit work schedules", Category: "scheduling"},
		{Key: PermissionManageEmployees, Label: "Manage employees", Description: "Create and edit employee records", Category: "employees"},
		{Key: PermissionManageRoles, Label: "Manage roles", Description: "Create and edit roles and what they can do", Category: "administration"},
		{Key: PermissionManageDepartments, Label: "Manage departments", Description: "Create and edit departments", Category: "administration"},
		{Key: PermissionManageLocations, Label: "Manage locations", Description: "Create and edit locations", Category: "administration"},
		{Key: PermissionManageHRConfig, Label: "Manage HR configuration", Description: "Edit company-wide HR settings", Category: "administration"},
		{Key: PermissionViewReports, Label: "View reports", Description: "Access HR reports", Category: "reports"},
		{Key: PermissionManagePayroll, Label: "Manage payroll", Description: "Run and edit payroll", Category: "payroll"},
		{Key: PermissionManageBenefits, Label: "Manage benefits", Description: "Edit employee benefits", Category: "payroll"},
		{Key: PermissionManageApprovalChains, Label: "Manage approval chains", Description: "Configure approval chains for requests", Category: "administration"},
		{Key: PermissionViewDirectory, Label: "View directory", Description: "See the company directory", Category: "general"},
	}
}
