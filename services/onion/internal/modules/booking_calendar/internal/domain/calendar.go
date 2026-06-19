package domain

import "context"

type Repository interface {
	FetchAppointments(ctx context.Context, businessID int64, from, to string) ([]CalendarAppointment, error)
}

type CalendarDay struct {
	Date         string                `json:"date"`
	Appointments []CalendarAppointment `json:"appointments"`
}

type CalendarAppointment struct {
	ID           int64  `json:"id"`
	BusinessID   int64  `json:"business_id"`
	ServiceID    int64  `json:"service_id"`
	CustomerName string `json:"customer_name"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	Status       string `json:"status"`
}
