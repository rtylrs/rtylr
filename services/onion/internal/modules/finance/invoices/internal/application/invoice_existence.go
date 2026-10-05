package application

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
)

func (s *Service) InvoiceExists(ctx context.Context, businessID, invoiceID int64) (bool, error) {
	_, err := s.repo.GetInvoiceByID(ctx, businessID, invoiceID)
	if errors.Is(err, domain.ErrInvoiceNotFound) {
		return false, nil
	}
	return err == nil, err
}
