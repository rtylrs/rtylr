package domain

import "time"

type Settings struct {
	ID                int64     `json:"id"`
	BusinessID        int64     `json:"businessId"`
	AutoAssign        bool      `json:"autoAssign"`
	SLAHours          int       `json:"slaHours"`
	BusinessHours     string    `json:"businessHours"`
	NotificationEmail string    `json:"notificationEmail"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type SettingsInput struct {
	BusinessID        int64
	AutoAssign        bool
	SLAHours          int
	BusinessHours     string
	NotificationEmail string
}
