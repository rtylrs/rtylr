package application

import (
	"context"
	"errors"
)

var ErrInvoiceExistenceUnavailable = errors.New("invoice verification unavailable")

// InvoiceExistenceProvider is the HR-owned port used to validate a Finance
// invoice before HR links its own time entries to that invoice.
type InvoiceExistenceProvider interface {
	InvoiceExists(ctx context.Context, businessID, invoiceID int64) (bool, error)
}

func (s *Service) SetInvoiceExistenceProvider(provider InvoiceExistenceProvider) {
	s.invoiceExistence = provider
}
