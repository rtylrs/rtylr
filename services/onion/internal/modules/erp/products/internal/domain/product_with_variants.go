package domain

// ProductWithVariants bundles a Product with its list of variant names.
// This is a lightweight projection for summary screens that need to show
// the variant count without fetching full variant objects.
type ProductWithVariants struct {
	Product      Product
	VariantCount int
	VariantNames []string
}

// HasVariants reports whether the product has at least one variant.
func (p ProductWithVariants) HasVariants() bool {
	return p.VariantCount > 0
}
