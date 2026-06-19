package domain

import (
	"database/sql"
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

func (r *Review) FromRow(row *sql.Rows) error {
	var (
		reviewerID   sql.NullInt64
		period       sql.NullString
		rating       sql.NullFloat64
		goalsRating  sql.NullFloat64
		valuesRating sql.NullFloat64
		skillsRating sql.NullFloat64
		notes        sql.NullString
		status       sql.NullString
	)

	err := row.Scan(
		&r.ID, &r.OwnerID, &r.StaffID, &reviewerID, &period,
		&rating, &goalsRating, &valuesRating, &skillsRating,
		&notes, &status, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if reviewerID.Valid {
		r.ReviewerID = reviewerID.Int64
	}
	if period.Valid {
		r.Period = period.String
	}
	if rating.Valid {
		r.Rating = rating.Float64
	}
	if goalsRating.Valid {
		r.GoalsRating = goalsRating.Float64
	}
	if valuesRating.Valid {
		r.ValuesRating = valuesRating.Float64
	}
	if skillsRating.Valid {
		r.SkillsRating = skillsRating.Float64
	}
	if notes.Valid {
		r.Notes = notes.String
	}
	if status.Valid {
		r.Status = status.String
	}

	return nil
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
