package domain

// ProductStatus represents the active/inactive state as a string label.
type ProductStatus string

const (
	ProductStatusActive   ProductStatus = "active"
	ProductStatusInactive ProductStatus = "inactive"
)

// Status returns a human-readable status string for the product.
func (p Product) Status() ProductStatus {
	if p.IsActive {
		return ProductStatusActive
	}
	return ProductStatusInactive
}
