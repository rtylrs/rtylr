package application

import (
	"context"
	"fmt"
	"html"
	"net/mail"
	"net/url"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// CampaignEmail is one rendered marketing email ready to queue.
type CampaignEmail struct {
	To        string
	Subject   string
	HTML      string
	Text      string
	From      string
	ReplyTo   string
	DedupeKey string
}

// campaignFromAddress is the verified address marketing email goes out from;
// the business's name is the display name and replies go to the business.
const campaignFromAddress = "noreply@voxire.com"

// CampaignDelivery is what sending needs from outside the module.
type CampaignDelivery struct {
	// Sender reads the business's display name, postal address and reply-to.
	Sender func(context.Context, int64) (domain.CampaignSender, error)
	// Enqueue queues one email; a repeated DedupeKey must be a no-op.
	Enqueue func(context.Context, CampaignEmail) error
	// UnsubscribeSecret signs unsubscribe links.
	UnsubscribeSecret []byte
	// PublicAPIURL is the origin unsubscribe links point at.
	PublicAPIURL string
}

func (s *Service) WithCampaignDelivery(delivery CampaignDelivery) *Service {
	s.delivery = &delivery
	return s
}

// CampaignAudience counts who an email campaign would reach right now.
func (s *Service) CampaignAudience(ctx context.Context, businessID int64, segmentID *int64) (domain.CampaignAudience, error) {
	members, total, err := s.repo.CampaignAudience(ctx, businessID, segmentID)
	if err != nil {
		return domain.CampaignAudience{}, err
	}
	return domain.CampaignAudience{Eligible: len(members), Total: total}, nil
}

// SendEmailCampaign sends an email campaign to every customer in its audience
// who gave marketing consent, and returns how many emails were queued.
//
// The campaign is claimed first, so a second request gets
// ErrCampaignNotSendable instead of a second copy. Each email is queued with a
// per-customer dedupe key, so retrying after a partial failure never sends a
// customer the campaign twice.
func (s *Service) SendEmailCampaign(ctx context.Context, businessID, campaignID int64) (int, error) {
	if s.delivery == nil || s.delivery.Sender == nil || s.delivery.Enqueue == nil {
		return 0, fmt.Errorf("campaign delivery is not configured")
	}
	sender, err := s.delivery.Sender(ctx, businessID)
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(sender.Address) == "" {
		return 0, domain.ErrSenderAddressRequired
	}

	campaign, err := s.repo.ClaimCampaignForSend(ctx, businessID, campaignID)
	if err != nil {
		return 0, err
	}
	release := func(cause error) (int, error) {
		if releaseErr := s.repo.ReleaseCampaignClaim(ctx, businessID, campaignID); releaseErr != nil {
			logger.Ctx(ctx).Error("failed to release campaign claim", logger.Error(releaseErr))
		}
		return 0, cause
	}
	if strings.TrimSpace(campaign.Subject) == "" || strings.TrimSpace(campaign.Content) == "" {
		return release(domain.ErrCampaignIncomplete)
	}

	members, total, err := s.repo.CampaignAudience(ctx, businessID, campaign.SegmentID)
	if err != nil {
		return release(err)
	}
	queued := make([]int64, 0, len(members))
	for _, member := range members {
		email := s.renderCampaignEmail(businessID, campaign, sender, member)
		if err := s.delivery.Enqueue(ctx, email); err != nil {
			if len(queued) == 0 {
				return release(err)
			}
			// Some emails are already queued; keep the campaign sent and
			// record them. A retry is safe because the dedupe keys hold.
			logger.Ctx(ctx).Error("campaign send stopped part-way", logger.Error(err),
				logger.Any("campaign_id", campaignID), logger.Int("queued", len(queued)))
			break
		}
		queued = append(queued, member.CustomerID)
	}
	if err := s.repo.RecordCampaignSends(ctx, businessID, campaignID, queued, total); err != nil {
		return len(queued), err
	}
	invalidateCustomerScope(ctx, businessID)
	return len(queued), nil
}

// Unsubscribe withdraws the marketing consent an unsubscribe link was signed
// for.
func (s *Service) Unsubscribe(ctx context.Context, token string) error {
	if s.delivery == nil {
		return domain.ErrInvalidUnsubscribe
	}
	businessID, customerID, err := domain.ParseUnsubscribeToken(s.delivery.UnsubscribeSecret, token)
	if err != nil {
		return err
	}
	if err := s.repo.UnsubscribeFromMarketing(ctx, businessID, customerID); err != nil {
		return err
	}
	invalidateCustomerScope(ctx, businessID)
	return nil
}

func (s *Service) renderCampaignEmail(businessID int64, campaign domain.CampaignToSend, sender domain.CampaignSender, member domain.AudienceMember) CampaignEmail {
	unsubscribeURL := s.delivery.PublicAPIURL + "/crm/public/unsubscribe?t=" +
		url.QueryEscape(domain.UnsubscribeToken(s.delivery.UnsubscribeSecret, businessID, member.CustomerID))
	values := map[string]string{
		"first_name":    member.FirstName,
		"last_name":     member.LastName,
		"points":        strconv.Itoa(member.Points),
		"tier":          member.Tier,
		"business_name": sender.Name,
	}
	subject := mergeTags(campaign.Subject, values)
	body := mergeTags(campaign.Content, values)

	from := mail.Address{Name: sender.Name, Address: campaignFromAddress}
	footerText := fmt.Sprintf("You are receiving this because you agreed to hear from %s.\n%s\n%s\nUnsubscribe: %s",
		sender.Name, sender.Name, sender.Address, unsubscribeURL)
	return CampaignEmail{
		To:        member.Email,
		Subject:   strings.ReplaceAll(subject, "\n", " "),
		HTML:      campaignHTML(body, sender, unsubscribeURL),
		Text:      body + "\n\n--\n" + footerText,
		From:      from.String(),
		ReplyTo:   sender.ReplyTo,
		DedupeKey: fmt.Sprintf("crm-campaign:%d:%d", campaign.ID, member.CustomerID),
	}
}

// mergeTags replaces {{name}} placeholders with the customer's values; an
// unknown placeholder is left as written.
func mergeTags(text string, values map[string]string) string {
	for key, value := range values {
		text = strings.ReplaceAll(text, "{{"+key+"}}", value)
		text = strings.ReplaceAll(text, "{{ "+key+" }}", value)
	}
	return text
}

// campaignHTML lays out a plain-text message as HTML paragraphs. The message
// is escaped: it is the business's text, not markup.
func campaignHTML(body string, sender domain.CampaignSender, unsubscribeURL string) string {
	var paragraphs strings.Builder
	for _, paragraph := range strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n\n") {
		if strings.TrimSpace(paragraph) == "" {
			continue
		}
		escaped := strings.ReplaceAll(html.EscapeString(paragraph), "\n", "<br>")
		paragraphs.WriteString(`<p style="margin:0 0 16px;line-height:1.6">` + escaped + "</p>")
	}
	name := html.EscapeString(sender.Name)
	return `<!doctype html><html><body style="margin:0;padding:24px;background:#f6f6f6;font-family:Helvetica,Arial,sans-serif;color:#111">` +
		`<div style="max-width:560px;margin:0 auto;background:#fff;border-radius:8px;padding:28px;font-size:15px">` +
		paragraphs.String() + `</div>` +
		`<p style="max-width:560px;margin:16px auto 0;font-size:12px;color:#777;line-height:1.5;text-align:center">` +
		`You are receiving this because you agreed to hear from ` + name + `.<br>` +
		name + ` · ` + html.EscapeString(sender.Address) + `<br>` +
		`<a href="` + html.EscapeString(unsubscribeURL) + `" style="color:#777">Unsubscribe</a></p>` +
		`</body></html>`
}
