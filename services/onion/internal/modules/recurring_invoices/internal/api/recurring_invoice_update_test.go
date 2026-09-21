package api

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	onekithttp "github.com/corezio/onekit/http"

	"github.com/voxire/onion/internal/modules/recurring_invoices/internal/application"
	"github.com/voxire/onion/internal/modules/recurring_invoices/internal/domain"
	"github.com/voxire/onion/internal/modules/recurring_invoices/internal/infrastructure"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/testutil"
	financerecurringinvoicev1 "github.com/voxire/voxie/gen/go/finance/recurring_invoice/v1"
)

// A dedicated business keeps this fixture away from the shared fixture tenant
// and from the seeded businesses other packages own.
const (
	recurringUpdateTestBusinessID int64 = 907301
	recurringUpdateTestUserID     int64 = 907201
)

func recurringStringPtr(value string) *string { return &value }

func recurringBoolPtr(value bool) *bool { return &value }

func recurringInvoiceStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if errors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

func seedRecurringUpdateTest(t *testing.T, sqlDB *sql.DB) (context.Context, *RecurringInvoiceAPI, *infrastructure.Repository, int64) {
	t.Helper()

	repo := infrastructure.NewRepository(sqlDB)
	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		UserID:     recurringUpdateTestUserID,
		BusinessID: recurringUpdateTestBusinessID,
	})

	id, err := repo.Create(ctx, recurringUpdateTestBusinessID, domain.RecurringInvoiceInput{
		TemplateInvoiceID: testutil.Int64Ptr(4242),
		Frequency:         "monthly",
		NextInvoiceDate:   "2026-06-01",
		EndDate:           recurringStringPtr("2026-12-31"),
		IsActive:          recurringBoolPtr(true),
	})
	if err != nil {
		t.Fatalf("seed recurring invoice: %v", err)
	}

	return ctx, NewRecurringInvoiceAPI(application.NewService(ctx, sqlDB)), repo, id
}

// Regression: a PUT that omits the optional is_active/end_date fields collapsed
// them to false/"" and paused the schedule and cleared its bound. Omitted
// fields must keep the stored values.
func TestUpdateRecurringInvoice_OmittedOptionalFields_KeepStoredValues(t *testing.T) {
	sqlDB := testutil.OpenDB(t)
	testutil.GrantBusinessAccess(t, sqlDB, recurringUpdateTestBusinessID, recurringUpdateTestUserID)
	ctx, api, repo, id := seedRecurringUpdateTest(t, sqlDB)

	if _, err := api.UpdateRecurringInvoice(ctx, &financerecurringinvoicev1.UpdateRecurringInvoiceRequest{
		Id: id,
		Body: &financerecurringinvoicev1.UpdateRecurringInvoiceBody{
			Frequency:       "quarterly",
			NextInvoiceDate: "2026-09-01",
		},
	}); err != nil {
		t.Fatalf("UpdateRecurringInvoice: %v", err)
	}

	items, err := repo.List(ctx, recurringUpdateTestBusinessID)
	if err != nil {
		t.Fatalf("List after update: %v", err)
	}
	for _, item := range items {
		if item.ID != id {
			continue
		}
		if item.Frequency != "quarterly" {
			t.Errorf("Frequency = %q, want quarterly", item.Frequency)
		}
		if !item.IsActive {
			t.Error("IsActive = false after an update that omitted is_active, want true")
		}
		wantEnd := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
		if !item.EndDate.Equal(wantEnd) {
			t.Errorf("EndDate = %v after an update that omitted end_date, want %v", item.EndDate, wantEnd)
		}
		if item.TemplateInvoiceID != 4242 {
			t.Errorf("TemplateInvoiceID = %d after an update that omitted it, want the stored 4242", item.TemplateInvoiceID)
		}
		return
	}
	t.Fatalf("recurring invoice id=%d not found after update", id)
}

// Regression: an update for a missing (or another business's) id answered 200
// "Recurring invoice updated successfully".
func TestUpdateRecurringInvoice_UnknownID_Returns404(t *testing.T) {
	sqlDB := testutil.OpenDB(t)
	testutil.GrantBusinessAccess(t, sqlDB, recurringUpdateTestBusinessID, recurringUpdateTestUserID)
	ctx, api, _, _ := seedRecurringUpdateTest(t, sqlDB)

	_, err := api.UpdateRecurringInvoice(ctx, &financerecurringinvoicev1.UpdateRecurringInvoiceRequest{
		Id: 999999999,
		Body: &financerecurringinvoicev1.UpdateRecurringInvoiceBody{
			Frequency:       "monthly",
			NextInvoiceDate: "2026-06-01",
		},
	})
	if got := recurringInvoiceStatus(err); got != 404 {
		t.Fatalf("update of unknown id = %d, want 404 (err=%v)", got, err)
	}
}

// Regression: a malformed end_date succeeded and was stored as open-ended.
func TestUpdateRecurringInvoice_MalformedEndDate_Returns400(t *testing.T) {
	sqlDB := testutil.OpenDB(t)
	testutil.GrantBusinessAccess(t, sqlDB, recurringUpdateTestBusinessID, recurringUpdateTestUserID)
	ctx, api, repo, id := seedRecurringUpdateTest(t, sqlDB)

	_, err := api.UpdateRecurringInvoice(ctx, &financerecurringinvoicev1.UpdateRecurringInvoiceRequest{
		Id: id,
		Body: &financerecurringinvoicev1.UpdateRecurringInvoiceBody{
			Frequency:       "monthly",
			NextInvoiceDate: "2026-06-01",
			EndDate:         recurringStringPtr("31/12/2026"),
		},
	})
	if got := recurringInvoiceStatus(err); got != 400 {
		t.Fatalf("update with malformed end_date = %d, want 400 (err=%v)", got, err)
	}

	items, err := repo.List(ctx, recurringUpdateTestBusinessID)
	if err != nil {
		t.Fatalf("List after rejected update: %v", err)
	}
	for _, item := range items {
		if item.ID != id {
			continue
		}
		wantEnd := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
		if !item.EndDate.Equal(wantEnd) {
			t.Errorf("EndDate = %v after a rejected update, want the stored %v", item.EndDate, wantEnd)
		}
		return
	}
	t.Fatalf("recurring invoice id=%d not found after rejected update", id)
}

// Regression: pausing a schedule that is already paused returned 404 because
// MySQL reports zero changed rows for the no-op update.
func TestPauseRecurringInvoice_AlreadyPaused_Succeeds(t *testing.T) {
	sqlDB := testutil.OpenDB(t)
	testutil.GrantBusinessAccess(t, sqlDB, recurringUpdateTestBusinessID, recurringUpdateTestUserID)
	ctx, api, _, id := seedRecurringUpdateTest(t, sqlDB)

	if _, err := api.PauseRecurringInvoice(ctx, &financerecurringinvoicev1.PauseRecurringInvoiceRequest{Id: id}); err != nil {
		t.Fatalf("first pause: %v", err)
	}
	if _, err := api.PauseRecurringInvoice(ctx, &financerecurringinvoicev1.PauseRecurringInvoiceRequest{Id: id}); err != nil {
		t.Fatalf("second pause on an already paused schedule: %v", err)
	}

	if _, err := api.PauseRecurringInvoice(ctx, &financerecurringinvoicev1.PauseRecurringInvoiceRequest{Id: 999999999}); recurringInvoiceStatus(err) != 404 {
		t.Fatalf("pause of unknown id = %d, want 404 (err=%v)", recurringInvoiceStatus(err), err)
	}
}
