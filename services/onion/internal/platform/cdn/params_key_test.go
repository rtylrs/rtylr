package cdn

import (
	"strings"
	"testing"
)

func TestParamsCacheKey_Deterministic(t *testing.T) {
	p := Params{Width: 256, Height: 256, Fit: "cover", Format: "webp", Quality: 85}
	first := p.CacheKey("uuid-1")
	second := p.CacheKey("uuid-1")
	if first != second {
		t.Errorf("CacheKey is not deterministic: %q != %q", first, second)
	}
}

func TestParamsCacheKey_SensitiveToEachField(t *testing.T) {
	base := Params{Width: 256, Height: 256, Fit: "cover", Format: "webp", Quality: 85}
	baseKey := base.CacheKey("uuid")

	variants := []Params{
		{Width: 128, Height: 256, Fit: "cover", Format: "webp", Quality: 85},
		{Width: 256, Height: 128, Fit: "cover", Format: "webp", Quality: 85},
		{Width: 256, Height: 256, Fit: "contain", Format: "webp", Quality: 85},
		{Width: 256, Height: 256, Fit: "cover", Format: "avif", Quality: 85},
		{Width: 256, Height: 256, Fit: "cover", Format: "webp", Quality: 70},
	}
	for i, v := range variants {
		if v.CacheKey("uuid") == baseKey {
			t.Errorf("variant %d produced the same cache key as base", i)
		}
	}
	if base.CacheKey("uuid-a") == base.CacheKey("uuid-b") {
		t.Error("CacheKey should differ across file UUIDs")
	}
}

func TestParamsStorageKey(t *testing.T) {
	p := Params{Format: "webp"}
	got := p.StorageKey("abc", "deadbeef")
	want := "transformed/abc/deadbeef.webp"
	if got != want {
		t.Errorf("StorageKey = %q, want %q", got, want)
	}
}

func TestParamsCLIArgs(t *testing.T) {
	// Both format and fit must reach the engine lowercased — the C++
	// parseFitMode only matches lowercase, so "COVER" would otherwise crash it.
	p := Params{Width: 256, Height: 128, Fit: "COVER", Format: "WEBP", Quality: 80}
	args := strings.Join(p.CLIArgs("/in.jpg", "/out.webp"), " ")
	for _, want := range []string{
		"resize", "--input /in.jpg", "--output /out.webp",
		"--format webp", "--quality 80", "--fit cover",
		"--width 256", "--height 128",
	} {
		if !strings.Contains(args, want) {
			t.Errorf("CLIArgs missing %q; got: %s", want, args)
		}
	}
}

func TestParamsCLIArgs_OmitsZeroDimensions(t *testing.T) {
	p := Params{Width: 256, Fit: "cover", Format: "webp", Quality: 80}
	args := strings.Join(p.CLIArgs("/in", "/out"), " ")
	if strings.Contains(args, "--height") {
		t.Errorf("CLIArgs should omit --height when height is 0; got: %s", args)
	}
	if !strings.Contains(args, "--width 256") {
		t.Errorf("CLIArgs should include --width 256; got: %s", args)
	}
}
