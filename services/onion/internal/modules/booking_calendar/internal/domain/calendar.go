package domain

type CalendarDay struct {
	Date         string            `json:"date"`
	Appointments []CalendarAppointment `json:"appointments"`
}

type CalendarAppointment struct {
	ID            int64  `json:"id"`
	ServiceID     int64  `json:"serviceId"`
	CustomerName  string `json:"customerName"`
	StartTime     string `json:"startTime"`
	EndTime       string `json:"endTime"`
	Status        string `json:"status"`
}
