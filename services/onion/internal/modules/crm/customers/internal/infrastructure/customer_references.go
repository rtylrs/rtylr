package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/infrastructure/persistence/sqlc"
)

// purgeCustomerReferences removes rows that only exist for a deleted customer
// and detaches the sales records that named them.
func purgeCustomerReferences(ctx context.Context, q *sqlc.Queries, businessID, customerID int64) error {
	nullable := sql.NullInt64{Int64: customerID, Valid: true}
	steps := []func() error{
		func() error {
			return q.DeleteCustomerCampaignSends(ctx, sqlc.DeleteCustomerCampaignSendsParams{BusinessID: businessID, CustomerID: customerID})
		},
		func() error {
			return q.DeleteCustomerFeedback(ctx, sqlc.DeleteCustomerFeedbackParams{BusinessID: businessID, CustomerID: customerID})
		},
		func() error {
			return q.DeleteCustomerLoyaltyTransactions(ctx, sqlc.DeleteCustomerLoyaltyTransactionsParams{BusinessID: businessID, CustomerID: customerID})
		},
		func() error {
			return q.DeleteCustomerTimeline(ctx, sqlc.DeleteCustomerTimelineParams{BusinessID: businessID, CustomerID: customerID})
		},
		func() error {
			return q.DetachCustomerFromContacts(ctx, sqlc.DetachCustomerFromContactsParams{BusinessID: businessID, CustomerID: nullable})
		},
		func() error {
			return q.DetachCustomerFromLeads(ctx, sqlc.DetachCustomerFromLeadsParams{BusinessID: businessID, CustomerID: nullable})
		},
		func() error {
			return q.DetachCustomerFromDeals(ctx, sqlc.DetachCustomerFromDealsParams{BusinessID: businessID, CustomerID: nullable})
		},
		func() error {
			return q.DetachCustomerFromActivities(ctx, sqlc.DetachCustomerFromActivitiesParams{BusinessID: businessID, CustomerID: nullable})
		},
		func() error {
			return q.DetachCustomerFromContracts(ctx, sqlc.DetachCustomerFromContractsParams{BusinessID: businessID, CustomerID: nullable})
		},
	}
	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}

// moveCustomerReferences points a merged-away customer's history and sales
// records at the surviving customer.
func moveCustomerReferences(ctx context.Context, q *sqlc.Queries, businessID, fromID, toID int64) error {
	from, to := sql.NullInt64{Int64: fromID, Valid: true}, sql.NullInt64{Int64: toID, Valid: true}
	steps := []func() error{
		func() error {
			return q.MoveCustomerCampaignSends(ctx, sqlc.MoveCustomerCampaignSendsParams{ToCustomerID: toID, FromCustomerID: fromID, BusinessID: businessID})
		},
		func() error {
			return q.MoveCustomerFeedback(ctx, sqlc.MoveCustomerFeedbackParams{ToCustomerID: toID, FromCustomerID: fromID, BusinessID: businessID})
		},
		func() error {
			return q.MoveCustomerContacts(ctx, sqlc.MoveCustomerContactsParams{ToCustomerID: to, FromCustomerID: from, BusinessID: businessID})
		},
		func() error {
			return q.MoveCustomerLeads(ctx, sqlc.MoveCustomerLeadsParams{ToCustomerID: to, FromCustomerID: from, BusinessID: businessID})
		},
		func() error {
			return q.MoveCustomerDeals(ctx, sqlc.MoveCustomerDealsParams{ToCustomerID: to, FromCustomerID: from, BusinessID: businessID})
		},
		func() error {
			return q.MoveCustomerActivities(ctx, sqlc.MoveCustomerActivitiesParams{ToCustomerID: to, FromCustomerID: from, BusinessID: businessID})
		},
		func() error {
			return q.MoveCustomerContracts(ctx, sqlc.MoveCustomerContractsParams{ToCustomerID: to, FromCustomerID: from, BusinessID: businessID})
		},
	}
	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}
