package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	"github.com/voxire/onion/internal/modules/invoices/internal/infrastructure"
)

// The status endpoint's own validation gate. This runs on the HTTP path before
// any database access, so "paid" and "partial" must never get past it.
func TestUpdateInvoiceStatusValidate_RejectsMoneyDerivedStatus(t *testing.T) {
	for _, status := range []string{domain.StatusPaid, domain.StatusPartial} {
		h := &UpdateInvoiceStatusHandler{}
		err := h.Validate(context.Background(), &UpdateInvoiceStatusRequest{ID: 1, Status: status})
		if got := financePermissionStatus(err); got != http.StatusBadRequest {
			t.Errorf("status %q = %d, want %d (err = %v)", status, got, http.StatusBadRequest, err)
		}
	}
}

func TestUpdateInvoiceStatusValidate_RejectsUnknownStatus(t *testing.T) {
	for _, status := range []string{"PAID", "settled", " refunded", "partially_paid"} {
		h := &UpdateInvoiceStatusHandler{}
		err := h.Validate(context.Background(), &UpdateInvoiceStatusRequest{ID: 1, Status: status})
		if got := financePermissionStatus(err); got != http.StatusBadRequest {
			t.Errorf("status %q = %d, want %d (err = %v)", status, got, http.StatusBadRequest, err)
		}
	}
}

func TestUpdateInvoiceStatusValidate_AllowsStatusesReachableWithoutMoney(t *testing.T) {
	for _, status := range []string{domain.StatusDraft, domain.StatusSent, domain.StatusViewed, domain.StatusOverdue, domain.StatusCancelled, domain.StatusVoid} {
		h := &UpdateInvoiceStatusHandler{}
		if err := h.Validate(context.Background(), &UpdateInvoiceStatusRequest{ID: 1, Status: status}); err != nil {
			t.Errorf("status %q rejected by validation: %v", status, err)
		}
	}
}

// A money rule that returns 500 is indistinguishable from a server fault in
// every dashboard and alert, so each of these must map onto its own status.
func TestInvoiceServiceError_MapsMoneyValidationToClientErrors(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"unknown status", domain.ErrInvalidInvoiceStatus, http.StatusBadRequest},
		{"paid without a payment", domain.ErrInvoiceStatusNotSettable, http.StatusBadRequest},
		{"due before issue", domain.ErrInvalidInvoiceDate, http.StatusBadRequest},
		{"discount over total", domain.ErrDiscountExceedsTotal, http.StatusBadRequest},
		{"non positive payment", domain.ErrInvalidPaymentAmount, http.StatusBadRequest},
		{"payment over balance", domain.ErrPaymentExceedsBalance, http.StatusBadRequest},
		{"missing invoice", domain.ErrInvoiceNotFound, http.StatusNotFound},
		{"missing customer", application.ErrCustomerNotFound, http.StatusNotFound},
		{"lifecycle conflict", domain.ErrInvoiceNotDraft, http.StatusConflict},
		{"wrapped lifecycle conflict", errors.Join(errors.New("update"), domain.ErrInvoiceNotDraft), http.StatusConflict},
		{"unrelated fault", errors.New("connection reset"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		err := invoiceServiceError(tc.err, "Failed to update invoice")
		if got := financePermissionStatus(err); got != tc.want {
			t.Errorf("%s: status = %d, want %d (err = %v)", tc.name, got, tc.want, err)
		}
	}
}

// Defence in depth for any caller that reaches the repository without going
// through the service: a status-only write cannot reach a money-derived status.
func TestRepositoryUpdateInvoiceStatus_RejectsMoneyDerivedStatus(t *testing.T) {
	repo := infrastructure.NewRepository(nil)
	for _, status := range []string{domain.StatusPaid, domain.StatusPartial, "settled"} {
		err := repo.UpdateInvoiceStatus(context.Background(), 1, 1, status)
		if err == nil {
			t.Errorf("repository accepted status %q without a payment", status)
		}
	}
}
