package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/products/internal/domain"
	"github.com/voxire/onion/internal/modules/products/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestUpdateProduct_ExplicitEmptyImageRemovesPrimaryImage(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateProduct(ctx, testutil.OwnerID, domain.ProductInput{
		Name:     "Image clear regression",
		ImageURL: "https://cdn.example/regression.png",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateProduct: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM erp_product_images WHERE business_id = ? AND product_id = ?`, testutil.OwnerID, id)
		_ = repo.DeleteProduct(ctx, testutil.OwnerID, id)
	})

	if err := repo.UpdateProduct(ctx, testutil.OwnerID, id, domain.ProductInput{
		Name:     "Image clear regression",
		IsActive: true,
	}); err != nil {
		t.Fatalf("UpdateProduct without image: %v", err)
	}

	var count int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM erp_product_images WHERE business_id = ? AND product_id = ? AND is_primary = TRUE`,
		testutil.OwnerID, id,
	).Scan(&count); err != nil {
		t.Fatalf("count primary images after omitted update: %v", err)
	}
	if count != 1 {
		t.Fatalf("primary image count after omitted update = %d, want 1", count)
	}

	if err := repo.UpdateProduct(ctx, testutil.OwnerID, id, domain.ProductInput{
		Name:        "Image clear regression",
		ImageURLSet: true,
		IsActive:    true,
	}); err != nil {
		t.Fatalf("UpdateProduct with image clear: %v", err)
	}

	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM erp_product_images WHERE business_id = ? AND product_id = ? AND is_primary = TRUE`,
		testutil.OwnerID, id,
	).Scan(&count); err != nil {
		t.Fatalf("count primary images after explicit clear: %v", err)
	}
	if count != 0 {
		t.Fatalf("primary image count after explicit clear = %d, want 0", count)
	}
}
