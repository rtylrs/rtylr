package cdn

import (
	"net/url"
	"strings"
	"testing"
)

func TestParseImageParams_FastlyStyleCropAutoFormatAndDPR(t *testing.T) {
	query := url.Values{}
	query.Set("width", "200")
	query.Set("dpr", "2")
	query.Set("crop", "1:1,smart")
	query.Set("format", "auto")
	query.Set("quality", "70")

	got, err := parseImageParams(query, "image/avif,image/webp", imageParamOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got.params.Format != "avif" {
		t.Fatalf("Format = %q", got.params.Format)
	}
	if got.params.Width != 400 || got.params.Height != 400 {
		t.Fatalf("dimensions = %dx%d", got.params.Width, got.params.Height)
	}
	if got.params.Quality != 70 {
		t.Fatalf("Quality = %d", got.params.Quality)
	}
	if !got.varyAccept {
		t.Fatal("expected Vary: Accept for format=auto")
	}
	for _, want := range []string{"w_400,h_400,c_fill,g_auto", "f_avif"} {
		if !strings.Contains(got.params.Chain, want) {
			t.Fatalf("chain %q missing %q", got.params.Chain, want)
		}
	}
}

func TestParseImageParams_FastlyStyleEffects(t *testing.T) {
	query := url.Values{}
	query.Set("width", "320")
	query.Set("height", "180")
	query.Set("fit", "crop")
	query.Set("blur", "200")
	query.Set("brightness", "10")
	query.Set("contrast", "-5")
	query.Set("saturation", "20")
	query.Set("bw", "true")
	query.Set("pad", "8")
	query.Set("bg-color", "ffffff")

	got, err := parseImageParams(query, "", imageParamOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"w_320,h_180,c_fill",
		"bo_8:ffffff",
		"e_brightness:10",
		"e_contrast:-5",
		"e_saturation:20",
		"e_grayscale",
		"e_blur:200",
		"f_webp",
	} {
		if !strings.Contains(got.params.Chain, want) {
			t.Fatalf("chain %q missing %q", got.params.Chain, want)
		}
	}
}

func TestParseImageParams_ClassAndAdvancedEffects(t *testing.T) {
	query := url.Values{}
	query.Set("class", "product")
	query.Set("duotone", "111111:eeeeee")
	query.Set("radius", "16")
	query.Set("opacity", "90")
	query.Set("text", "New Arrival")
	query.Set("tf", "sans 32")
	query.Set("tc", "111111")
	query.Set("tg", "south")

	got, err := parseImageParams(query, "image/webp", imageParamOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"w_900,h_900,c_pad,b_ffffff,f_auto",
		"e_duotone:111111:eeeeee",
		"r_16",
		"o_90",
		"txt_New_Arrival",
		"font_sans_32",
		"tc_111111",
		"tg_south",
	} {
		if !strings.Contains(got.params.Chain, want) {
			t.Fatalf("chain %q missing %q", got.params.Chain, want)
		}
	}
}

func TestParseImageParams_AspectAutoQualityAndWatermark(t *testing.T) {
	query := url.Values{}
	query.Set("width", "400")
	query.Set("aspect-ratio", "4:3")
	query.Set("quality", "auto:eco")
	query.Set("watermark", "Voxire POS")
	query.Set("wm-gravity", "south_east")
	query.Set("wm-opacity", "60")
	query.Set("wm-color", "ffffff")

	got, err := parseImageParams(query, "", imageParamOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if got.params.Width != 400 || got.params.Height != 300 {
		t.Fatalf("dimensions = %dx%d, want 400x300", got.params.Width, got.params.Height)
	}
	if got.params.Quality != 75 {
		t.Fatalf("Quality = %d, want 75", got.params.Quality)
	}
	for _, want := range []string{
		"w_400,h_300,c_fill",
		"txt_Voxire_POS",
		"tg_south_east",
		"to_60",
		"tc_ffffff",
	} {
		if !strings.Contains(got.params.Chain, want) {
			t.Fatalf("chain %q missing %q", got.params.Chain, want)
		}
	}
}

func TestParseImageParams_RejectsRemoteWatermark(t *testing.T) {
	query := url.Values{}
	query.Set("width", "400")
	query.Set("watermark", "https://example.com/logo.png")
	if _, err := parseImageParams(query, "", imageParamOptions{}); err == nil {
		t.Fatal("expected remote watermark to be rejected")
	}
}

func TestParseImageParams_ExplicitCropOffsets(t *testing.T) {
	query := url.Values{}
	query.Set("width", "320")
	query.Set("height", "200")
	query.Set("fit", "crop")
	query.Set("x", "10")
	query.Set("y", "20")

	got, err := parseImageParams(query, "", imageParamOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"w_320,h_200,c_crop", "x_10", "y_20"} {
		if !strings.Contains(got.params.Chain, want) {
			t.Fatalf("chain %q missing %q", got.params.Chain, want)
		}
	}
}

func TestParseChainParamsAutoQuality(t *testing.T) {
	format, quality := parseChainParams("w_400,c_fill/f_auto,q_auto:low")
	if format != "auto" {
		t.Fatalf("format = %q, want auto", format)
	}
	if quality != 60 {
		t.Fatalf("quality = %d, want 60", quality)
	}
}

func TestParseImageParams_UnknownClassRejected(t *testing.T) {
	query := url.Values{}
	query.Set("class", "prodcut")
	if _, err := parseImageParams(query, "", imageParamOptions{}); err == nil {
		t.Fatal("expected unknown image class to be rejected")
	}
}

func TestUploadImageReadURLPreservesTransformParams(t *testing.T) {
	c := &CDN{baseURL: "https://cdn.voxire.com"}
	query := url.Values{}
	query.Set("url", "https://example.com/source.png")
	query.Set("width", "300")
	query.Set("crop", "1:1,smart")

	got := c.uploadImageReadURL("4c85c397-9fc3-4bd8-adf5-9d1ef64f5e8b", query)
	if !strings.Contains(got, "width=300") || !strings.Contains(got, "crop=1%3A1%2Csmart") {
		t.Fatalf("transform params not preserved in %q", got)
	}
	if !strings.Contains(got, "format=webp") {
		t.Fatalf("default webp format missing in %q", got)
	}
	if strings.Contains(got, "url=") {
		t.Fatalf("source url leaked into image read URL: %q", got)
	}
}

func TestOrientStepBareLettersMatchAliases(t *testing.T) {
	tests := map[string]string{
		"h":     "fx_flop",
		"flop":  "fx_flop",
		"hflip": "fx_flop",
		"v":     "fx_flip",
		"flip":  "fx_flip",
		"vflip": "fx_flip",
	}
	for raw, want := range tests {
		if got := orientStep(raw); got != want {
			t.Errorf("orientStep(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestResponsiveWidths(t *testing.T) {
	got := responsiveWidths(200, 1000, 5)
	want := []int{200, 400, 600, 800, 1000}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("widths = %v, want %v", got, want)
		}
	}
	if got := responsiveWidths(100, 800, 1); len(got) != 1 || got[0] != 800 {
		t.Fatalf("single width = %v", got)
	}
}

func TestUploadURLTransformStorageKeyScopesBySource(t *testing.T) {
	p := Params{Chain: "f_webp", Format: "webp", Quality: 85}
	a := uploadURLTransformStorageKey("remote:https://example.com/a.png", "variant", p)
	b := uploadURLTransformStorageKey("remote:https://example.com/b.png", "variant", p)
	if a == b {
		t.Fatal("different source seeds should not share a storage prefix")
	}
	if !strings.HasPrefix(a, "upload-url-transformed/") || !strings.HasSuffix(a, ".webp") {
		t.Fatalf("unexpected storage key %q", a)
	}
}
