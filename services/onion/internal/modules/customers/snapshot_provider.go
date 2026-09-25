package customers

import (
	"context"
	"strings"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/platform/logger"
)

// CustomerSnapshotProvider is the port invoices defines and this module
// implements: verify a customer belongs to the business and return the
// issue-time snapshot fields (#840). Declared here (the implementer) rather
// than duplicated in invoices so the two packages never import each other.
type CustomerSnapshotProvider interface {
	CustomerSnapshot(ctx context.Context, businessID, customerID int64) (name, email, address string, err error)
}

// snapshotProvider adapts this module's customer reads to that port.
type snapshotProvider struct {
	service *application.Service
}

// CustomerSnapshot verifies the customer belongs to the business and returns
// the issue-time fields an invoice snapshots. CRM customers carry no tax id,
// so that field stays caller-supplied.
func (p snapshotProvider) CustomerSnapshot(ctx context.Context, businessID, customerID int64) (name, email, address string, err error) {
	c, err := p.service.GetCustomer(ctx, businessID, customerID)
	if err != nil {
		logger.Ctx(ctx).Warn("customer snapshot lookup failed",
			logger.Any("business_id", businessID), logger.Any("customer_id", customerID), logger.Error(err))
		return "", "", "", err
	}
	name = strings.TrimSpace(c.FirstName + " " + c.LastName)
	return name, c.Email, c.Address, nil
}

// SnapshotProvider exposes the port for main.go wiring.
func (m *Module) SnapshotProvider() CustomerSnapshotProvider {
	return snapshotProvider{service: m.service}
}
