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
	ID          int64             `json:"id"`
	OwnerID     int64             `json:"ownerId"`
	Provider    string            `json:"provider"`
	Status      IntegrationStatus `json:"status"`
	Config      json.RawMessage   `json:"config"`
	ConnectedAt *time.Time        `json:"connectedAt"`
	CreatedAt   time.Time         `json:"createdAt"`
}
