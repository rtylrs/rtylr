package domain

import (
	"encoding/json"
	"time"
)

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
