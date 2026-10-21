package domain

import "fmt"

// ValidateMemberLocationInput holds data needed to create a member-location association.
type MemberLocationInput struct {
	BusinessID         int64
	UserID             int64
	LocationID         int64
	CanManagePOSUsers  bool
	CanManageTerminals bool
	CanViewReports     bool
}

// Validate returns an error when the input is malformed.
func (m MemberLocationInput) Validate() error {
	if m.BusinessID == 0 {
		return fmt.Errorf("business id is required")
	}
	if m.UserID == 0 {
		return fmt.Errorf("user id is required")
	}
	if m.LocationID == 0 {
		return fmt.Errorf("location id is required")
	}
	return nil
}

// HasManagementPrivileges reports whether the member can perform any management action.
func (m MemberLocation) HasManagementPrivileges() bool {
	return m.CanManagePOSUsers || m.CanManageTerminals
}
