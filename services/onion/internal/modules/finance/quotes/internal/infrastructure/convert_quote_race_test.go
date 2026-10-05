package infrastructure_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/finance/invoices"
	"github.com/voxire/onion/internal/modules/finance/quotes/internal/domain"
	"github.com/voxire/onion/internal/modules/finance/quotes/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	financeinvoicesv1 "github.com/voxire/voxie/gen/go/finance/invoices/v1"
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

	prefix := fmt.Sprintf("INV-%d-", year)
	if _, err := tx.ExecContext(ctx, "INSERT INTO finance_invoice_sequences (business_id, prefix, last_sequence) VALUES (?, ?, 0) ON DUPLICATE KEY UPDATE last_sequence = last_sequence", testutil.OwnerID, prefix); err != nil {
		t.Fatal(err)
	}
	var sequence int64
	if err := tx.QueryRowContext(ctx, "SELECT last_sequence FROM finance_invoice_sequences WHERE business_id = ? AND prefix = ? FOR UPDATE", testutil.OwnerID, prefix).Scan(&sequence); err != nil {
		t.Fatal(err)
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
		t.Fatalf("conversion raced past the held invoice sequence lock: invoice=%d err=%v", res.invoiceID, res.err)
	case <-time.After(300 * time.Millisecond):
	}

	// The competing conversion takes the next number and commits it.
	var existing int64
	if err := tx.QueryRowContext(ctx, "SELECT COALESCE(MAX(CAST(SUBSTRING_INDEX(invoice_number, '-', -1) AS UNSIGNED)), 0) FROM finance_invoices WHERE business_id = ? AND invoice_number LIKE CONCAT(?, '%')", testutil.OwnerID, prefix).Scan(&existing); err != nil {
		t.Fatal(err)
	}
	if existing > sequence {
		sequence = existing
	}
	competingNumber := fmt.Sprintf("%s%04d", prefix, sequence+1)
	if _, err := tx.ExecContext(ctx, "UPDATE finance_invoice_sequences SET last_sequence = ? WHERE business_id = ? AND prefix = ?", sequence+1, testutil.OwnerID, prefix); err != nil {
		t.Fatal(err)
	}

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

func TestInvoiceCreationAndQuoteConversionShareNumberAllocation(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	quotes := infrastructure.NewRepository(db)
	mux := http.NewServeMux()
	invoices.NewModule(ctx, db).RegisterAPI(mux)
	invoiceCtx := auth.WithAuthUser(ctx, &auth.AuthUser{UserID: testutil.OwnerID, BusinessID: testutil.OwnerID,
		EffectiveAccess: &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{AppID: authorization.AppFinance, Licensed: true, HasAccess: true, Permissions: []string{"finance:invoices:create"}}}},
	})
	const writes = 8
	quoteIDs := make([]int64, writes/2)
	for i := range quoteIDs {
		quoteIDs[i] = createConvertibleQuote(t, db, quotes, ctx, fmt.Sprintf("Mixed Writer %d", i))
	}
	type result struct {
		id  int64
		err error
	}
	results := make(chan result, writes)
	start := make(chan struct{})
	for _, quoteID := range quoteIDs {
		go func(id int64) {
			<-start
			invoiceID, err := quotes.ConvertToInvoice(ctx, testutil.OwnerID, id)
			results <- result{invoiceID, err}
		}(quoteID)
	}
	type createdInvoice struct {
		ID            int64
		InvoiceNumber string
	}
	createInvoice := func() (createdInvoice, error) {
		date := time.Now().UTC().Format(time.DateOnly)
		body, err := json.Marshal(&financeinvoicesv1.CreateInvoiceRequest{ClientName: "Synthetic Mixed Writer", IssueDate: date, DueDate: date,
			Currency: "USD", DiscountType: "fixed", Items: []*financeinvoicesv1.InvoiceItem{{Description: "Service", Quantity: 1, UnitPrice: 100}},
		})
		if err != nil {
			return createdInvoice{}, err
		}
		request := httptest.NewRequest(http.MethodPost, "/finance/invoices/v1/invoices", bytes.NewReader(body)).WithContext(invoiceCtx)
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			return createdInvoice{}, fmt.Errorf("create invoice: status %d, %s", response.Code, response.Body.String())
		}
		var decoded financeinvoicesv1.CreateInvoiceResponse
		if err := json.Unmarshal(response.Body.Bytes(), &decoded); err != nil {
			return createdInvoice{}, err
		}
		return createdInvoice{decoded.Invoice.Id, decoded.Invoice.InvoiceNumber}, nil
	}

	for range writes / 2 {
		go func() { <-start; invoice, err := createInvoice(); results <- result{invoice.ID, err} }()
	}
	close(start)
	numbers := make(map[string]int64, writes)
	var lastID int64
	for range writes {
		res := <-results
		if res.err != nil {
			t.Errorf("concurrent invoice write: %v", res.err)
			continue
		}
		deleteInvoice(t, db, ctx, res.id)
		var number string
		if err := db.QueryRow("SELECT invoice_number FROM finance_invoices WHERE id = ?", res.id).Scan(&number); err != nil {
			t.Fatal(err)
		}
		if first, duplicate := numbers[number]; duplicate {
			t.Fatalf("invoice number %s reused for %d and %d", number, first, res.id)
		}
		numbers[number] = res.id
		lastID = res.id
	}
	if len(numbers) != writes {
		t.Fatalf("created %d distinct invoices, want %d", len(numbers), writes)
	}
	if _, err := db.Exec("DELETE FROM finance_invoice_items WHERE invoice_id = ?", lastID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("DELETE FROM finance_invoices WHERE id = ?", lastID); err != nil {
		t.Fatal(err)
	}
	next, err := createInvoice()
	if err != nil {
		t.Fatal(err)
	}
	deleteInvoice(t, db, ctx, next.ID)
	if _, reused := numbers[next.InvoiceNumber]; reused {
		t.Fatalf("deleting an invoice recycled %s", next.InvoiceNumber)
	}
}
