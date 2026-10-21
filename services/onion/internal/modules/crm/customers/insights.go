package customers

import (
	"context"
	"time"

	crmcontract "github.com/voxire/onion/internal/contracts/crm"
)

// Insights reports read customer data through these methods rather than
// querying CRM tables.

func (m *Module) InsightsCustomerCount(ctx context.Context, businessID int64) (int64, error) {
	return m.repository.InsightsCustomerCount(ctx, businessID)
}

func (m *Module) InsightsCustomersAsOf(ctx context.Context, businessID int64, asOf time.Time) (int64, error) {
	return m.repository.InsightsCustomersAsOf(ctx, businessID, asOf)
}

func (m *Module) InsightsNewCustomers(ctx context.Context, businessID int64, from, to time.Time) (int64, error) {
	return m.repository.InsightsNewCustomers(ctx, businessID, from, to)
}

func (m *Module) InsightsCustomerContacts(ctx context.Context, businessID int64, ids []int64) ([]crmcontract.CustomerContact, error) {
	return m.repository.InsightsCustomerContacts(ctx, businessID, ids)
}

func (m *Module) InsightsLastChange(ctx context.Context, businessID int64) (time.Time, error) {
	return m.repository.InsightsLastChange(ctx, businessID)
}
