package snips

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

func TestValidateImagesAcceptsSmallPNGAndRejectsMismatchedOrExcessiveInput(t *testing.T) {
	imageData := image.NewRGBA(image.Rect(0, 0, 8, 8))
	imageData.Set(0, 0, color.White)
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, imageData); err != nil {
		t.Fatal(err)
	}
	valid := ImageInput{Name: "receipt.png", DataURL: "data:image/png;base64," + base64.StdEncoding.EncodeToString(encoded.Bytes())}
	if err := ValidateImages([]ImageInput{valid}); err != nil {
		t.Fatalf("valid PNG rejected: %v", err)
	}
	if err := ValidateImages([]ImageInput{{Name: "fake.png", DataURL: strings.Replace(valid.DataURL, "data:image/png", "data:image/jpeg", 1)}}); err == nil {
		t.Fatal("expected declared/encoded MIME mismatch to fail")
	}
	if err := ValidateImages([]ImageInput{valid, valid, valid, valid, valid}); err == nil {
		t.Fatal("expected five images to fail the count limit")
	}
	if _, _, err := DecodeImage(ImageInput{Name: "x", DataURL: "data:image/svg+xml;base64,PHN2Zy8+"}); err == nil {
		t.Fatal("expected SVG input to be rejected")
	}
}
