package infrastructure

import (
	"context"
	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type ExpenseByCategoryRow = domain.ExpenseByCategoryRow

func (r *Repository) GetExpensesByCategory(ctx context.Context, businessID int64, from, to time.Time) ([]ExpenseByCategoryRow, error) {
	// cross-schema read from finance_expenses + finance_expense_categories — category-grouped expense totals
	rows, err := r.queries.GetExpensesByCategory(ctx, sqlc.GetExpensesByCategoryParams{
		BusinessID:      businessID,
		FromExpenseDate: from,
		ToExpenseDate:   to,
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: expenses by category query failed", logger.Error(err))
		return nil, err
	}
	out := make([]ExpenseByCategoryRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, ExpenseByCategoryRow{
			CategoryID: row.CatID, CategoryName: row.CatName,
			AmountMinorUnits: decimalToMinorUnits(row.Amount),
		})
	}
	return out, nil
}
