package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/quotes/internal/domain"
	"github.com/voxire/onion/internal/modules/quotes/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{queries: sqlc.New(dbclient.Pool())}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Quote, error) {
	rows, err := r.queries.ListQuotes(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list quotes", logger.Error(err))
		return nil, err
	}

	items := make([]domain.Quote, 0, len(rows))
	for _, row := range rows {
		items = append(items, quoteFromSqlc(row))
	}

	return items, nil
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.QuoteInput) (int64, error) {
	status := input.Status
	if status == "" {
		status = "draft"
	}

	res, err := r.queries.CreateQuote(ctx, sqlc.CreateQuoteParams{
		OwnerID:       ownerID,
		CustomerName:  input.CustomerName,
		CustomerEmail: nullableString(input.CustomerEmail),
		Status:        sqlc.NullFinanceQuotesStatus{FinanceQuotesStatus: sqlc.FinanceQuotesStatus(status), Valid: true},
		Items:         rawJSON(input.Items),
		Subtotal:      nullableString(fmtFloat(input.Subtotal)),
		Tax:           nullableString(fmtFloat(input.Tax)),
		Discount:      nullableString(fmtFloat(input.Discount)),
		Total:         nullableString(fmtFloat(input.Total)),
		ValidUntil:    nullTimeStr(input.ValidUntil),
		Notes:         nullableString(input.Notes),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create quote", logger.Error(err))
		return 0, err
	}

	return res.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, ownerID, id int64, input domain.QuoteInput) error {
	err := r.queries.UpdateQuote(ctx, sqlc.UpdateQuoteParams{
		CustomerName:  input.CustomerName,
		CustomerEmail: nullableString(input.CustomerEmail),
		Status:        sqlc.NullFinanceQuotesStatus{FinanceQuotesStatus: sqlc.FinanceQuotesStatus(input.Status), Valid: input.Status != ""},
		Items:         rawJSON(input.Items),
		Subtotal:      nullableString(fmtFloat(input.Subtotal)),
		Tax:           nullableString(fmtFloat(input.Tax)),
		Discount:      nullableString(fmtFloat(input.Discount)),
		Total:         nullableString(fmtFloat(input.Total)),
		ValidUntil:    nullTimeStr(input.ValidUntil),
		Notes:         nullableString(input.Notes),
		ID:            id,
		OwnerID:       ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update quote", logger.Error(err))
		return err
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, ownerID, id int64) error {
	err := r.queries.DeleteQuote(ctx, sqlc.DeleteQuoteParams{
		ID:      id,
		OwnerID: ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete quote", logger.Error(err))
		return err
	}

	return nil
}

// ---------------------------------------------------------------------------
// Mapping helpers
// ---------------------------------------------------------------------------

func quoteFromSqlc(row sqlc.FinanceQuote) domain.Quote {
	q := domain.Quote{
		ID:            row.ID,
		OwnerID:       row.OwnerID,
		CustomerName:  row.CustomerName,
		CustomerEmail: nullStr(row.CustomerEmail),
		Status:        string(row.Status.FinanceQuotesStatus),
		Items:         string(row.Items),
		Subtotal:      parseFloatNull(row.Subtotal),
		Tax:           parseFloatNull(row.Tax),
		Discount:      parseFloatNull(row.Discount),
		Total:         parseFloatNull(row.Total),
		Notes:         nullStr(row.Notes),
		CreatedAt:     nullTime(row.CreatedAt),
		UpdatedAt:     nullTime(row.UpdatedAt),
	}
	if row.ValidUntil.Valid {
		q.ValidUntil = row.ValidUntil.Time
	}
	return q
}

// ---------------------------------------------------------------------------
// Conversion helpers
// ---------------------------------------------------------------------------

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullStr(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func nullTime(t sql.NullTime) time.Time {
	if t.Valid {
		return t.Time
	}
	return time.Time{}
}

func nullTimeStr(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{}
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func fmtFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func parseFloatNull(s sql.NullString) float64 {
	if !s.Valid {
		return 0
	}
	v, _ := strconv.ParseFloat(s.String, 64)
	return v
}

func rawJSON(s string) json.RawMessage {
	if s == "" {
		return nil
	}
	return json.RawMessage(s)
}
