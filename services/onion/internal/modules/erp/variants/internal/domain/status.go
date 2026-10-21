package domain

// VariantStatus represents the active state as a string label.
type VariantStatus string

const (
	VariantStatusActive   VariantStatus = "active"
	VariantStatusInactive VariantStatus = "inactive"
)

// Status returns the human-readable status of the variant.
func (v Variant) Status() VariantStatus {
	if v.IsActive {
		return VariantStatusActive
	}
	return VariantStatusInactive
}
