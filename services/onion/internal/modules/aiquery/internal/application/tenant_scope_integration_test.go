package application_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/aiquery"
	"github.com/voxire/onion/internal/testutil"
)

// siblingBusinessID is a second business of the same owner as the fixture
// business. The pair is the whole point: owner_id is identical for both, so any
// query scoped by owner hands both businesses' rows to whoever asked.
const siblingBusinessID = testutil.BusinessID + 90001

// memberUserID belongs to the fixture business only. It has no access to
// siblingBusinessID, which is the caller the leak reaches.
const memberUserID = testutil.OwnerID + 90001

func TestRunJSON_LeavesSiblingBusinessInvisibleToItsMembers(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	// The owner must really own both businesses and the member must really be
	// limited to one of them, or the assertions below prove nothing.
	testutil.EnsureBusiness(t, db, siblingBusinessID, testutil.OwnerID)
	testutil.EnsureBusiness(t, db, testutil.BusinessID, testutil.OwnerID)
	grantMemberAccess(t, db, testutil.BusinessID)

	marker := fmt.Sprintf("aiq-%d", time.Now().UnixNano())
	mine := seedInvoice(t, db, marker+"-mine", testutil.BusinessID)
	theirs := seedInvoice(t, db, marker+"-sibling", siblingBusinessID)

	// The fixture business is shared across the suite, so the read is narrowed to
	// the rows this test seeded rather than asserting on the whole table.
	rows, sql, err := aiquery.NewModule(ctx).RunJSON(ctx, testutil.BusinessID, memberUserID, fmt.Sprintf(
		`{"view":"ai_invoices","columns":["invoice_number","client_name"],"where":[{"col":"client_name","op":"like","value":"%s%%"}],"limit":50}`,
		marker,
	))
	if err != nil {
		t.Fatalf("RunJSON: %v", err)
	}
	if !strings.Contains(sql, "business_id = ?") {
		t.Fatalf("query is not business scoped: %s", sql)
	}

	seen := map[string]bool{}
	for _, row := range rows {
		seen[fmt.Sprint(row["invoice_number"])] = true
	}
	if seen[theirs] {
		t.Fatalf("member of business %d received business %d's invoice %q: %v",
			testutil.BusinessID, siblingBusinessID, theirs, seen)
	}
	if !seen[mine] {
		t.Fatalf("expected the caller's own invoice %q in rows, got %v", mine, seen)
	}
}

// The seeded pair must differ only by business, so an empty result cannot be
// mistaken for isolation.
func TestRunJSON_SiblingBusinessSeesOnlyItsOwnRows(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	testutil.EnsureBusiness(t, db, siblingBusinessID, testutil.OwnerID)
	testutil.EnsureBusiness(t, db, testutil.BusinessID, testutil.OwnerID)

	marker := fmt.Sprintf("aiq-%d", time.Now().UnixNano())
	mine := seedInvoice(t, db, marker+"-mine", testutil.BusinessID)
	theirs := seedInvoice(t, db, marker+"-sibling", siblingBusinessID)

	rows, _, err := aiquery.NewModule(ctx).RunJSON(ctx, siblingBusinessID, testutil.OwnerID, fmt.Sprintf(
		`{"view":"ai_invoices","columns":["invoice_number"],"where":[{"col":"client_name","op":"like","value":"%s%%"}],"limit":50}`,
		marker,
	))
	if err != nil {
		t.Fatalf("RunJSON: %v", err)
	}

	seen := map[string]bool{}
	for _, row := range rows {
		seen[fmt.Sprint(row["invoice_number"])] = true
	}
	if !seen[theirs] {
		t.Fatalf("expected the sibling business's own invoice %q, got %v", theirs, seen)
	}
	if seen[mine] {
		t.Fatalf("business %d received business %d's invoice %q: %v",
			siblingBusinessID, testutil.BusinessID, mine, seen)
	}
}

func grantMemberAccess(t *testing.T, db *sql.DB, businessID int64) {
	t.Helper()
	ctx := context.Background()
	if _, err := db.ExecContext(ctx,
		`INSERT IGNORE INTO user_business_access (user_id, business_id, role) VALUES (?, ?, 'staff')`,
		memberUserID, businessID,
	); err != nil {
		t.Fatalf("grant member access to business %d: %v", businessID, err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(),
			`DELETE FROM user_business_access WHERE user_id = ? AND business_id = ?`, memberUserID, businessID)
	})
}

func seedInvoice(t *testing.T, db *sql.DB, invoiceNumber string, businessID int64) string {
	t.Helper()
	ctx := context.Background()
	if _, err := db.ExecContext(ctx,
		`INSERT INTO finance_invoices
		   (owner_id, business_id, invoice_number, client_name, issue_date, due_date, total, amount_paid, amount_due)
		 VALUES (?, ?, ?, 'aiq-probe', CURRENT_DATE, CURRENT_DATE, 10, 0, 10)`,
		testutil.OwnerID, businessID, invoiceNumber,
	); err != nil {
		t.Fatalf("seed finance_invoices (%d, %s): %v", businessID, invoiceNumber, err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(),
			`DELETE FROM finance_invoices WHERE invoice_number = ?`, invoiceNumber)
	})
	return invoiceNumber
}
