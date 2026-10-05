package domain

import "time"

type Feedback struct {
	ID           int64
	OwnerID      int64
	BusinessID   int64
	CustomerID   int64
	CustomerName string
	Type         string
	Score        int
	Comment      string
	Source       string
	CreatedAt    time.Time
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
