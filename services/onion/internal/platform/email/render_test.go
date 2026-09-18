package email

import (
	"strings"
	"testing"
)

// TestRender ensures every template parses, the base layout is applied, and the
// named body is injected with its data. It guards the embed pattern + name→path
// resolution so a templates-repo restructure can't silently break rendering.
//
// The cases are keyed by Template. TestAllTemplatesCovered asserts this map
// covers every entry in AllTemplates, so adding a const without a test case fails
// the build rather than shipping an untested (and possibly broken) template.
var renderCases = map[Template]struct {
	data any
	want []string // substrings that must appear in the rendered HTML
}{
	TemplateWelcome: {
		data: WelcomeData{Subject: "Welcome", Name: "Sam"},
		// the hub URL verifies the {{appURL "hub"}} func resolved.
		want: []string{"Hey Sam", "welcome to rtylr", "https://dash.rtylr.com"},
	},
	TemplateInvitation: {
		data: InvitationData{Subject: "Invite", InviterName: "Alex", BusinessName: "Acme", InviteLink: "https://x.test/i"},
		want: []string{"invited to rtylr", "Alex", "Acme", "https://x.test/i"},
	},
	TemplatePasswordReset: {
		data: PasswordResetData{Subject: "Reset", ResetLink: "https://x.test/r"},
		want: []string{"https://x.test/r"},
	},
	TemplateInvoice: {
		data: InvoiceData{Subject: "Invoice", CustomerName: "Jordan", InvoiceNumber: "INV-42", Amount: "$10", DueDate: "soon", ViewLink: "https://x.test/v"},
		want: []string{"Invoice INV-42", "Jordan", "$10", "https://x.test/v"},
	},
	TemplateAccessRequest: {
		data: AccessRequestData{Subject: "Access", RequesterName: "Pat", RequesterEmail: "pat@x.test", App: "ERP", TeamLink: "https://x.test/team"},
		want: []string{"Access request", "Pat", "pat@x.test", "ERP", "https://x.test/team"},
	},
}

func TestRender(t *testing.T) {
	for tmpl, tc := range renderCases {
		t.Run(string(tmpl), func(t *testing.T) {
			out, err := Render(tmpl, tc.data)
			if err != nil {
				t.Fatalf("Render(%q) error: %v", tmpl, err)
			}
			// Base layout must always wrap the body.
			if !strings.Contains(out, `class="card"`) || !strings.Contains(out, "rtylr.com") {
				t.Errorf("Render(%q): base layout not applied; got:\n%s", tmpl, out)
			}
			for _, w := range tc.want {
				if !strings.Contains(out, w) {
					t.Errorf("Render(%q): missing %q in output", tmpl, w)
				}
			}
		})
	}
}

func TestHTMLToTextRemovesMarkupAndUnescapesEntities(t *testing.T) {
	got := htmlToText(`<p>Hello <strong>Sam</strong> &amp; welcome.</p>`)
	if got != "Hello  Sam  & welcome." {
		t.Fatalf("htmlToText() = %q", got)
	}
}

// TestWiredTemplatesCovered fails if a Go-wired Template has no render case,
// keeping the typed templates and their tests in lockstep.
func TestWiredTemplatesCovered(t *testing.T) {
	for _, tmpl := range wiredTemplates {
		if _, ok := renderCases[tmpl]; !ok {
			t.Errorf("template %q is wired but has no case in renderCases", tmpl)
		}
	}
}

// TestWiredTemplatesExist ensures every wired template actually resolves in the
// embedded FS (catches a renamed/missing file behind a const).
func TestWiredTemplatesExist(t *testing.T) {
	for _, tmpl := range wiredTemplates {
		if !Exists(tmpl) {
			t.Errorf("wired template %q has no file in the embedded FS", tmpl)
		}
	}
}

// TestAvailableFromFS confirms the embedded set is discovered (catches a broken
// embed glob).
func TestAvailableFromFS(t *testing.T) {
	all := Available()
	if len(all) < len(wiredTemplates) {
		t.Fatalf("expected at least %d embedded templates, got only %d", len(wiredTemplates), len(all))
	}
	if !Exists(TemplateWelcome) {
		t.Error("expected a wired template to be discoverable")
	}
	if Exists("does/not_exist") || Exists("../etc/passwd") {
		t.Error("Exists must reject missing/traversal paths")
	}
}
