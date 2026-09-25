package domain

import "fmt"

// ValidateAuditLogInput returns an error when the audit log input is malformed.
func ValidateAuditLogInput(input AuditLogInput) error {
	if input.BusinessID == 0 {
		return fmt.Errorf("business id is required")
	}
	if input.BranchID == 0 {
		return fmt.Errorf("branch id is required")
	}
	if input.EmployeeID == 0 {
		return fmt.Errorf("employee id is required")
	}
	if input.Action == "" {
		return fmt.Errorf("action is required")
	}
	if input.EntityType == "" {
		return fmt.Errorf("entity type is required")
	}
	return nil
}

// ValidateAuditLogFilter returns an error when the filter is malformed.
func ValidateAuditLogFilter(filter AuditLogFilter) error {
	if filter.BusinessID == 0 {
		return fmt.Errorf("business id is required for audit log filter")
	}
	return nil
}
