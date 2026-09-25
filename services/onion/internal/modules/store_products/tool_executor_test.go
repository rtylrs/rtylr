package store_products

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/store_products/internal/domain"
)

func TestParseListStoreProductsToolInput(t *testing.T) {
	parsed, err := parseListStoreProductsToolInput(map[string]any{
		"search":     "whey",
		"product_id": json.Number("4"),
		"published":  "published",
		"page":       2,
		"limit":      "10",
	})
	if err != nil {
		t.Fatalf("parseListStoreProductsToolInput: %v", err)
	}
	if parsed.Query != "whey" {
		t.Fatalf("Query = %q", parsed.Query)
	}
	if parsed.ProductID != 4 {
		t.Fatalf("ProductID = %d", parsed.ProductID)
	}
	if parsed.Published == nil || !*parsed.Published {
		t.Fatalf("Published = %#v, want true", parsed.Published)
	}
	if parsed.Page != 2 || parsed.Limit != 10 {
		t.Fatalf("page/limit = %d/%d", parsed.Page, parsed.Limit)
	}
}

func TestParseListStoreProductsToolInputDefaultsAndCapsLimit(t *testing.T) {
	parsed, err := parseListStoreProductsToolInput(map[string]any{
		"limit": 500,
	})
	if err != nil {
		t.Fatalf("parseListStoreProductsToolInput: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.Limit != maxListStoreProductsToolLimit {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
}

func TestParseListStoreProductsToolInputRejectsBadPublished(t *testing.T) {
	_, err := parseListStoreProductsToolInput(map[string]any{
		"published": "sometimes",
	})
	if err == nil || !strings.Contains(err.Error(), "published") {
		t.Fatalf("error = %v, want published validation", err)
	}
}

func TestParseGetStoreProductToolInputAcceptsAliases(t *testing.T) {
	cases := []map[string]any{
		{"store_product_id": int64(7)},
		{"storefront_product_id": json.Number("7")},
		{"id": "#7"},
		{"id": "store-product-7"},
		{"id": "store_product_7"},
		{"product_id": "product-7"},
		{"slug": "whey-protein"},
		{"title": "Whey Protein"},
	}

	for _, input := range cases {
		parsed, err := parseGetStoreProductToolInput(input)
		if err != nil {
			t.Fatalf("parseGetStoreProductToolInput(%v) returned error: %v", input, err)
		}
		if parsed.StoreProductID == 0 && parsed.ProductID == 0 && parsed.Slug == "" && parsed.Title == "" {
			t.Fatalf("parsed empty input for %v: %+v", input, parsed)
		}
	}
}

func TestParseGetStoreProductToolInputRejectsMissingLookup(t *testing.T) {
	_, err := parseGetStoreProductToolInput(map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "store_product_id") {
		t.Fatalf("error = %v, want lookup validation", err)
	}
}

func TestFilterStoreProductsForTool(t *testing.T) {
	published := true
	products := []domain.StoreProduct{
		{ID: 1, ProductID: 4, Title: "Whey Protein", IsPublished: true},
		{ID: 2, ProductID: 4, Title: "Whey Old", IsPublished: false},
		{ID: 3, ProductID: 5, Title: "Whey Isolate", IsPublished: true},
		{ID: 4, ProductID: 4, Title: "Creatine", IsPublished: true},
	}

	filtered := filterStoreProductsForTool(products, listStoreProductsToolInput{
		Query:     "whey",
		ProductID: 4,
		Published: &published,
	})
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d", len(filtered))
	}
	if filtered[0].ID != 1 {
		t.Fatalf("filtered[0].ID = %d", filtered[0].ID)
	}
}

func TestFindStoreProductForTool(t *testing.T) {
	products := []domain.StoreProduct{
		{ID: 1, ProductID: 10, Title: "Whey Protein", Slug: "whey-protein"},
		{ID: 2, ProductID: 11, Title: "Creatine", Slug: "creatine"},
	}

	cases := []struct {
		name  string
		input getStoreProductToolInput
		want  int64
	}{
		{name: "store product id", input: getStoreProductToolInput{StoreProductID: 2}, want: 2},
		{name: "inventory product id", input: getStoreProductToolInput{ProductID: 10}, want: 1},
		{name: "slug", input: getStoreProductToolInput{Slug: "CREATINE"}, want: 2},
		{name: "title", input: getStoreProductToolInput{Title: "whey protein"}, want: 1},
	}

	for _, tc := range cases {
		got, err := findStoreProductForTool(products, tc.input)
		if err != nil {
			t.Fatalf("%s: findStoreProductForTool returned error: %v", tc.name, err)
		}
		if got.ID != tc.want {
			t.Fatalf("%s: ID = %d, want %d", tc.name, got.ID, tc.want)
		}
	}
}

func TestFindStoreProductForToolNotFound(t *testing.T) {
	_, err := findStoreProductForTool([]domain.StoreProduct{{ID: 1}}, getStoreProductToolInput{StoreProductID: 99})
	if err == nil {
		t.Fatal("expected not found error")
	}
}

func TestStoreProductToolPayload(t *testing.T) {
	createdAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := storeProductToolPayload(domain.StoreProduct{
		ID:             7,
		OwnerID:        2,
		ProductID:      4,
		Title:          "Whey",
		Slug:           "whey",
		Price:          15,
		CompareAtPrice: 20,
		IsPublished:    true,
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
	})

	if payload["price"] != float64(15) {
		t.Fatalf("price = %#v", payload["price"])
	}
	if payload["created_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("created_at = %#v", payload["created_at"])
	}
}
