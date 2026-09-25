package http

import (
	"testing"

	"github.com/voxire/onion/internal/platform/status"
)

func TestRedactStatusDetails(t *testing.T) {
	systemStatus := status.SystemStatus{
		Components: map[string]status.ComponentStatus{
			"database": {
				Message: "database ping failed: secret-host:3306",
				Details: map[string]interface{}{"host": "secret-host", "password": "never"},
			},
		},
	}

	redactStatusDetails(&systemStatus)
	component := systemStatus.Components["database"]
	if component.Message != "" {
		t.Fatalf("Message = %q, want empty", component.Message)
	}
	if component.Details != nil {
		t.Fatalf("Details = %#v, want nil", component.Details)
	}
}
