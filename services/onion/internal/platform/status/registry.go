package status

import (
	"context"
	"sync"
	"time"
)

type Registry struct {
	providers map[string]Provider
	mu        sync.RWMutex
}

var (
	registry     *Registry
	registryOnce sync.Once
)

func GetRegistry() *Registry {
	registryOnce.Do(func() {
		registry = &Registry{
			providers: make(map[string]Provider),
		}
	})

	return registry
}

func (r *Registry) RegisterProvider(provider Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.providers[provider.GetName()] = provider
}

func (r *Registry) GetSystemStatus(ctx context.Context) SystemStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	components := make(map[string]ComponentStatus)
	overallHealthy := true

	for name, provider := range r.providers {
		status := provider.GetStatus(ctx)
		components[name] = status

		if status.Status != "healthy" {
			overallHealthy = false
		}
	}

	overallStatus := "healthy"
	if !overallHealthy {
		overallStatus = "unhealthy"
	}

	return SystemStatus{
		OverallStatus: overallStatus,
		Components:    components,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}
}

// Convenience functions.
func RegisterProvider(provider Provider) {
	GetRegistry().RegisterProvider(provider)
}

func GetSystemStatus(ctx context.Context) SystemStatus {
	return GetRegistry().GetSystemStatus(ctx)
}
