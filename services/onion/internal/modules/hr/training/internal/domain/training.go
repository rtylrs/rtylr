package domain

import (
	"time"
)

type TrainingCourse struct {
	ID            int64
	OwnerID       int64
	BusinessID    int64
	Name          string
	Description   string
	DurationHours float64
	IsRequired    bool
	CreatedAt     time.Time
	Progress      []TrainingProgress
}
type TrainingProgress struct {
	ID          int64
	CourseID    int64
	StaffID     int64
	Status      string
	CompletedAt time.Time
}

type TrainingCourseInput struct {
	Name          string
	Description   string
	DurationHours float64
	IsRequired    bool
}
