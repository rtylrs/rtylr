package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"image/color"
	"io"
	"net/http"
	"strings"

	"github.com/fogleman/gg"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"

	"github.com/voxire/onion/internal/modules/recruit/careers/internal/application"
	"github.com/voxire/onion/internal/modules/recruit/careers/internal/domain"
)

// The OG card endpoint renders a 1200×630 PNG on the fly from a
// business's branding. It's the image that shows up when a careers
// URL gets pasted into LinkedIn / Slack / iMessage. Generating it
// server-side is cheaper than asking every customer to design one,
// and keeps the preview locked to whatever their current branding is.
//
// The serve flow:
//   1. Resolve the business by slug.
//   2. Compute an ETag from the brand fields that feed the image.
//   3. If the client's If-None-Match matches, return 304.
//   4. Otherwise, render and cache at the edge for an hour.
//
// We don't write to disk or R2 — the render is a few milliseconds
// and Cloudflare's CDN in front of api.voxire.dev handles repeat
// traffic via the Cache-Control header.

// ─────────────────────────────────────────────
// Route declarations
// ─────────────────────────────────────────────

func PublicOGRoutes() []string { return []string{"/recruit/public/og"} }

// ─────────────────────────────────────────────
// GET /recruit/public/og?slug=<slug>   (anonymous)
//
// Returns image/png directly. This handler is structurally different
// from every other handler in this module — most use the typed
// Request/Response + JSON wrapper from platform/server/http, but a
// PNG response doesn't fit that shape, so we register this one via
// raw chi below and do our own auth (none; it's anonymous).
// ─────────────────────────────────────────────

func PublicOGHandler(careersService *application.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := strings.TrimSpace(r.URL.Query().Get("slug"))
		if slug == "" {
			http.Error(w, "slug is required", http.StatusBadRequest)
			return
		}

		custom, err := careersService.GetCustomizationBySlug(r.Context(), slug)
		if err != nil {
			http.Error(w, "Careers page not found", http.StatusNotFound)
			return
		}

		etag := ogCardETag(custom)
		if match := r.Header.Get("If-None-Match"); match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		png, err := renderOGCard(r.Context(), custom)
		if err != nil {
			http.Error(w, "Failed to render preview", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=3600, s-maxage=86400")
		w.Header().Set("ETag", etag)
		if _, writeErr := io.Copy(w, bytes.NewReader(png)); writeErr != nil {
			return
		}
	}
}

// ogCardETag hashes the fields that materially affect the rendered
// image. If any of them change (brand name, tagline, colors) the
// caches flush naturally because the ETag flips.
func ogCardETag(c domain.Customization) string {
	h := sha256.New()
	h.Write([]byte(c.BrandName))
	h.Write([]byte{0})
	h.Write([]byte(c.Tagline))
	h.Write([]byte{0})
	h.Write([]byte(c.PrimaryColor))
	h.Write([]byte{0})
	h.Write([]byte(c.AccentColor))
	h.Write([]byte{0})
	h.Write([]byte(c.BgColor))
	h.Write([]byte{0})
	h.Write([]byte(c.TextColor))
	h.Write([]byte{0})
	h.Write([]byte(c.MutedTextColor))
	h.Write([]byte{0})
	h.Write([]byte(c.Template))
	return `"` + hex.EncodeToString(h.Sum(nil)) + `"`
}

// renderOGCard draws the preview in memory and returns PNG bytes.
// Layout: left-aligned brand name (bold), tagline beneath (regular),
// accent bar running down the left edge. Simple, readable at every
// social-card size, and survives aggressive cropping on mobile.
func renderOGCard(_ context.Context, c domain.Customization) ([]byte, error) {
	const (
		w, h     = 1200, 630
		sidePad  = 96
		barWidth = 18
	)

	dc := gg.NewContext(w, h)

	// Background.
	bg := parseHexOr(c.BgColor, color.RGBA{R: 0xfa, G: 0xfa, B: 0xf9, A: 0xff})
	dc.SetColor(bg)
	dc.Clear()

	// Accent vertical bar, left edge. Falls back to a near-black if
	// neither accent nor primary is set.
	accent := parseHexOr(c.AccentColor, parseHexOr(c.PrimaryColor, color.RGBA{R: 17, G: 17, B: 17, A: 0xff}))
	dc.SetColor(accent)
	dc.DrawRectangle(0, 0, barWidth, h)
	dc.Fill()

	// Typography — Go's built-in bold for the brand name + tagline
	// label, regular for the body. This avoids shipping a TTF inside
	// the binary and stays legally clean (Go fonts are Apache-licensed).
	boldFont, err := opentype.Parse(gobold.TTF)
	if err != nil {
		return nil, err
	}
	regFont, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil, err
	}

	brandName := c.BrandName
	if brandName == "" {
		brandName = "Open roles"
	}
	tagline := c.Tagline
	if tagline == "" {
		tagline = "Careers at " + brandName
	}

	fg := parseHexOr(c.TextColor, color.RGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff})
	muted := parseHexOr(c.MutedTextColor, color.RGBA{R: 0x6b, G: 0x72, B: 0x80, A: 0xff})

	// Eyebrow label above the brand name — signals what this card
	// represents before the reader parses the big type.
	eyebrowFace, err := opentype.NewFace(regFont, &opentype.FaceOptions{Size: 22, DPI: 72})
	if err != nil {
		return nil, err
	}
	dc.SetFontFace(eyebrowFace)
	dc.SetColor(muted)
	dc.DrawStringAnchored(strings.ToUpper("Careers · Powered by rtylr"), sidePad, 140, 0, 0.5)

	// Brand name — big bold headline. Scaled down if overflowing the
	// available width. 88px handles most brand names up to ~18 chars;
	// we shrink by 8px steps for longer names.
	maxWidth := float64(w - sidePad*2)
	size := 88.0
	var brandFace interface {
		Close() error
	}
	for size >= 56 {
		face, faceErr := opentype.NewFace(boldFont, &opentype.FaceOptions{Size: size, DPI: 72})
		if faceErr != nil {
			return nil, faceErr
		}
		dc.SetFontFace(face)
		tw, _ := dc.MeasureString(brandName)
		if tw <= maxWidth {
			brandFace = face
			break
		}
		_ = face.Close()
		size -= 8
	}
	if brandFace == nil {
		// Safety net — never loop below 56px without setting something.
		face, faceErr := opentype.NewFace(boldFont, &opentype.FaceOptions{Size: 56, DPI: 72})
		if faceErr != nil {
			return nil, faceErr
		}
		dc.SetFontFace(face)
		brandFace = face
	}
	defer func() { _ = brandFace.Close() }()
	dc.SetColor(fg)
	dc.DrawStringAnchored(brandName, sidePad, 260, 0, 0.5)

	// Tagline — softer, wraps at the card width. We manually wrap
	// because gg.DrawStringWrapped centers by default.
	taglineFace, err := opentype.NewFace(regFont, &opentype.FaceOptions{Size: 36, DPI: 72})
	if err != nil {
		return nil, err
	}
	defer func() { _ = taglineFace.Close() }()
	dc.SetFontFace(taglineFace)
	dc.SetColor(fg)
	for i, line := range wrapLines(dc, tagline, maxWidth) {
		if i >= 3 {
			break
		}
		dc.DrawStringAnchored(line, sidePad, 360+float64(i)*52, 0, 0.5)
	}

	// Footer — the slug lives in the URL but we surface it so the
	// reader can find the careers page later even if the link scroll
	// off.
	footerFace, err := opentype.NewFace(regFont, &opentype.FaceOptions{Size: 22, DPI: 72})
	if err != nil {
		return nil, err
	}
	defer func() { _ = footerFace.Close() }()
	dc.SetFontFace(footerFace)
	dc.SetColor(muted)
	dc.DrawStringAnchored("recruit.rtylr.com/"+c.Slug, sidePad, float64(h)-80, 0, 0.5)

	var buf bytes.Buffer
	if err := dc.EncodePNG(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// wrapLines is a minimal word-wrapper — fogleman/gg's built-in
// WrapLines returns a single string, which we'd then have to split.
// Doing it manually keeps the control we want (truncation, max line
// count) in one place.
func wrapLines(dc *gg.Context, text string, maxWidth float64) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	current := words[0]
	for _, word := range words[1:] {
		trial := current + " " + word
		tw, _ := dc.MeasureString(trial)
		if tw <= maxWidth {
			current = trial
		} else {
			lines = append(lines, current)
			current = word
		}
	}
	lines = append(lines, current)
	return lines
}

// parseHexOr reads a #rrggbb string and returns an RGBA, falling back
// to the supplied default if the value is empty or malformed.
func parseHexOr(hexStr string, fallback color.RGBA) color.RGBA {
	hexStr = strings.TrimPrefix(strings.TrimSpace(hexStr), "#")
	if len(hexStr) != 6 {
		return fallback
	}
	bytesBuf, err := hex.DecodeString(hexStr)
	if err != nil {
		return fallback
	}
	return color.RGBA{R: bytesBuf[0], G: bytesBuf[1], B: bytesBuf[2], A: 0xff}
}
