package api

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/erp/products/internal/domain"
	erpproductv1 "github.com/voxire/voxie/gen/go/erp/product/v1"
)

func TestProductUpdateInputRetainsExplicitZeroFalseAndImageClear(t *testing.T) {
	body := &erpproductv1.UpdateProductBody{
		Name:          "Product",
		StockQuantity: 0,
		IsActive:      false,
		ImageUrl:      "",
	}
	present := map[string]json.RawMessage{
		domain.ProductFieldName:          json.RawMessage(`"Product"`),
		domain.ProductFieldStockQuantity: json.RawMessage(`0`),
		domain.ProductFieldIsActive:      json.RawMessage(`false`),
		domain.ProductFieldImageURL:      json.RawMessage(`""`),
	}

	input := productUpdateInput(body, present)

	if !input.StockQuantitySet {
		t.Fatal("explicit stock_quantity=0 was treated as omitted")
	}
	if !input.ImageURLSet {
		t.Fatal("explicit empty image_url was treated as omitted")
	}
	if !input.FieldPresent(domain.ProductFieldIsActive) {
		t.Fatal("explicit is_active=false was treated as omitted")
	}
}

func TestProductUpdateInputTreatsOmittedStockAsAbsent(t *testing.T) {
	input := productUpdateInput(&erpproductv1.UpdateProductBody{Name: "Product"}, map[string]json.RawMessage{
		domain.ProductFieldName: json.RawMessage(`"Product"`),
	})

	if input.StockQuantitySet {
		t.Fatal("omitted stock_quantity was marked as an update")
	}
}

func TestDecodeProductUpdateBodyRetainsJSONPresence(t *testing.T) {
	req := httptest.NewRequest("PUT", "/erp/product/v1/products/1", strings.NewReader(`{"name":"Product","stock_quantity":0,"is_active":false,"image_url":""}`))
	req.Header.Set("Content-Type", "application/json")

	body, present, err := decodeProductUpdateBody(httptest.NewRecorder(), req)
	if err != nil {
		t.Fatalf("decodeProductUpdateBody: %v", err)
	}
	if body.GetStockQuantity() != 0 || body.GetIsActive() || body.GetImageUrl() != "" {
		t.Fatalf("decoded values = stock:%v active:%v image:%q", body.GetStockQuantity(), body.GetIsActive(), body.GetImageUrl())
	}
	for _, field := range []string{
		domain.ProductFieldStockQuantity,
		domain.ProductFieldIsActive,
		domain.ProductFieldImageURL,
	} {
		if _, ok := present[field]; !ok {
			t.Errorf("presence map omitted %q", field)
		}
	}
}
