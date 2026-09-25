package infrastructure_test

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/quotes/internal/domain"
	"github.com/voxire/onion/internal/modules/quotes/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func createConvertibleQuote(t *testing.T, db *sql.DB, repo *infrastructure.Repository, ctx context.Context, customer string) int64 {
	t.Helper()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.QuoteInput{
		CustomerName: customer,
		Status:       "draft",
		Subtotal:     100,
		Total:        100,
	})
	if err != nil {
		t.Fatalf("quotes.Create(%s): %v", customer, err)
	}
	// Conversion flips the quote to 'converted', so it can no longer be
	// removed through the repository.
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM finance_quotes WHERE id = ? AND business_id = ?", id, testutil.OwnerID)
	})
	return id
}

func deleteInvoice(t *testing.T, db *sql.DB, ctx context.Context, invoiceID int64) {
	t.Helper()

	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM finance_invoice_items WHERE invoice_id = ?", invoiceID)
		_, _ = db.ExecContext(ctx, "DELETE FROM finance_invoices WHERE id = ? AND business_id = ?", invoiceID, testutil.OwnerID)
	})
}

// TestConvertToInvoice_ConcurrentConversionsMintDistinctInvoiceNumbers converts
// several quotes of one business at the same time. Invoice numbers are minted
// from COUNT(*)+1, so without per-business serialization two conversions can
// pick the same number and the loser fails the unique
// (business_id, invoice_number) insert with a duplicate-key error.
func TestConvertToInvoice_ConcurrentConversionsMintDistinctInvoiceNumbers(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const conversions = 4
	quoteIDs := make([]int64, 0, conversions)
	for i := 0; i < conversions; i++ {
		quoteIDs = append(quoteIDs, createConvertibleQuote(t, db, repo, ctx, fmt.Sprintf("Concurrent Customer %d", i)))
	}

	type result struct {
		quoteID   int64
		invoiceID int64
		err       error
	}

	start := make(chan struct{})
	results := make(chan result, conversions)
	var wg sync.WaitGroup
	for _, quoteID := range quoteIDs {
		wg.Add(1)
		go func(quoteID int64) {
			defer wg.Done()
			<-start
			invoiceID, err := repo.ConvertToInvoice(ctx, testutil.OwnerID, quoteID)
			results <- result{quoteID: quoteID, invoiceID: invoiceID, err: err}
		}(quoteID)
	}
	close(start)
	wg.Wait()
	close(results)

	numbers := make(map[string]int64, conversions)
	for res := range results {
		if res.err != nil {
			t.Errorf("ConvertToInvoice(quote=%d): unexpected error: %v", res.quoteID, res.err)
			continue
		}
		deleteInvoice(t, db, ctx, res.invoiceID)

		var number string
		if err := db.QueryRowContext(ctx,
			"SELECT invoice_number FROM finance_invoices WHERE id = ? AND business_id = ?",
			res.invoiceID, testutil.OwnerID,
		).Scan(&number); err != nil {
			t.Errorf("invoice %d for quote %d not readable: %v", res.invoiceID, res.quoteID, err)
			continue
		}
		if first, dup := numbers[number]; dup {
			t.Errorf("invoice number %q minted for both quote %d and quote %d", number, first, res.quoteID)
			continue
		}
		numbers[number] = res.quoteID
	}

	if len(numbers) != conversions {
		t.Fatalf("expected %d distinct invoices, got %d", conversions, len(numbers))
	}
}

// TestConvertToInvoice_SerializesPerBusiness locks the business row the way a
// competing conversion does and verifies the conversion waits for it and then
// mints the *next* number. A conversion that reads the count before the lock
// would mint the already-taken number and fail the unique insert.
func TestConvertToInvoice_SerializesPerBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	year := time.Now().UTC().Year()

	quoteID := createConvertibleQuote(t, db, repo, ctx, "Serialized Customer")

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback() })

	var lockedBusinessID int64
	if err = tx.QueryRowContext(ctx,
		"SELECT id FROM businesses WHERE id = ? FOR UPDATE", testutil.OwnerID,
	).Scan(&lockedBusinessID); err != nil {
		t.Fatalf("lock business row: %v", err)
	}

	type result struct {
		invoiceID int64
		err       error
	}
	done := make(chan result, 1)
	go func() {
		invoiceID, err := repo.ConvertToInvoice(ctx, testutil.OwnerID, quoteID)
		done <- result{invoiceID: invoiceID, err: err}
	}()

	select {
	case res := <-done:
		if res.invoiceID > 0 {
			deleteInvoice(t, db, ctx, res.invoiceID)
		}
		t.Fatalf("conversion raced past the held business lock: invoice=%d err=%v", res.invoiceID, res.err)
	case <-time.After(300 * time.Millisecond):
	}

	// The competing conversion takes the next number and commits it.
	var sequence int64
	if err = tx.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM finance_invoices WHERE business_id = ? AND YEAR(created_at) = ?",
		testutil.OwnerID, year,
	).Scan(&sequence); err != nil {
		t.Fatalf("count invoice sequence: %v", err)
	}
	competingNumber := fmt.Sprintf("INV-%d-%04d", year, sequence+1)
	res, err := tx.ExecContext(ctx,
		`INSERT INTO finance_invoices
		   (owner_id, business_id, invoice_number, client_name, status, issue_date, due_date)
		 VALUES (?, ?, ?, ?, 'draft', CURDATE(), CURDATE())`,
		testutil.OwnerID, testutil.OwnerID, competingNumber, "Competing Conversion",
	)
	if err != nil {
		t.Fatalf("insert competing invoice %s: %v", competingNumber, err)
	}
	competingID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("competing invoice LastInsertId: %v", err)
	}
	deleteInvoice(t, db, ctx, competingID)
	if err = tx.Commit(); err != nil {
		t.Fatalf("commit competing conversion: %v", err)
	}

	got := <-done
	if got.err != nil {
		t.Fatalf("ConvertToInvoice after competing commit: %v", got.err)
	}
	deleteInvoice(t, db, ctx, got.invoiceID)

	var number string
	if err = db.QueryRowContext(ctx,
		"SELECT invoice_number FROM finance_invoices WHERE id = ? AND business_id = ?",
		got.invoiceID, testutil.OwnerID,
	).Scan(&number); err != nil {
		t.Fatalf("read converted invoice: %v", err)
	}
	if number == competingNumber {
		t.Fatalf("conversion minted %q, already taken by the competing conversion", number)
	}
}
