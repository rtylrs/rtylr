package cdn

import (
	"strings"
	"testing"
)

func TestParseChainParams(t *testing.T) {
	cases := []struct {
		chain   string
		wantFmt string
		wantQ   int
	}{
		{"w_400,c_fill", "", 0},
		{"w_400,c_fill,f_avif", "avif", 0},
		{"w_400/e_blur:300,f_webp,q_70", "webp", 70},
		{"f_jpeg/f_png", "png", 0}, // last wins
		{"q_55", "", 55},
	}
	for _, tc := range cases {
		f, q := parseChainParams(tc.chain)
		if f != tc.wantFmt || q != tc.wantQ {
			t.Errorf("parseChainParams(%q) = (%q,%d), want (%q,%d)", tc.chain, f, q, tc.wantFmt, tc.wantQ)
		}
	}
}

func TestParamsValidate_ChainMode(t *testing.T) {
	// In chain mode, width/height are not required (they live in the chain).
	p := Params{Chain: "w_400,c_fill,e_blur:300", Format: "webp", Quality: 80}
	if err := p.Validate(); err != nil {
		t.Errorf("valid chain params rejected: %v", err)
	}
	// But the output format/quality still must be sane.
	bad := Params{Chain: "w_400", Format: "bmp", Quality: 80}
	if bad.Validate() == nil {
		t.Error("expected rejection of unsupported chain output format")
	}
}

func TestParamsValidate_ChainModeRejectsUnboundedGeometryAndEffects(t *testing.T) {
	for _, chain := range []string{
		"w_20000",
		"z_30",
		"e_blur:5000",
		"e_median:99",
	} {
		p := Params{Chain: chain, Format: "webp", Quality: 80}
		if err := p.Validate(); err == nil {
			t.Errorf("chain %q should be rejected", chain)
		}
	}
}

func TestParamsCLIArgs_ChainMode(t *testing.T) {
	p := Params{Chain: "w_400,c_fill,e_blur:300", Format: "WEBP", Quality: 80}
	args := strings.Join(p.CLIArgs("/in", "/out"), " ")
	for _, want := range []string{"transform", "--chain w_400,c_fill,e_blur:300", "--format webp", "--quality 80"} {
		if !strings.Contains(args, want) {
			t.Errorf("chain CLIArgs missing %q; got: %s", want, args)
		}
	}
	if strings.Contains(args, "resize") {
		t.Errorf("chain mode should not use resize: %s", args)
	}
}

func TestParamsCacheKey_VariesByChain(t *testing.T) {
	a := Params{Chain: "w_400,c_fill", Format: "webp", Quality: 80}
	b := Params{Chain: "w_400,c_fit", Format: "webp", Quality: 80}
	if a.CacheKey("uuid") == b.CacheKey("uuid") {
		t.Error("cache key should differ when the chain differs")
	}
}
