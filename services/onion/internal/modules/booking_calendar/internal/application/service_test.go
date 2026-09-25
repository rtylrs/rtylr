package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/booking_calendar/internal/domain"
)

type mockCalendarRepo struct {
	appts []domain.CalendarAppointment
	err   error
}

func (m *mockCalendarRepo) FetchAppointments(_ context.Context, _ int64, _, _ string) ([]domain.CalendarAppointment, error) {
	return m.appts, m.err
}

func newCalendarService(r domain.Repository) *Service {
	return &Service{repo: r}
}

func TestGetCalendar_ReturnsGroupedDays(t *testing.T) {
	appts := []domain.CalendarAppointment{
		{ID: 1, StartTime: "2026-06-01T09:00:00Z"},
		{ID: 2, StartTime: "2026-06-01T11:00:00Z"},
		{ID: 3, StartTime: "2026-06-02T10:00:00Z"},
	}
	svc := newCalendarService(&mockCalendarRepo{appts: appts})
	days, err := svc.fetchCalendar(context.Background(), 1, "2026-06-01", "2026-06-02")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(days) != 2 {
		t.Fatalf("expected 2 days, got %d", len(days))
	}
}

func TestGetCalendar_PropagatesError(t *testing.T) {
	svc := newCalendarService(&mockCalendarRepo{err: errors.New("db down")})
	_, err := svc.fetchCalendar(context.Background(), 1, "2026-06-01", "2026-06-02")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetCalendar_EmptyResult(t *testing.T) {
	svc := newCalendarService(&mockCalendarRepo{appts: nil})
	days, err := svc.fetchCalendar(context.Background(), 1, "2026-06-01", "2026-06-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(days) != 0 {
		t.Fatalf("expected 0 days, got %d", len(days))
	}
}

func TestGetCalendar_SingleAppointment(t *testing.T) {
	appts := []domain.CalendarAppointment{
		{ID: 42, StartTime: "2026-06-15T14:30:00Z"},
	}
	svc := newCalendarService(&mockCalendarRepo{appts: appts})
	days, err := svc.fetchCalendar(context.Background(), 1, "2026-06-15", "2026-06-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(days) != 1 {
		t.Fatalf("expected 1 day, got %d", len(days))
	}
	if len(days[0].Appointments) != 1 {
		t.Fatalf("expected 1 appointment on the day, got %d", len(days[0].Appointments))
	}
}

func TestGetCalendar_AppointmentsGroupedByDate(t *testing.T) {
	appts := []domain.CalendarAppointment{
		{ID: 1, StartTime: "2026-07-01T08:00:00Z"},
		{ID: 2, StartTime: "2026-07-01T09:00:00Z"},
		{ID: 3, StartTime: "2026-07-01T10:00:00Z"},
	}
	svc := newCalendarService(&mockCalendarRepo{appts: appts})
	days, err := svc.fetchCalendar(context.Background(), 1, "2026-07-01", "2026-07-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(days) != 1 {
		t.Fatalf("expected 1 day, got %d", len(days))
	}
	if len(days[0].Appointments) != 3 {
		t.Fatalf("expected 3 appointments, got %d", len(days[0].Appointments))
	}
}

func TestGetCalendar_DateKeyExtraction(t *testing.T) {
	appts := []domain.CalendarAppointment{
		{ID: 1, StartTime: "2026-06-10T00:00:00Z"},
	}
	svc := newCalendarService(&mockCalendarRepo{appts: appts})
	days, err := svc.fetchCalendar(context.Background(), 1, "2026-06-10", "2026-06-10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if days[0].Date != "2026-06-10" {
		t.Errorf("expected date 2026-06-10, got %q", days[0].Date)
	}
}

func TestGetCalendar_ShortStartTimeBecomesUnknown(t *testing.T) {
	appts := []domain.CalendarAppointment{
		{ID: 1, StartTime: "2026"},
	}
	svc := newCalendarService(&mockCalendarRepo{appts: appts})
	days, err := svc.fetchCalendar(context.Background(), 1, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(days) != 1 {
		t.Fatalf("expected 1 day bucket, got %d", len(days))
	}
	if days[0].Date != "unknown" {
		t.Errorf("expected date 'unknown', got %q", days[0].Date)
	}
}

func TestGetCalendar_MultipleOwnersAreIndependent(t *testing.T) {
	appts := []domain.CalendarAppointment{
		{ID: 100, StartTime: "2026-06-01T09:00:00Z"},
	}
	repo := &mockCalendarRepo{appts: appts}
	svc := newCalendarService(repo)

	days1, err := svc.fetchCalendar(context.Background(), 1, "2026-06-01", "2026-06-01")
	if err != nil {
		t.Fatalf("unexpected error owner1: %v", err)
	}
	days2, err := svc.fetchCalendar(context.Background(), 2, "2026-06-01", "2026-06-01")
	if err != nil {
		t.Fatalf("unexpected error owner2: %v", err)
	}
	if len(days1) != len(days2) {
		t.Errorf("expected same count, got %d vs %d", len(days1), len(days2))
	}
}

func TestGetCalendar_FourDays(t *testing.T) {
	appts := []domain.CalendarAppointment{
		{ID: 1, StartTime: "2026-08-01T09:00:00Z"},
		{ID: 2, StartTime: "2026-08-02T09:00:00Z"},
		{ID: 3, StartTime: "2026-08-03T09:00:00Z"},
		{ID: 4, StartTime: "2026-08-04T09:00:00Z"},
	}
	svc := newCalendarService(&mockCalendarRepo{appts: appts})
	days, err := svc.fetchCalendar(context.Background(), 1, "2026-08-01", "2026-08-04")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(days) != 4 {
		t.Fatalf("expected 4 days, got %d", len(days))
	}
}

func TestGetCalendar_ErrorReturnsNilSlice(t *testing.T) {
	svc := newCalendarService(&mockCalendarRepo{err: errors.New("timeout")})
	days, err := svc.fetchCalendar(context.Background(), 99, "2026-01-01", "2026-01-31")
	if err == nil {
		t.Fatal("expected error")
	}
	if days != nil {
		t.Errorf("expected nil days on error, got %v", days)
	}
}

func TestGetCalendar_AppointmentFieldsPreserved(t *testing.T) {
	appts := []domain.CalendarAppointment{
		{ID: 7, ServiceID: 3, CustomerName: "Alice", StartTime: "2026-06-20T10:00:00Z", EndTime: "2026-06-20T11:00:00Z", Status: "confirmed"},
	}
	svc := newCalendarService(&mockCalendarRepo{appts: appts})
	days, err := svc.fetchCalendar(context.Background(), 1, "2026-06-20", "2026-06-20")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	a := days[0].Appointments[0]
	if a.ID != 7 || a.CustomerName != "Alice" || a.Status != "confirmed" {
		t.Errorf("appointment fields not preserved: %+v", a)
	}
}
