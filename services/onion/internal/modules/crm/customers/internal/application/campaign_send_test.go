package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
)

// campaignRepo implements the delivery calls; anything else panics.
type campaignRepo struct {
	repository
	campaign domain.CampaignToSend
	members  []domain.AudienceMember
	claimed  int
	released int
	recorded []int64
}

func (r *campaignRepo) ClaimCampaignForSend(context.Context, int64, int64) (domain.CampaignToSend, error) {
	r.claimed++
	return r.campaign, nil
}

func (r *campaignRepo) ReleaseCampaignClaim(context.Context, int64, int64) error {
	r.released++
	return nil
}

func (r *campaignRepo) CampaignAudience(context.Context, int64, *int64) ([]domain.AudienceMember, int, error) {
	return r.members, len(r.members) + 1, nil
}

func (r *campaignRepo) RecordCampaignSends(_ context.Context, _, _ int64, customerIDs []int64, _ int) error {
	r.recorded = customerIDs
	return nil
}

func newCampaignService(repo *campaignRepo, address string, enqueue func(CampaignEmail) error) *Service {
	return NewService(context.Background(), repo).WithCampaignDelivery(CampaignDelivery{
		Sender: func(context.Context, int64) (domain.CampaignSender, error) {
			return domain.CampaignSender{Name: "Cedar & Co", Address: address, ReplyTo: "hi@cedar.test"}, nil
		},
		Enqueue:           func(_ context.Context, email CampaignEmail) error { return enqueue(email) },
		UnsubscribeSecret: []byte("secret"),
		PublicAPIURL:      "https://api.example.test",
	})
}

func TestSendEmailCampaign_RefusesWithoutAPostalAddress(t *testing.T) {
	repo := &campaignRepo{}
	service := newCampaignService(repo, "  ", func(CampaignEmail) error { return nil })

	if _, err := service.SendEmailCampaign(context.Background(), 1, 9); !errors.Is(err, domain.ErrSenderAddressRequired) {
		t.Fatalf("err = %v, want ErrSenderAddressRequired", err)
	}
	if repo.claimed != 0 {
		t.Fatal("the campaign was claimed although it could not be sent")
	}
}

func TestSendEmailCampaign_RendersEachEmailForItsCustomer(t *testing.T) {
	repo := &campaignRepo{
		campaign: domain.CampaignToSend{ID: 9, Subject: "{{first_name}}, a treat", Content: "Hi {{first_name}}, you have {{points}} points.\n\n<b>Thanks</b>"},
		members: []domain.AudienceMember{
			{CustomerID: 1, FirstName: "Rania", Email: "rania@example.test", Points: 120},
			{CustomerID: 2, FirstName: "<Sam>", Email: "sam@example.test"},
		},
	}
	var sent []CampaignEmail
	service := newCampaignService(repo, "Hamra St 12, Beirut", func(email CampaignEmail) error {
		sent = append(sent, email)
		return nil
	})

	recipients, err := service.SendEmailCampaign(context.Background(), 1, 9)
	if err != nil || recipients != 2 {
		t.Fatalf("SendEmailCampaign = %d, %v; want 2, nil", recipients, err)
	}
	first := sent[0]
	if first.Subject != "Rania, a treat" || !strings.Contains(first.Text, "you have 120 points") {
		t.Errorf("merge tags not filled: %q / %q", first.Subject, first.Text)
	}
	if first.DedupeKey != "crm-campaign:9:1" || first.ReplyTo != "hi@cedar.test" || !strings.Contains(first.From, "Cedar & Co") {
		t.Errorf("headers = %q %q %q", first.DedupeKey, first.ReplyTo, first.From)
	}
	for _, part := range []string{"&lt;b&gt;Thanks&lt;/b&gt;", "Hamra St 12, Beirut", "https://api.example.test/crm/public/unsubscribe?t="} {
		if !strings.Contains(first.HTML, part) {
			t.Errorf("HTML is missing %q", part)
		}
	}
	if strings.Contains(sent[1].HTML, "<Sam>") {
		t.Error("a customer's name was put into the HTML unescaped")
	}
	if len(repo.recorded) != 2 {
		t.Errorf("recorded sends = %v, want both customers", repo.recorded)
	}
}

func TestSendEmailCampaign_ReleasesTheClaimWhenNothingCouldBeQueued(t *testing.T) {
	repo := &campaignRepo{
		campaign: domain.CampaignToSend{ID: 9, Subject: "Hi", Content: "Hello"},
		members:  []domain.AudienceMember{{CustomerID: 1, Email: "rania@example.test"}},
	}
	service := newCampaignService(repo, "Hamra St 12", func(CampaignEmail) error { return errors.New("queue down") })

	if _, err := service.SendEmailCampaign(context.Background(), 1, 9); err == nil {
		t.Fatal("expected the queue error")
	}
	if repo.released != 1 {
		t.Fatal("the campaign stayed marked sent although nothing was queued")
	}
}

func TestSendEmailCampaign_RefusesAnEmptyCampaign(t *testing.T) {
	repo := &campaignRepo{campaign: domain.CampaignToSend{ID: 9, Subject: "Hi"}}
	service := newCampaignService(repo, "Hamra St 12", func(CampaignEmail) error { return nil })

	if _, err := service.SendEmailCampaign(context.Background(), 1, 9); !errors.Is(err, domain.ErrCampaignIncomplete) {
		t.Fatalf("err = %v, want ErrCampaignIncomplete", err)
	}
	if repo.released != 1 {
		t.Fatal("an incomplete campaign was left marked sent")
	}
}
