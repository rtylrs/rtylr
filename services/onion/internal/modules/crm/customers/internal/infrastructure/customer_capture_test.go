package infrastructure_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	crmcontract "github.com/voxire/onion/internal/contracts/crm"
	"github.com/voxire/onion/internal/modules/organization/business"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/eventbus"
	"github.com/voxire/onion/internal/testutil"
)

func TestCaptureCustomerContactIsIdempotentAndPublishesLink(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	sourceID := time.Now().UnixNano()
	email := fmt.Sprintf("capture-%d@example.test", sourceID)
	contact := crmcontract.CustomerContactCaptured{
		Name:       "Capture Test",
		Email:      email,
		TotalSpent: 12.5,
		CountVisit: true,
	}

	if err := eventbus.NewPublisher(db).Record(ctx, eventbus.Event{
		BusinessID: testutil.BusinessID, Type: crmcontract.CustomerContactCapturedType,
		SubjectType: "capturetest", SubjectID: sourceID, Payload: contact,
	}); err != nil {
		t.Fatalf("record contact capture: %v", err)
	}
	var eventID uint64
	if err := db.QueryRowContext(ctx, `
		SELECT id FROM platform_events
		WHERE business_id = ? AND event_type = ? AND subject_type = ? AND subject_id = ?`,
		testutil.BusinessID, crmcontract.CustomerContactCapturedType, "capturetest", sourceID,
	).Scan(&eventID); err != nil {
		t.Fatalf("find captured event: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM crm_customer_capture_events WHERE event_id = ?`, eventID)
		_, _ = db.ExecContext(ctx, `DELETE FROM crm_outbox WHERE business_id = ? AND dedupe_key LIKE ?`, testutil.BusinessID, fmt.Sprintf("crm_customer_linked:capturetest:%d:%%", sourceID))
		_, _ = db.ExecContext(ctx, `DELETE FROM crm_customers WHERE business_id = ? AND email = ?`, testutil.BusinessID, email)
		_, _ = db.ExecContext(ctx, `DELETE FROM platform_events WHERE id = ?`, eventID)
	})

	event := eventbus.Event{ID: eventID, BusinessID: testutil.BusinessID, SubjectType: "capturetest", SubjectID: sourceID}
	repo := infrastructure.NewRepository(db).WithReportingTimezone(business.NewModule(context.Background(), db).ReportingTimezone)
	for i := 0; i < 2; i++ {
		if err := repo.CaptureCustomerContact(ctx, event, testutil.OwnerID, contact); err != nil {
			t.Fatalf("capture attempt %d: %v", i+1, err)
		}
	}

	var customerID int64
	var totalSpent string
	var visits int
	if err := db.QueryRowContext(ctx, `
		SELECT id, total_spent, visit_count FROM crm_customers
		WHERE business_id = ? AND email = ?`, testutil.BusinessID, email,
	).Scan(&customerID, &totalSpent, &visits); err != nil {
		t.Fatalf("find captured customer: %v", err)
	}
	if customerID <= 0 || totalSpent != "12.50" || visits != 1 {
		t.Fatalf("captured customer id=%d spend=%s visits=%d, want positive id, 12.50, 1", customerID, totalSpent, visits)
	}

	var linkedCount int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM crm_outbox
		WHERE business_id = ? AND event_type = ? AND dedupe_key = ?`,
		testutil.BusinessID,
		crmcontract.CustomerLinkedType,
		crmcontract.CustomerLinkDedupeKey("capturetest", sourceID, customerID),
	).Scan(&linkedCount); err != nil {
		t.Fatalf("count customer link events: %v", err)
	}
	if linkedCount != 1 {
		t.Fatalf("customer link event count = %d, want 1", linkedCount)
	}
	var processedCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM crm_customer_capture_events WHERE event_id = ?`, eventID).Scan(&processedCount); err != nil {
		t.Fatalf("verify capture event record: %v", err)
	}
	if processedCount != 1 {
		t.Fatalf("capture event record count = %d, want 1", processedCount)
	}
}
