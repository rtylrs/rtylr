package infrastructure_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/voxire/onion/internal/modules/erp/products/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/products/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/testutil"
)

func TestProductMutationsRollBackWhenFactCannotBeRecorded(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := auth.WithAuthUser(context.Background(), &auth.AuthUser{UserID: 7654321, IP: "203.0.113.7"})
	businessID := time.Now().UnixNano() % 1_000_000_000_000
	repo := infrastructure.NewRepository(db)
	trigger := "fail_fact_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	t.Cleanup(func() {
		_, _ = db.Exec("DROP TRIGGER IF EXISTS " + trigger)
		_, _ = db.Exec("DELETE FROM erp_product_attributes WHERE business_id = ?", businessID)
		_, _ = db.Exec("DELETE FROM erp_products WHERE business_id = ?", businessID)
		_, _ = db.Exec("DELETE FROM erp_outbox WHERE business_id = ?", businessID)
	})
	input := domain.ProductInput{Name: "Original", SKU: uuid.NewString(), SalePrice: 12.5, IsActive: true}
	id, err := repo.CreateProduct(ctx, businessID, input)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.ReplaceProductAttributes(ctx, businessID, id, []domain.ProductAttributeInput{{Name: "Color", Value: "Blue"}}); err != nil {
		t.Fatal(err)
	}
	var before int
	if err := db.QueryRow("SELECT COUNT(*) FROM erp_outbox WHERE business_id = ?", businessID).Scan(&before); err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(fmt.Sprintf("CREATE TRIGGER %s BEFORE INSERT ON erp_outbox FOR EACH ROW BEGIN IF NEW.business_id = %d THEN SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'synthetic fact failure'; END IF; END", trigger, businessID))
	if err != nil {
		t.Fatal(err)
	}
	failedCreate := input
	failedCreate.SKU = uuid.NewString()
	if _, err := repo.CreateProduct(ctx, businessID, failedCreate); err == nil {
		t.Fatal("create succeeded despite failed fact")
	}
	update := domain.ProductInput{Name: "Changed", PresentFields: map[string]bool{"name": true}}
	if err := repo.UpdateProduct(ctx, businessID, id, update); err == nil {
		t.Fatal("update succeeded despite failed fact")
	}
	if err := repo.ReplaceProductAttributes(ctx, businessID, id, []domain.ProductAttributeInput{{Name: "Color", Value: "Red"}}); err == nil {
		t.Fatal("attribute replacement succeeded despite failed fact")
	}
	if err := repo.DeleteProduct(ctx, businessID, id); err == nil {
		t.Fatal("delete succeeded despite failed fact")
	}
	p, err := repo.GetProduct(ctx, businessID, id)
	if err != nil || p.Name != "Original" || p.SalePrice != 12.5 {
		t.Fatalf("product changed after rollback: %+v, %v", p, err)
	}
	attrs, err := repo.ListProductAttributes(ctx, businessID, id)
	if err != nil || len(attrs) != 1 || attrs[0].Value != "Blue" {
		t.Fatalf("attributes changed after rollback: %+v, %v", attrs, err)
	}
	var products, facts int
	if err := db.QueryRow("SELECT COUNT(*) FROM erp_products WHERE business_id = ?", businessID).Scan(&products); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow("SELECT COUNT(*) FROM erp_outbox WHERE business_id = ?", businessID).Scan(&facts); err != nil {
		t.Fatal(err)
	}
	if products != 1 || facts != before {
		t.Fatalf("failed mutation escaped rollback: products=%d facts=%d, want 1/%d", products, facts, before)
	}
	if _, err := db.Exec("DROP TRIGGER " + trigger); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpdateProduct(ctx, businessID, id, update); err != nil {
		t.Fatal(err)
	}
	p, err = repo.GetProduct(ctx, businessID, id)
	if err != nil || p.Name != "Changed" || p.SalePrice != 12.5 || p.SKU != input.SKU {
		t.Fatalf("partial update lost existing fields: %+v, %v", p, err)
	}
	var raw []byte
	if err := db.QueryRow("SELECT payload FROM erp_outbox WHERE business_id = ? AND subject_type = 'product' AND subject_id = ? ORDER BY id DESC LIMIT 1", businessID, id).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var fact struct {
		ActorID int64  `json:"actor_id"`
		IP      string `json:"ip_address"`
		Old     struct {
			Name string `json:"name"`
		} `json:"old_values"`
		New struct {
			Name      string  `json:"name"`
			SalePrice float64 `json:"sale_price"`
		} `json:"new_values"`
	}
	if err := json.Unmarshal(raw, &fact); err != nil {
		t.Fatal(err)
	}
	if fact.ActorID != 7654321 || fact.IP != "203.0.113.7" || fact.Old.Name != "Original" || fact.New.Name != "Changed" || fact.New.SalePrice != 12.5 {
		t.Fatalf("incorrect committed mutation fact: %+v", fact)
	}
}
