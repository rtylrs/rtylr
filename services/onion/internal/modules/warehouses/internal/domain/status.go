package domain

// WarehouseStatus represents the active state as a string label.
type WarehouseStatus string

const (
	WarehouseStatusActive   WarehouseStatus = "active"
	WarehouseStatusInactive WarehouseStatus = "inactive"
)

// Status returns the human-readable status of the warehouse.
func (w Warehouse) Status() WarehouseStatus {
	if w.IsActive {
		return WarehouseStatusActive
	}
	return WarehouseStatusInactive
}
