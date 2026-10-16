package domain

import (
	"fmt"
	"strings"
)

// ValidItemStatuses lists all valid KDS item status values.
var ValidItemStatuses = []string{"pending", "preparing", "ready", "served", "cancelled"}

// NormalizeItemStatus maps incoming status values and common client aliases
// (e.g. "cooking" -> "preparing", "new" -> "pending") to canonical backend item status strings.
func NormalizeItemStatus(status string) (string, error) {
	norm := strings.ToLower(strings.TrimSpace(status))
	switch norm {
	case "cooking":
		return "preparing", nil
	case "new":
		return "pending", nil
	case "pending", "preparing", "ready", "served", "cancelled":
		return norm, nil
	default:
		return "", fmt.Errorf("invalid status %q", status)
	}
}

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
	if _, err := NormalizeItemStatus(input.Status); err != nil {
		return err
	}
	return nil
}

// DeriveOrderStatusFromItemStatuses computes the parent order workflow status
// based on the statuses of all its active line items.
func DeriveOrderStatusFromItemStatuses(itemStatuses []string) string {
	if len(itemStatuses) == 0 {
		return "confirmed"
	}
	allServed := true
	allReadyOrServed := true
	anyPreparingOrReady := false

	for _, s := range itemStatuses {
		norm, err := NormalizeItemStatus(s)
		if err != nil || norm == "cancelled" {
			continue
		}
		if norm != "served" {
			allServed = false
		}
		if norm != "ready" && norm != "served" {
			allReadyOrServed = false
		}
		if norm == "preparing" || norm == "ready" {
			anyPreparingOrReady = true
		}
	}

	if allServed {
		return "served"
	}
	if allReadyOrServed {
		return "ready"
	}
	if anyPreparingOrReady {
		return "preparing"
	}
	return "confirmed"
}

// IsPending reports whether the KDS item has pending status.
func (k KDSItem) IsPending() bool {
	return k.Status == "pending"
}

// IsPreparing reports whether the KDS item is currently being prepared.
func (k KDSItem) IsPreparing() bool {
	return k.Status == "preparing" || k.Status == "cooking"
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
