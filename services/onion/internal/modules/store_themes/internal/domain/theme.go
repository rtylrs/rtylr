package domain

import "time"

type Theme struct {
	ID           int64     `json:"id"`
	OwnerID      int64     `json:"ownerId"`
	Name         string    `json:"name"`
	PrimaryColor string    `json:"primaryColor"`
	Font         string    `json:"font"`
	HeaderStyle  string    `json:"headerStyle"`
	FooterStyle  string    `json:"footerStyle"`
	CustomCSS    string    `json:"customCss"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type ThemeInput struct {
	Name         string `json:"name"`
	PrimaryColor string `json:"primaryColor"`
	Font         string `json:"font"`
	HeaderStyle  string `json:"headerStyle"`
	FooterStyle  string `json:"footerStyle"`
	CustomCSS    string `json:"customCss"`
}
