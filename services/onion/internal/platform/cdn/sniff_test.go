package cdn

import (
	"io"
	"strings"
	"testing"
)

func TestSniffMIME_DetectsAndReplays(t *testing.T) {
	// Minimal PNG signature.
	png := "\x89PNG\r\n\x1a\n" + strings.Repeat("\x00", 64)
	mime, r, err := sniffMIME(strings.NewReader(png))
	if err != nil {
		t.Fatalf("sniffMIME: %v", err)
	}
	if mime != "image/png" {
		t.Errorf("detected %q, want image/png", mime)
	}
	// The returned reader must replay the entire original stream.
	got, _ := io.ReadAll(r)
	if string(got) != png {
		t.Errorf("replayed stream differs from original (len %d vs %d)", len(got), len(png))
	}
}

func TestSniffMIME_StripsCharset(t *testing.T) {
	mime, _, err := sniffMIME(strings.NewReader("<!DOCTYPE html><html></html>"))
	if err != nil {
		t.Fatalf("sniffMIME: %v", err)
	}
	if strings.Contains(mime, ";") {
		t.Errorf("detected type should not include a charset param: %q", mime)
	}
	if mime != "text/html" {
		t.Errorf("detected %q, want text/html", mime)
	}
}

func TestReconcileMIME(t *testing.T) {
	cases := []struct {
		name     string
		declared string
		sniffed  string
		want     string
		wantErr  bool
	}{
		{"sniffed allowed wins", "image/png", "image/jpeg", "image/jpeg", false},
		{"html disguised as png rejected", "image/png", "text/html", "", true},
		{"octet-stream falls back to declared avif", "image/avif", "application/octet-stream", "image/avif", false},
		{"octet-stream falls back to declared pdf", "application/pdf", "application/octet-stream", "application/pdf", false},
		{"case-insensitive", "IMAGE/PNG", "IMAGE/PNG", "image/png", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := reconcileMIME(tc.declared, tc.sniffed)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
