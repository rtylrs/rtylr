package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/money"
)

func (r *Repository) ListSignedContracts(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Contract, error) {
	rows, err := r.queries.ListSignedContracts(ctx, sqlc.ListSignedContractsParams{
		BusinessID: businessID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list signed contracts", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	out := make([]domain.Contract, 0, len(rows))
	for _, row := range rows {
		out = append(out, contractFromFields(
			row.ID, row.OwnerID, row.BusinessID, row.LeadID, row.Plan, row.Mrr,
			int(row.Seats), contractTermsRow{row.Value, row.Billing, row.StartDate, row.EndDate, row.CustomerID},
			row.SignedAt, row.SignedByName, row.ClosedBy, row.Notes,
			row.CreatedAt, row.UpdatedAt,
		))
	}
	return out, nil
}

func (r *Repository) CountSignedContracts(ctx context.Context, businessID int64) (int64, error) {
	return r.queries.CountSignedContracts(ctx, businessID)
}

func (r *Repository) ListUnsignedContracts(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Contract, error) {
	rows, err := r.queries.ListUnsignedContracts(ctx, sqlc.ListUnsignedContractsParams{
		BusinessID: businessID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list unsigned contracts", logger.Any("business_id", businessID), logger.Error(err))
		return nil, err
	}
	out := make([]domain.Contract, 0, len(rows))
	for _, row := range rows {
		out = append(out, contractFromFields(
			row.ID, row.OwnerID, row.BusinessID, row.LeadID, row.Plan, row.Mrr,
			int(row.Seats), contractTermsRow{row.Value, row.Billing, row.StartDate, row.EndDate, row.CustomerID},
			row.SignedAt, row.SignedByName, row.ClosedBy, row.Notes,
			row.CreatedAt, row.UpdatedAt,
		))
	}
	return out, nil
}

func (r *Repository) CountUnsignedContracts(ctx context.Context, businessID int64) (int64, error) {
	return r.queries.CountUnsignedContracts(ctx, businessID)
}

func contractFromRow(row sqlc.ListContractsRow) domain.Contract {
	return contractFromFields(
		row.ID, row.OwnerID, row.BusinessID, row.LeadID, row.Plan, row.Mrr,
		int(row.Seats), contractTermsRow{row.Value, row.Billing, row.StartDate, row.EndDate, row.CustomerID},
		row.SignedAt, row.SignedByName, row.ClosedBy, row.Notes,
		row.CreatedAt, row.UpdatedAt,
	)
}

func contractFromGetRow(row sqlc.GetContractRow) domain.Contract {
	return contractFromFields(
		row.ID, row.OwnerID, row.BusinessID, row.LeadID, row.Plan, row.Mrr,
		int(row.Seats), contractTermsRow{row.Value, row.Billing, row.StartDate, row.EndDate, row.CustomerID},
		row.SignedAt, row.SignedByName, row.ClosedBy, row.Notes,
		row.CreatedAt, row.UpdatedAt,
	)
}

func contractFromFields(
	id, ownerID, businessID, leadID int64,
	plan, mrr string,
	seats int,
	terms contractTermsRow,
	signedAt sql.NullTime,
	signedByName string,
	closedBy sql.NullInt64,
	notes string,
	createdAt, updatedAt time.Time,
) domain.Contract {
	c := domain.Contract{
		ID:           id,
		OwnerID:      ownerID,
		BusinessID:   businessID,
		LeadID:       leadID,
		Plan:         plan,
		MRR:          parseDecimalString(mrr),
		Seats:        seats,
		Value:        parseDecimalString(terms.value),
		Billing:      domain.Billing(terms.billing),
		StartDate:    datePtr(terms.startDate),
		EndDate:      datePtr(terms.endDate),
		SignedByName: signedByName,
		Notes:        notes,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
	if signedAt.Valid {
		s := signedAt.Time.Format("2006-01-02")
		c.SignedAt = &s
	}
	if closedBy.Valid {
		c.ClosedBy = &closedBy.Int64
	}
	if terms.customerID.Valid {
		c.CustomerID = &terms.customerID.Int64
	}
	return c
}

// contractTermsRow carries the value, billing, term and customer columns,
// which every contract query selects the same way.
type contractTermsRow struct {
	value      string
	billing    sqlc.CrmContractsBilling
	startDate  sql.NullTime
	endDate    sql.NullTime
	customerID sql.NullInt64
}

func datePtr(value sql.NullTime) *string {
	if !value.Valid {
		return nil
	}
	formatted := value.Time.Format("2006-01-02")
	return &formatted
}

func parsePipelineMRR(v interface{}) money.Amount {
	switch n := v.(type) {
	case money.Amount:
		return n
	case float64:
		amount, err := money.FromFloat64(n)
		if err == nil {
			return amount
		}
	case []byte:
		return parseDecimalString(string(n))
	case string:
		return parseDecimalString(n)
	}
	return 0
}

func parseDecimalString(s string) money.Amount {
	if s == "" {
		return 0
	}
	amount, err := money.FromString(s)
	if err != nil {
		return 0
	}
	return amount
}

// ---------------------------------------------------------------------------
// SQL null helpers
// ---------------------------------------------------------------------------

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullableDatePtr(s *string) (sql.NullTime, error) {
	if s == nil || *s == "" {
		return sql.NullTime{}, nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return sql.NullTime{}, fmt.Errorf("invalid date %q: want YYYY-MM-DD", *s)
	}
	return sql.NullTime{Time: t, Valid: true}, nil
}

func nullableInt64(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

func nullableInt64Value(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v != 0}
}

func nullableDecimalString(v *money.Amount) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: v.String(), Valid: true}
}
