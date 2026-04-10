package domain

import "time"

type Feedback struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"ownerId"`
	CustomerID int64     `json:"customerId"`
	Type       string    `json:"type"`
	Score      int       `json:"score"`
	Comment    string    `json:"comment"`
	Source     string    `json:"source"`
	CreatedAt  time.Time `json:"createdAt"`
}

type FeedbackInput struct {
	OwnerID    int64
	CustomerID int64
	Type       string
	Score      int
	Comment    string
	Source     string
}
