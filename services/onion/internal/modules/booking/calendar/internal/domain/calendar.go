package domain

import "context"

type Repository interface {
	FetchAppointments(ctx context.Context, businessID int64, from, to string) ([]CalendarAppointment, error)
}

type CalendarDay struct {
	Date         string
	Appointments []CalendarAppointment
}

type CalendarAppointment struct {
	ID           int64
	BusinessID   int64
	ServiceID    int64
	CustomerName string
	StartTime    string
	EndTime      string
	Status       string
}
