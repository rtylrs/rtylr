package email

// Template identifies an email template by its path under the email-templates
// submodule's templates/ directory, without the .html extension. These constants
// are the single source of truth for template names — pass one to SendTemplate or
// EnqueueTemplate rather than hardcoding the path string anywhere else.
//
// Adding a template: drop the .html in the submodule under templates/<domain>/,
// add a const here, add it to AllTemplates, and add a case to render_test.go.
type Template string

const (
	// TemplateWelcome is sent on signup.
	TemplateWelcome Template = "welcome/welcome"
	// TemplateInvitation is sent when a user is invited to a business.
	TemplateInvitation Template = "team/invitation"
	// TemplatePasswordReset is sent on a password-reset request.
	TemplatePasswordReset Template = "auth/password_reset"
	// TemplateInvoice is sent with an invoice.
	TemplateInvoice Template = "billing/invoice"
	// TemplateAccessRequest notifies a business owner that a user requested app access.
	TemplateAccessRequest Template = "team/access_request"
)

// AllTemplates lists every registered template. render_test.go iterates it, so a
// new const that isn't added here (or whose file is missing) fails the build.
var AllTemplates = []Template{
	TemplateWelcome,
	TemplateInvitation,
	TemplatePasswordReset,
	TemplateInvoice,
	TemplateAccessRequest,
}

// The data shapes each template expects. Every template's base layout reads
// .Subject for the <title>; the rest are the body's fields.

// WelcomeData is the data for TemplateWelcome.
type WelcomeData struct {
	Subject string
	Name    string
}

// InvitationData is the data for TemplateInvitation.
type InvitationData struct {
	Subject      string
	InviterName  string
	BusinessName string
	InviteLink   string
}

// PasswordResetData is the data for TemplatePasswordReset.
type PasswordResetData struct {
	Subject   string
	ResetLink string
}

// InvoiceData is the data for TemplateInvoice.
type InvoiceData struct {
	Subject       string
	CustomerName  string
	InvoiceNumber string
	Amount        string
	DueDate       string
	ViewLink      string
}

// AccessRequestData is the data for TemplateAccessRequest.
type AccessRequestData struct {
	Subject        string
	RequesterName  string
	RequesterEmail string
	App            string
	TeamLink       string
}
