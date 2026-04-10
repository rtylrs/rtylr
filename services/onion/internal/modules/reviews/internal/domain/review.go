package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrReviewNotFound = errors.New("review not found")

type Review struct {
	ID           int64     `json:"id"`
	OwnerID      int64     `json:"ownerId"`
	StaffID      int64     `json:"staffId"`
	ReviewerID   int64     `json:"reviewerId"`
	Period       string    `json:"period"`
	Rating       float64   `json:"rating"`
	GoalsRating  float64   `json:"goalsRating"`
	ValuesRating float64   `json:"valuesRating"`
	SkillsRating float64   `json:"skillsRating"`
	Notes        string    `json:"notes"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
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
	StaffID      int64   `json:"staffId"`
	ReviewerID   int64   `json:"reviewerId"`
	Period       string  `json:"period"`
	Rating       float64 `json:"rating"`
	GoalsRating  float64 `json:"goalsRating"`
	ValuesRating float64 `json:"valuesRating"`
	SkillsRating float64 `json:"skillsRating"`
	Notes        string  `json:"notes"`
	Status       string  `json:"status"`
}
