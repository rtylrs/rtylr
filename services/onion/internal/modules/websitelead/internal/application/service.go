package application

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/websitelead/internal/domain"
	"github.com/voxire/onion/internal/modules/websitelead/internal/infrastructure"
	emailPkg "github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/logger"
	sheetsPkg "github.com/voxire/onion/internal/platform/sheets"
)

// notifyEmails receives the internal lead notification.
var notifyEmails = []string{"info@voxire.com", "amounehabed@gmail.com"}

type repository interface {
	Create(ctx context.Context, input domain.LeadInput) (int64, error)
}

type Service struct {
	repo repository
}

const asyncLeadTaskTimeout = 10 * time.Second

func asyncLeadContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), asyncLeadTaskTimeout)
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{
		repo: infrastructure.NewRepository(db),
	}
}

func (s *Service) Submit(ctx context.Context, input domain.LeadInput) (int64, error) {
	id, err := s.repo.Create(ctx, input)
	if err != nil {
		return 0, err
	}

	go func() {
		taskCtx, cancel := asyncLeadContext(ctx)
		defer cancel()
		s.sendNotification(taskCtx, id, input)
	}()
	go func() {
		taskCtx, cancel := asyncLeadContext(ctx)
		defer cancel()
		s.sendAutoReply(taskCtx, input)
	}()
	go func() {
		taskCtx, cancel := asyncLeadContext(ctx)
		defer cancel()
		s.logToSheets(taskCtx, id, input)
	}()

	return id, nil
}

// ─── Internal helpers ────────────────────────────────────────────────────────

// fieldRow renders a label + value pair for the notification email table.
func fieldRow(label, value string) string {
	if value == "" {
		value = `<span style="color:rgba(240,240,246,0.3);font-style:italic;">not provided</span>`
	}
	return fmt.Sprintf(`
      <tr>
        <td style="padding:10px 0 10px 0;color:rgba(240,240,246,0.5);font-size:12px;font-weight:500;text-transform:uppercase;letter-spacing:0.06em;width:110px;vertical-align:top;">%s</td>
        <td style="padding:10px 0 10px 12px;font-size:15px;color:#f0f0f6;vertical-align:top;">%s</td>
      </tr>`, label, value)
}

// priorityBadge returns an HTML badge based on budget signal.
func priorityBadge(lead domain.LeadInput) string {
	budget := strings.ToLower(lead.Budget)
	switch {
	case strings.Contains(budget, "10,000") ||
		strings.Contains(budget, "10000") ||
		strings.Contains(budget, "15,000") ||
		strings.Contains(budget, "20,000") ||
		strings.Contains(budget, "25,000") ||
		strings.Contains(budget, "$10k") ||
		strings.Contains(budget, "$15k") ||
		strings.Contains(budget, "$20k"):
		return `<span style="display:inline-block;background:#ef4444;color:#fff;font-size:11px;font-weight:700;padding:3px 10px;border-radius:20px;letter-spacing:0.05em;text-transform:uppercase;">HOT</span>`
	case strings.Contains(budget, "5,000") ||
		strings.Contains(budget, "5000") ||
		strings.Contains(budget, "$5k") ||
		strings.Contains(budget, "7,000") ||
		strings.Contains(budget, "$7k"):
		return `<span style="display:inline-block;background:#f59e0b;color:#111;font-size:11px;font-weight:700;padding:3px 10px;border-radius:20px;letter-spacing:0.05em;text-transform:uppercase;">WARM</span>`
	default:
		return `<span style="display:inline-block;background:rgba(99,198,189,0.2);color:#63c6bd;font-size:11px;font-weight:700;padding:3px 10px;border-radius:20px;letter-spacing:0.05em;text-transform:uppercase;">NEW</span>`
	}
}

// ctaButton renders an email CTA anchor.
func ctaButton(href, label, bg, fg string) string {
	return fmt.Sprintf(
		`<a href="%s" style="display:inline-block;background:%s;color:%s;padding:11px 22px;border-radius:8px;text-decoration:none;font-size:13px;font-weight:600;margin-right:10px;">%s</a>`,
		href, bg, fg, label,
	)
}

// ─── Notification email ───────────────────────────────────────────────────────

func (s *Service) sendNotification(ctx context.Context, id int64, lead domain.LeadInput) {
	// Determine lead type label for the subject and badge.
	leadType := lead.Service
	if leadType == "contact" {
		leadType = "Contact message"
	}

	// Quick-action buttons.
	replyBtn := ctaButton(
		fmt.Sprintf("mailto:%s?subject=Re%%3A%%20Your%%20Voxire%%20Enquiry", lead.Email),
		"Reply to "+lead.Name,
		"#63c6bd", "#111118",
	)
	var phoneButtons string
	if lead.Phone != "" {
		cleanPhone := strings.ReplaceAll(lead.Phone, " ", "")
		callBtn := ctaButton("tel:"+cleanPhone, "Call "+lead.Name, "rgba(255,255,255,0.08)", "#f0f0f6")
		waBtn := ctaButton(
			fmt.Sprintf("https://wa.me/%s?text=Hi+%s%%2C+this+is+Abed+from+Voxire!", strings.TrimPrefix(cleanPhone, "+"), lead.Name),
			"WhatsApp",
			"#25d366", "#111118",
		)
		phoneButtons = callBtn + waBtn
	}

	// Build field rows.
	var rows strings.Builder
	rows.WriteString(fieldRow("Name", lead.Name))
	rows.WriteString(fieldRow("Email", fmt.Sprintf(`<a href="mailto:%s" style="color:#63c6bd;text-decoration:none;">%s</a>`, lead.Email, lead.Email)))
	if lead.Phone != "" {
		rows.WriteString(fieldRow("Phone", fmt.Sprintf(`<a href="tel:%s" style="color:#63c6bd;text-decoration:none;">%s</a>`, lead.Phone, lead.Phone)))
	}
	if lead.Company != "" {
		rows.WriteString(fieldRow("Company", lead.Company))
	}
	rows.WriteString(fieldRow("Service", lead.Service))
	if lead.Budget != "" {
		rows.WriteString(fieldRow("Budget", lead.Budget))
	}
	if lead.Source != "" {
		rows.WriteString(fieldRow("Source", lead.Source))
	}

	// Project brief block (only if non-empty).
	briefBlock := ""
	if strings.TrimSpace(lead.Project) != "" {
		briefBlock = fmt.Sprintf(`
    <div style="margin-top:24px;padding:18px 20px;background:rgba(99,198,189,0.07);border-radius:10px;border-left:3px solid #63c6bd;">
      <p style="font-size:11px;color:rgba(240,240,246,0.45);margin:0 0 8px 0;text-transform:uppercase;letter-spacing:0.06em;font-weight:500;">Message / Project Brief</p>
      <p style="font-size:14px;color:#f0f0f6;margin:0;line-height:1.7;white-space:pre-wrap;">%s</p>
    </div>`, lead.Project)
	}

	now := time.Now().UTC().Format("2 Jan 2006, 15:04 UTC")

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<body style="margin:0;padding:0;background:#0d0d14;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
<div style="max-width:620px;margin:0 auto;padding:40px 16px;">

  <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:24px;">
    <span style="font-size:18px;font-weight:700;color:#63c6bd;letter-spacing:-0.01em;">Voxire</span>
    <span style="font-size:12px;color:rgba(240,240,246,0.35);">%s</span>
  </div>

  <div style="background:#111118;border-radius:14px;padding:32px;border:1px solid rgba(255,255,255,0.06);">

    <div style="margin-bottom:6px;">
      <h2 style="font-size:20px;font-weight:700;margin:0 0 6px 0;color:#f0f0f6;">New %s &nbsp; %s</h2>
      <p style="font-size:13px;color:rgba(240,240,246,0.4);margin:0 0 28px 0;">Lead #%d via voxire.com</p>
    </div>

    <div style="height:1px;background:rgba(255,255,255,0.07);margin-bottom:20px;"></div>

    <table style="width:100%%;border-collapse:collapse;">
      %s
    </table>

    %s

    <div style="height:1px;background:rgba(255,255,255,0.07);margin:28px 0;"></div>

    <div>
      %s
      %s
    </div>

  </div>

  <p style="font-size:12px;color:rgba(240,240,246,0.2);text-align:center;margin-top:24px;">
    Voxire lead management &mdash; voxire.com
  </p>

</div>
</body>
</html>`,
		now,
		leadType,
		priorityBadge(lead),
		id,
		rows.String(),
		briefBlock,
		replyBtn,
		phoneButtons,
	)

	subject := fmt.Sprintf("[Lead #%d] %s from %s", id, leadType, lead.Name)
	if lead.Budget != "" {
		subject = fmt.Sprintf("[Lead #%d] %s from %s - %s", id, leadType, lead.Name, lead.Budget)
	}

	_, emailErr := emailPkg.Send(ctx, emailPkg.SendParams{
		To:      notifyEmails,
		Subject: subject,
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

// ─── Auto-reply to submitter ──────────────────────────────────────────────────

func (s *Service) sendAutoReply(ctx context.Context, lead domain.LeadInput) {
	// Build a brief summary of what they submitted.
	var summaryRows strings.Builder
	if lead.Service != "" && lead.Service != "contact" {
		fmt.Fprintf(&summaryRows, `<tr><td style="padding:7px 0;color:rgba(0,0,0,0.45);font-size:13px;width:120px;">Service</td><td style="padding:7px 0;font-size:14px;color:#111;font-weight:500;">%s</td></tr>`,
			lead.Service)
	}
	if lead.Budget != "" {
		fmt.Fprintf(&summaryRows, `<tr><td style="padding:7px 0;color:rgba(0,0,0,0.45);font-size:13px;">Budget</td><td style="padding:7px 0;font-size:14px;color:#111;">%s</td></tr>`,
			lead.Budget)
	}

	summaryBlock := ""
	if summaryRows.Len() > 0 {
		summaryBlock = fmt.Sprintf(`
    <div style="margin:24px 0;padding:16px 20px;background:#f5f5f7;border-radius:10px;">
      <p style="font-size:12px;color:rgba(0,0,0,0.45);margin:0 0 8px 0;text-transform:uppercase;letter-spacing:0.06em;font-weight:500;">Your request summary</p>
      <table style="width:100%%;border-collapse:collapse;">%s</table>
    </div>`, summaryRows.String())
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<body style="margin:0;padding:0;background:#f0f0f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
<div style="max-width:560px;margin:0 auto;padding:40px 16px;">

  <div style="background:#fff;border-radius:14px;padding:36px;box-shadow:0 2px 16px rgba(0,0,0,0.07);">

    <p style="font-size:20px;font-weight:700;color:#63c6bd;margin:0 0 28px 0;letter-spacing:-0.01em;">Voxire</p>

    <h2 style="font-size:22px;font-weight:700;color:#111;margin:0 0 12px 0;line-height:1.3;">
      Got it, %s.
    </h2>
    <p style="font-size:15px;color:#444;line-height:1.7;margin:0 0 20px 0;">
      Thanks for reaching out. We have received your message and one of our team will get back to you within <strong>24 hours</strong> (usually much faster).
    </p>

    %s

    <p style="font-size:14px;color:#555;line-height:1.7;margin:0 0 28px 0;">
      If your project is time-sensitive, you can reach Abed directly on WhatsApp:
    </p>

    <a href="https://wa.me/9613940708?text=Hi+Abed%%2C+I+just+submitted+a+request+on+voxire.com"
       style="display:inline-block;background:#25d366;color:#fff;padding:12px 24px;border-radius:8px;text-decoration:none;font-size:14px;font-weight:600;">
      WhatsApp Abed
    </a>

  </div>

  <p style="font-size:12px;color:rgba(0,0,0,0.3);text-align:center;margin-top:20px;">
    Voxire Digital Agency &middot; Beirut, Lebanon &middot;
    <a href="https://voxire.com" style="color:rgba(0,0,0,0.3);text-decoration:none;">voxire.com</a>
  </p>

</div>
</body>
</html>`,
		lead.Name,
		summaryBlock,
	)

	_, emailErr := emailPkg.Send(ctx, emailPkg.SendParams{
		To:      []string{lead.Email},
		Subject: "We received your message - Voxire",
		HTML:    html,
		From:    "Abed at Voxire <info@voxire.com>",
		ReplyTo: "info@voxire.com",
	})
	if emailErr != nil {
		logger.Ctx(ctx).Warn("Failed to send auto-reply to lead",
			logger.Error(emailErr),
		)
	}
}

// ─── Google Sheets logging ────────────────────────────────────────────────────

func (s *Service) logToSheets(ctx context.Context, id int64, lead domain.LeadInput) {
	if !sheetsPkg.IsEnabled() {
		return
	}

	now := time.Now().UTC().Format("2006-01-02 15:04:05")

	sheetsPkg.AppendLeadRow(ctx, []interface{}{
		id,
		now,
		lead.Name,
		lead.Email,
		lead.Phone,
		lead.Company,
		lead.Service,
		lead.Budget,
		lead.Source,
		lead.Project,
	})
}
