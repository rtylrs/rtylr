package domain

import (
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrNotFound        = errors.New("integration not found")
	ErrInvalidProvider = errors.New("unsupported integration provider")
	ErrInvalidConfig   = errors.New("integration config must be a JSON object")
)

// Providers lists the integrations a business can connect.
var Providers = map[string]struct{}{"stripe": {}}

type IntegrationStatus string

const (
	StatusConnected    IntegrationStatus = "connected"
	StatusDisconnected IntegrationStatus = "disconnected"
	StatusError        IntegrationStatus = "error"
)

type Integration struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	Provider    string
	Status      IntegrationStatus
	Config      json.RawMessage
	ConnectedAt *time.Time
	CreatedAt   time.Time
}
