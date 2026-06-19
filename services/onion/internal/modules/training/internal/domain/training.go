package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrTrainingCourseNotFound = errors.New("training course not found")

type TrainingCourse struct {
	ID            int64              `json:"id"`
	OwnerID       int64              `json:"owner_id"`
	BusinessID    int64              `json:"business_id"`
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	DurationHours float64            `json:"duration_hours"`
	IsRequired    bool               `json:"is_required"`
	CreatedAt     time.Time          `json:"created_at"`
	Progress      []TrainingProgress `json:"progress,omitempty"`
}

func (t *TrainingCourse) FromRow(row *sql.Rows) error {
	var (
		description   sql.NullString
		durationHours sql.NullFloat64
	)

	err := row.Scan(&t.ID, &t.OwnerID, &t.BusinessID, &t.Name, &description, &durationHours, &t.IsRequired, &t.CreatedAt)
	if err != nil {
		return err
	}

	if description.Valid {
		t.Description = description.String
	}
	if durationHours.Valid {
		t.DurationHours = durationHours.Float64
	}

	return nil
}

type TrainingProgress struct {
	ID          int64     `json:"id"`
	CourseID    int64     `json:"course_id"`
	StaffID     int64     `json:"staff_id"`
	Status      string    `json:"status"`
	CompletedAt time.Time `json:"completed_at"`
}

type TrainingCourseInput struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	DurationHours float64 `json:"duration_hours"`
	IsRequired    bool    `json:"is_required"`
}
