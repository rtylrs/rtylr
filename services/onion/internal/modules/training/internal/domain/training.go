package domain

import (
	"time"
)

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
