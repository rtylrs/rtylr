package application

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
)

var ErrBillableTimeUnavailable = errors.New("billable time provider unavailable")

// BillableTimeReader is the invoices-owned port for approved HR time that can
// be attached to an invoice. HR owns the query and its persistence.
type BillableTimeReader interface {
	ListUnbilledBillableTimeEntries(ctx context.Context, businessID int64, limit, offset int) ([]domain.BillableTimeEntry, int64, int64, error)
}

func (s *Service) SetBillableTimeReader(reader BillableTimeReader) {
	s.billableTime = reader
}
