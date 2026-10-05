package domain

import "time"

type Project struct {
	ID        string
	Name      string
	Icon      string
	Color     string
	Context   string
	UpdatedAt time.Time
}
