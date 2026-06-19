package domain

import "time"

type Theme struct {
	ID           int64     `json:"id"`
	OwnerID      int64     `json:"owner_id"`
	BusinessID   int64     `json:"business_id"`
	Name         string    `json:"name"`
	PrimaryColor string    `json:"primary_color"`
	Font         string    `json:"font"`
	HeaderStyle  string    `json:"header_style"`
	FooterStyle  string    `json:"footer_style"`
	CustomCSS    string    `json:"custom_css"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ThemeInput struct {
	Name         string `json:"name"`
	PrimaryColor string `json:"primary_color"`
	Font         string `json:"font"`
	HeaderStyle  string `json:"header_style"`
	FooterStyle  string `json:"footer_style"`
	CustomCSS    string `json:"custom_css"`
}
