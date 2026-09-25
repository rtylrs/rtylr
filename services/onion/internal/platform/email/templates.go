package email

import (
	"io/fs"
	"sort"
	"strings"
)

// Template identifies an email template by its path under the embedded
// templates/ directory, without the .html extension (e.g. "welcome/welcome").
// Templates are discovered from the embedded FS at runtime — dropping an .html
// under templates/<domain>/ registers it automatically, no const needed. The
// consts below are the templates wired into Go code (they have typed data
// structs); everything embedded is reachable via Available/Exists and the
// generic SendTemplate/EnqueueTemplate with a data map.
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

// wiredTemplates are the templates referenced from Go code — each has a typed
// data struct and a render_test case. render_test.go iterates this, so a new
// const here without a test case fails the build.
var wiredTemplates = []Template{
	TemplateWelcome,
	TemplateInvitation,
	TemplatePasswordReset,
	TemplateInvoice,
	TemplateAccessRequest,
}

// fsTemplateRoot is where body templates live in the embedded FS: the templates
// directory keeps its own templates/<domain>/ subdir.
const fsTemplateRoot = "templates/templates"

// Exists reports whether tmpl resolves to an embedded template file.
func Exists(tmpl Template) bool {
	if tmpl == "" || strings.Contains(string(tmpl), "..") {
		return false
	}
	_, err := fs.Stat(templateFS, fsTemplateRoot+"/"+string(tmpl)+".html")
	return err == nil
}

// Available returns every embedded template (domain/name), sorted. Derived from
// the embedded FS, so it always reflects what's actually shippable.
func Available() []Template {
	var out []Template
	_ = fs.WalkDir(templateFS, fsTemplateRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}
		name := strings.TrimSuffix(strings.TrimPrefix(path, fsTemplateRoot+"/"), ".html")
		out = append(out, Template(name))
		return nil
	})
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
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
