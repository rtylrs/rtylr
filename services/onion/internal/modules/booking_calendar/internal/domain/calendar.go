package domain

type CalendarDay struct {
	Date         string                `json:"date"`
	Appointments []CalendarAppointment `json:"appointments"`
}

type CalendarAppointment struct {
	ID           int64  `json:"id"`
	ServiceID    int64  `json:"service_id"`
	CustomerName string `json:"customer_name"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	Status       string `json:"status"`
}
