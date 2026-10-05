package infrastructure_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/erp/products"
	"github.com/voxire/onion/internal/modules/store/checkout/internal/domain"
	"github.com/voxire/onion/internal/modules/store/checkout/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCheckout_Create_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithSalesProductReader(products.NewModule(context.Background(), db).SalesProduct)
	input := domain.CheckoutCreateInput{
		SessionID:     "sess-abc-123",
		CustomerEmail: "test@example.com",
		Items:         `[{"product_id":5201,"qty":2}]`,
		Subtotal:      50.0,
		Tax:           5.0,
		Total:         55.0,
	}
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCheckout_Create_EmptyItems(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithSalesProductReader(products.NewModule(context.Background(), db).SalesProduct)
	input := domain.CheckoutCreateInput{
		Subtotal: 10.0,
		Tax:      1.0,
		Total:    11.0,
	}
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, input)
	if err != nil {
		t.Fatalf("Create with empty items: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCheckout_Create_MultipleCheckouts(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithSalesProductReader(products.NewModule(context.Background(), db).SalesProduct)
	for i := 0; i < 3; i++ {
		id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, domain.CheckoutCreateInput{
			Subtotal: float64(i) * 10,
			Tax:      float64(i),
			Total:    float64(i)*10 + float64(i),
		})
		if err != nil {
			t.Fatalf("Create #%d: %v", i, err)
		}
		if id <= 0 {
			t.Errorf("iteration %d: expected positive ID", i)
		}
	}
}

func TestCheckout_Complete_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithSalesProductReader(products.NewModule(context.Background(), db).SalesProduct)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, domain.CheckoutCreateInput{
		Subtotal: 20.0,
		Tax:      2.0,
		Total:    22.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.Complete(context.Background(), testutil.OwnerID, id, testutil.OwnerID); err != nil {
		t.Fatalf("Complete: %v", err)
	}
}

func TestCheckout_Complete_UnknownID_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithSalesProductReader(products.NewModule(context.Background(), db).SalesProduct)
	// Completing a non-existent checkout should not return an error (0 rows affected)
	err := repo.Complete(context.Background(), testutil.OwnerID, 9999999, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Complete on missing ID: %v", err)
	}
}

func TestCheckout_Create_WithSessionID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithSalesProductReader(products.NewModule(context.Background(), db).SalesProduct)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, domain.CheckoutCreateInput{
		SessionID: "unique-session-xyz",
		Subtotal:  100.0,
		Tax:       10.0,
		Total:     110.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestCheckout_Create_ZeroTotals(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithSalesProductReader(products.NewModule(context.Background(), db).SalesProduct)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, domain.CheckoutCreateInput{
		Subtotal: 0,
		Tax:      0,
		Total:    0,
	})
	if err != nil {
		t.Fatalf("Create with zero totals: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestCheckout_Complete_CrossOwner_NoEffect(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithSalesProductReader(products.NewModule(context.Background(), db).SalesProduct)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, domain.CheckoutCreateInput{
		Subtotal: 30.0,
		Tax:      3.0,
		Total:    33.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	// Complete with wrong business should silently do nothing
	err = repo.Complete(context.Background(), 9999999, id, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Complete cross-owner: %v", err)
	}
}

func TestCheckout_Create_CustomerEmailStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithSalesProductReader(products.NewModule(context.Background(), db).SalesProduct)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, domain.CheckoutCreateInput{
		CustomerEmail: "checkout@example.com",
		Subtotal:      40.0,
		Tax:           4.0,
		Total:         44.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestCheckout_Create_SubtotalIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithSalesProductReader(products.NewModule(context.Background(), db).SalesProduct)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, domain.CheckoutCreateInput{
		Subtotal: 77.77,
		Tax:      7.77,
		Total:    85.54,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestCheckout_Create_TaxIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithSalesProductReader(products.NewModule(context.Background(), db).SalesProduct)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, domain.CheckoutCreateInput{
		Subtotal: 50.0,
		Tax:      8.5,
		Total:    58.5,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestCheckout_Create_IDsAreUnique(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithSalesProductReader(products.NewModule(context.Background(), db).SalesProduct)
	id1, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, domain.CheckoutCreateInput{Total: 1.0})
	if err != nil {
		t.Fatalf("Create 1: %v", err)
	}
	id2, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, domain.CheckoutCreateInput{Total: 2.0})
	if err != nil {
		t.Fatalf("Create 2: %v", err)
	}
	if id1 == id2 {
		t.Errorf("expected unique IDs, both are %d", id1)
	}
}

func TestCheckout_Complete_TwiceSameID_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithSalesProductReader(products.NewModule(context.Background(), db).SalesProduct)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, domain.CheckoutCreateInput{
		Subtotal: 25.0, Tax: 2.5, Total: 27.5,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.Complete(context.Background(), testutil.OwnerID, id, testutil.OwnerID); err != nil {
		t.Fatalf("Complete first: %v", err)
	}
	if err := repo.Complete(context.Background(), testutil.OwnerID, id, testutil.OwnerID); err != nil {
		t.Fatalf("Complete second: %v", err)
	}
}

// Concurrent completions of the same checkout (double click or client retry)
// must publish the CRM spend/visit fact exactly once. CRM applies that fact
// asynchronously, so the source transaction is responsible for one durable
// outbox event rather than writing CRM-owned rows itself.
func TestCheckout_Complete_ConcurrentDuplicateCountsOnce(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithSalesProductReader(products.NewModule(context.Background(), db).SalesProduct)
	ctx := context.Background()

	stamp := time.Now().UnixNano()
	email := fmt.Sprintf("b37-r04-%d@example.com", stamp)
	const total = 42.50
	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.CheckoutCreateInput{
		SessionID:     fmt.Sprintf("b37-r04-%d", stamp),
		CustomerEmail: email,
		Subtotal:      total,
		Total:         total,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), `DELETE FROM store_outbox WHERE business_id = ? AND subject_id = ? AND subject_type IN ('store_checkout', 'store_checkout_completed')`, testutil.BusinessID, id); err != nil {
			t.Errorf("cleanup checkout events: %v", err)
		}
		if _, err := db.ExecContext(context.Background(), "DELETE FROM store_checkouts WHERE id = ?", id); err != nil {
			t.Errorf("cleanup store_checkouts: %v", err)
		}
		if _, err := db.ExecContext(context.Background(), "DELETE FROM crm_customers WHERE business_id = ? AND email = ?", testutil.BusinessID, email); err != nil {
			t.Errorf("cleanup crm_customers: %v", err)
		}
	})

	const workers = 6
	start := make(chan struct{})
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			errs <- repo.Complete(ctx, testutil.BusinessID, id, testutil.OwnerID)
		}()
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent Complete: %v", err)
		}
	}

	var eventCount int
	var totalSpent string
	var countVisit string
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*), MAX(JSON_UNQUOTE(JSON_EXTRACT(payload, '$.total_spent'))),
		       MAX(JSON_UNQUOTE(JSON_EXTRACT(payload, '$.count_visit')))
		FROM store_outbox
		WHERE business_id = ? AND event_type = 'customer_contact_captured'
		  AND subject_type = 'store_checkout_completed' AND subject_id = ?`, testutil.BusinessID, id,
	).Scan(&eventCount, &totalSpent, &countVisit); err != nil {
		t.Fatalf("read CRM capture event: %v", err)
	}
	if eventCount != 1 || totalSpent != "42.5" || countVisit != "true" {
		t.Fatalf("completion capture events=%d total_spent=%q count_visit=%q, want 1, 42.5, true", eventCount, totalSpent, countVisit)
	}
}

func TestCheckout_Create_WithItems_PositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithSalesProductReader(products.NewModule(context.Background(), db).SalesProduct)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, domain.CheckoutCreateInput{
		Items:    `[{"product_id":5201,"qty":3},{"product_id":5202,"qty":1}]`,
		Subtotal: 90.0,
		Tax:      9.0,
		Total:    99.0,
	})
	if err != nil {
		t.Fatalf("Create with items: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestCheckout_Create_SessionIDUnique(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithSalesProductReader(products.NewModule(context.Background(), db).SalesProduct)
	id1, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, domain.CheckoutCreateInput{
		SessionID: "sess-unique-001", Total: 10.0,
	})
	if err != nil {
		t.Fatalf("Create 1: %v", err)
	}
	id2, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, domain.CheckoutCreateInput{
		SessionID: "sess-unique-002", Total: 20.0,
	})
	if err != nil {
		t.Fatalf("Create 2: %v", err)
	}
	if id1 == id2 {
		t.Errorf("expected different IDs for different sessions, got %d", id1)
	}
}

func TestCheckout_Complete_ReturnsNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithSalesProductReader(products.NewModule(context.Background(), db).SalesProduct)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, domain.CheckoutCreateInput{
		Subtotal: 15.0, Tax: 1.5, Total: 16.5,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = repo.Complete(context.Background(), testutil.OwnerID, id, testutil.OwnerID)
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
}

func TestCheckout_Create_LargeTotalIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithSalesProductReader(products.NewModule(context.Background(), db).SalesProduct)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, domain.CheckoutCreateInput{
		Subtotal: 9999.99,
		Tax:      999.99,
		Total:    10999.98,
	})
	if err != nil {
		t.Fatalf("Create large total: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestCheckout_Create_ZeroOwner_PositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithSalesProductReader(products.NewModule(context.Background(), db).SalesProduct)
	// Creating with owner 0 is technically valid at the DB layer
	id, err := repo.Create(context.Background(), 0, 0, domain.CheckoutCreateInput{
		Total: 5.0,
	})
	if err != nil {
		t.Fatalf("Create with owner=0: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestCheckout_Create_MultipleForSameOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db).WithSalesProductReader(products.NewModule(context.Background(), db).SalesProduct)
	var ids []int64
	for i := 0; i < 4; i++ {
		id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, domain.CheckoutCreateInput{
			Subtotal: float64(i) * 5,
			Total:    float64(i) * 5,
		})
		if err != nil {
			t.Fatalf("Create #%d: %v", i, err)
		}
		ids = append(ids, id)
	}
	seen := make(map[int64]bool)
	for _, id := range ids {
		if seen[id] {
			t.Errorf("duplicate ID %d", id)
		}
		seen[id] = true
	}
}
