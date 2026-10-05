package api

import (
	"github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/domain"
	"github.com/voxire/onion/internal/platform/money"
	hremployeespayrollrunsv1 "github.com/voxire/voxie/gen/go/hr/employees/payroll/runs/v1"
)

func payrollRunToProto(r domain.PayrollRunRecord) *hremployeespayrollrunsv1.PayrollRunPayload {
	var approvedBy *int64
	var exportedAt *int64

	if r.ApprovedBy != 0 {
		approvedBy = &r.ApprovedBy
	}
	if !r.ExportedAt.IsZero() {
		ts := r.ExportedAt.Unix()
		exportedAt = &ts
	}

	// The generated proto still carries these amounts as float64 dollars
	// (that contract is owned by Voxire and out of scope for this fix), so
	// the internally-stored cents are converted at this boundary only.
	return &hremployeespayrollrunsv1.PayrollRunPayload{
		Id:          r.ID,
		OwnerId:     r.OwnerID,
		PeriodStart: r.PeriodStart,
		PeriodEnd:   r.PeriodEnd,
		Status:      r.Status,
		CreatedBy:   r.CreatedBy,
		ApprovedBy:  approvedBy,
		ExportedAt:  exportedAt,
		TotalGross:  money.CentsToDollars(r.TotalGrossCents),
		TotalNet:    money.CentsToDollars(r.TotalNetCents),
		CreatedAt:   r.CreatedAt.Unix(),
		UpdatedAt:   r.UpdatedAt.Unix(),
	}
}

func payrollEntryToProto(e domain.PayrollEntryRecord) *hremployeespayrollrunsv1.PayrollEntryPayload {
	// Same wire boundary conversion as payrollRunToProto above.
	return &hremployeespayrollrunsv1.PayrollEntryPayload{
		Id:            e.ID,
		OwnerId:       e.OwnerID,
		PayrollRunId:  e.PayrollRunID,
		EmployeeId:    e.EmployeeID,
		BasePay:       money.CentsToDollars(e.BasePayCents),
		OvertimePay:   money.CentsToDollars(e.OvertimePayCents),
		Deductions:    money.CentsToDollars(e.DeductionsCents),
		Bonuses:       money.CentsToDollars(e.BonusesCents),
		NetPay:        money.CentsToDollars(e.NetPayCents),
		HoursWorked:   e.HoursWorked,
		OvertimeHours: e.OvertimeHours,
		CreatedAt:     e.CreatedAt.Unix(),
		UpdatedAt:     e.UpdatedAt.Unix(),
	}
}
