package domain

import "fmt"

// ValidItemStatuses lists all valid KDS item status values.
var ValidItemStatuses = []string{"pending", "preparing", "ready", "served", "cancelled"}

// ValidateUpdateItemStatusInput returns an error when the input is malformed.
func ValidateUpdateItemStatusInput(input UpdateItemStatusInput) error {
	if input.OwnerID == 0 {
		return fmt.Errorf("owner id is required")
	}
	if input.OrderID == 0 {
		return fmt.Errorf("order id is required")
	}
	if input.ItemID == 0 {
		return fmt.Errorf("item id is required")
	}
	if input.Status == "" {
		return fmt.Errorf("status is required")
	}
	for _, valid := range ValidItemStatuses {
		if input.Status == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid status %q", input.Status)
}

// IsPending reports whether the KDS item has pending status.
func (k KDSItem) IsPending() bool {
	return k.Status == "pending"
}

// IsPreparing reports whether the KDS item is currently being prepared.
func (k KDSItem) IsPreparing() bool {
	return k.Status == "preparing"
}

// IsReady reports whether the KDS item is ready for pickup.
func (k KDSItem) IsReady() bool {
	return k.Status == "ready"
}

// AllItemsReady reports whether all items on the KDS order are ready.
func (o KDSOrder) AllItemsReady() bool {
	if len(o.Items) == 0 {
		return false
	}
	for _, item := range o.Items {
		if item.Status != "ready" && item.Status != "served" {
			return false
		}
	}
	return true
}
