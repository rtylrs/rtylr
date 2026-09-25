package application

import (
	"context"
	"database/sql"
	"fmt"
	"html"

	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
	"github.com/voxire/onion/internal/modules/voxhire/internal/infrastructure"
	emailPkg "github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/logger"
)

// notifyEmails are the recipients for every VoxHire submission. Keep
// this short; loud notifications matter for a service in launch mode.
var notifyEmails = []string{"hire@voxire.com", "info@voxire.com", "amounehabed@gmail.com"}

// sendNotificationEmail delivers a background notification email. It is a
// variable so tests can assert that the context handed to the notification
// goroutine outlives the HTTP request that triggered it.
var sendNotificationEmail = emailPkg.Send

type repository interface {
	CreateCompanyLead(ctx context.Context, input domain.CompanyLeadInput) (int64, error)
	CreateTalentApplication(ctx context.Context, input domain.TalentApplicationInput) (int64, error)

	// Recruiter-admin surface.
	CreateAdmin(ctx context.Context, email, passwordHash, name string) (int64, error)
	GetAdminByEmail(ctx context.Context, email string) (*domain.Admin, error)
	GetAdminByID(ctx context.Context, id int64) (*domain.Admin, error)
	ListAdmins(ctx context.Context) ([]domain.Admin, error)
	CountAdmins(ctx context.Context) (int64, error)
	ListCompanyLeads(ctx context.Context) ([]domain.Lead, error)
	ListOpenings(ctx context.Context) ([]domain.Opening, error)
	ListOpeningsPage(ctx context.Context, limit, offset int32) (domain.OpeningPage, error)
	GetOpening(ctx context.Context, id int64) (*domain.Opening, error)
	GetAnalytics(ctx context.Context) (*domain.Analytics, error)
	CreateCareer(ctx context.Context, in domain.CareerInput) (*domain.Opening, error)
	UpdateCareer(ctx context.Context, in domain.CareerInput) (*domain.Opening, error)
	SetCareerActive(ctx context.Context, id int64, active bool) (*domain.Opening, error)
	ListApplications(ctx context.Context, f domain.ApplicationFilter) ([]domain.Application, int64, error)
	ListCandidates(ctx context.Context, f domain.CandidateFilter) ([]domain.Candidate, int64, error)
	ListCVJobs(ctx context.Context, f domain.CandidateFilter) ([]domain.CVJob, error)
	RetryCVProcessing(ctx context.Context, source string, id int64) error
	UpdateApplicationStatus(ctx context.Context, id int64, status string) error
	UpdateApplicationReview(ctx context.Context, id int64, in domain.ApplicationReviewInput) error
	GetApplicationCVKey(ctx context.Context, id int64) (string, error)
	ListApplicationTimeline(ctx context.Context, id int64) ([]domain.TimelineEvent, error)
	ListTasks(ctx context.Context, status string) ([]domain.Task, error)
	UpsertTask(ctx context.Context, in domain.TaskInput) (*domain.Task, error)
	UpdateTask(ctx context.Context, in domain.TaskInput) (*domain.Task, error)
	ListShortlists(ctx context.Context, jobID int64) ([]domain.Shortlist, error)
	CreateShortlist(ctx context.Context, in domain.ShortlistInput) (*domain.Shortlist, error)
	AddShortlistItem(ctx context.Context, in domain.ShortlistItemInput) (*domain.ShortlistItem, error)
	CreateCareerApplication(ctx context.Context, in domain.CareerApplicationInput) (int64, error)
	ListPendingCVProcessingTargets(ctx context.Context, limit int) ([]domain.CVProcessingTarget, error)
	UpdateCVProcessingResult(ctx context.Context, result domain.CVProcessingResult) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

// SubmitCompanyLead persists a company hiring inquiry and fires a
// notification email in the background. Email failures don't block the
// response.
func (s *Service) SubmitCompanyLead(ctx context.Context, input domain.CompanyLeadInput) (int64, error) {
	id, err := s.repo.CreateCompanyLead(ctx, input)
	if err != nil {
		return 0, err
	}

	// Detach from the request context: the goroutine outlives the handler and
	// would otherwise be canceled before the email is sent.
	go s.notifyCompanyLead(context.WithoutCancel(ctx), id, input)

	return id, nil
}

// SubmitTalentApplication persists a talent application and fires a
// notification email in the background.
func (s *Service) SubmitTalentApplication(ctx context.Context, input domain.TalentApplicationInput) (int64, error) {
	id, err := s.repo.CreateTalentApplication(ctx, input)
	if err != nil {
		return 0, err
	}

	// See SubmitCompanyLead: the notification must outlive the request.
	go s.notifyTalentApplication(context.WithoutCancel(ctx), id, input)

	return id, nil
}

func (s *Service) notifyCompanyLead(ctx context.Context, id int64, lead domain.CompanyLeadInput) {
	emailHTML := fmt.Sprintf(`
<div style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;max-width:600px;margin:0 auto;padding:40px 20px;background:#f9f9f9;">
  <div style="background:#111118;border-radius:12px;padding:32px;color:#f0f0f6;">
    <h2 style="font-size:20px;font-weight:600;margin:0 0 4px 0;color:#63c6bd;">New VoxHire hiring inquiry</h2>
    <p style="font-size:13px;color:rgba(240,240,246,0.55);margin:0 0 28px 0;">Lead #%d via hire.voxire.com</p>

    <table style="width:100%%;border-collapse:collapse;">
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;width:140px;">Company</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;font-weight:500;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Contact</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Work email</td><td style="padding:8px 0;font-size:15px;"><a href="mailto:%s" style="color:#63c6bd;text-decoration:none;">%s</a></td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Phone</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Role</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;font-weight:500;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Headcount</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Timeline</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Service</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Agency opt-in</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
    </table>

    <div style="margin-top:24px;padding:16px;background:rgba(255,255,255,0.05);border-radius:8px;border-left:3px solid #63c6bd;">
      <p style="font-size:13px;color:rgba(240,240,246,0.55);margin:0 0 6px 0;">Their note</p>
      <p style="font-size:14px;color:#f0f0f6;margin:0;line-height:1.6;">%s</p>
    </div>

    <a href="mailto:%s?subject=Re%%3A%%20Your%%20VoxHire%%20Inquiry" style="display:inline-block;margin-top:24px;background:#63c6bd;color:#111118;padding:12px 24px;border-radius:8px;text-decoration:none;font-size:14px;font-weight:600;">
      Reply to %s
    </a>
  </div>
</div>`,
		id,
		html.EscapeString(lead.CompanyName),
		html.EscapeString(lead.ContactName),
		formatTitle(lead.ContactTitle),
		html.EscapeString(lead.WorkEmail), html.EscapeString(lead.WorkEmail),
		html.EscapeString(orDash(lead.Phone)),
		html.EscapeString(lead.Role),
		html.EscapeString(orDash(lead.Headcount)),
		html.EscapeString(orDash(lead.Timeline)),
		html.EscapeString(orDash(lead.Service)),
		boolLabel(lead.AgencyOptIn, "Yes - also wants to talk to Voxire agency", "No"),
		html.EscapeString(orDash(lead.Message)),
		html.EscapeString(lead.WorkEmail),
		html.EscapeString(lead.ContactName),
	)

	subject := fmt.Sprintf("[VoxHire] New hiring inquiry from %s", lead.CompanyName)

	if _, err := sendNotificationEmail(ctx, emailPkg.SendParams{
		To:      notifyEmails,
		Subject: subject,
		HTML:    emailHTML,
		From:    "VoxHire Leads <noreply@voxire.com>",
		ReplyTo: lead.WorkEmail,
	}); err != nil {
		logger.Ctx(ctx).Warn("Failed to send VoxHire company lead notification email",
			logger.Error(err),
			logger.Any("lead_id", id),
		)
	}
}

func (s *Service) notifyTalentApplication(ctx context.Context, id int64, app domain.TalentApplicationInput) {
	cvLine := "Not attached"
	if app.CVKey != "" {
		cvLine = fmt.Sprintf("Stored at R2 key: <code>%s</code>", html.EscapeString(app.CVKey))
	}

	emailHTML := fmt.Sprintf(`
<div style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;max-width:600px;margin:0 auto;padding:40px 20px;background:#f9f9f9;">
  <div style="background:#111118;border-radius:12px;padding:32px;color:#f0f0f6;">
    <h2 style="font-size:20px;font-weight:600;margin:0 0 4px 0;color:#63c6bd;">New VoxHire talent application</h2>
    <p style="font-size:13px;color:rgba(240,240,246,0.55);margin:0 0 28px 0;">Application #%d via hire.voxire.com</p>

    <table style="width:100%%;border-collapse:collapse;">
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;width:160px;">Name</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;font-weight:500;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Email</td><td style="padding:8px 0;font-size:15px;"><a href="mailto:%s" style="color:#63c6bd;text-decoration:none;">%s</a></td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Phone</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Country</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Nationality</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Track</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;font-weight:500;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Years experience</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Notice period</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Open to</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Current role</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Desired roles</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Salary expectation</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Authorization/location notes</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">LinkedIn</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">Portfolio</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
      <tr><td style="padding:8px 0;color:rgba(240,240,246,0.55);font-size:13px;">CV</td><td style="padding:8px 0;font-size:15px;color:#f0f0f6;">%s</td></tr>
    </table>

    <div style="margin-top:24px;padding:16px;background:rgba(255,255,255,0.05);border-radius:8px;border-left:3px solid #63c6bd;">
      <p style="font-size:13px;color:rgba(240,240,246,0.55);margin:0 0 6px 0;">Why they are a strong match</p>
      <p style="font-size:14px;color:#f0f0f6;margin:0;line-height:1.6;">%s</p>
    </div>

    <div style="margin-top:16px;padding:16px;background:rgba(255,255,255,0.05);border-radius:8px;border-left:3px solid #63c6bd;">
      <p style="font-size:13px;color:rgba(240,240,246,0.55);margin:0 0 6px 0;">Work history</p>
      <p style="font-size:14px;color:#f0f0f6;margin:0;line-height:1.6;">%s</p>
    </div>

    <div style="margin-top:16px;padding:16px;background:rgba(255,255,255,0.05);border-radius:8px;border-left:3px solid #63c6bd;">
      <p style="font-size:13px;color:rgba(240,240,246,0.55);margin:0 0 6px 0;">Education and certifications</p>
      <p style="font-size:14px;color:#f0f0f6;margin:0;line-height:1.6;">%s</p>
    </div>

    <a href="mailto:%s?subject=VoxHire%%20-%%20Following%%20up%%20on%%20your%%20application" style="display:inline-block;margin-top:24px;background:#63c6bd;color:#111118;padding:12px 24px;border-radius:8px;text-decoration:none;font-size:14px;font-weight:600;">
      Reply to %s
    </a>
  </div>
</div>`,
		id,
		html.EscapeString(app.FullName),
		html.EscapeString(app.Email), html.EscapeString(app.Email),
		html.EscapeString(orDash(app.Phone)),
		html.EscapeString(orDash(app.CountryOfResidence)),
		html.EscapeString(orDash(app.Nationality)),
		html.EscapeString(orDash(app.Track)),
		html.EscapeString(orDash(app.YearsExperience)),
		html.EscapeString(orDash(app.NoticePeriod)),
		html.EscapeString(orDash(app.OpenTo)),
		html.EscapeString(orDash(app.CurrentRole)),
		html.EscapeString(orDash(app.DesiredRoles)),
		html.EscapeString(orDash(app.SalaryExpectation)),
		html.EscapeString(orDash(app.WorkAuthorization)),
		linkOrDash(app.LinkedinURL),
		linkOrDash(app.PortfolioURL),
		cvLine,
		html.EscapeString(orDash(app.Pitch)),
		html.EscapeString(orDash(app.WorkHistory)),
		html.EscapeString(orDash(app.Education)),
		html.EscapeString(app.Email),
		html.EscapeString(app.FullName),
	)

	subject := fmt.Sprintf("[VoxHire] New talent application from %s (%s)", app.FullName, app.Track)

	if _, err := sendNotificationEmail(ctx, emailPkg.SendParams{
		To:      notifyEmails,
		Subject: subject,
		HTML:    emailHTML,
		From:    "VoxHire Talent <noreply@voxire.com>",
		ReplyTo: app.Email,
	}); err != nil {
		logger.Ctx(ctx).Warn("Failed to send VoxHire talent application notification email",
			logger.Error(err),
			logger.Any("application_id", id),
		)
	}
}

func formatTitle(title string) string {
	if title == "" {
		return ""
	}
	return " (" + html.EscapeString(title) + ")"
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func linkOrDash(url string) string {
	if url == "" {
		return "-"
	}
	safe := html.EscapeString(url)
	return fmt.Sprintf(`<a href="%s" style="color:#63c6bd;text-decoration:none;">%s</a>`, safe, safe)
}

func boolLabel(b bool, yes, no string) string {
	if b {
		return yes
	}
	return no
}
