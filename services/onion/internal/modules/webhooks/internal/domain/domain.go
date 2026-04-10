package domain

import "time"

type Webhook struct {
	ID              int64      `json:"id"`
	OwnerID         int64      `json:"ownerId"`
	URL             string     `json:"url"`
	Events          []string   `json:"events"`
	Secret          string     `json:"-"`
	IsActive        bool       `json:"isActive"`
	LastTriggeredAt *time.Time `json:"lastTriggeredAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

type CreateWebhookInput struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
}
