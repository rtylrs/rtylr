package customersync_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	customersync "github.com/voxire/onion/internal/platform/customer_sync"
	"github.com/voxire/onion/internal/testutil"
)

// TestUpsertCRMCustomer_EmailOnlySyncKeepsStoredName guards the update branch:
// a checkout that carries only the customer email must not replace the stored
// CRM name with the email local part derived for fresh rows.
func TestUpsertCRMCustomer_EmailOnlySyncKeepsStoredName(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	businessID, ownerID := fixtureBusiness(t, db)
	email := fmt.Sprintf("name-preserve-%d@example.com", time.Now().UnixNano())
	customerID := insertCRMCustomer(t, db, ownerID, businessID, "John", "Doe", email)

	if _, err := customersync.UpsertCRMCustomer(ctx, db, businessID, customersync.Contact{Email: email}); err != nil {
		t.Fatalf("UpsertCRMCustomer: %v", err)
	}

	firstName, lastName := readCustomerName(t, db, customerID)
	if firstName != "John" {
		t.Errorf("first_name was overwritten by the email fallback: got %q, want %q", firstName, "John")
	}
	if lastName != "Doe" {
		t.Errorf("last_name = %q, want %q", lastName, "Doe")
	}
}

// TestUpsertCRMCustomer_InsertDerivesNameFromEmail keeps the fallback for fresh
// rows: with no stored name to protect, the email local part is still used.
func TestUpsertCRMCustomer_InsertDerivesNameFromEmail(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	businessID, _ := fixtureBusiness(t, db)
	email := fmt.Sprintf("fallback-%d@example.com", time.Now().UnixNano())

	customerID, err := customersync.UpsertCRMCustomer(ctx, db, businessID, customersync.Contact{Email: email})
	if err != nil {
		t.Fatalf("UpsertCRMCustomer: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM crm_customers WHERE id = ?`, customerID)
	})

	firstName, _ := readCustomerName(t, db, customerID)
	want := email[:len(email)-len("@example.com")]
	if firstName != want {
		t.Errorf("first_name = %q, want email-derived %q", firstName, want)
	}
}

func fixtureBusiness(t *testing.T, db *sql.DB) (businessID, ownerID int64) {
	t.Helper()

	err := db.QueryRow(
		`SELECT id, owner_id FROM businesses WHERE id = 1 AND is_active = TRUE`,
	).Scan(&businessID, &ownerID)
	if err != nil {
		t.Fatalf("read fixture business: %v", err)
	}
	return businessID, ownerID
}

func insertCRMCustomer(t *testing.T, db *sql.DB, ownerID, businessID int64, firstName, lastName, email string) int64 {
	t.Helper()

	result, err := db.Exec(
		`INSERT INTO crm_customers (owner_id, business_id, first_name, last_name, email, total_spent, visit_count)
		 VALUES (?, ?, ?, ?, ?, 0, 0)`,
		ownerID, businessID, firstName, lastName, email,
	)
	if err != nil {
		t.Fatalf("insert crm customer: %v", err)
	}
	customerID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("read crm customer id: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DELETE FROM crm_customers WHERE id = ?`, customerID)
	})
	return customerID
}

func readCustomerName(t *testing.T, db *sql.DB, customerID int64) (string, string) {
	t.Helper()

	var firstName string
	var lastName sql.NullString
	if err := db.QueryRow(
		`SELECT first_name, last_name FROM crm_customers WHERE id = ?`, customerID,
	).Scan(&firstName, &lastName); err != nil {
		t.Fatalf("read crm customer: %v", err)
	}
	return firstName, lastName.String
}
