package domain

import "testing"

func TestProductWithVariants_HasVariants_True(t *testing.T) {
	pw := ProductWithVariants{VariantCount: 3}
	if !pw.HasVariants() {
		t.Error("expected true when variant count > 0")
	}
}

func TestProductWithVariants_HasVariants_False(t *testing.T) {
	pw := ProductWithVariants{VariantCount: 0}
	if pw.HasVariants() {
		t.Error("expected false when variant count == 0")
	}
}

func TestProductWithVariants_ProductFields(t *testing.T) {
	pw := ProductWithVariants{
		Product:      Product{ID: 42, Name: "Burger"},
		VariantCount: 2,
		VariantNames: []string{"Small", "Large"},
	}
	if pw.Product.ID != 42 {
		t.Errorf("expected product ID=42, got %d", pw.Product.ID)
	}
	if pw.Product.Name != "Burger" {
		t.Errorf("expected product name='Burger', got %q", pw.Product.Name)
	}
	if pw.VariantCount != 2 {
		t.Errorf("expected variant count=2, got %d", pw.VariantCount)
	}
	if len(pw.VariantNames) != 2 {
		t.Errorf("expected 2 variant names, got %d", len(pw.VariantNames))
	}
}
