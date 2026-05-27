package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

//go:embed templates/*.html
var templateFS embed.FS

// Render executes the named email template (e.g. "welcome") with data and
// returns the full HTML string ready to send. The base layout is always
// applied; the named template must define a "body" block.
func Render(name string, data any) (string, error) {
	t, err := template.New("base").ParseFS(templateFS,
		"templates/base.html",
		fmt.Sprintf("templates/%s.html", name),
	)
	if err != nil {
		return "", fmt.Errorf("email: parse template %q: %w", name, err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("email: render template %q: %w", name, err)
	}

	return buf.String(), nil
}
