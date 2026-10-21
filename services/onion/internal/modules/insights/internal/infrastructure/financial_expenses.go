package infrastructure

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type ExpenseByCategoryRow = domain.ExpenseByCategoryRow

func (r *Repository) GetExpensesByCategory(ctx context.Context, businessID int64, from, to time.Time) ([]ExpenseByCategoryRow, error) {
	rows, err := r.sources.Finance.InsightsExpensesByCategory(ctx, businessID, from, to)
	if err != nil {
		logger.Ctx(ctx).Error("insights: expenses by category query failed", logger.Error(err))
		return nil, err
	}
	out := make([]ExpenseByCategoryRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, ExpenseByCategoryRow{
			CategoryID: row.CategoryID, CategoryName: row.CategoryName,
			AmountMinorUnits: row.AmountMinorUnits,
		})
	}
	return out, nil
}
