package domain

import "time"

type WaitlistEntry struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"ownerId"`
	ServiceID     int64     `json:"serviceId"`
	CustomerName  string    `json:"customerName"`
	CustomerEmail string    `json:"customerEmail"`
	CustomerPhone string    `json:"customerPhone"`
	PreferredDate string    `json:"preferredDate"`
	PreferredTime string    `json:"preferredTime"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
}

type WaitlistInput struct {
	ServiceID     int64  `json:"serviceId"`
	CustomerName  string `json:"customerName"`
	CustomerEmail string `json:"customerEmail"`
	CustomerPhone string `json:"customerPhone"`
	PreferredDate string `json:"preferredDate"`
	PreferredTime string `json:"preferredTime"`
	Status        string `json:"status"`
}
