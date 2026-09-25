package providers

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/platform/status"
	"github.com/voxire/onion/internal/platform/trino"
)

type TrinoStatusProvider struct{}

func NewTrinoStatusProvider() *TrinoStatusProvider {
	return &TrinoStatusProvider{}
}

func (t *TrinoStatusProvider) GetName() string {
	return "trino"
}

func (t *TrinoStatusProvider) GetStatus(ctx context.Context) status.ComponentStatus {
	if !config.IsTrinoEnabled() {
		return status.ComponentStatus{
			Name:        "Trino",
			Status:      healthyStatus,
			Message:     "Trino analytics disabled",
			LastChecked: time.Now().UTC().Format(time.RFC3339),
			Details: map[string]interface{}{
				"enabled": false,
			},
		}
	}

	details := map[string]interface{}{
		"enabled":   true,
		"host":      config.GetTrinoHost(),
		"port":      config.GetTrinoPort(),
		"catalog":   config.GetTrinoCatalog(),
		"schema":    config.GetTrinoSchema(),
		"tls":       config.IsTrinoTLS(),
		"connected": trino.Available(),
	}

	if !config.IsTrinoConfigured() {
		return status.ComponentStatus{
			Name:        "Trino",
			Status:      healthyStatus,
			Message:     "Trino enabled but missing TRINO_HOST or TRINO_USER",
			LastChecked: time.Now().UTC().Format(time.RFC3339),
			Details:     details,
		}
	}

	if !trino.Available() {
		return status.ComponentStatus{
			Name:        "Trino",
			Status:      healthyStatus,
			Message:     "Trino configured but not connected (optional analytics)",
			LastChecked: time.Now().UTC().Format(time.RFC3339),
			Details:     details,
		}
	}

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	pingErr := trino.Ping(pingCtx)
	statusStr := healthyStatus
	message := "Trino is reachable"

	if pingErr != nil {
		statusStr = "degraded"
		message = "Trino ping failed (optional analytics): " + pingErr.Error()
		details["ping_error"] = pingErr.Error()
	}

	return status.ComponentStatus{
		Name:        "Trino",
		Status:      statusStr,
		Message:     message,
		LastChecked: time.Now().UTC().Format(time.RFC3339),
		Details:     details,
	}
}
