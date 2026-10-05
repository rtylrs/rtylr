package api

import (
	"testing"

	hrcore "github.com/voxire/onion/internal/modules/hr/hr_core"
)

func TestToConfigPayloadDefaultsTimeTrackingModel(t *testing.T) {
	payload := toConfigPayload(nil, &hrcore.Config{})
	if payload.TimeTrackingModel != "both" {
		t.Fatalf("time_tracking_model = %q, want both", payload.TimeTrackingModel)
	}
}
