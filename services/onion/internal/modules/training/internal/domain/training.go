package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrTrainingCourseNotFound = errors.New("training course not found")

type TrainingCourse struct {
	ID            int64              `json:"id"`
	OwnerID       int64              `json:"ownerId"`
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	DurationHours float64            `json:"durationHours"`
	IsRequired    bool               `json:"isRequired"`
	CreatedAt     time.Time          `json:"createdAt"`
	Progress      []TrainingProgress `json:"progress,omitempty"`
}

func (t *TrainingCourse) FromRow(row *sql.Rows) error {
	var (
		description   sql.NullString
		durationHours sql.NullFloat64
	)

	err := row.Scan(&t.ID, &t.OwnerID, &t.Name, &description, &durationHours, &t.IsRequired, &t.CreatedAt)
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
	CourseID    int64     `json:"courseId"`
	StaffID     int64     `json:"staffId"`
	Status      string    `json:"status"`
	CompletedAt time.Time `json:"completedAt"`
}

type TrainingCourseInput struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	DurationHours float64 `json:"durationHours"`
	IsRequired    bool    `json:"isRequired"`
}
