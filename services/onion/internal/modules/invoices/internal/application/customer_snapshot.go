package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/voxire/onion/internal/platform/logger"
)

// Sentinels the HTTP layer maps to status codes. Verification failing must
// not look like a generic 500: an unlinked customer is a caller-fixable 404,
// and an unwired port is a server misconfiguration callers cannot fix.
var (
	ErrCustomerVerificationUnavailable = errors.New("customer verification unavailable")
	ErrCustomerNotFound                = errors.New("customer not found")
)

// CustomerSnapshotProvider is the port invoices use to verify a CRM customer
// exists in the caller's business and to read the name/email/address that get
// snapshotted onto the invoice at issue time (#840). Implemented by the
// customers module; neither module imports the other.
type CustomerSnapshotProvider interface {
	CustomerSnapshot(ctx context.Context, businessID, customerID int64) (name, email, address string, err error)
}

// SetCustomerSnapshotProvider wires the customers module implementation.
func (s *Service) SetCustomerSnapshotProvider(p CustomerSnapshotProvider) {
	if p == nil {
		return
	}
	s.customerSnapshots = p
}

// resolveCustomerSnapshot verifies the link and returns the snapshot fields.
// A missing or unwired provider fails loudly: an invoice must never carry a
// customer_id this server could not verify belongs to the business.
func (s *Service) resolveCustomerSnapshot(ctx context.Context, businessID, customerID int64) (name, email, address string, err error) {
	if customerID <= 0 {
		return "", "", "", nil
	}
	if s.customerSnapshots == nil {
		return "", "", "", ErrCustomerVerificationUnavailable
	}
	name, email, address, err = s.customerSnapshots.CustomerSnapshot(ctx, businessID, customerID)
	if err != nil {
		logger.Ctx(ctx).Warn("customer snapshot lookup failed",
			logger.Any("business_id", businessID), logger.Any("customer_id", customerID), logger.Error(err))
		return "", "", "", fmt.Errorf("%w: %d", ErrCustomerNotFound, customerID)
	}
	return name, email, address, nil
}
