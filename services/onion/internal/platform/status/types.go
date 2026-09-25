package status

import "context"

type ComponentStatus struct {
	Name        string                 `json:"name"`
	Status      string                 `json:"status"` // "healthy", "unhealthy", "degraded", "unknown"
	Message     string                 `json:"message,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	LastChecked string                 `json:"last_checked"`
}

type SystemStatus struct {
	OverallStatus string                     `json:"overall_status"`
	Components    map[string]ComponentStatus `json:"components"`
	Timestamp     string                     `json:"timestamp"`
}

type Provider interface {
	GetStatus(ctx context.Context) ComponentStatus
	GetName() string
}
