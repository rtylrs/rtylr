package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/booking/dashboard/internal/application"
	"github.com/voxire/onion/internal/modules/booking/dashboard/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/testutil"
	bookingdashboardv1 "github.com/voxire/voxie/gen/go/booking/dashboard/v1"
)

func TestDashboardHTTPKeepsRevenueAndBusinessIsolation(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	const businessID, ownerID = int64(9968101), int64(9968102)
	testutil.GrantBusinessAccess(t, db, businessID, ownerID)
	t.Cleanup(func() { _, _ = db.Exec("DELETE FROM users WHERE id = ?", ownerID) })
	res, err := db.Exec("INSERT INTO booking_services (owner_id, business_id, name, price) VALUES (?, ?, 'Transport check', 1234.56)", ownerID, businessID)
	if err != nil {
		t.Fatal(err)
	}
	serviceID, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = db.Exec("DELETE FROM booking_services WHERE id = ?", serviceID) })
	res, err = db.Exec("INSERT INTO booking_appointments (owner_id, business_id, service_id, start_time, end_time, status) VALUES (?, ?, ?, ?, ?, 'completed')", ownerID, businessID, serviceID, time.Now(), time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	appointmentID, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = db.Exec("DELETE FROM booking_appointments WHERE id = ?", appointmentID) })
	mux := http.NewServeMux()
	if err := bookingdashboardv1.RegisterDashboardServiceServer(NewDashboardAPI(application.NewService(ctx, infrastructure.NewRepository(db))), bookingdashboardv1.WithMux(mux)); err != nil {
		t.Fatal(err)
	}
	for _, user := range []auth.AuthUser{{UserID: ownerID, BusinessID: businessID}, {UserID: testutil.OwnerID, BusinessID: testutil.BusinessID}} {
		req := httptest.NewRequest(http.MethodGet, "/booking/dashboard/v1/dashboard", nil)
		req = req.WithContext(context.WithValue(req.Context(), auth.AuthUserKey, &user))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, req)
		if response.Code != http.StatusOK {
			t.Fatalf("dashboard status=%d: %s", response.Code, response.Body.String())
		}
		var payload map[string]json.RawMessage
		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
			t.Fatal(err)
		}
		var revenue float64
		if raw, exists := payload["revenue"]; exists {
			if err := json.Unmarshal(raw, &revenue); err != nil {
				t.Fatal(err)
			}
		}
		if user.BusinessID == businessID && revenue != 1234.56 {
			t.Fatalf("revenue=%v", revenue)
		}
		if user.BusinessID != businessID && revenue == 1234.56 {
			t.Fatal("foreign business revenue leaked")
		}
		if _, exists := payload["Revenue"]; exists {
			t.Fatal("domain representation leaked into HTTP")
		}
	}
}
