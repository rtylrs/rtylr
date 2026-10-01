package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
)

func TestInviteUser_RequiresEmail(t *testing.T) {
	s := &Service{}
	if _, _, err := s.InviteUser(context.Background(), 1, 0, domain.InviteUserInput{}); err == nil {
		t.Fatalf("expected error for missing email")
	}
}

func TestUpdatePermissions_RequiresUserID(t *testing.T) {
	s := &Service{}
	if err := s.UpdatePermissions(context.Background(), 1, 0, domain.UpdatePermissionsInput{}); err == nil {
		t.Fatalf("expected error for missing user id")
	}
}

func TestDeactivateUser_RequiresUserID(t *testing.T) {
	s := &Service{}
	if err := s.DeactivateUser(context.Background(), 1, 0, 0); err == nil {
		t.Fatalf("expected error for missing user id")
	}
}

func TestCreateGroup_RequiresName(t *testing.T) {
	s := &Service{}
	if _, err := s.CreateGroup(context.Background(), 1, 1, 1, domain.CreateGroupInput{Name: ""}); err == nil {
		t.Fatal("expected error for empty group name")
	}
}

func TestCreateGroup_NameRequired_GivesError(t *testing.T) {
	s := &Service{}
	_, err := s.CreateGroup(context.Background(), 1, 1, 1, domain.CreateGroupInput{Name: ""})
	if err == nil {
		t.Fatal("expected error for empty group name")
	}
}

func TestUpdateGroup_RequiresName(t *testing.T) {
	s := &Service{}
	if err := s.UpdateGroup(context.Background(), 1, 1, 1, 5, domain.UpdateGroupInput{Name: ""}); err == nil {
		t.Fatal("expected error for empty group name")
	}
}

func TestAddGroupMembers_RequiresUserIDs(t *testing.T) {
	s := &Service{}
	if err := s.AddGroupMembers(context.Background(), 1, 5, 1, 1, nil); err == nil {
		t.Fatal("expected error when no user IDs provided")
	}
}

func TestAddGroupMembers_RequiresNonEmptyUserIDs(t *testing.T) {
	s := &Service{}
	if err := s.AddGroupMembers(context.Background(), 1, 5, 1, 1, []int64{}); err == nil {
		t.Fatal("expected error for empty user IDs slice")
	}
}

func TestBulkInvite_SkipsEmptyEmails(t *testing.T) {
	s := &Service{}
	results, _ := s.BulkInvite(context.Background(), 1, 0, domain.BulkInviteInput{
		Emails:     []string{"", "  ", ""},
		MemberType: "member",
	})
	// All emails are blank, so all should be skipped and results be nil/empty.
	if len(results) != 0 {
		t.Errorf("expected 0 results for all-empty emails, got %d", len(results))
	}
}

func TestInviteUser_EmptyEmailStringFails(t *testing.T) {
	// Service checks Email == "" before any repo call.
	s := &Service{}
	_, _, err := s.InviteUser(context.Background(), 1, 0, domain.InviteUserInput{Email: ""})
	if err == nil {
		t.Fatal("expected error for empty email string")
	}
}

func TestInviteUser_RejectsOwnerMemberType(t *testing.T) {
	s := &Service{}
	_, _, err := s.InviteUser(context.Background(), 1, 0, domain.InviteUserInput{
		Email:      "new-owner@example.com",
		MemberType: "owner",
	})
	if err == nil || err.Error() != "invalid member type" {
		t.Fatalf("expected invalid member type for owner invite, got %v", err)
	}
}

func TestInviteUser_RejectsInvalidMemberType(t *testing.T) {
	s := &Service{}
	_, _, err := s.InviteUser(context.Background(), 1, 0, domain.InviteUserInput{
		Email:      "bad-type@example.com",
		MemberType: "super_admin",
	})
	if err == nil {
		t.Fatal("expected error for invalid invite role")
	}
}

func TestPaginateTeam_DefaultsPage(t *testing.T) {
	limit, offset := paginateTeam(0, 10)
	if limit != 10 || offset != 0 {
		t.Errorf("page=0 should default to page 1: limit=%d offset=%d", limit, offset)
	}
}

func TestPaginateTeam_DefaultsPageSize(t *testing.T) {
	limit, offset := paginateTeam(1, 0)
	if limit != 50 || offset != 0 {
		t.Errorf("pageSize=0 should default to 50: limit=%d offset=%d", limit, offset)
	}
}

func TestPaginateTeam_CapsPageSize(t *testing.T) {
	limit, _ := paginateTeam(1, 9999)
	if limit != 500 {
		t.Errorf("pageSize should be capped at 500, got %d", limit)
	}
}

func TestPaginateTeam_Page2Offset(t *testing.T) {
	limit, offset := paginateTeam(2, 25)
	if limit != 25 || offset != 25 {
		t.Errorf("page=2 pageSize=25: limit=%d offset=%d", limit, offset)
	}
}

func TestUpdatePermissions_ErrorMessage(t *testing.T) {
	s := &Service{}
	err := s.UpdatePermissions(context.Background(), 1, 0, domain.UpdatePermissionsInput{UserID: 0})
	if err == nil || err.Error() == "" {
		t.Fatal("expected non-empty error message")
	}
}

func TestDeactivateUser_ErrorMessage(t *testing.T) {
	s := &Service{}
	err := s.DeactivateUser(context.Background(), 1, 0, 0)
	if err == nil || err.Error() == "" {
		t.Fatal("expected non-empty error message")
	}
}

func TestInviteUser_ErrorMessage(t *testing.T) {
	s := &Service{}
	_, _, err := s.InviteUser(context.Background(), 1, 0, domain.InviteUserInput{})
	if err == nil || err.Error() == "" {
		t.Fatal("expected non-empty error message")
	}
}

func TestGetBusinessInfo_ZeroBusinessID(t *testing.T) {
	s := &Service{}
	name, ownerID, isOwner := s.GetBusinessInfo(context.Background(), 0, 1)
	if name != "" || ownerID != 0 || isOwner {
		t.Errorf("expected empty name, zero owner, and not owner for businessID=0: got name=%q ownerID=%d isOwner=%v", name, ownerID, isOwner)
	}
}

func TestCreateGroup_EmptyNameAlwaysErrors(t *testing.T) {
	s := &Service{}
	_, err := s.CreateGroup(context.Background(), 0, 0, 0, domain.CreateGroupInput{Name: ""})
	if err == nil {
		t.Fatal("expected error for empty name with zero ownerID")
	}
}

func TestAcceptInvite_RejectsWeakPassword(t *testing.T) {
	s := NewService(context.Background(), nil)
	for _, password := range []string{"", "short1A", "alllowercase1", "NOLOWERCASE1", "NoDigitsHere"} {
		if _, err := s.AcceptInvite(context.Background(), "any-token", password); !errors.Is(err, passwordutils.ErrWeakPassword) {
			t.Errorf("AcceptInvite(password=%q): got err=%v, want ErrWeakPassword", password, err)
		}
	}
}

type fakeSeatChangeStore struct {
	count     int
	recordErr error
	recorded  int
	calls     int
}

func (f *fakeSeatChangeStore) CountActiveMembers(context.Context, int64) (int, error) {
	return f.count, nil
}

func (f *fakeSeatChangeStore) RecordSeatChange(_ context.Context, _ int64, newCount int) error {
	f.calls++
	f.recorded = newCount
	return f.recordErr
}

func seedCachedSubscription(t *testing.T, ownerID int64) cacher.Cache[string] {
	t.Helper()
	ctx := context.Background()
	if err := cacher.InitializeLocalCache(ctx); err != nil {
		t.Fatalf("InitializeLocalCache: %v", err)
	}

	// The subscriptions module stores its payload under this key family; the
	// typed value does not matter here, only that the key is invalidated.
	cache := cacher.Of[string]()
	cache.Put(ctx, cacheKeys.Subscription(ownerID), "stale subscription payload",
		cacher.MemoryOnly(), cacher.WithExpiry(ctx, time.Minute))
	if _, ok := cache.Get(ctx, cacheKeys.Subscription(ownerID), cacher.MemoryOnly()); !ok {
		t.Fatal("failed to seed cached subscription")
	}
	return cache
}

// A team-driven seat change must invalidate the owner's cached subscription:
// the payload embeds seat counts and GET /hub/subscription/me otherwise serves
// the pre-change usage until the cache TTL expires.
func TestRecordSeatChange_InvalidatesCachedSubscription(t *testing.T) {
	const ownerID int64 = 4242
	cache := seedCachedSubscription(t, ownerID)

	store := &fakeSeatChangeStore{count: 5}
	if err := recordSeatChange(context.Background(), store, ownerID); err != nil {
		t.Fatalf("recordSeatChange: %v", err)
	}
	if store.calls != 1 || store.recorded != 5 {
		t.Fatalf("recorded seat count = %d after %d calls, want 5 after 1", store.recorded, store.calls)
	}

	if got, ok := cache.Get(context.Background(), cacheKeys.Subscription(ownerID), cacher.MemoryOnly()); ok {
		t.Fatalf("cached subscription survived the seat change: %+v", got)
	}
}

// A failed seat-change write leaves the database untouched, so the cached
// subscription must be kept (invalidating would only cause a pointless reload).
func TestRecordSeatChange_FailureKeepsCachedSubscription(t *testing.T) {
	const ownerID int64 = 4243
	cache := seedCachedSubscription(t, ownerID)

	store := &fakeSeatChangeStore{count: 5, recordErr: errors.New("write failed")}
	if err := recordSeatChange(context.Background(), store, ownerID); err == nil {
		t.Fatal("expected recordSeatChange to surface the write error")
	}

	if _, ok := cache.Get(context.Background(), cacheKeys.Subscription(ownerID), cacher.MemoryOnly()); !ok {
		t.Fatal("cached subscription was invalidated even though the seat change failed")
	}
}
