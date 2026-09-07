package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_time/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func MarkTimeEntriesInvoicedRoutes() []string {
	return []string{"/hr/time-entries/mark-invoiced"}
}

type MarkTimeEntriesInvoicedHandler struct{ Service *application.Service }

type MarkTimeEntriesInvoicedRequest struct {
	EntryIDs  []int64 `json:"entry_ids"`
	InvoiceID int64   `json:"invoice_id"`
}

type MarkTimeEntriesInvoicedResponse struct {
	MarkedCount  int32 `json:"marked_count"`
	SkippedCount int32 `json:"skipped_count"`
}

func (h *MarkTimeEntriesInvoicedHandler) Validate(_ context.Context, req *MarkTimeEntriesInvoicedRequest) error {
	if req.InvoiceID <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "invoice_id must be positive")
	}
	if len(req.EntryIDs) == 0 {
		return onekiterrors.New(http.StatusBadRequest, "entry_ids must not be empty")
	}
	if len(req.EntryIDs) > 500 {
		return onekiterrors.New(http.StatusBadRequest, "entry_ids exceeds the batch limit of 500")
	}
	for _, id := range req.EntryIDs {
		if id <= 0 {
			return onekiterrors.New(http.StatusBadRequest, "entry_ids must all be positive")
		}
	}
	return nil
}

// Serve closes the billable loop (#836/#791): after an invoice is raised for
// chosen hours, the caller stamps those entries with the invoice id so the
// unbilled-billable query stops offering them. Idempotent per entry; already
// invoiced or ineligible ids are reported via skipped_count instead of failing
// the batch.
func (h *MarkTimeEntriesInvoicedHandler) Serve(ctx context.Context, req *MarkTimeEntriesInvoicedRequest) (MarkTimeEntriesInvoicedResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return MarkTimeEntriesInvoicedResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return MarkTimeEntriesInvoicedResponse{}, onekiterrors.PermissionDenied()
	}

	result, err := h.Service.MarkEntriesInvoiced(ctx, businessID, req.InvoiceID, req.EntryIDs)
	if err != nil {
		return MarkTimeEntriesInvoicedResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to mark time entries invoiced")
	}
	return MarkTimeEntriesInvoicedResponse{
		MarkedCount:  result.MarkedCount,
		SkippedCount: result.SkippedCount,
	}, nil
}
