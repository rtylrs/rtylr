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
		want: []string{"Hey Sam", "welcome to rtylr", "https://portal.rtylr.com"},
	},
	TemplateInvitation: {
		data: InvitationData{Subject: "Invite", InviterName: "Alex", BusinessName: "Acme", InviteLink: "https://x.test/i", ValidDays: 7},
		want: []string{"Join Acme on rtylr", "<strong>Alex</strong> invited you", "https://x.test/i", "Accept invitation", "expires in 7 days"},
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
		want: []string{"Pat wants access to ERP", "pat@x.test", "https://x.test/team"},
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

func TestHTMLToTextDropsStyleAndScriptContents(t *testing.T) {
	got := htmlToText(`<style>body { color: #123456; }</style><script>alert("secret")</script><p>Readable message</p>`)
	if strings.Contains(got, "color") || strings.Contains(got, "alert") {
		t.Fatalf("htmlToText() leaked raw CSS/script: %q", got)
	}
	if !strings.Contains(got, "Readable message") {
		t.Fatalf("htmlToText() = %q, want readable body", got)
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

// An invitation without names still reads naturally, and never says
// "rtylr on rtylr".
func TestInvitationWithoutNames(t *testing.T) {
	out, err := Render(TemplateInvitation, InvitationData{Subject: "Invite", InviteLink: "https://x.test/i"})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"You're invited to rtylr", "You've been invited", "join a team on rtylr"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(out, "expires in") {
		t.Error("an unknown expiry should be left out")
	}
}

func TestInvitationSubject(t *testing.T) {
	cases := []struct{ inviter, business, want string }{
		{"Sam", "Acme", "Sam invited you to join Acme on rtylr"},
		{"", "Acme", "You're invited to join Acme on rtylr"},
		{"Sam", "", "Sam invited you to rtylr"},
		{"", "", "You're invited to rtylr"},
	}
	for _, c := range cases {
		if got := InvitationSubject(c.inviter, c.business); got != c.want {
			t.Errorf("InvitationSubject(%q, %q) = %q, want %q", c.inviter, c.business, got, c.want)
		}
	}
}

// The shared button is a table cell with an inline fill, so clients that strip
// <style> still show a button; its link is escaped like any other.
func TestButtonIsBulletproofAndEscaped(t *testing.T) {
	out, err := Render(TemplatePasswordReset, PasswordResetData{Subject: "Reset", ResetLink: "https://x.test/r?a=1&b=2"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `bgcolor="#131313"`) || !strings.Contains(out, `href="https://x.test/r?a=1&amp;b=2"`) {
		t.Errorf("button not rendered as expected:\n%s", out)
	}
	unsafe, err := Render(TemplatePasswordReset, PasswordResetData{Subject: "Reset", ResetLink: "javascript:alert(1)"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(unsafe, "javascript:alert") {
		t.Error("an unsafe link must be filtered")
	}
}
