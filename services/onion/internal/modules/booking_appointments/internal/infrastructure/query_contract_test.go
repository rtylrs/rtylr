package infrastructure_test

import (
	"os"
	"strings"
	"testing"
)

func TestAppointmentListQueriesHaveServerBounds(t *testing.T) {
	query, err := os.ReadFile("../../sql/queries/booking_appointments.sql")
	if err != nil {
		t.Fatalf("read appointment queries: %v", err)
	}
	if got := strings.Count(string(query), "LIMIT 500"); got != 2 {
		t.Fatalf("expected two bounded appointment list queries, found %d", got)
	}
}
