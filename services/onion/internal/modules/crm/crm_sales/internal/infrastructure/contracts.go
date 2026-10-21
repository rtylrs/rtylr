package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/infrastructure/persistence/sqlc"
)

func contractsFromRows(rows []sqlc.CrmContract) []domain.Contract {
	contracts := make([]domain.Contract, 0, len(rows))
	for _, row := range rows {
		contracts = append(contracts, contractFromRow(row))
	}
	return contracts
}

func (r *Repository) ListContracts(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Contract, int64, error) {
	rows, err := r.queries.ListContracts(ctx, sqlc.ListContractsParams{BusinessID: businessID, Limit: limit, Offset: offset})
	if err != nil {
		return nil, 0, err
	}
	total, err := r.queries.CountContracts(ctx, businessID)
	return contractsFromRows(rows), total, err
}

func (r *Repository) ListSignedContracts(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Contract, int64, error) {
	rows, err := r.queries.ListSignedContracts(ctx, sqlc.ListSignedContractsParams{BusinessID: businessID, Limit: limit, Offset: offset})
	if err != nil {
		return nil, 0, err
	}
	total, err := r.queries.CountSignedContracts(ctx, businessID)
	return contractsFromRows(rows), total, err
}

func (r *Repository) ListUnsignedContracts(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Contract, int64, error) {
	rows, err := r.queries.ListUnsignedContracts(ctx, sqlc.ListUnsignedContractsParams{BusinessID: businessID, Limit: limit, Offset: offset})
	if err != nil {
		return nil, 0, err
	}
	total, err := r.queries.CountUnsignedContracts(ctx, businessID)
	return contractsFromRows(rows), total, err
}

func (r *Repository) GetContract(ctx context.Context, businessID, id int64) (domain.Contract, error) {
	row, err := r.queries.GetContract(ctx, sqlc.GetContractParams{ID: id, BusinessID: businessID})
	if err != nil {
		return domain.Contract{}, notFound(err, domain.ErrContractNotFound)
	}
	return contractFromRow(row), nil
}

// CreateContract drafts a contract, for a deal when one is named; the deal's
// lead and customer carry over unless the input names a customer.
func (r *Repository) CreateContract(ctx context.Context, businessID, actorID int64, in domain.ContractInput) (domain.Contract, error) {
	var id int64
	err := r.inTx(ctx, func(q *sqlc.Queries, _ *sql.Tx) error {
		var leadID *int64
		if in.DealID != nil {
			row, err := q.GetDeal(ctx, sqlc.GetDealParams{ID: *in.DealID, BusinessID: businessID})
			if err != nil {
				return notFound(err, domain.ErrDealNotFound)
			}
			deal := dealFromRow(row)
			leadID = deal.LeadID
			if in.CustomerID == nil {
				in.CustomerID = deal.CustomerID
			}
		}
		var err error
		id, err = insertContract(ctx, q, businessID, actorID, leadID, in)
		return err
	})
	if err != nil {
		return domain.Contract{}, err
	}
	return r.GetContract(ctx, businessID, id)
}

func insertContract(ctx context.Context, q *sqlc.Queries, businessID, actorID int64, leadID *int64, in domain.ContractInput) (int64, error) {
	result, err := q.CreateContract(ctx, sqlc.CreateContractParams{
		OwnerID:      actorID,
		BusinessID:   businessID,
		LeadID:       nullInt64(leadID),
		DealID:       nullInt64(in.DealID),
		Plan:         in.Plan,
		Mrr:          domain.MonthlyEquivalent(in.Value, in.Billing).String(),
		Seats:        int32(in.Seats), // #nosec G115 -- seat counts are small.
		Value:        in.Value.String(),
		Billing:      sqlc.CrmContractsBilling(in.Billing),
		StartDate:    nullDatePtr(in.StartDate),
		EndDate:      nullDatePtr(in.EndDate),
		CustomerID:   nullInt64(in.CustomerID),
		SignedAt:     nullDatePtr(in.SignedAt),
		SignedByName: nullString(in.SignedByName),
		ClosedBy:     someInt64(actorID),
		Notes:        nullString(in.Notes),
	})
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateContract replaces a contract's terms; an edit that names no customer
// keeps the linked one.
func (r *Repository) UpdateContract(ctx context.Context, businessID, id int64, in domain.ContractInput) (domain.Contract, error) {
	existing, err := r.GetContract(ctx, businessID, id)
	if err != nil {
		return domain.Contract{}, err
	}
	if in.CustomerID == nil {
		in.CustomerID = existing.CustomerID
	}
	if err := r.queries.UpdateContract(ctx, sqlc.UpdateContractParams{
		Plan:         in.Plan,
		Mrr:          domain.MonthlyEquivalent(in.Value, in.Billing).String(),
		Seats:        int32(in.Seats), // #nosec G115 -- seat counts are small.
		Value:        in.Value.String(),
		Billing:      sqlc.CrmContractsBilling(in.Billing),
		StartDate:    nullDatePtr(in.StartDate),
		EndDate:      nullDatePtr(in.EndDate),
		CustomerID:   nullInt64(in.CustomerID),
		SignedAt:     nullDatePtr(in.SignedAt),
		SignedByName: nullString(in.SignedByName),
		ClosedBy:     nullInt64(existing.ClosedBy),
		Notes:        nullString(in.Notes),
		ID:           id,
		BusinessID:   businessID,
	}); err != nil {
		return domain.Contract{}, err
	}
	return r.GetContract(ctx, businessID, id)
}
