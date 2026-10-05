package cdn

import (
	"strings"
	"testing"
)

// TestParamsPipeline exercises the full sequence a transform request goes
// through: raw query params -> Defaults -> Validate -> CacheKey/CLIArgs, the
// same path serveTransformed and transform follow.
func TestParamsPipeline(t *testing.T) {
	// Simulate ?w=300 with everything else defaulted (and mixed case).
	p := Params{Width: 300, Fit: "Cover", Format: "WebP"}
	p.Defaults()

	if err := p.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}

	// CacheKey is stable across repeated calls for the same inputs.
	first := p.CacheKey("uuid")
	second := p.CacheKey("uuid")
	if first != second {
		t.Errorf("cache key is not stable: %q != %q", first, second)
	}

	// CLIArgs hands the engine lowercased format/fit and omits the unset height.
	args := strings.Join(p.CLIArgs("/in", "/out"), " ")
	if !strings.Contains(args, "--format webp") {
		t.Errorf("expected lowercased format in args: %s", args)
	}
	if !strings.Contains(args, "--fit cover") {
		t.Errorf("expected lowercased fit in args: %s", args)
	}
	if strings.Contains(args, "--height") {
		t.Errorf("expected no --height when height is unset: %s", args)
	}
	if !strings.Contains(args, "--quality 85") {
		t.Errorf("expected the default quality in args: %s", args)
	}
}
