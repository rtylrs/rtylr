package domain

import (
	"errors"
	"net/url"
	"strings"
	"time"
)

// Link kinds recognised from a URL alone, with or without an integration.
const (
	LinkKindGitHubPR     = "github_pr"
	LinkKindGitHubIssue  = "github_issue"
	LinkKindGitLabMR     = "gitlab_mr"
	LinkKindGitLabIssue  = "gitlab_issue"
	LinkKindFigma        = "figma"
	LinkKindVercel       = "vercel"
	LinkKindNetlify      = "netlify"
	LinkKindSentry       = "sentry"
	LinkKindGoogleDocs   = "google_docs"
	LinkKindGeneric      = "generic"
	maxTaskLinkURLLength = 2048
)

// ErrInvalidTaskLink rejects a URL that is not an absolute http(s) address.
var ErrInvalidTaskLink = errors.New("link must be a full http or https URL")

type TaskLink struct {
	ID         int64
	TaskID     int64
	BusinessID int64
	URL        string
	Kind       string
	Title      string
	Ref        string
	State      string
	AddedBy    int64
	CreatedAt  time.Time
}

// ClassifiedLink is what can be known about a link from its URL alone.
type ClassifiedLink struct {
	URL  string
	Kind string
	Ref  string
	// For GitHub PRs and issues, the parts needed to ask GitHub for the title.
	Owner, Repo, Number string
}

// ClassifyLink validates a pasted URL and recognises its type. Only http and
// https are accepted, so a stored link can never become a javascript: href.
func ClassifyLink(raw string) (ClassifiedLink, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > maxTaskLinkURLLength {
		return ClassifiedLink{}, ErrInvalidTaskLink
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil {
		return ClassifiedLink{}, ErrInvalidTaskLink
	}
	host := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	segs := pathSegments(u.Path)
	out := ClassifiedLink{URL: u.String(), Kind: LinkKindGeneric, Ref: host}

	switch {
	case host == "github.com" && len(segs) >= 4 && (segs[2] == "pull" || segs[2] == "issues") && isNumber(segs[3]):
		out.Kind = LinkKindGitHubIssue
		if segs[2] == "pull" {
			out.Kind = LinkKindGitHubPR
		}
		out.Owner, out.Repo, out.Number = segs[0], segs[1], segs[3]
		out.Ref = segs[0] + "/" + segs[1] + "#" + segs[3]
	case isGitLabHost(host):
		if i := indexOf(segs, "-"); i >= 2 && len(segs) > i+2 && isNumber(segs[i+2]) {
			project := strings.Join(segs[:i], "/")
			switch segs[i+1] {
			case "merge_requests":
				out.Kind, out.Ref = LinkKindGitLabMR, project+"!"+segs[i+2]
			case "issues":
				out.Kind, out.Ref = LinkKindGitLabIssue, project+"#"+segs[i+2]
			}
		}
	case host == "figma.com" || strings.HasSuffix(host, ".figma.com"):
		out.Kind = LinkKindFigma
	case host == "vercel.com" || strings.HasSuffix(host, ".vercel.app"):
		out.Kind = LinkKindVercel
	case host == "netlify.com" || strings.HasSuffix(host, ".netlify.app"):
		out.Kind = LinkKindNetlify
	case host == "sentry.io" || strings.HasSuffix(host, ".sentry.io"):
		out.Kind = LinkKindSentry
	case host == "docs.google.com":
		out.Kind = LinkKindGoogleDocs
	}
	return out, nil
}

func isGitLabHost(host string) bool {
	return host == "gitlab.com" || strings.HasPrefix(host, "gitlab.")
}

func pathSegments(p string) []string {
	var out []string
	for _, s := range strings.Split(p, "/") {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func indexOf(segs []string, want string) int {
	for i, s := range segs {
		if s == want {
			return i
		}
	}
	return -1
}

func isNumber(s string) bool {
	if s == "" || len(s) > 12 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
