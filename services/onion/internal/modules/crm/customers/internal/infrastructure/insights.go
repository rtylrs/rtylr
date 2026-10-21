package infrastructure

import (
	"context"
	"database/sql"
	"time"

	crmcontract "github.com/voxire/onion/internal/contracts/crm"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/infrastructure/persistence/sqlc"
)

// The reads below serve Insights reports, so Insights never queries CRM
// tables itself. Period arguments are calendar days.

func (r *Repository) InsightsCustomerCount(ctx context.Context, businessID int64) (int64, error) {
	return r.queries.InsightsCountCustomers(ctx, businessID)
}

func (r *Repository) InsightsCustomersAsOf(ctx context.Context, businessID int64, asOf time.Time) (int64, error) {
	return r.queries.InsightsCountCustomersAsOf(ctx, sqlc.InsightsCountCustomersAsOfParams{BusinessID: businessID, AsOf: asOf})
}

func (r *Repository) InsightsNewCustomers(ctx context.Context, businessID int64, from, to time.Time) (int64, error) {
	return r.queries.InsightsCountNewCustomers(ctx, sqlc.InsightsCountNewCustomersParams{
		BusinessID: businessID,
		DateFrom:   sql.NullTime{Time: from, Valid: true},
		DateTo:     to,
	})
}

// InsightsCustomerContacts returns the business's customers among ids.
// Deleted customers and other businesses' IDs are left out.
func (r *Repository) InsightsCustomerContacts(ctx context.Context, businessID int64, ids []int64) ([]crmcontract.CustomerContact, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := r.queries.InsightsCustomerContacts(ctx, sqlc.InsightsCustomerContactsParams{BusinessID: businessID, Ids: ids})
	if err != nil {
		return nil, err
	}
	out := make([]crmcontract.CustomerContact, 0, len(rows))
	for _, row := range rows {
		out = append(out, crmcontract.CustomerContact{ID: row.ID, FirstName: row.FirstName, LastName: row.LastName, Email: row.Email})
	}
	return out, nil
}

func (r *Repository) InsightsLastChange(ctx context.Context, businessID int64) (time.Time, error) {
	return r.queries.InsightsCRMLastChange(ctx, businessID)
}
