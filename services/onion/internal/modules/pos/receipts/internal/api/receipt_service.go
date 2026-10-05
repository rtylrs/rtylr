package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/pos/receipts/internal/application"
	"github.com/voxire/onion/internal/modules/pos/receipts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	posreceiptv1 "github.com/voxire/voxie/gen/go/pos/receipt/v1"
)

type ReceiptAPI struct {
	Service *application.Service
}

func NewReceiptAPI(service *application.Service) *ReceiptAPI {
	return &ReceiptAPI{Service: service}
}

func (a *ReceiptAPI) ListReceipts(ctx context.Context, req *posreceiptv1.ListReceiptsRequest) (*posreceiptv1.ListReceiptsResponse, error) {
	businessID, err := authenticatedReceiptBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	const dateLayout = "2006-01-02"
	if from := req.GetFromDate(); from != "" {
		if _, parseErr := time.Parse(dateLayout, from); parseErr != nil {
			return nil, onekiterrors.New(http.StatusBadRequest, "from must be a valid date in YYYY-MM-DD format")
		}
	}
	if to := req.GetToDate(); to != "" {
		if _, parseErr := time.Parse(dateLayout, to); parseErr != nil {
			return nil, onekiterrors.New(http.StatusBadRequest, "to must be a valid date in YYYY-MM-DD format")
		}
	}

	receipts, listErr := a.Service.List(ctx, businessID, req.GetFromDate(), req.GetToDate())
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list receipts")
	}

	payloads := make([]*posreceiptv1.Receipt, len(receipts))
	for i, rc := range receipts {
		payloads[i] = toProtoReceipt(rc)
	}
	return &posreceiptv1.ListReceiptsResponse{Receipts: payloads}, nil
}

func (a *ReceiptAPI) EmailReceipt(ctx context.Context, req *posreceiptv1.EmailReceiptRequest) (*posreceiptv1.EmailReceiptResponse, error) {
	businessID, err := authenticatedReceiptBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetId() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid receipt ID")
	}

	email := strings.TrimSpace(req.GetEmail())
	if email == "" || !strings.Contains(email, "@") {
		return nil, onekiterrors.New(http.StatusBadRequest, "A valid email is required")
	}

	if sendErr := a.Service.EmailReceipt(ctx, businessID, req.GetId(), email); sendErr != nil {
		if errors.Is(sendErr, domain.ErrReceiptNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Receipt not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to send receipt email")
	}
	return &posreceiptv1.EmailReceiptResponse{Message: "Receipt sent"}, nil
}

func authenticatedReceiptBusinessID(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return 0, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}

func toProtoReceipt(r domain.Receipt) *posreceiptv1.Receipt {
	return &posreceiptv1.Receipt{
		Id:            r.ID,
		OrderNumber:   r.OrderNumber,
		CustomerName:  r.CustomerName,
		Total:         r.Total,
		PaymentMethod: r.PaymentMethod,
		PaidAt:        r.PaidAt.Format("2006-01-02T15:04:05Z"),
	}
}
