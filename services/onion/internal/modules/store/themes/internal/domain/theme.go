package domain

import "time"

type Theme struct {
	ID           int64
	OwnerID      int64
	BusinessID   int64
	Name         string
	PrimaryColor string
	Font         string
	HeaderStyle  string
	FooterStyle  string
	CustomCSS    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ThemeInput struct {
	Name         string
	PrimaryColor string
	Font         string
	HeaderStyle  string
	FooterStyle  string
	CustomCSS    string
}
