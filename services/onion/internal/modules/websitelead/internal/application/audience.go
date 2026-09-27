package application

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/websitelead/internal/infrastructure"
	emailPkg "github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/logger"
)

// The audience mirror is a job, not a hook on Subscribe: the row is the
// source of truth, and a job that walks rows with resend_synced_at IS NULL
// backfills the list that existed before the audience did, retries transient
// provider failures, and cannot lose a signup to a dropped goroutine.
const (
	audienceSyncBatch       = 200
	audienceSyncMaxAttempts = 5
)

// addAudienceContact is the outbound seam for tests.
var addAudienceContact = emailPkg.AddAudienceContact

type audienceRepository interface {
	ListUnsyncedSubscribers(ctx context.Context, maxAttempts, limit int) ([]infrastructure.PendingSubscriber, error)
	MarkSubscriberSynced(ctx context.Context, id int64, contactID string) error
	MarkSubscriberSyncFailed(ctx context.Context, id int64, reason string) error
}

// SyncResendAudience mirrors unsynced subscribers into the configured Resend
// audience. With no audience id configured it is a no-op, so self-hosted
// deployments and local runs stay quiet.
func (s *Service) SyncResendAudience(ctx context.Context) error {
	if s.audienceID == "" {
		return nil
	}
	repo, ok := s.repo.(audienceRepository)
	if !ok {
		return errors.New("websitelead repository does not support audience sync")
	}
	pending, err := repo.ListUnsyncedSubscribers(ctx, audienceSyncMaxAttempts, audienceSyncBatch)
	if err != nil {
		return err
	}
	if len(pending) == 0 {
		return nil
	}
	synced, failed := 0, 0
	for _, sub := range pending {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		contactID, err := addAudienceContact(ctx, s.audienceID, sub.Email)
		if err != nil {
			if errors.Is(err, emailPkg.ErrAudienceDisabled) {
				return err
			}
			failed++
			logger.Ctx(ctx).Warn("newsletter audience sync failed for subscriber",
				logger.Any("subscriber_id", sub.ID), logger.Error(err))
			if markErr := repo.MarkSubscriberSyncFailed(ctx, sub.ID, err.Error()); markErr != nil {
				return markErr
			}
			continue
		}
		if err := repo.MarkSubscriberSynced(ctx, sub.ID, contactID); err != nil {
			return err
		}
		synced++
	}
	logger.Ctx(ctx).Info("newsletter audience sync finished",
		logger.Int("synced", synced), logger.Int("failed", failed), logger.Int("pending", len(pending)))
	return nil
}
