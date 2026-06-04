package handlers

import (
	"github.com/voxire/onion/internal/modules/hr_payroll/internal/domain"
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

	return &hremployeespayrollrunsv1.PayrollRunPayload{
		Id:          r.ID,
		OwnerId:     r.OwnerID,
		PeriodStart: r.PeriodStart,
		PeriodEnd:   r.PeriodEnd,
		Status:      r.Status,
		CreatedBy:   r.CreatedBy,
		ApprovedBy:  approvedBy,
		ExportedAt:  exportedAt,
		TotalGross:  r.TotalGross,
		TotalNet:    r.TotalNet,
		CreatedAt:   r.CreatedAt.Unix(),
		UpdatedAt:   r.UpdatedAt.Unix(),
	}
}

func payrollEntryToProto(e domain.PayrollEntryRecord) *hremployeespayrollrunsv1.PayrollEntryPayload {
	return &hremployeespayrollrunsv1.PayrollEntryPayload{
		Id:            e.ID,
		OwnerId:       e.OwnerID,
		PayrollRunId:  e.PayrollRunID,
		EmployeeId:    e.EmployeeID,
		BasePay:       e.BasePay,
		OvertimePay:   e.OvertimePay,
		Deductions:    e.Deductions,
		Bonuses:       e.Bonuses,
		NetPay:        e.NetPay,
		HoursWorked:   e.HoursWorked,
		OvertimeHours: e.OvertimeHours,
		CreatedAt:     e.CreatedAt.Unix(),
		UpdatedAt:     e.UpdatedAt.Unix(),
	}
}
