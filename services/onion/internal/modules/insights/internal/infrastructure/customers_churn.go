package infrastructure

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type ChurnRateRow = domain.ChurnRateRow

// GetChurnRates classifies customers with a paid order before the period
// ended: those ordering inside the period are active, the rest are churned.
func (r *Repository) GetChurnRates(ctx context.Context, businessID int64, from, to time.Time) ([]ChurnRateRow, error) {
	from = dateOnly(from)
	to = dateOnly(to)
	active, err := r.sources.POS.InsightsBuyerCount(ctx, businessID, from, to)
	if err != nil {
		logger.Ctx(ctx).Error("insights: churn active count failed", logger.Error(err))
		return nil, err
	}
	churned, err := r.sources.POS.InsightsLapsedBuyerCount(ctx, businessID, from, to)
	if err != nil {
		logger.Ctx(ctx).Error("insights: churn lapsed count failed", logger.Error(err))
		return nil, err
	}

	total := active + churned
	var pct float64
	if total > 0 {
		pct = float64(churned) / float64(total) * 100
	}
	windowDays := int32(to.Sub(from).Hours()/24 + 1)
	return []ChurnRateRow{{
		WindowDays:   windowDays,
		ChurnedCount: churned,
		ActiveCount:  active,
		ChurnPct:     pct,
	}}, nil
}
