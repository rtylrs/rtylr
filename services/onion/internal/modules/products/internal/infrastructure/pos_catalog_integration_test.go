package infrastructure_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/pos_menu"
	"github.com/voxire/onion/internal/modules/products/internal/domain"
	"github.com/voxire/onion/internal/modules/products/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	posmenuitemv1 "github.com/voxire/voxie/gen/go/pos/menu_item/v1"
)

func TestERPProductLifecycleAppearsInPOSCatalogForPOSOnlyMember(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := infrastructure.NewRepository(db)
	input := domain.ProductInput{
		Name: "Cross-app product", SKU: fmt.Sprintf("CROSS-%d", time.Now().UnixNano()),
		CategoryID: 5002, CostPrice: 3.25, SalePrice: 8.5, StockQuantity: 12,
		IsActive: true, TrackInventory: true, TaxRate: 10,
		ImageURL: "https://example.test/product.png",
	}
	id, err := repo.CreateProduct(ctx, testutil.BusinessID, input)
	if err != nil {
		t.Fatalf("create ERP product: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM erp_product_images WHERE business_id = ? AND product_id = ?", testutil.BusinessID, id)
		if err := repo.DeleteProduct(ctx, testutil.BusinessID, id); err != nil {
			t.Errorf("clean up ERP product: %v", err)
		}
	})

	mux := http.NewServeMux()
	pos_menu.NewModule(ctx, db).RegisterAPI(mux)
	user := &auth.AuthUser{
		UserID: 1002, BusinessID: testutil.BusinessID,
		EffectiveAccess: &authorization.EffectiveAccess{Apps: []authorization.EffectiveAppAccess{{
			AppID: authorization.AppPOS, Licensed: true, HasAccess: true,
			Permissions: []string{"pos:orders:create"},
			PermissionScopes: map[string][]authorization.Scope{
				"pos:orders:create": {{Type: "pos_location", ID: "3001"}},
			},
		}}},
	}
	read := func(user *auth.AuthUser, path string) (int, []*posmenuitemv1.MenuItem) {
		t.Helper()
		request := httptest.NewRequest(http.MethodGet, path, nil).WithContext(auth.WithAuthUser(ctx, user))
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			return response.Code, nil
		}
		var payload posmenuitemv1.ListMenuItemsResponse
		if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode catalog: %v", err)
		}
		return response.Code, payload.Items
	}
	assertItem := func(wantName string, wantPrice, wantStock float64) {
		t.Helper()
		status, items := read(user, "/pos/menu-item/v1/catalog")
		if status != http.StatusOK {
			t.Fatalf("POS-only member catalog status = %d, want 200", status)
		}
		for _, item := range items {
			if item.GetId() != id {
				continue
			}
			if item.GetProductId() != id || item.GetName() != wantName || item.GetSalePrice() != wantPrice || item.GetStock() != wantStock || item.GetImageUrl() != input.ImageURL || item.GetCategory() != "Mains" || item.GetTaxRate() != 10 {
				t.Fatalf("POS catalog differs from ERP product: %+v", item)
			}
			if item.UnitCost != nil {
				t.Fatal("POS-only operator can see ERP purchase cost")
			}
			return
		}
		t.Fatalf("ERP product %d missing from POS catalog", id)
	}
	assertItem(input.Name, input.SalePrice, input.StockQuantity)

	input.Name, input.SalePrice, input.StockQuantity = "Updated cross-app product", 9.75, 7
	input.StockQuantitySet = true
	if err := repo.UpdateProduct(ctx, testutil.BusinessID, id, input); err != nil {
		t.Fatalf("update ERP product: %v", err)
	}
	assertItem(input.Name, input.SalePrice, input.StockQuantity)

	input.IsActive = false
	if err := repo.UpdateProduct(ctx, testutil.BusinessID, id, input); err != nil {
		t.Fatalf("deactivate ERP product: %v", err)
	}
	status, items := read(user, "/pos/menu-item/v1/catalog")
	if status != http.StatusOK {
		t.Fatalf("catalog status after deactivation = %d", status)
	}
	for _, item := range items {
		if item.GetId() == id {
			t.Fatal("deactivated ERP product remains sellable in POS")
		}
	}
	if status, _ := read(user, "/pos/menu-item/v1/menu-items"); status != http.StatusForbidden {
		t.Fatalf("POS-only member gained Menu administration access: status %d", status)
	}
	user.BusinessID = 999999
	if status, _ := read(user, "/pos/menu-item/v1/catalog"); status != http.StatusForbidden {
		t.Fatalf("unrelated business catalog status = %d, want 403", status)
	}
}
