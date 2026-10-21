package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type SlowMoverRow = domain.SlowMoverRow

func (r *Repository) GetSlowMovers(ctx context.Context, businessID int64, days int32) ([]SlowMoverRow, error) {
	if days <= 0 {
		days = 60
	}
	if days > 365 {
		days = 365
	}
	cutoff := nowFn().AddDate(0, 0, -int(days))
	sold, err := r.sources.POS.InsightsProductsSoldSince(ctx, businessID, cutoff)
	if err != nil {
		logger.Ctx(ctx).Error("insights: slow movers sales failed", logger.Error(err))
		return nil, err
	}
	products, err := r.sources.ERP.InsightsActiveProducts(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("insights: slow movers products failed", logger.Error(err))
		return nil, err
	}
	recent := make(map[int64]bool, len(sold))
	for _, id := range sold {
		recent[id] = true
	}
	out := make([]SlowMoverRow, 0)
	for _, product := range products {
		if recent[product.ID] {
			continue
		}
		out = append(out, SlowMoverRow{ProductID: product.ID, Name: product.Name, StockQuantity: product.StockQuantity})
		if len(out) == slowMoverLimit {
			break
		}
	}
	return out, nil
}

// slowMoverLimit caps the slow-mover list.
const slowMoverLimit = 500
