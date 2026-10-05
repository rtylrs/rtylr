package domain

// SupplierStatus represents the active state as a string label.
type SupplierStatus string

const (
	SupplierStatusActive   SupplierStatus = "active"
	SupplierStatusInactive SupplierStatus = "inactive"
)

// Status returns the human-readable status of the supplier.
func (s Supplier) Status() SupplierStatus {
	if s.IsActive {
		return SupplierStatusActive
	}
	return SupplierStatusInactive
}
