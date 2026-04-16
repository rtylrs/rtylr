package handlers

import (
	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"

	"github.com/voxire/onion/internal/modules/hr_payroll/internal/domain"
)

func payrollRunToProto(r domain.PayrollRunRecord) *hrv1.PayrollRunPayload {
	var approvedBy *int64
	var exportedAt *int64

	if r.ApprovedBy != 0 {
		approvedBy = &r.ApprovedBy
	}
	if !r.ExportedAt.IsZero() {
		ts := r.ExportedAt.Unix()
		exportedAt = &ts
	}

	return &hrv1.PayrollRunPayload{
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

func payrollEntryToProto(e domain.PayrollEntryRecord) *hrv1.PayrollEntryPayload {
	return &hrv1.PayrollEntryPayload{
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
