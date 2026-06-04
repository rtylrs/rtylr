package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

// The email-templates submodule is mounted at templates/; inside it, base.html
// sits at the root and the body templates live under its own templates/<domain>/
// subdir — hence templates/base.html and templates/templates/<domain>/*.html.
//
//go:embed templates/base.html
//go:embed templates/templates/*/*.html
var templateFS embed.FS

// Render executes the given template (e.g. TemplateWelcome) with data and returns
// the full HTML string ready to send. tmpl is the template's path under the
// submodule's templates/ dir without the .html extension. The base layout is
// always applied; the named template must define a "body" block.
func Render(tmpl Template, data any) (string, error) {
	t, err := template.New("base").ParseFS(templateFS,
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
