package api

import (
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

// invoiceServiceError maps a domain or application error onto the status this
// module already uses for it, with the same onekiterrors codes as the rest of
// the API. A rejected money rule must not fall through to 500: once it does, a
// validation failure is indistinguishable from a server fault in every
// dashboard and alert. fallbackMessage is the 500 text of the calling
// operation.
func invoiceServiceError(err error, fallbackMessage string) error {
	switch {
	case errors.Is(err, domain.ErrInvalidInvoiceStatus):
		return onekiterrors.New(http.StatusBadRequest, "Invalid invoice status")
	case errors.Is(err, domain.ErrInvoiceStatusNotSettable):
		return onekiterrors.New(http.StatusBadRequest, "Invoice status is set by recording a payment")
	case errors.Is(err, domain.ErrInvalidInvoiceDate):
		return onekiterrors.New(http.StatusBadRequest, domain.ErrInvalidInvoiceDate.Error())
	case errors.Is(err, domain.ErrDiscountExceedsTotal):
		return onekiterrors.New(http.StatusBadRequest, domain.ErrDiscountExceedsTotal.Error())
	case errors.Is(err, domain.ErrInvalidPaymentAmount):
		return onekiterrors.New(http.StatusBadRequest, domain.ErrInvalidPaymentAmount.Error())
	case errors.Is(err, domain.ErrPaymentExceedsBalance):
		return onekiterrors.New(http.StatusBadRequest, domain.ErrPaymentExceedsBalance.Error())
	case errors.Is(err, domain.ErrInvoiceNotFound):
		return onekiterrors.New(http.StatusNotFound, "Invoice not found")
	case errors.Is(err, application.ErrCustomerNotFound):
		return onekiterrors.New(http.StatusNotFound, "Customer not found")
	case errors.Is(err, domain.ErrInvoiceNotDraft):
		return onekiterrors.New(http.StatusConflict, "Invoice lifecycle does not allow this change")
	default:
		return onekiterrors.New(http.StatusInternalServerError, fallbackMessage)
	}
}
