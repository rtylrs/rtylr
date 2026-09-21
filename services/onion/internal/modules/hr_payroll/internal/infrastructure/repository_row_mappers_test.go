package infrastructure

import (
	"database/sql"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr_payroll/internal/infrastructure/persistence/sqlc"
)

// GetPayrollRun feeds ExportPayrollToFinance, which passes run.BusinessID to the
// finance journal poster as the journal's tenant scope. The mapper used to drop
// the column, so every authorized export posted business_id 0 and the journal
// rejected it. Owner and business are deliberately different here: the business
// scope must come from business_id, not from owner_id.
func TestPayrollRunFromRowCarriesBusinessID(t *testing.T) {
	const (
		ownerID    int64 = 1001
		businessID int64 = 2002
	)
	row := sqlc.HrPayrollRun{
		ID:          4301,
		OwnerID:     ownerID,
		BusinessID:  businessID,
		PeriodStart: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC),
		Status:      "approved",
		CreatedBy:   ownerID,
		ApprovedBy:  sql.NullInt64{Int64: ownerID, Valid: true},
		TotalGross:  "2400.00",
		TotalNet:    "2200.00",
		CreatedAt:   time.Date(2026, 3, 31, 12, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 3, 31, 12, 0, 0, 0, time.UTC),
	}

	got := payrollRunFromRow(row)
	if got.BusinessID != businessID {
		t.Errorf("BusinessID = %d, want %d; the finance journal poster rejects business_id 0", got.BusinessID, businessID)
	}
	if got.OwnerID != ownerID {
		t.Errorf("OwnerID = %d, want %d", got.OwnerID, ownerID)
	}
	if got.Status != "approved" || got.TotalGrossCents != 240000 || got.TotalNetCents != 220000 {
		t.Errorf("row mapping changed: %+v", got)
	}
}
