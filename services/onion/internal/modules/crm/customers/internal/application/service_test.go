package application

import (
	"context"
	"testing"

	crmcontract "github.com/voxire/onion/internal/contracts/crm"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
)

func TestRecalculateTier_WithExplicitConfig(t *testing.T) {
	cfg := domain.LoyaltyConfig{
		SilverMin:   100,
		GoldMin:     500,
		PlatinumMin: 2000,
	}
	cases := []struct {
		points int
		want   string
	}{
		{0, domain.TierBronze},
		{50, domain.TierBronze},
		{99, domain.TierBronze},
		{100, domain.TierSilver},
		{499, domain.TierSilver},
		{500, domain.TierGold},
		{1999, domain.TierGold},
		{2000, domain.TierPlatinum},
		{1_000_000, domain.TierPlatinum},
	}
	for _, tc := range cases {
		if got := RecalculateTier(tc.points, cfg); got != tc.want {
			t.Errorf("RecalculateTier(%d) = %q, want %q", tc.points, got, tc.want)
		}
	}
}

func TestRecalculateTier_DefaultsWhenConfigZero(t *testing.T) {
	// Empty config: defaults are silver=500, gold=1500, platinum=5000.
	cfg := domain.LoyaltyConfig{}
	cases := []struct {
		points int
		want   string
	}{
		{0, domain.TierBronze},
		{499, domain.TierBronze},
		{500, domain.TierSilver},
		{1500, domain.TierGold},
		{5000, domain.TierPlatinum},
	}
	for _, tc := range cases {
		if got := RecalculateTier(tc.points, cfg); got != tc.want {
			t.Errorf("default RecalculateTier(%d) = %q, want %q", tc.points, got, tc.want)
		}
	}
}

func TestLoyaltyPointsFromCents_RoundsCurrencyBeforeRate(t *testing.T) {
	if got := loyaltyPointsFromCents(99.99, 0.1); got != 10 {
		t.Fatalf("loyaltyPointsFromCents(99.99, 0.1) = %d, want 10", got)
	}
}

type fakePublicContactRepo struct {
	repository
	loyaltyCfg domain.LoyaltyConfig
	awarded    []domain.LoyaltyTransaction
	upsertedID int64
}

func (f *fakePublicContactRepo) UpsertDirectContact(_ context.Context, _, _ int64, _ crmcontract.CustomerContactCaptured) (int64, error) {
	return f.upsertedID, nil
}

func (f *fakePublicContactRepo) GetLoyaltyConfig(_ context.Context, _ int64) (domain.LoyaltyConfig, error) {
	return f.loyaltyCfg, nil
}

func (f *fakePublicContactRepo) ApplyLoyaltyDelta(_ context.Context, tx domain.LoyaltyTransaction, _ int, _ func(int) string, onCommitted func()) error {
	f.awarded = append(f.awarded, tx)
	if onCommitted != nil {
		onCommitted()
	}
	return nil
}

func (f *fakePublicContactRepo) InsertTimelineEntry(_ context.Context, _ domain.TimelineEntry) error {
	return nil
}

type stubOwnerResolver struct{ ownerID int64 }

func (r stubOwnerResolver) GetBusinessOwnerID(_ context.Context, _ int64) (int64, error) {
	return r.ownerID, nil
}

func TestCapturePublicContactAwardsUnifiedLoyaltyPointsOnOrder(t *testing.T) {
	repo := &fakePublicContactRepo{
		upsertedID: 77,
		loyaltyCfg: domain.LoyaltyConfig{
			IsEnabled:             true,
			PointsPerCurrencyUnit: 2,
			SilverMin:             100,
			GoldMin:               500,
			PlatinumMin:           1000,
		},
	}
	svc := NewService(context.Background(), repo).WithBusinessOwnerResolver(stubOwnerResolver{ownerID: 5})

	// Browse-time "Stay in touch" signup (CountVisit=false) must NOT award points.
	id, err := svc.CapturePublicContact(context.Background(), 101, crmcontract.CustomerContactCaptured{
		Name:  "Browse Guest",
		Phone: "+15550001111",
	})
	if err != nil || id != 77 {
		t.Fatalf("browse capture failed: id=%d err=%v", id, err)
	}
	if len(repo.awarded) != 0 {
		t.Fatalf("expected 0 loyalty awards on browse signup, got %+v", repo.awarded)
	}

	// Online menu order (CountVisit=true, TotalSpent=12.50) awards 25 unified CRM points.
	id, err = svc.CapturePublicContact(context.Background(), 101, crmcontract.CustomerContactCaptured{
		Name:       "Order Guest",
		Phone:      "+15550002222",
		CountVisit: true,
		TotalSpent: 12.50,
	})
	if err != nil || id != 77 {
		t.Fatalf("order capture failed: id=%d err=%v", id, err)
	}
	if len(repo.awarded) != 1 || repo.awarded[0].Points != 25 || repo.awarded[0].CustomerID != 77 {
		t.Fatalf("expected 1 loyalty award of 25 points for customer 77, got %+v", repo.awarded)
	}
}
