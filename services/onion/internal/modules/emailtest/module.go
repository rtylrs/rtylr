// Package emailtest exposes authenticated non-production endpoints for manually
// testing email template enqueueing end-to-end.
//
// The endpoints only exist for a run that has explicitly declared a
// development environment; anything else, including an unset or unrecognised
// APP_ENV, resolves to production and the routes 404.
package emailtest

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	serverHttp "github.com/voxire/onion/internal/platform/server/http"
)

type Module struct {
	db *sql.DB
}

func NewModule(_ context.Context, db *sql.DB) *Module {
	return &Module{db: db}
}

func (m *Module) RegisterRoutes(ctx context.Context, router *chi.Mux, _ ...interface{}) {
	serverHttp.RegisterPOSTWithValidation[TestEmailRequest, TestEmailResponse](
		ctx, router, sendRoutes(), &TestEmailHandler{DB: m.db},
	)
	serverHttp.RegisterGET[ListTemplatesRequest, ListTemplatesResponse](
		ctx, router, listRoutes(), &ListTemplatesHandler{},
	)
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

func sendRoutes() []string { return []string{"/test/email"} }
func listRoutes() []string { return []string{"/test/email/templates"} }

// ── GET /test/email/templates ── lists every sendable template ──────────────

type ListTemplatesRequest struct{}

type ListTemplatesResponse struct {
	Count     int      `json:"count"`
	Templates []string `json:"templates"`
}

type ListTemplatesHandler struct{}

func (h *ListTemplatesHandler) Serve(ctx context.Context, _ *ListTemplatesRequest) (ListTemplatesResponse, error) {
	if err := requireAuthenticatedNonProduction(ctx); err != nil {
		return ListTemplatesResponse{}, err
	}
	names := templateNames()
	return ListTemplatesResponse{Count: len(names), Templates: names}, nil
}

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

func (h *TestEmailHandler) Serve(ctx context.Context, req *TestEmailRequest) (TestEmailResponse, error) {
	if err := requireAuthenticatedNonProduction(ctx); err != nil {
		return TestEmailResponse{}, err
	}
	if req.To == "" || req.Template == "" {
		return TestEmailResponse{}, badRequest("'to' and 'template' are required; GET /test/email/templates lists valid templates")
	}

	tmpl := email.Template(req.Template)
	if !email.Exists(tmpl) {
		return TestEmailResponse{}, badRequest("unknown template " + req.Template + "; GET /test/email/templates lists valid templates")
	}

	if h.DB == nil {
		return TestEmailResponse{}, onekiterrors.New(http.StatusServiceUnavailable, "email queue unavailable (no database)")
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
		logger.Ctx(ctx).Error("test email enqueue failed",
			logger.String("template", req.Template), logger.Error(err))
		return TestEmailResponse{}, onekiterrors.New(http.StatusInternalServerError, "failed to enqueue test email")
	}

	return TestEmailResponse{Queued: true, Template: req.Template, To: req.To}, nil
}

func badRequest(msg string) error {
	return onekiterrors.New(http.StatusBadRequest, msg)
}

// requireAuthenticatedNonProduction refuses the request unless the run has
// explicitly declared itself a development environment. It asks the allowlist
// question directly (config.IsNonProduction) rather than negating
// config.IsProduction so that the safety of this endpoint does not depend on
// the polarity of the production check staying correct elsewhere.
//
// This matters more than the usual dev-only endpoint: /test/email renders the
// real branded templates through the platform's sending domain and posts them
// to a caller-supplied address, so treating an unset or misspelled APP_ENV as
// development is a phishing primitive rather than a small convenience leak.
func requireAuthenticatedNonProduction(ctx context.Context) error {
	if !config.IsNonProduction() {
		return onekiterrors.New(http.StatusNotFound, "Not found")
	}
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return onekiterrors.PermissionDenied()
	}
	return nil
}

func templateNames() []string {
	all := email.Available()
	out := make([]string, 0, len(all))
	for _, t := range all {
		out = append(out, string(t))
	}
	return out
}
