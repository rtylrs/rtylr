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
	"encoding/json"
	"fmt"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	emailtestv1 "github.com/voxire/voxie/gen/go/devtools/emailtest/v1"
)

type Module struct {
	db *sql.DB
}

func NewModule(_ context.Context, db *sql.DB) *Module {
	return &Module{db: db}
}

// RegisterAPI serves devtools.emailtest.v1.
func (m *Module) RegisterAPI(mux *http.ServeMux) {
	if err := emailtestv1.RegisterEmailTestServiceServer(&EmailTestAPI{DB: m.db}, emailtestv1.WithMux(mux)); err != nil {
		panic(fmt.Errorf("register email test service: %w", err))
	}
}

func (m *Module) GetAnonymousRoutes() []string {
	return []string{}
}

// EmailTestAPI implements devtools.emailtest.v1.
type EmailTestAPI struct{ DB *sql.DB }

// ListTemplates lists every sendable template.
func (a *EmailTestAPI) ListTemplates(ctx context.Context, _ *emailtestv1.ListTemplatesRequest) (*emailtestv1.ListTemplatesResponse, error) {
	if err := requireAuthenticatedNonProduction(ctx); err != nil {
		return nil, err
	}
	names := templateNames()
	return &emailtestv1.ListTemplatesResponse{Count: int32(len(names)), Templates: names}, nil // #nosec G115 -- the template catalog is a small fixed list.
}

// SendTestEmail enqueues one rendered template to the given address.
func (a *EmailTestAPI) SendTestEmail(ctx context.Context, req *emailtestv1.SendTestEmailRequest) (*emailtestv1.SendTestEmailResponse, error) {
	if err := requireAuthenticatedNonProduction(ctx); err != nil {
		return nil, err
	}
	if req.GetTo() == "" || req.GetTemplate() == "" {
		return nil, badRequest("'to' and 'template' are required; GET /test/email/templates lists valid templates")
	}

	tmpl := email.Template(req.GetTemplate())
	if !email.Exists(tmpl) {
		return nil, badRequest("unknown template " + req.GetTemplate() + "; GET /test/email/templates lists valid templates")
	}

	if a.DB == nil {
		return nil, onekiterrors.New(http.StatusServiceUnavailable, "email queue unavailable (no database)")
	}

	subject := req.GetSubject()
	if subject == "" {
		subject = "rtylr test email"
	}
	data := map[string]any{}
	if raw := req.GetData(); len(raw) > 0 && string(raw) != "null" {
		if err := json.Unmarshal(raw, &data); err != nil {
			return nil, badRequest("'data' must be a JSON object")
		}
	}
	if _, ok := data["Subject"]; !ok {
		data["Subject"] = subject
	}

	if err := email.EnqueueTemplate(ctx, a.DB, email.PriorityNormal, req.GetTo(), subject, tmpl, data); err != nil {
		logger.Ctx(ctx).Error("test email enqueue failed",
			logger.String("template", req.GetTemplate()), logger.Error(err))
		return nil, onekiterrors.New(http.StatusInternalServerError, "failed to enqueue test email")
	}

	return &emailtestv1.SendTestEmailResponse{Queued: true, Template: req.GetTemplate(), To: req.GetTo()}, nil
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
