package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/websitelead/internal/infrastructure"
	emailPkg "github.com/voxire/onion/internal/platform/email"
)

// audienceRepo satisfies both repository and audienceRepository.
type audienceRepo struct {
	mockRepo
	pending []infrastructure.PendingSubscriber
	synced  map[int64]string
	failed  map[int64]string
}

func newAudienceRepo(pending ...infrastructure.PendingSubscriber) *audienceRepo {
	return &audienceRepo{pending: pending, synced: map[int64]string{}, failed: map[int64]string{}}
}

func (r *audienceRepo) ListUnsyncedSubscribers(_ context.Context, maxAttempts, limit int) ([]infrastructure.PendingSubscriber, error) {
	if maxAttempts <= 0 || limit <= 0 {
		return nil, errors.New("bad bounds")
	}
	return r.pending, nil
}

func (r *audienceRepo) MarkSubscriberSynced(_ context.Context, id int64, contactID string) error {
	r.synced[id] = contactID
	return nil
}

func (r *audienceRepo) MarkSubscriberSyncFailed(_ context.Context, id int64, reason string) error {
	r.failed[id] = reason
	return nil
}

func withAudienceSeam(t *testing.T, fn func(ctx context.Context, audienceID, email string) (string, error)) {
	t.Helper()
	prev := addAudienceContact
	addAudienceContact = fn
	t.Cleanup(func() { addAudienceContact = prev })
}

func TestSyncResendAudience_NoAudienceConfiguredIsNoop(t *testing.T) {
	calls := 0
	withAudienceSeam(t, func(context.Context, string, string) (string, error) { calls++; return "c", nil })
	repo := newAudienceRepo(infrastructure.PendingSubscriber{ID: 1, Email: "a@b.co"})
	svc := &Service{repo: repo}
	if err := svc.SyncResendAudience(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 0 || len(repo.synced) != 0 {
		t.Fatalf("expected no provider calls without an audience id, got %d calls, %d synced", calls, len(repo.synced))
	}
}

func TestSyncResendAudience_MarksSyncedAndFailedPerRow(t *testing.T) {
	withAudienceSeam(t, func(_ context.Context, audienceID, email string) (string, error) {
		if audienceID != "aud_1" {
			t.Fatalf("audience id = %q", audienceID)
		}
		if email == "bad@example.com" {
			return "", errors.New("provider says no")
		}
		return "contact_" + email, nil
	})
	repo := newAudienceRepo(
		infrastructure.PendingSubscriber{ID: 1, Email: "one@example.com"},
		infrastructure.PendingSubscriber{ID: 2, Email: "bad@example.com"},
		infrastructure.PendingSubscriber{ID: 3, Email: "three@example.com"},
	)
	svc := &Service{repo: repo, audienceID: "aud_1"}
	if err := svc.SyncResendAudience(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.synced[1] != "contact_one@example.com" || repo.synced[3] != "contact_three@example.com" {
		t.Fatalf("synced = %v", repo.synced)
	}
	if _, ok := repo.synced[2]; ok {
		t.Fatalf("row 2 must not be marked synced")
	}
	if repo.failed[2] != "provider says no" {
		t.Fatalf("failed = %v", repo.failed)
	}
}

func TestSyncResendAudience_DisabledProviderStopsWithoutBurningAttempts(t *testing.T) {
	withAudienceSeam(t, func(context.Context, string, string) (string, error) { return "", emailPkg.ErrAudienceDisabled })
	repo := newAudienceRepo(infrastructure.PendingSubscriber{ID: 1, Email: "one@example.com"})
	svc := &Service{repo: repo, audienceID: "aud_1"}
	err := svc.SyncResendAudience(context.Background())
	if !errors.Is(err, emailPkg.ErrAudienceDisabled) {
		t.Fatalf("expected ErrAudienceDisabled, got %v", err)
	}
	if len(repo.failed) != 0 {
		t.Fatalf("a missing key must not count as a per-row failure, got %v", repo.failed)
	}
}

func TestSyncResendAudience_RepoWithoutAudienceSupport(t *testing.T) {
	svc := &Service{repo: &mockRepo{}, audienceID: "aud_1"}
	if err := svc.SyncResendAudience(context.Background()); err == nil {
		t.Fatal("expected an error for a repository without audience support")
	}
}
