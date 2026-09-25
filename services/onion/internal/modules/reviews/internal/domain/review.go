package domain

import (
	"errors"
	"time"
)

var ErrReviewNotFound = errors.New("review not found")

type Review struct {
	ID           int64     `json:"id"`
	OwnerID      int64     `json:"owner_id"`
	BusinessID   int64     `json:"business_id"`
	StaffID      int64     `json:"staff_id"`
	ReviewerID   int64     `json:"reviewer_id"`
	Period       string    `json:"period"`
	Rating       float64   `json:"rating"`
	GoalsRating  float64   `json:"goals_rating"`
	ValuesRating float64   `json:"values_rating"`
	SkillsRating float64   `json:"skills_rating"`
	Notes        string    `json:"notes"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
type ReviewInput struct {
	StaffID      int64   `json:"staff_id"`
	ReviewerID   int64   `json:"reviewer_id"`
	Period       string  `json:"period"`
	Rating       float64 `json:"rating"`
	GoalsRating  float64 `json:"goals_rating"`
	ValuesRating float64 `json:"values_rating"`
	SkillsRating float64 `json:"skills_rating"`
	Notes        string  `json:"notes"`
	Status       string  `json:"status"`
}
