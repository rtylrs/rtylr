package handlers

import (
	"context"
	"net/http"

	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"

	"github.com/voxire/onion/internal/modules/payroll/internal/application"
	"github.com/voxire/onion/internal/modules/payroll/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func ListPayrollRoutes() []string      { return []string{"/hr/payroll"} }
func RunPayrollRoutes() []string       { return []string{"/hr/payroll/run"} }
func MarkPayrollPaidRoutes() []string  { return []string{"/hr/payroll/mark-paid"} }

func newPayrollPayload(p domain.PayrollRecord) *hrv1.PayrollRecord {
	return &hrv1.PayrollRecord{
		Id: p.ID, OwnerId: p.OwnerID, StaffId: p.StaffID,
		PeriodStart: p.PeriodStart.Format("2006-01-02"), PeriodEnd: p.PeriodEnd.Format("2006-01-02"),
		PeriodLabel: p.PeriodLabel, RegularHours: p.RegularHours, OvertimeHours: p.OvertimeHours,
		BasePay: p.BasePay, OvertimePay: p.OvertimePay, Bonuses: p.Bonuses,
		Deductions: p.Deductions, TaxWithheld: p.TaxWithheld, NetPay: p.NetPay,
		Status: p.Status, PaidAt: p.PaidAt.Format("2006-01-02T15:04:05Z"),
		Notes: p.Notes, CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ─── List ────────────────────────────────────────────────────────────────────

type ListPayrollHandler struct{ Service *application.Service }
type ListPayrollRequest struct {
	PeriodLabel string `json:"period_label"`
}
type ListPayrollResponse struct {
	Records []*hrv1.PayrollRecord `json:"records"`
}

func (h *ListPayrollHandler) Serve(ctx context.Context, req *ListPayrollRequest) (ListPayrollResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListPayrollResponse{}, errorModels.PermissionDeniedServerError()
	}
	records, err := h.Service.ListPayrollRecords(ctx, device.User.UserID, req.PeriodLabel)
	if err != nil {
		return ListPayrollResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list payroll records"}
	}
	payloads := make([]*hrv1.PayrollRecord, 0, len(records))
	for _, p := range records {
		payloads = append(payloads, newPayrollPayload(p))
	}
	return ListPayrollResponse{Records: payloads}, nil
}

// ─── Run Payroll ─────────────────────────────────────────────────────────────

type RunPayrollHandler struct{ Service *application.Service }

type RunPayrollRequest struct {
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
	PeriodLabel string `json:"period_label"`
}

type RunPayrollResponse struct {
	RecordsCreated int `json:"records_created"`
}

func (h *RunPayrollHandler) Validate(_ context.Context, req *RunPayrollRequest) *errorModels.ServerError {
	if req.PeriodStart == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "period_start is required"}
	}
	if req.PeriodEnd == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "period_end is required"}
	}
	if req.PeriodLabel == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "period_label is required"}
	}
	return nil
}

func (h *RunPayrollHandler) Serve(ctx context.Context, req *RunPayrollRequest) (RunPayrollResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return RunPayrollResponse{}, errorModels.PermissionDeniedServerError()
	}
	input := domain.PayrollRunInput{
		OwnerID: device.User.UserID, PeriodStart: req.PeriodStart,
		PeriodEnd: req.PeriodEnd, PeriodLabel: req.PeriodLabel,
	}
	count, err := h.Service.RunPayroll(ctx, device.User.UserID, input)
	if err != nil {
		return RunPayrollResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to run payroll: " + err.Error()}
	}
	return RunPayrollResponse{RecordsCreated: count}, nil
}

// ─── Mark Payroll Paid ───────────────────────────────────────────────────────

type MarkPayrollPaidHandler struct{ Service *application.Service }

type MarkPayrollPaidRequest struct {
	ID int64 `json:"id"`
}

type MarkPayrollPaidResponse struct {
	Success bool `json:"success"`
}

func (h *MarkPayrollPaidHandler) Validate(_ context.Context, req *MarkPayrollPaidRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "id is required"}
	}
	return nil
}

func (h *MarkPayrollPaidHandler) Serve(ctx context.Context, req *MarkPayrollPaidRequest) (MarkPayrollPaidResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return MarkPayrollPaidResponse{}, errorModels.PermissionDeniedServerError()
	}
	if err := h.Service.MarkPayrollPaid(ctx, device.User.UserID, req.ID); err != nil {
		return MarkPayrollPaidResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to mark payroll as paid: " + err.Error()}
	}
	return MarkPayrollPaidResponse{Success: true}, nil
}
