package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/organization/user/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestUser_AvatarRoundTripAndBusinessList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	t.Cleanup(func() { _ = repo.SetUserAvatar(ctx, testutil.OwnerID, "") })

	if got, err := repo.GetUserAvatar(ctx, testutil.OwnerID); err != nil || got != "" {
		t.Fatalf("default avatar = %q, %v; want empty", got, err)
	}

	stored := `{"kind":"emoji","emoji":"🗿","color":2}`
	if err := repo.SetUserAvatar(ctx, testutil.OwnerID, stored); err != nil {
		t.Fatalf("SetUserAvatar: %v", err)
	}
	if got, err := repo.GetUserAvatar(ctx, testutil.OwnerID); err != nil || got != stored {
		t.Fatalf("stored avatar = %q, %v; want %q", got, err, stored)
	}

	avatars, err := repo.ListBusinessUserAvatars(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("ListBusinessUserAvatars: %v", err)
	}
	if avatars[testutil.OwnerID] != stored {
		t.Fatalf("business list = %v; want the owner's avatar", avatars)
	}

	// Another business does not see this owner.
	other, err := repo.ListBusinessUserAvatars(ctx, testutil.BusinessID+999999)
	if err != nil {
		t.Fatalf("ListBusinessUserAvatars other business: %v", err)
	}
	if _, leaked := other[testutil.OwnerID]; leaked {
		t.Fatal("an unrelated business listed the owner's avatar")
	}

	if err := repo.SetUserAvatar(ctx, testutil.OwnerID, ""); err != nil {
		t.Fatalf("reset: %v", err)
	}
	if got, _ := repo.GetUserAvatar(ctx, testutil.OwnerID); got != "" {
		t.Fatalf("after reset = %q; want empty", got)
	}
}
