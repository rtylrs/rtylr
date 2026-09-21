package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// Dashboard
// ─────────────────────────────────────────────────────────────────────────────

func (r *Repository) GetDashboardStats(ctx context.Context, businessID int64) (domain.DashboardStats, error) {
	var stats domain.DashboardStats

	salesRow, err := r.queries.GetDashboardSalesToday(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get dashboard stats", logger.Error(err))
		return stats, err
	}

	stats.TotalSales = interfaceToFloat(salesRow.TotalSales)
	stats.OrderCount = int(salesRow.OrderCount)
	stats.AvgOrder = interfaceToFloat(salesRow.AvgOrder)

	turnover, err := r.queries.GetTableTurnover(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get table turnover", logger.Error(err))
		return stats, err
	}
	if turnover.TableCount > 0 {
		stats.TableTurnover = interfaceToFloat(turnover.TurnedTables) / float64(turnover.TableCount)
	}

	return stats, nil
}

// GetOrderSessionID returns the session associated with an order.
func (r *Repository) GetOrderSessionID(ctx context.Context, orderID int64) (int64, error) {
	sessionID, err := r.queries.GetOrderSessionID(ctx, orderID)
	if err != nil {
		return 0, err
	}
	if sessionID.Valid {
		return sessionID.Int64, nil
	}
	return 0, nil
}
