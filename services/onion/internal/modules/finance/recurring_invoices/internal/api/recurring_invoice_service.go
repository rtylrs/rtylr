package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/finance/recurring_invoices/internal/application"
	"github.com/voxire/onion/internal/modules/finance/recurring_invoices/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	financerecurringinvoicev1 "github.com/voxire/voxie/gen/go/finance/recurring_invoice/v1"
)

type RecurringInvoiceAPI struct {
	Service *application.Service
}

func NewRecurringInvoiceAPI(service *application.Service) *RecurringInvoiceAPI {
	return &RecurringInvoiceAPI{Service: service}
}

func requireRecurringInvoicePermission(ctx context.Context, permission string) error {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil || !authPkg.HasAppPermission(device.User, authorization.AppFinance, permission) {
		return onekiterrors.PermissionDenied()
	}
	return nil
}

func recurringInvoiceBusinessID(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return 0, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}

func recurringInvoiceToProto(row domain.RecurringInvoice) *financerecurringinvoicev1.RecurringInvoice {
	var endDate *string
	if !row.EndDate.IsZero() {
		v := row.EndDate.Format("2006-01-02")
		endDate = &v
	}
	return &financerecurringinvoicev1.RecurringInvoice{
		Id:                row.ID,
		OwnerId:           row.OwnerID,
		TemplateInvoiceId: row.TemplateInvoiceID,
		Frequency:         row.Frequency,
		NextInvoiceDate:   row.NextInvoiceDate.Format("2006-01-02"),
		EndDate:           endDate,
		IsActive:          row.IsActive,
		CreatedAt:         row.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func recurringInvoiceInput(templateID *int64, frequency, nextInvoiceDate string, endDate *string, isActive *bool) (domain.RecurringInvoiceInput, error) {
	frequency = strings.TrimSpace(frequency)
	nextInvoiceDate = strings.TrimSpace(nextInvoiceDate)
	// A nil template_invoice_id means the update did not include it; the
	// repository then keeps the stored template.
	if templateID != nil && *templateID <= 0 {
		return domain.RecurringInvoiceInput{}, onekiterrors.New(http.StatusBadRequest, "template_invoice_id must be greater than zero")
	}
	if frequency == "" {
		return domain.RecurringInvoiceInput{}, onekiterrors.New(http.StatusBadRequest, "frequency is required")
	}
	if nextInvoiceDate == "" {
		return domain.RecurringInvoiceInput{}, onekiterrors.New(http.StatusBadRequest, "next_invoice_date is required")
	}
	if _, err := time.Parse("2006-01-02", nextInvoiceDate); err != nil {
		return domain.RecurringInvoiceInput{}, onekiterrors.New(http.StatusBadRequest, "next_invoice_date must be in YYYY-MM-DD format")
	}
	// end_date stays nil when the caller omitted it, so update keeps the stored
	// bound instead of widening the schedule to open-ended. A malformed value
	// is rejected rather than silently dropped.
	if endDate != nil {
		trimmed := strings.TrimSpace(*endDate)
		if trimmed != "" {
			if _, err := time.Parse("2006-01-02", trimmed); err != nil {
				return domain.RecurringInvoiceInput{}, onekiterrors.New(http.StatusBadRequest, "end_date must be in YYYY-MM-DD format")
			}
		}
		endDate = &trimmed
	}
	return domain.RecurringInvoiceInput{
		TemplateInvoiceID: templateID,
		Frequency:         frequency,
		NextInvoiceDate:   nextInvoiceDate,
		EndDate:           endDate,
		IsActive:          isActive,
	}, nil
}

func (a *RecurringInvoiceAPI) ListRecurringInvoices(ctx context.Context, req *financerecurringinvoicev1.ListRecurringInvoicesRequest) (*financerecurringinvoicev1.ListRecurringInvoicesResponse, error) {
	if err := requireRecurringInvoicePermission(ctx, "finance:recurring_invoices:view"); err != nil {
		return nil, err
	}
	businessID, err := recurringInvoiceBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	page, pageSize := req.GetPage(), req.GetPageSize()
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 500 {
		pageSize = 500
	}
	rows, hasMore, err := a.Service.List(ctx, businessID, page, pageSize)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list recurring invoices")
	}
	out := make([]*financerecurringinvoicev1.RecurringInvoice, 0, len(rows))
	for _, row := range rows {
		out = append(out, recurringInvoiceToProto(row))
	}
	return &financerecurringinvoicev1.ListRecurringInvoicesResponse{
		RecurringInvoices: out,
		Page:              page,
		PageSize:          pageSize,
		HasMore:           hasMore,
	}, nil
}

func (a *RecurringInvoiceAPI) CreateRecurringInvoice(ctx context.Context, req *financerecurringinvoicev1.CreateRecurringInvoiceRequest) (*financerecurringinvoicev1.CreateRecurringInvoiceResponse, error) {
	if err := requireRecurringInvoicePermission(ctx, "finance:recurring_invoices:manage"); err != nil {
		return nil, err
	}
	businessID, err := recurringInvoiceBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	// template_invoice_id is optional on update but required to create.
	if req.TemplateInvoiceId == nil || *req.TemplateInvoiceId <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "template_invoice_id is required")
	}
	input, err := recurringInvoiceInput(req.TemplateInvoiceId, req.GetFrequency(), req.GetNextInvoiceDate(), req.EndDate, req.IsActive)
	if err != nil {
		return nil, err
	}
	id, err := a.Service.Create(ctx, businessID, input)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create recurring invoice")
	}
	return &financerecurringinvoicev1.CreateRecurringInvoiceResponse{Id: id, Message: "Recurring invoice created successfully"}, nil
}

func (a *RecurringInvoiceAPI) GetRecurringInvoice(ctx context.Context, req *financerecurringinvoicev1.GetRecurringInvoiceRequest) (*financerecurringinvoicev1.GetRecurringInvoiceResponse, error) {
	if err := requireRecurringInvoicePermission(ctx, "finance:recurring_invoices:view"); err != nil {
		return nil, err
	}
	businessID, err := recurringInvoiceBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := a.Service.Get(ctx, businessID, req.GetId())
	if errors.Is(err, domain.ErrRecurringInvoiceNotFound) {
		return nil, onekiterrors.New(http.StatusNotFound, "Recurring invoice not found")
	}
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get recurring invoice")
	}
	return &financerecurringinvoicev1.GetRecurringInvoiceResponse{Invoice: recurringInvoiceToProto(row)}, nil
}

func (a *RecurringInvoiceAPI) UpdateRecurringInvoice(ctx context.Context, req *financerecurringinvoicev1.UpdateRecurringInvoiceRequest) (*financerecurringinvoicev1.UpdateRecurringInvoiceResponse, error) {
	if err := requireRecurringInvoicePermission(ctx, "finance:recurring_invoices:manage"); err != nil {
		return nil, err
	}
	businessID, err := recurringInvoiceBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	body := req.GetBody()
	if body == nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Recurring invoice body is required")
	}
	input, err := recurringInvoiceInput(body.TemplateInvoiceId, body.GetFrequency(), body.GetNextInvoiceDate(), body.EndDate, body.IsActive)
	if err != nil {
		return nil, err
	}
	if err := a.Service.Update(ctx, businessID, req.GetId(), input); err != nil {
		if errors.Is(err, domain.ErrRecurringInvoiceNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Recurring invoice not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update recurring invoice")
	}
	return &financerecurringinvoicev1.UpdateRecurringInvoiceResponse{Message: "Recurring invoice updated successfully"}, nil
}

func (a *RecurringInvoiceAPI) PauseRecurringInvoice(ctx context.Context, req *financerecurringinvoicev1.PauseRecurringInvoiceRequest) (*financerecurringinvoicev1.PauseRecurringInvoiceResponse, error) {
	if err := requireRecurringInvoicePermission(ctx, "finance:recurring_invoices:manage"); err != nil {
		return nil, err
	}
	businessID, err := recurringInvoiceBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.Service.Pause(ctx, businessID, req.GetId()); err != nil {
		if errors.Is(err, domain.ErrRecurringInvoiceNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Recurring invoice not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to pause recurring invoice")
	}
	return &financerecurringinvoicev1.PauseRecurringInvoiceResponse{Message: "Recurring invoice paused"}, nil
}

func (a *RecurringInvoiceAPI) ResumeRecurringInvoice(ctx context.Context, req *financerecurringinvoicev1.ResumeRecurringInvoiceRequest) (*financerecurringinvoicev1.ResumeRecurringInvoiceResponse, error) {
	if err := requireRecurringInvoicePermission(ctx, "finance:recurring_invoices:manage"); err != nil {
		return nil, err
	}
	businessID, err := recurringInvoiceBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.Service.Resume(ctx, businessID, req.GetId()); err != nil {
		if errors.Is(err, domain.ErrRecurringInvoiceNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Recurring invoice not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to resume recurring invoice")
	}
	return &financerecurringinvoicev1.ResumeRecurringInvoiceResponse{Message: "Recurring invoice resumed"}, nil
}
