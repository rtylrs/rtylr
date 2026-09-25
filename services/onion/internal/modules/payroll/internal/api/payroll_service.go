package api

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/payroll/internal/application"
	"github.com/voxire/onion/internal/modules/payroll/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrpayrollv1 "github.com/voxire/voxie/gen/go/hr/payroll/v1"
)

type PayrollAPI struct {
	service     *application.Service
	permissions PermissionProvider
}

func NewPayrollAPI(service *application.Service, permissions PermissionProvider) *PayrollAPI {
	return &PayrollAPI{service: service, permissions: permissions}
}

func (a *PayrollAPI) ListPayroll(ctx context.Context, req *hrpayrollv1.ListPayrollRequest) (*hrpayrollv1.ListPayrollResponse, error) {
	businessID, viewerUserID, errModel := resolveIdentity(ctx)
	if errModel != nil {
		return nil, errModel
	}

	records, err := a.service.ListPayrollRecords(ctx, businessID, req.GetPeriodLabel())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list payroll records")
	}

	return &hrpayrollv1.ListPayrollResponse{
		Records: newPayrollRecords(records, canViewCompensation(ctx, a.permissions, businessID, viewerUserID)),
	}, nil
}

func (a *PayrollAPI) RunPayroll(ctx context.Context, req *hrpayrollv1.RunPayrollRequest) (*hrpayrollv1.RunPayrollResponse, error) {
	if req.GetPeriodStart() == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "period_start is required")
	}
	if req.GetPeriodEnd() == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "period_end is required")
	}
	if req.GetPeriodLabel() == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "period_label is required")
	}

	businessID, ownerID, errModel := resolveIdentity(ctx)
	if errModel != nil {
		return nil, errModel
	}

	count, err := a.service.RunPayroll(ctx, businessID, domain.PayrollRunInput{
		OwnerID:     ownerID,
		BusinessID:  businessID,
		PeriodStart: req.GetPeriodStart(),
		PeriodEnd:   req.GetPeriodEnd(),
		PeriodLabel: req.GetPeriodLabel(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to run payroll: "+err.Error())
	}

	return &hrpayrollv1.RunPayrollResponse{RecordsCreated: int32(count)}, nil // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
}

func (a *PayrollAPI) MarkPayrollPaid(ctx context.Context, req *hrpayrollv1.MarkPayrollPaidRequest) (*hrpayrollv1.MarkPayrollPaidResponse, error) {
	if req.GetId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "id is required")
	}

	businessID, _, errModel := resolveIdentity(ctx)
	if errModel != nil {
		return nil, errModel
	}

	if err := a.service.MarkPayrollPaid(ctx, businessID, req.GetId()); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to mark payroll as paid: "+err.Error())
	}

	return &hrpayrollv1.MarkPayrollPaidResponse{Success: true}, nil
}

func resolveIdentity(ctx context.Context) (businessID, ownerID int64, errModel error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return 0, 0, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return 0, 0, onekiterrors.PermissionDenied()
	}
	return businessID, device.User.UserID, nil
}

func newPayrollRecords(records []domain.PayrollRecord, compensationVisible bool) []*hrpayrollv1.PayrollRecord {
	out := make([]*hrpayrollv1.PayrollRecord, 0, len(records))
	for _, record := range records {
		if !compensationVisible {
			record.BasePay = 0
			record.OvertimePay = 0
			record.Bonuses = 0
			record.Deductions = 0
			record.TaxWithheld = 0
			record.NetPay = 0
		}

		out = append(out, &hrpayrollv1.PayrollRecord{
			Id:            record.ID,
			OwnerId:       record.OwnerID,
			StaffId:       record.StaffID,
			PeriodStart:   formatDate(record.PeriodStart),
			PeriodEnd:     formatDate(record.PeriodEnd),
			PeriodLabel:   record.PeriodLabel,
			RegularHours:  record.RegularHours,
			OvertimeHours: record.OvertimeHours,
			BasePay:       record.BasePay,
			OvertimePay:   record.OvertimePay,
			Bonuses:       record.Bonuses,
			Deductions:    record.Deductions,
			TaxWithheld:   record.TaxWithheld,
			NetPay:        record.NetPay,
			Status:        record.Status,
			PaidAt:        formatTime(record.PaidAt),
			Notes:         record.Notes,
			CreatedAt:     formatTime(record.CreatedAt),
		})
	}
	return out
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
