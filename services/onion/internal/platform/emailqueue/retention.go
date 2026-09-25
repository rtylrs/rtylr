package emailqueue

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

// retentionInterval is how often the pool purges finished jobs.
const retentionInterval = time.Hour

// defaultRetention keeps a week of delivery history for troubleshooting
// before purging. Sent and failed rows are the only terminal states; rows
// holding reset or invite links therefore stop being readable from dumps
// within days of delivery instead of forever (#754).
const defaultRetention = 7 * 24 * time.Hour

// CleanupOld deletes sent/failed jobs older than maxAge. It runs on the
// pool's loop so every deployment that sends mail also prunes it.
func (p *Pool) CleanupOld(ctx context.Context, maxAge time.Duration) {
	cutoff := time.Now().UTC().Add(-maxAge)
	n, err := p.q.DeleteFinishedJobsOlderThan(ctx, cutoff)
	if err != nil {
		logger.Ctx(ctx).Error("email job retention purge failed", logger.Error(err))
		return
	}
	if n > 0 {
		logger.Ctx(ctx).Info("email job retention purge", logger.Any("deleted", n))
	}
}
