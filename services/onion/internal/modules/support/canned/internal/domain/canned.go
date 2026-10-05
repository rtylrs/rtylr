package domain

import "time"

type CannedResponse struct {
	ID         int64
	BusinessID int64
	Name       string
	Body       string
	Category   string
	Shortcut   string
	CreatedAt  time.Time
}

type CannedResponseInput struct {
	BusinessID int64
	Name       string
	Body       string
	Category   string
	Shortcut   string
}
