package domain

import "time"

type CannedResponse struct {
	ID         int64     `json:"id"`
	BusinessID int64     `json:"businessId"`
	Name       string    `json:"name"`
	Body       string    `json:"body"`
	Category   string    `json:"category"`
	Shortcut   string    `json:"shortcut"`
	CreatedAt  time.Time `json:"createdAt"`
}

type CannedResponseInput struct {
	BusinessID int64
	Name       string
	Body       string
	Category   string
	Shortcut   string
}
