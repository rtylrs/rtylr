package domain

import "time"

type Addon struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	Icon             string  `json:"icon"`
	Price            float64 `json:"price"`
	IsFree           bool    `json:"isFree"`
	IsDefaultEnabled bool    `json:"isDefaultEnabled"`
	SortOrder        int     `json:"sortOrder"`
}

type BusinessAddon struct {
	AddonID     string     `json:"addonId"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Icon        string     `json:"icon"`
	Price       float64    `json:"price"`
	IsFree      bool       `json:"isFree"`
	IsEnabled   bool       `json:"isEnabled"`
	EnabledAt   *time.Time `json:"enabledAt,omitempty"`
	DisabledAt  *time.Time `json:"disabledAt,omitempty"`
}
