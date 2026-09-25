package domain

import "time"

type Feedback struct {
	ID           int64     `json:"id"`
	OwnerID      int64     `json:"owner_id"`
	BusinessID   int64     `json:"business_id"`
	CustomerID   int64     `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	Type         string    `json:"type"`
	Score        int       `json:"score"`
	Comment      string    `json:"comment"`
	Source       string    `json:"source"`
	CreatedAt    time.Time `json:"created_at"`
}

type FeedbackInput struct {
	OwnerID    int64
	BusinessID int64
	CustomerID int64
	Type       string
	Score      int
	Comment    string
	Source     string
}

// ListFilters describes optional filters applied to a feedback list query.
// A score filter is only applied when MinScoreSet or MaxScoreSet is true.
// Source is applied when non-empty.
type ListFilters struct {
	Type        string
	Source      string
	MinScore    int
	MaxScore    int
	MinScoreSet bool
	MaxScoreSet bool
}
