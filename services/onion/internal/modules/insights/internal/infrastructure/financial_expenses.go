package infrastructure

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

// ExpenseByCategoryRow is one bucket of expense spend, grouped by the
// category label on finance_expense_categories. Uncategorised rows
// collapse to "Uncategorised" so the chart never has a blank segment.
type ExpenseByCategoryRow struct {
	CategoryID       int64
	CategoryName     string
	AmountMinorUnits int64
}

// GetExpensesByCategory returns category-grouped expense totals for the
// window. The join to finance_expense_categories is left so expenses
// without a category remain visible.
func (r *Repository) GetExpensesByCategory(ctx context.Context, businessID int64, from, to time.Time) ([]ExpenseByCategoryRow, error) {
	// cross-schema read from finance_expenses + finance_expense_categories — category-grouped expense totals
	const q = `
		SELECT COALESCE(c.id, 0) AS cat_id,
		       COALESCE(c.name, 'Uncategorised') AS cat_name,
		       COALESCE(SUM(e.amount), 0)
		FROM finance_expenses e
		LEFT JOIN finance_expense_categories c ON c.id = e.category_id AND c.business_id = e.business_id
		WHERE e.business_id = ?
		  AND e.expense_date BETWEEN ? AND ?
		GROUP BY c.id, c.name
		ORDER BY SUM(e.amount) DESC
	`
	rows, err := r.pool.QueryContext(ctx, q, businessID, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if err != nil {
		logger.Ctx(ctx).Error("insights: expenses by category query failed", logger.Error(err))
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Ctx(ctx).Warn("insights: rows.Close failed", logger.Error(closeErr))
		}
	}()

	out := make([]ExpenseByCategoryRow, 0)
	for rows.Next() {
		var row ExpenseByCategoryRow
		var amount interface{}
		if scanErr := rows.Scan(&row.CategoryID, &row.CategoryName, &amount); scanErr != nil {
			logger.Ctx(ctx).Error("insights: expenses by category scan failed", logger.Error(scanErr))
			return nil, scanErr
		}
		row.AmountMinorUnits = decimalToMinorUnits(amount)
		out = append(out, row)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}
	return out, nil
}
