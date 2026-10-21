package domain

import (
	"errors"
	"time"
)

var ErrReviewNotFound = errors.New("review not found")

type Review struct {
	ID           int64
	OwnerID      int64
	BusinessID   int64
	StaffID      int64
	ReviewerID   int64
	Period       string
	Rating       float64
	GoalsRating  float64
	ValuesRating float64
	SkillsRating float64
	Notes        string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
type ReviewInput struct {
	StaffID      int64
	ReviewerID   int64
	Period       string
	Rating       float64
	GoalsRating  float64
	ValuesRating float64
	SkillsRating float64
	Notes        string
	Status       string
}
