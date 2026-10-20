package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"github.com/voxire/onion/internal/platform/config"
)

// templateFuncs exposes the rtylr app base URLs to templates as {{appURL "hub"}},
// so links live in the templates without hardcoding "https://<app>.rtylr.com".
// Unknown names fall back to the landing site.
var templateFuncs = template.FuncMap{
	// action pairs a link with its label for the base layout's "button"
	// template: {{template "button" (action .Link "Label")}}.
	"action": func(url, label string) map[string]string {
		return map[string]string{"URL": url, "Label": label}
	},
	"appURL": func(name string) string {
		switch name {
		case "auth":
			return config.AuthAppURL
		case "hub":
			return config.HubAppURL
		case "erp":
			return config.ERPAppURL
		case "hr":
			return config.HRAppURL
		case "crm":
			return config.CRMAppURL
		case "finance":
			return config.FinanceAppURL
		case "flow":
			return config.FlowAppURL
		case "insights":
			return config.InsightsAppURL
		default:
			return config.LandingURL
		}
	},
}

// The embedded templates/ directory holds base.html at its root; body templates
// live under its own templates/<domain>/ subdir — hence templates/base.html and
// templates/templates/<domain>/*.html.
//
//go:embed templates/base.html
//go:embed templates/templates/*/*.html
var templateFS embed.FS

// Render executes the given template (e.g. TemplateWelcome) with data and returns
// the full HTML string ready to send. tmpl is the template's path under the
// embedded templates/ dir without the .html extension. The base layout is
// always applied; the named template must define a "body" block.
func Render(tmpl Template, data any) (string, error) {
	t, err := template.New("base").Funcs(templateFuncs).ParseFS(templateFS,
		"templates/base.html",
		fmt.Sprintf("templates/templates/%s.html", tmpl),
	)
	if err != nil {
		return "", fmt.Errorf("email: parse template %q: %w", tmpl, err)
	}

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "base.html", data); err != nil {
		return "", fmt.Errorf("email: render template %q: %w", tmpl, err)
	}

	return buf.String(), nil
}
