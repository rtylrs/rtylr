package api

import (
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/domain"
)

func TestNewLeaveBalancesResponseIncludesDeclaredMetadata(t *testing.T) {
	response := newLeaveBalancesResponse([]domain.LeaveBalanceRecord{{
		ID: 1, EmployeeID: 2, LeaveTypeID: 3, Year: 2026, EntitledDays: 10,
	}}, 7)
	if len(response.Balances) != 1 || response.Total != 7 || response.Limit != leaveBalancesPageLimit || response.Offset != 0 {
		t.Fatalf("metadata = total %d limit %d offset %d balances %d", response.Total, response.Limit, response.Offset, len(response.Balances))
	}
}
