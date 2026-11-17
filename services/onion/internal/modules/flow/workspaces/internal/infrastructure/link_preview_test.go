package infrastructure

import (
	"context"
	"net/netip"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

func TestExtractHTMLTitlePrefersOpenGraph(t *testing.T) {
	page := `<html><head><title>Fallback</title><meta property="og:title" content="  Checkout &amp; billing  "></head><body>x</body></html>`
	if got := extractHTMLTitle(strings.NewReader(page)); got != "Checkout & billing" {
		t.Errorf("title = %q", got)
	}
	if got := extractHTMLTitle(strings.NewReader("<title>\n Plain\n page </title>")); got != "Plain page" {
		t.Errorf("title = %q", got)
	}
}

func TestLinkPreviewRefusesNonPublicAddresses(t *testing.T) {
	for _, addr := range []string{"127.0.0.1:80", "10.0.0.5:443", "169.254.169.254:80", "[::1]:443", "8.8.8.8:22"} {
		if err := linkPreviewDialControl("tcp", addr, nil); err == nil {
			t.Errorf("dial to %s was allowed", addr)
		}
	}
	if err := linkPreviewDialControl("tcp", "8.8.8.8:443", nil); err != nil {
		t.Errorf("public 443 refused: %v", err)
	}
	if isPublicLinkAddr(netip.MustParseAddr("fd00::1")) {
		t.Errorf("unique-local IPv6 treated as public")
	}
}

func TestPreviewOfLocalURLFetchesNothing(t *testing.T) {
	link, err := domain.ClassifyLink("http://127.0.0.1/admin")
	if err != nil {
		t.Fatal(err)
	}
	if title, _ := NewLinkPreviewer().Preview(context.Background(), link); title != "" {
		t.Errorf("fetched a loopback page: %q", title)
	}
}
