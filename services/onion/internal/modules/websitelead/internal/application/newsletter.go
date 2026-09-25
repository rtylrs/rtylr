package application

import (
	"context"
	"fmt"
	"html"
	"net/url"
	"time"

	"github.com/voxire/onion/internal/modules/websitelead/internal/domain"
	emailPkg "github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/logger"
)

// Subscribe stores a newsletter signup and tells the team about it. The row
// is the source of truth; the notification runs detached like lead emails.
func (s *Service) Subscribe(ctx context.Context, input domain.SubscriberInput) (int64, error) {
	id, err := s.repo.CreateSubscriber(ctx, input)
	if err != nil {
		return 0, err
	}
	go func() {
		taskCtx, cancel := asyncLeadContext(ctx)
		defer cancel()
		s.sendSubscriberNotification(taskCtx, id, input)
	}()
	return id, nil
}

func (s *Service) sendSubscriberNotification(ctx context.Context, id int64, sub domain.SubscriberInput) {
	esc := html.EscapeString
	rows := fieldRow("Email", fmt.Sprintf(`<a href="mailto:%s" style="color:#63c6bd;text-decoration:none;">%s</a>`, esc(sub.Email), esc(sub.Email)))
	rows += fieldRow("Source", esc(sub.Source))
	if sub.PostTitle != "" {
		rows += fieldRow("Post", esc(sub.PostTitle))
	}
	if sub.PostSlug != "" {
		rows += fieldRow("Slug", fmt.Sprintf(`<a href="https://voxire.com/blog/%s/" style="color:#63c6bd;text-decoration:none;">%s</a>`, esc(url.PathEscape(sub.PostSlug)), esc(sub.PostSlug)))
	}
	if sub.PostCategory != "" {
		rows += fieldRow("Category", esc(sub.PostCategory))
	}
	rows += fieldRow("Language", esc(sub.Language))

	now := time.Now().UTC().Format("2 Jan 2006, 15:04 UTC")
	body := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<body style="margin:0;padding:0;background:#0d0d14;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
<div style="max-width:620px;margin:0 auto;padding:40px 16px;">
  <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:24px;">
    <span style="font-size:18px;font-weight:700;color:#63c6bd;letter-spacing:-0.01em;">Voxire</span>
    <span style="font-size:12px;color:rgba(240,240,246,0.35);">%s</span>
  </div>
  <div style="background:#111118;border-radius:14px;padding:32px;border:1px solid rgba(255,255,255,0.06);">
    <h2 style="font-size:20px;font-weight:700;margin:0 0 6px 0;color:#f0f0f6;">New newsletter subscriber</h2>
    <p style="font-size:13px;color:rgba(240,240,246,0.4);margin:0 0 28px 0;">Subscriber #%d via voxire.com</p>
    <table style="width:100%%;border-collapse:collapse;">%s</table>
  </div>
  <p style="font-size:12px;color:rgba(240,240,246,0.2);text-align:center;margin-top:24px;">Voxire newsletter: voxire.com</p>
</div>
</body>
</html>`, now, id, rows)

	_, err := sendEmail(ctx, emailPkg.SendParams{
		To:      s.notify.NewsletterEmails,
		Subject: fmt.Sprintf("[Newsletter] %s via %s", sub.Email, sub.Source),
		HTML:    body,
		From:    "Voxire Website <noreply@voxire.com>",
		ReplyTo: sub.Email,
	})
	if err != nil {
		logger.Ctx(ctx).Warn("Failed to send newsletter subscriber notification",
			logger.Error(err),
			logger.Any("subscriber_id", id),
		)
	}
}
