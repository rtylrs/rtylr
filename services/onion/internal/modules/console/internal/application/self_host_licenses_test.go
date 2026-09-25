package application

import (
	"testing"

	"github.com/voxire/onion/internal/modules/console/internal/infrastructure"
)

func TestToSelfHostLicenseDTO_DoesNotExposeKeyMaterial(t *testing.T) {
	dto := toSelfHostLicenseDTO(infrastructure.SelfHostLicenseRow{
		ID:         42,
		OwnerEmail: "owner@example.com",
		LicenseKey: "rtylr_license_v1_sensitive-key-material",
	})

	if dto.KeyPreview != "" {
		t.Fatalf("KeyPreview = %q, want empty because license keys are one-time secrets", dto.KeyPreview)
	}
}
