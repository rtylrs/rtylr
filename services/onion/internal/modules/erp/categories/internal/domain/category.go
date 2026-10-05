package domain

import (
	"time"
)

type Category struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
type CategoryInput struct {
	Name        string
	Description string
}
