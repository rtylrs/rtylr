// Package emailtest exposes an unauthenticated endpoint that enqueues any
// registered email template to an arbitrary address, to verify the email queue
// end-to-end (POST -> email_jobs -> worker -> Resend) by curl.
//
// SECURITY: this is effectively an open email relay (e.g. it can send a genuine
// rtylr password-reset email with a caller-supplied link — a phishing vector), so
// it is OFF by default and only registers when EMAIL_TEST_ENDPOINT_ENABLED=true.
// Enable it transiently to test, then unset the flag (or delete this module). It
// only accepts templates in email.AllTemplates and is covered by the global rate
// limiter.
package emailtest

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/errors"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

// enabled reports whether the test endpoint is switched on. It is off unless
// EMAIL_TEST_ENDPOINT_ENABLED=true, so the unauthenticated relay never exists in
// a normal (incl. production) deployment.
func enabled() bool {
	return config.GlobalConfig != nil && config.GlobalConfig.EmailTestEndpointEnabled
}

type Module struct {
	db *sql.DB
}

func NewModule(_ context.Context, db *sql.DB) *Module {
	return &Module{db: db}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	if !enabled() {
		return
	}
	serverHttp.RegisterPOST[TestEmailRequest, TestEmailResponse](
		ctx, router, TestEmailRoutes(), &TestEmailHandler{DB: m.db},
	)
}

func (m *Module) GetAnonymousRoutes() []string {
	if !enabled() {
		return []string{}
	}
	return TestEmailRoutes()
}

func TestEmailRoutes() []string { return []string{"/test/email"} }

// TestEmailRequest is the JSON body. `data` is optional — missing template fields
// just render empty, so a bare {"template","to"} is enough to test delivery.
type TestEmailRequest struct {
	Template string         `json:"template"`
	To       string         `json:"to"`
	Subject  string         `json:"subject"`
	Data     map[string]any `json:"data"`
}

type TestEmailResponse struct {
	Queued   bool   `json:"queued"`
	Template string `json:"template"`
	To       string `json:"to"`
}

type TestEmailHandler struct{ DB *sql.DB }

func (h *TestEmailHandler) Serve(ctx context.Context, req *TestEmailRequest) (TestEmailResponse, *errors.ServerError) {
	if req.To == "" || req.Template == "" {
		return TestEmailResponse{}, badRequest("both 'to' and 'template' are required. known templates: " + strings.Join(templateNames(), ", "))
	}

	tmpl := email.Template(req.Template)
	if !isKnownTemplate(tmpl) {
		return TestEmailResponse{}, badRequest("unknown template " + req.Template + ". known templates: " + strings.Join(templateNames(), ", "))
	}

	if h.DB == nil {
		return TestEmailResponse{}, &errors.ServerError{
			HTTPCode: http.StatusServiceUnavailable,
			Code:     http.StatusServiceUnavailable,
			Message:  "email queue unavailable (no database)",
		}
	}

	subject := req.Subject
	if subject == "" {
		subject = "rtylr test email"
	}
	data := req.Data
	if data == nil {
		data = map[string]any{}
	}
	if _, ok := data["Subject"]; !ok {
		data["Subject"] = subject
	}

	if err := email.EnqueueTemplate(ctx, h.DB, email.PriorityNormal, req.To, subject, tmpl, data); err != nil {
		return TestEmailResponse{}, &errors.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "failed to enqueue: " + err.Error(),
		}
	}

	return TestEmailResponse{Queued: true, Template: req.Template, To: req.To}, nil
}

func badRequest(msg string) *errors.ServerError {
	return &errors.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: msg}
}

func isKnownTemplate(t email.Template) bool {
	for _, k := range email.AllTemplates {
		if k == t {
			return true
		}
	}
	return false
}

func templateNames() []string {
	out := make([]string, 0, len(email.AllTemplates))
	for _, t := range email.AllTemplates {
		out = append(out, string(t))
	}
	return out
}
