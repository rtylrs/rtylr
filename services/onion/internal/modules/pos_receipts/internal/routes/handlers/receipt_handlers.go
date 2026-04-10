package handlers

import (
	"context"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/pos_receipts/internal/application"
	"github.com/voxire/onion/internal/modules/pos_receipts/internal/domain"
)

type ReceiptPayload struct {
	ID            int64   `json:"id"`
	OrderNumber   string  `json:"orderNumber"`
	CustomerName  string  `json:"customerName"`
	Total         float64 `json:"total"`
	PaymentMethod string  `json:"paymentMethod"`
	PaidAt        string  `json:"paidAt"`
}

func newReceiptPayload(r domain.Receipt) ReceiptPayload {
	return ReceiptPayload{
		ID:            r.ID,
		OrderNumber:   r.OrderNumber,
		CustomerName:  r.CustomerName,
		Total:         r.Total,
		PaymentMethod: r.PaymentMethod,
		PaidAt:        r.PaidAt.Format("2006-01-02T15:04:05Z"),
	}
}

// --- List ---

func ListReceiptsRoutes() []string { return []string{"/pos/receipts"} }

type ListReceiptsHandler struct{ Service *application.Service }

type ListReceiptsRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type ListReceiptsResponse struct {
	Receipts []ReceiptPayload `json:"receipts"`
}

func (h *ListReceiptsHandler) Serve(ctx context.Context, req *ListReceiptsRequest) (ListReceiptsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListReceiptsResponse{}, errorModels.PermissionDeniedServerError()
	}

	receipts, err := h.Service.List(ctx, device.User.UserID, req.From, req.To)
	if err != nil {
		return ListReceiptsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list receipts"}
	}
	payloads := make([]ReceiptPayload, len(receipts))
	for i, rc := range receipts {
		payloads[i] = newReceiptPayload(rc)
	}
	return ListReceiptsResponse{Receipts: payloads}, nil
}
