package domain

import "time"

type ShippingMethod struct {
	ID            int64
	OwnerID       int64
	BusinessID    int64
	Name          string
	Price         float64
	EstimatedDays int
	IsActive      bool
	CreatedAt     time.Time
}

type ShippingMethodInput struct {
	Name          string
	Price         float64
	EstimatedDays int
	IsActive      bool
}
