package infrastructure_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/organization/custom_domains/internal/domain"
	"github.com/voxire/onion/internal/modules/organization/custom_domains/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func claim(businessID int64, name, token string) domain.CustomDomain {
	return domain.CustomDomain{OwnerID: businessID, BusinessID: businessID, Domain: name, VerificationToken: token}
}

func TestClaimsRaceOnlyTheVerifiedOneWins(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	base := time.Now().UnixNano()
	squatter, owner := base, base+1
	name := fmt.Sprintf("claims-%d.example.com", base)
	t.Cleanup(func() {
		_ = repo.Remove(ctx, squatter)
		_ = repo.Remove(ctx, owner)
	})

	if _, err := repo.Register(ctx, claim(squatter, name, "squatter-token")); err != nil {
		t.Fatalf("squatter claim: %v", err)
	}
	if _, err := repo.Register(ctx, claim(owner, name, "owner-token")); err != nil {
		t.Fatalf("an unverified claim blocked the real owner: %v", err)
	}

	if err := repo.Verify(ctx, owner, "owner-token"); err != nil {
		t.Fatalf("owner verify: %v", err)
	}
	holder, err := repo.GetVerifiedByDomain(ctx, name)
	if err != nil || holder.BusinessID != owner {
		t.Fatalf("verified holder = %+v, err %v", holder, err)
	}
	if _, err := repo.Get(ctx, squatter); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("the competing pending claim survived verification: %v", err)
	}
}

func TestTwoBusinessesCannotBothBeVerified(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	base := time.Now().UnixNano()
	first, second := base, base+1
	name := fmt.Sprintf("both-%d.example.com", base)
	t.Cleanup(func() {
		_ = repo.Remove(ctx, first)
		_ = repo.Remove(ctx, second)
	})

	if _, err := repo.Register(ctx, claim(first, name, "t1")); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Register(ctx, claim(second, name, "t2")); err != nil {
		t.Fatal(err)
	}
	// Force both rows verified out of band: the unique index must refuse the second.
	if _, err := db.ExecContext(ctx, `UPDATE custom_domains SET status = 'verified' WHERE business_id = ?`, first); err != nil {
		t.Fatalf("first verified: %v", err)
	}
	if err := repo.Verify(ctx, second, "t2"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("second verify = %v, want ErrConflict", err)
	}
}

func TestChangingTheDomainRestartsTheChallenge(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	business := time.Now().UnixNano()
	t.Cleanup(func() { _ = repo.Remove(ctx, business) })
	oldName := fmt.Sprintf("old-%d.example.com", business)
	newName := fmt.Sprintf("new-%d.example.com", business)

	if _, err := repo.Register(ctx, claim(business, oldName, "t1")); err != nil {
		t.Fatal(err)
	}
	if err := repo.Verify(ctx, business, "t1"); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Register(ctx, claim(business, newName, "t2"))
	if err != nil {
		t.Fatalf("changing domain: %v", err)
	}
	if got.Domain != newName || got.Status == domain.StatusVerified || got.VerificationToken != "t2" {
		t.Fatalf("after change = %+v", got)
	}
}
