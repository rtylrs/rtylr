package cdn

import "testing"

func TestParamsDefaults(t *testing.T) {
	var p Params
	p.Defaults()
	if p.Quality != 85 {
		t.Errorf("default quality = %d, want 85", p.Quality)
	}
	if p.Format != "webp" {
		t.Errorf("default format = %q, want webp", p.Format)
	}
	if p.Fit != "cover" {
		t.Errorf("default fit = %q, want cover", p.Fit)
	}
}

func TestParamsDefaults_PreservesNonZero(t *testing.T) {
	p := Params{Quality: 70, Format: "avif", Fit: "contain"}
	p.Defaults()
	if p.Quality != 70 || p.Format != "avif" || p.Fit != "contain" {
		t.Errorf("Defaults overwrote explicit values: %+v", p)
	}
}

func TestParamsValidate(t *testing.T) {
	cases := []struct {
		name    string
		p       Params
		wantErr bool
	}{
		{"valid", Params{Width: 256, Height: 256, Quality: 85, Format: "webp", Fit: "cover"}, false},
		{"width only", Params{Width: 256, Quality: 85, Format: "webp", Fit: "cover"}, false},
		{"both zero", Params{Quality: 85, Format: "webp", Fit: "cover"}, true},
		{"negative width", Params{Width: -1, Height: 10, Quality: 85, Format: "webp", Fit: "cover"}, true},
		{"width too large", Params{Width: 8001, Quality: 85, Format: "webp", Fit: "cover"}, true},
		{"height too large", Params{Height: 9000, Quality: 85, Format: "webp", Fit: "cover"}, true},
		{"quality zero", Params{Width: 10, Quality: 0, Format: "webp", Fit: "cover"}, true},
		{"quality too high", Params{Width: 10, Quality: 101, Format: "webp", Fit: "cover"}, true},
		{"bad format", Params{Width: 10, Quality: 85, Format: "bmp", Fit: "cover"}, true},
		{"bad fit", Params{Width: 10, Quality: 85, Format: "webp", Fit: "squish"}, true},
		{"uppercase format ok", Params{Width: 10, Quality: 85, Format: "WEBP", Fit: "COVER"}, false},
		{"area too large", Params{Width: 8000, Height: 8000, Quality: 85, Format: "webp", Fit: "cover"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.p.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() err=%v, wantErr=%v", err, tc.wantErr)
			}
		})
	}
}
