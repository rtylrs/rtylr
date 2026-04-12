package domain

import "time"

type WaitlistEntry struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"owner_id"`
	ServiceID     int64     `json:"service_id"`
	CustomerName  string    `json:"customer_name"`
	CustomerEmail string    `json:"customer_email"`
	CustomerPhone string    `json:"customer_phone"`
	PreferredDate string    `json:"preferred_date"`
	PreferredTime string    `json:"preferred_time"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type WaitlistInput struct {
	ServiceID     int64  `json:"service_id"`
	CustomerName  string `json:"customer_name"`
	CustomerEmail string `json:"customer_email"`
	CustomerPhone string `json:"customer_phone"`
	PreferredDate string `json:"preferred_date"`
	PreferredTime string `json:"preferred_time"`
	Status        string `json:"status"`
}
