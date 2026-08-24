package infrastructure_test

import (
	"testing"

	"github.com/voxire/onion/internal/testutil"
)

func TestAppointmentListQueriesHaveServerBounds(t *testing.T) {
	testutil.RequireBoundedListQueries(t, "../../sql/queries/booking_appointments.sql")
}
