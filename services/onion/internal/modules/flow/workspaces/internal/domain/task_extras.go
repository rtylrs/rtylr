package domain

import "time"

type TaskAttachment struct {
	ID         int64
	TaskID     int64
	OwnerID    int64
	BusinessID int64
	Name       string
	URL        string
	Size       int64
	UploadedBy int64
	CreatedAt  time.Time
}

type TaskAttachmentInput struct {
	TaskID     int64
	OwnerID    int64
	BusinessID int64
	Name       string
	URL        string
	Size       int64
	UploadedBy int64
}

type TimeLog struct {
	ID              int64
	OwnerID         int64
	BusinessID      int64
	TaskID          int64
	UserID          int64
	DurationMinutes int
	Description     string
	LoggedAt        time.Time
	CreatedAt       time.Time
}

type TimeLogInput struct {
	OwnerID         int64
	BusinessID      int64
	TaskID          int64
	UserID          int64
	DurationMinutes int
	Description     string
}
