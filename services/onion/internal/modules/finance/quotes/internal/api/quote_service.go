package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/finance/quotes/internal/application"
	"github.com/voxire/onion/internal/modules/finance/quotes/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	financequotev1 "github.com/voxire/voxie/gen/go/finance/quote/v1"
)

type QuoteAPI struct {
	Service *application.Service
}

func NewQuoteAPI(service *application.Service) *QuoteAPI {
	return &QuoteAPI{Service: service}
}

func (a *QuoteAPI) ListQuotes(ctx context.Context, req *financequotev1.ListQuotesRequest) (*financequotev1.ListQuotesResponse, error) {
	if err := requireQuotePermission(ctx, "finance:quotes:view"); err != nil {
		return nil, err
	}
	businessID, errModel := quoteBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	page := req.GetPage()
	if page < 1 {
		page = 1
	}
	pageSize := req.GetPageSize()
	if pageSize < 1 {
		pageSize = 50
	}

	quotes, hasMore, err := a.Service.List(ctx, businessID, page, pageSize)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list quotes")
	}

	payloads := make([]*financequotev1.Quote, 0, len(quotes))
	for _, q := range quotes {
		payloads = append(payloads, newQuotePayload(ctx, q))
	}

	return &financequotev1.ListQuotesResponse{
		Quotes:   payloads,
		Page:     page,
		PageSize: pageSize,
		HasMore:  hasMore,
	}, nil
}

func (a *QuoteAPI) CreateQuote(ctx context.Context, req *financequotev1.CreateQuoteRequest) (*financequotev1.CreateQuoteResponse, error) {
	if err := requireQuotePermission(ctx, "finance:quotes:manage"); err != nil {
		return nil, err
	}
	businessID, errModel := quoteBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	status := req.GetStatus()
	if status == "" {
		status = "draft"
	}

	input := domain.QuoteInput{
		CustomerName:  req.GetCustomerName(),
		CustomerEmail: req.GetCustomerEmail(),
		Status:        status,
		Subtotal:      req.GetSubtotal(),
		Tax:           req.GetTax(),
		Discount:      req.GetDiscount(),
		Total:         req.GetTotal(),
		ValidUntil:    req.GetValidUntil(),
		Notes:         req.GetNotes(),
	}
	if err := domain.ValidateQuoteInput(input); err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, err.Error())
	}

	itemsJSON, err := marshalQuoteLines(req.GetLines())
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid quote lines")
	}
	input.Items = itemsJSON

	id, err := a.Service.Create(ctx, businessID, input)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create quote")
	}

	quote, err := a.Service.GetByID(ctx, businessID, id)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Quote created but failed to load")
	}

	return &financequotev1.CreateQuoteResponse{Quote: newQuotePayload(ctx, quote)}, nil
}

func (a *QuoteAPI) GetQuote(ctx context.Context, req *financequotev1.GetQuoteRequest) (*financequotev1.GetQuoteResponse, error) {
	if err := requireQuotePermission(ctx, "finance:quotes:view"); err != nil {
		return nil, err
	}
	businessID, errModel := quoteBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	quote, err := a.Service.GetByID(ctx, businessID, req.GetId())
	if err != nil {
		return nil, onekiterrors.New(http.StatusNotFound, "Quote not found")
	}

	return &financequotev1.GetQuoteResponse{Quote: newQuotePayload(ctx, quote)}, nil
}

func (a *QuoteAPI) UpdateQuoteStatus(ctx context.Context, req *financequotev1.UpdateQuoteStatusRequest) (*financequotev1.UpdateQuoteStatusResponse, error) {
	if err := requireQuotePermission(ctx, "finance:quotes:manage"); err != nil {
		return nil, err
	}
	businessID, errModel := quoteBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	body := req.GetBody()
	if body == nil || body.GetStatus() == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "status is required")
	}
	if _, ok := domain.ValidQuoteStatuses[body.GetStatus()]; !ok {
		return nil, onekiterrors.New(http.StatusBadRequest, "invalid quote status")
	}

	if err := a.Service.UpdateStatus(ctx, businessID, req.GetId(), body.GetStatus()); err != nil {
		if errors.Is(err, domain.ErrQuoteNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Quote not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update quote status")
	}

	return &financequotev1.UpdateQuoteStatusResponse{Message: "Quote status updated successfully"}, nil
}

func (a *QuoteAPI) ConvertQuote(ctx context.Context, req *financequotev1.ConvertQuoteRequest) (*financequotev1.ConvertQuoteResponse, error) {
	if err := requireQuotePermission(ctx, "finance:quotes:manage"); err != nil {
		return nil, err
	}
	businessID, errModel := quoteBusinessID(ctx)
	if errModel != nil {
		return nil, errModel
	}

	invoiceID, err := a.Service.ConvertToInvoice(ctx, businessID, req.GetId())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to convert quote")
	}

	return &financequotev1.ConvertQuoteResponse{InvoiceId: invoiceID, Message: "Quote converted to invoice successfully"}, nil
}

func newQuotePayload(ctx context.Context, q domain.Quote) *financequotev1.Quote {
	return &financequotev1.Quote{
		Id:            q.ID,
		OwnerId:       q.OwnerID,
		CustomerName:  q.CustomerName,
		CustomerEmail: q.CustomerEmail,
		Status:        q.Status,
		Lines:         unmarshalQuoteLines(ctx, q.Items),
		Subtotal:      q.Subtotal,
		Tax:           q.Tax,
		Discount:      q.Discount,
		Total:         q.Total,
		ValidUntil:    q.ValidUntil.Format("2006-01-02"),
		Notes:         q.Notes,
		CreatedAt:     q.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     q.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func marshalQuoteLines(lines []*financequotev1.QuoteLine) (string, error) {
	type quoteLineJSON struct {
		Description string  `json:"description"`
		Quantity    int32   `json:"quantity"`
		UnitPrice   float64 `json:"unit_price"`
		TaxRate     float64 `json:"tax_rate"`
		Total       float64 `json:"total"`
	}

	out := make([]quoteLineJSON, 0, len(lines))
	for _, line := range lines {
		out = append(out, quoteLineJSON{
			Description: line.GetDescription(),
			Quantity:    line.GetQuantity(),
			UnitPrice:   line.GetUnitPrice(),
			TaxRate:     line.GetTaxRate(),
			Total:       line.GetTotal(),
		})
	}

	raw, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func unmarshalQuoteLines(ctx context.Context, items string) []*financequotev1.QuoteLine {
	if items == "" {
		return nil
	}

	type quoteLineJSON struct {
		Description string  `json:"description"`
		Quantity    int32   `json:"quantity"`
		UnitPrice   float64 `json:"unit_price"`
		TaxRate     float64 `json:"tax_rate"`
		Total       float64 `json:"total"`
	}

	var parsed []quoteLineJSON
	if err := json.Unmarshal([]byte(items), &parsed); err != nil {
		logger.Ctx(ctx).Warn("Failed to parse quote items JSON", logger.Error(err))
		return nil
	}

	lines := make([]*financequotev1.QuoteLine, 0, len(parsed))
	for _, item := range parsed {
		lines = append(lines, &financequotev1.QuoteLine{
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TaxRate:     item.TaxRate,
			Total:       item.Total,
		})
	}
	return lines
}

func requireQuotePermission(ctx context.Context, permission string) error {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil || !authPkg.HasAppPermission(device.User, authorization.AppFinance, permission) {
		return onekiterrors.PermissionDenied()
	}
	return nil
}

func quoteBusinessID(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return 0, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}
