package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/payroll/internal/application"
	"github.com/voxire/onion/internal/modules/payroll/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ─── List Payroll ─────────────────────────────────────────────────────────────

func ListPayrollRoutes() []string { return []string{"/hr/payroll"} }

type ListPayrollHandler struct {
	Service *application.Service
}

type ListPayrollRequest struct {
	PeriodLabel string `json:"period_label"`
}

type PayrollPayload struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"owner_id"`
	StaffID       int64     `json:"staff_id"`
	PeriodStart   time.Time `json:"period_start"`
	PeriodEnd     time.Time `json:"period_end"`
	PeriodLabel   string    `json:"period_label"`
	RegularHours  float64   `json:"regular_hours"`
	OvertimeHours float64   `json:"overtime_hours"`
	BasePay       float64   `json:"base_pay"`
	OvertimePay   float64   `json:"overtime_pay"`
	Bonuses       float64   `json:"bonuses"`
	Deductions    float64   `json:"deductions"`
	TaxWithheld   float64   `json:"tax_withheld"`
	NetPay        float64   `json:"net_pay"`
	Status        string    `json:"status"`
	PaidAt        time.Time `json:"paid_at"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
}

type ListPayrollResponse struct {
	Records []PayrollPayload `json:"records"`
}

func newPayrollPayload(p domain.PayrollRecord) PayrollPayload {
	return PayrollPayload{
		ID:            p.ID,
		OwnerID:       p.OwnerID,
		StaffID:       p.StaffID,
		PeriodStart:   p.PeriodStart,
		PeriodEnd:     p.PeriodEnd,
		PeriodLabel:   p.PeriodLabel,
		RegularHours:  p.RegularHours,
		OvertimeHours: p.OvertimeHours,
		BasePay:       p.BasePay,
		OvertimePay:   p.OvertimePay,
		Bonuses:       p.Bonuses,
		Deductions:    p.Deductions,
		TaxWithheld:   p.TaxWithheld,
		NetPay:        p.NetPay,
		Status:        p.Status,
		PaidAt:        p.PaidAt,
		Notes:         p.Notes,
		CreatedAt:     p.CreatedAt,
	}
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

	payloads := make([]PayrollPayload, 0, len(records))
	for _, p := range records {
		payloads = append(payloads, newPayrollPayload(p))
	}

	return ListPayrollResponse{Records: payloads}, nil
}

// ─── Run Payroll ──────────────────────────────────────────────────────────────

func RunPayrollRoutes() []string { return []string{"/hr/payroll/run"} }

type RunPayrollHandler struct {
	Service *application.Service
}

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
		OwnerID:     device.User.UserID,
		PeriodStart: req.PeriodStart,
		PeriodEnd:   req.PeriodEnd,
		PeriodLabel: req.PeriodLabel,
	}

	count, err := h.Service.RunPayroll(ctx, device.User.UserID, input)
	if err != nil {
		return RunPayrollResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to run payroll: " + err.Error()}
	}

	return RunPayrollResponse{RecordsCreated: count}, nil
}

// ─── Mark Payroll Paid ────────────────────────────────────────────────────────

func MarkPayrollPaidRoutes() []string { return []string{"/hr/payroll/mark-paid"} }

type MarkPayrollPaidHandler struct {
	Service *application.Service
}

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
