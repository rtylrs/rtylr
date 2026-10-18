package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/infrastructure"
	"github.com/voxire/onion/internal/modules/organization/business"
	"github.com/voxire/onion/internal/testutil"
)

func TestCampaignDelivery_ReachesOnlyConsentingCustomersOnce(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	ctx := context.Background()

	customer := func(name, email, consent string) int64 {
		id, err := repo.CreateCustomer(ctx, domain.CustomerInput{
			OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
			FirstName: name, Email: email, Consent: []byte(consent),
		})
		if err != nil {
			t.Fatalf("CreateCustomer %s: %v", name, err)
		}
		return id
	}
	optedIn := customer("Opted", "opted-in@send.test", `{"email":true,"marketing":true}`)
	noConsent := customer("Silent", "silent@send.test", `{"email":true}`)
	noEmail := customer("Offline", "", `{"marketing":true}`)

	members, total, err := repo.CampaignAudience(ctx, testutil.BusinessID, nil)
	if err != nil {
		t.Fatalf("CampaignAudience: %v", err)
	}
	reached := map[int64]bool{}
	for _, member := range members {
		reached[member.CustomerID] = true
	}
	if !reached[optedIn] || reached[noConsent] || reached[noEmail] {
		t.Fatalf("audience = %v; want the opted-in customer only (opted %d, silent %d, offline %d)",
			reached, optedIn, noConsent, noEmail)
	}
	if total < len(members)+2 {
		t.Errorf("total = %d, want it to count customers without consent too", total)
	}

	campaignID, err := repo.CreateCampaign(ctx, domain.CampaignInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID,
		Name: "Send once", Channel: "email", Subject: "Hi", Content: "Hello",
	})
	if err != nil {
		t.Fatalf("CreateCampaign: %v", err)
	}
	if _, err := repo.ClaimCampaignForSend(ctx, testutil.BusinessID, campaignID); err != nil {
		t.Fatalf("first claim: %v", err)
	}
	if _, err := repo.ClaimCampaignForSend(ctx, testutil.BusinessID, campaignID); !errors.Is(err, domain.ErrCampaignNotSendable) {
		t.Fatalf("second claim err = %v, want ErrCampaignNotSendable", err)
	}

	if err := repo.UnsubscribeFromMarketing(ctx, testutil.BusinessID, optedIn); err != nil {
		t.Fatalf("UnsubscribeFromMarketing: %v", err)
	}
	members, _, err = repo.CampaignAudience(ctx, testutil.BusinessID, nil)
	if err != nil {
		t.Fatalf("CampaignAudience after unsubscribe: %v", err)
	}
	for _, member := range members {
		if member.CustomerID == optedIn {
			t.Fatal("an unsubscribed customer is still in the audience")
		}
	}
	var emailConsent bool
	if err := db.QueryRowContext(ctx,
		"SELECT JSON_EXTRACT(consent, '$.email') = CAST('true' AS JSON) FROM crm_customers WHERE id = ?", optedIn,
	).Scan(&emailConsent); err != nil || !emailConsent {
		t.Fatalf("unsubscribing from marketing changed email consent: %v %v", emailConsent, err)
	}
}
