package providers

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/status"
)

type DatabaseStatusProvider struct{}

func NewDatabaseStatusProvider() *DatabaseStatusProvider {
	return &DatabaseStatusProvider{}
}

func (d *DatabaseStatusProvider) GetName() string {
	return "database"
}

func (d *DatabaseStatusProvider) GetStatus(ctx context.Context) status.ComponentStatus {
	db := dbclient.Pool()
	if db == nil {
		return status.ComponentStatus{
			Name:        "Database",
			Status:      "unhealthy",
			Message:     "connection pool not initialised",
			LastChecked: time.Now().UTC().Format(time.RFC3339),
		}
	}

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	pingErr := db.PingContext(pingCtx)
	stats := db.Stats()

	statusStr := healthyStatus
	message := "Database is reachable and pool is healthy"

	if pingErr != nil {
		statusStr = "unhealthy"
		message = "Database ping failed: " + pingErr.Error()
	} else if stats.WaitCount > 0 && stats.OpenConnections >= stats.MaxOpenConnections {
		statusStr = "degraded"
		message = "Connection pool is saturated — queries are waiting for a connection"
	}

	return status.ComponentStatus{
		Name:    "Database",
		Status:  statusStr,
		Message: message,
		Details: map[string]interface{}{
			"open_connections":     stats.OpenConnections,
			"in_use":               stats.InUse,
			"idle":                 stats.Idle,
			"max_open_connections": stats.MaxOpenConnections,
			"wait_count":           stats.WaitCount,
			"wait_duration_ms":     stats.WaitDuration.Milliseconds(),
			"max_idle_closed":      stats.MaxIdleClosed,
			"max_lifetime_closed":  stats.MaxLifetimeClosed,
		},
		LastChecked: time.Now().UTC().Format(time.RFC3339),
	}
}
