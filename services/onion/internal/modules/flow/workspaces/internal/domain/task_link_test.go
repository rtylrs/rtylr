package domain

import (
	"errors"
	"testing"
)

func TestClassifyLink(t *testing.T) {
	cases := []struct {
		url, kind, ref string
	}{
		{"https://github.com/rtylrs/rtylr/pull/1615", LinkKindGitHubPR, "rtylrs/rtylr#1615"},
		{"https://github.com/rtylrs/rtylr/pull/1615/files", LinkKindGitHubPR, "rtylrs/rtylr#1615"},
		{"https://github.com/rtylrs/rtylr/issues/12", LinkKindGitHubIssue, "rtylrs/rtylr#12"},
		{"https://github.com/rtylrs/rtylr", LinkKindGeneric, "github.com"},
		{"https://gitlab.com/group/sub/app/-/merge_requests/7", LinkKindGitLabMR, "group/sub/app!7"},
		{"https://gitlab.example.com/team/app/-/issues/3", LinkKindGitLabIssue, "team/app#3"},
		{"https://www.figma.com/design/abc/Board", LinkKindFigma, "figma.com"},
		{"https://app-git-main-acme.vercel.app/", LinkKindVercel, "app-git-main-acme.vercel.app"},
		{"https://deploy-preview-4--site.netlify.app", LinkKindNetlify, "deploy-preview-4--site.netlify.app"},
		{"https://acme.sentry.io/issues/123/", LinkKindSentry, "acme.sentry.io"},
		{"https://docs.google.com/document/d/x/edit", LinkKindGoogleDocs, "docs.google.com"},
		{"http://example.com/a", LinkKindGeneric, "example.com"},
	}
	for _, c := range cases {
		got, err := ClassifyLink(c.url)
		if err != nil {
			t.Fatalf("ClassifyLink(%q): %v", c.url, err)
		}
		if got.Kind != c.kind || got.Ref != c.ref {
			t.Errorf("ClassifyLink(%q) = %s %q, want %s %q", c.url, got.Kind, got.Ref, c.kind, c.ref)
		}
	}
}

func TestClassifyLinkRejectsNonWebURLs(t *testing.T) {
	for _, raw := range []string{"", "javascript:alert(1)", "ftp://example.com/x", "example.com", "/relative", "https://user:pw@example.com/"} {
		if _, err := ClassifyLink(raw); !errors.Is(err, ErrInvalidTaskLink) {
			t.Errorf("ClassifyLink(%q) err = %v, want ErrInvalidTaskLink", raw, err)
		}
	}
}
