package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/html"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

const (
	linkPreviewTimeout  = 5 * time.Second
	linkPreviewMaxBytes = 256 << 10
	linkTitleMaxRunes   = 300
)

var errLinkPreviewBlocked = errors.New("link preview: address is not public")

// LinkPreviewer reads a pasted link's title once, when it is added. It is
// reachable from user input, so every connection is checked against the
// address actually dialled: only public IPs on ports 80 and 443, no proxy,
// at most three redirects, and a capped body.
type LinkPreviewer struct {
	client    *http.Client
	githubAPI string
}

func NewLinkPreviewer() *LinkPreviewer {
	dialer := &net.Dialer{Timeout: linkPreviewTimeout, Control: linkPreviewDialControl}
	return &LinkPreviewer{
		githubAPI: "https://api.github.com",
		client: &http.Client{
			Timeout: linkPreviewTimeout,
			Transport: &http.Transport{
				Proxy:                 nil,
				DialContext:           dialer.DialContext,
				TLSHandshakeTimeout:   linkPreviewTimeout,
				ResponseHeaderTimeout: linkPreviewTimeout,
				MaxIdleConns:          8,
				IdleConnTimeout:       30 * time.Second,
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 3 {
					return errors.New("too many redirects")
				}
				if req.URL.Scheme != "https" && req.URL.Scheme != "http" {
					return errors.New("redirect to an unsupported scheme")
				}
				return nil
			},
		},
	}
}

// Preview returns the link's title and, for GitHub PRs and issues, its state.
// Any failure returns empty values: a link without a title is still useful.
func (p *LinkPreviewer) Preview(ctx context.Context, link domain.ClassifiedLink) (title, state string) {
	if link.Kind == domain.LinkKindGitHubPR || link.Kind == domain.LinkKindGitHubIssue {
		if title, state, ok := p.githubPreview(ctx, link); ok {
			return title, state
		}
		// A private repository answers 404 and its page is a login wall, so
		// the HTML title would only say "GitHub". Leave it empty instead.
		return "", ""
	}
	return p.htmlTitle(ctx, link.URL), ""
}

func (p *LinkPreviewer) githubPreview(ctx context.Context, link domain.ClassifiedLink) (string, string, bool) {
	resource := "issues"
	if link.Kind == domain.LinkKindGitHubPR {
		resource = "pulls"
	}
	endpoint := p.githubAPI + "/repos/" + link.Owner + "/" + link.Repo + "/" + resource + "/" + link.Number
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", "", false
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "rtylr-flow-links")
	resp, err := p.client.Do(req)
	if err != nil {
		return "", "", false
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", "", false
	}
	var body struct {
		Title    string  `json:"title"`
		State    string  `json:"state"`
		Draft    bool    `json:"draft"`
		MergedAt *string `json:"merged_at"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, linkPreviewMaxBytes)).Decode(&body); err != nil {
		return "", "", false
	}
	state := body.State
	switch {
	case body.MergedAt != nil && *body.MergedAt != "":
		state = "merged"
	case body.Draft && state == "open":
		state = "draft"
	}
	return cleanLinkTitle(body.Title), state, true
}

func (p *LinkPreviewer) htmlTitle(ctx context.Context, rawURL string) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; rtylr-flow-links)")
	resp, err := p.client.Do(req)
	if err != nil {
		return ""
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK || !strings.Contains(resp.Header.Get("Content-Type"), "html") {
		return ""
	}
	return extractHTMLTitle(io.LimitReader(resp.Body, linkPreviewMaxBytes))
}

// extractHTMLTitle prefers og:title and falls back to <title>.
func extractHTMLTitle(r io.Reader) string {
	z := html.NewTokenizer(r)
	var title string
	inTitle := false
	for {
		switch z.Next() {
		case html.ErrorToken:
			return cleanLinkTitle(title)
		case html.StartTagToken, html.SelfClosingTagToken:
			name, hasAttr := z.TagName()
			switch string(name) {
			case "title":
				inTitle = title == ""
			case "meta":
				var property, content string
				for hasAttr {
					var k, v []byte
					k, v, hasAttr = z.TagAttr()
					switch strings.ToLower(string(k)) {
					case "property", "name":
						property = strings.ToLower(string(v))
					case "content":
						content = string(v)
					}
				}
				if property == "og:title" && strings.TrimSpace(content) != "" {
					return cleanLinkTitle(content)
				}
			case "body":
				return cleanLinkTitle(title)
			}
		case html.TextToken:
			if inTitle {
				title += string(z.Text())
			}
		case html.EndTagToken:
			if name, _ := z.TagName(); string(name) == "title" {
				inTitle = false
			}
		}
	}
}

func cleanLinkTitle(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > linkTitleMaxRunes {
		s = string(r[:linkTitleMaxRunes])
	}
	return s
}

func linkPreviewDialControl(_, address string, _ syscall.RawConn) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	if port != "80" && port != "443" {
		return errLinkPreviewBlocked
	}
	ip, err := netip.ParseAddr(host)
	if err != nil || !isPublicLinkAddr(ip) {
		return errLinkPreviewBlocked
	}
	return nil
}

func isPublicLinkAddr(ip netip.Addr) bool {
	ip = ip.Unmap()
	if !ip.IsValid() || ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast() || ip.IsMulticast() {
		return false
	}
	for _, prefix := range nonPublicLinkPrefixes {
		if prefix.Contains(ip) {
			return false
		}
	}
	return true
}

var nonPublicLinkPrefixes = []netip.Prefix{
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("fc00::/7"),
}
