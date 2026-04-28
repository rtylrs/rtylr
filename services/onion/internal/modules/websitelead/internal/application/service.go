package application

import (
	"context"
	"fmt"

	"github.com/voxire/onion/internal/modules/websitelead/internal/domain"
	"github.com/voxire/onion/internal/modules/websitelead/internal/infrastructure"
	emailPkg "github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/logger"
)

const notifyEmail = "info@voxire.com"

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{
		repo: infrastructure.NewRepository(ctx),
	}
}

func (s *Service) Submit(ctx context.Context, input domain.LeadInput) (int64, error) {
	id, err := s.repo.Create(ctx, input)
	if err != nil {
		return 0, err
	}

	go s.sendNotification(ctx, id, input)

	return id, nil
}

func (s *Service) sendNotification(ctx context.Context, id int64, lead domain.LeadInput) {
	html := fmt.Sprintf(`
<div style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;max-width:600px;margin:0 auto;padding:40px 20px;background:#f9f9f9;">
  <div style="background:#111118;border-radius:12px;padding:32px;color:#f0f0f6;">
    <h2 style="font-size:20px;font-weight:600;margin:0 0 4px 0;color:#63c6bd;">New quote request</h2>
    <p style="font-size:13px;color:rgba(240,240,246,0.55);margin:0 0 28px 0;">Lead #%d via voxire.com</p>

    <table style="width:100%%;border-collapse:collapse;">
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;width:120px;">Name</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;font-weight:500;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Email</td><td style="padding:8px 0;font-size:15px;"><a href="mailto:%s" style="color:#63c6bd;text-decoration:none;">%s</a></td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Company</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Service</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Budget</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Source</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
    </table>

    <div style="margin-top:24px;padding:16px;background:rgba(255,255,255,0.05);border-radius:8px;border-left:3px solid #63c6bd;">
      <p style="font-size:13px;color:rgba(240,240,246,0.55);margin:0 0 6px 0;">Project brief</p>
      <p style="font-size:14px;color:#f0f0f6;margin:0;line-height:1.6;">%s</p>
    </div>

    <a href="mailto:%s?subject=Re%%3A%%20Your%%20Voxire%%20Quote%%20Request" style="display:inline-block;margin-top:24px;background:#63c6bd;color:#111118;padding:12px 24px;border-radius:8px;text-decoration:none;font-size:14px;font-weight:600;">
      Reply to %s
    </a>
  </div>
</div>`,
		id,
		lead.Name,
		lead.Email, lead.Email,
		lead.Company,
		lead.Service,
		lead.Budget,
		lead.Source,
		lead.Project,
		lead.Email,
		lead.Name,
	)

	_, emailErr := emailPkg.Send(ctx, emailPkg.SendParams{
		To:      []string{notifyEmail},
		Subject: fmt.Sprintf("New quote request from %s (%s)", lead.Name, lead.Service),
		HTML:    html,
		From:    "Voxire Leads <noreply@voxire.com>",
		ReplyTo: lead.Email,
	})
	if emailErr != nil {
		logger.Ctx(ctx).Warn("Failed to send lead notification email",
			logger.Error(emailErr),
			logger.Any("lead_id", id),
		)
	}
}
