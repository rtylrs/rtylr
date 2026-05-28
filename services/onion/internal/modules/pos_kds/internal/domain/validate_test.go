package domain

import "testing"

func TestValidateUpdateItemStatusInput_RejectsMissingOwnerID(t *testing.T) {
	if err := ValidateUpdateItemStatusInput(UpdateItemStatusInput{OrderID: 1, ItemID: 2, Status: "ready"}); err == nil {
		t.Fatal("expected error for missing owner id")
	}
}

func TestValidateUpdateItemStatusInput_RejectsMissingOrderID(t *testing.T) {
	if err := ValidateUpdateItemStatusInput(UpdateItemStatusInput{OwnerID: 1, ItemID: 2, Status: "ready"}); err == nil {
		t.Fatal("expected error for missing order id")
	}
}

func TestValidateUpdateItemStatusInput_RejectsMissingItemID(t *testing.T) {
	if err := ValidateUpdateItemStatusInput(UpdateItemStatusInput{OwnerID: 1, OrderID: 2, Status: "ready"}); err == nil {
		t.Fatal("expected error for missing item id")
	}
}

func TestValidateUpdateItemStatusInput_RejectsBlankStatus(t *testing.T) {
	if err := ValidateUpdateItemStatusInput(UpdateItemStatusInput{OwnerID: 1, OrderID: 2, ItemID: 3}); err == nil {
		t.Fatal("expected error for blank status")
	}
}

func TestValidateUpdateItemStatusInput_RejectsInvalidStatus(t *testing.T) {
	if err := ValidateUpdateItemStatusInput(UpdateItemStatusInput{OwnerID: 1, OrderID: 2, ItemID: 3, Status: "unknown"}); err == nil {
		t.Fatal("expected error for unknown status")
	}
}

func TestValidateUpdateItemStatusInput_AcceptsAllValidStatuses(t *testing.T) {
	for _, status := range ValidItemStatuses {
		err := ValidateUpdateItemStatusInput(UpdateItemStatusInput{OwnerID: 1, OrderID: 2, ItemID: 3, Status: status})
		if err != nil {
			t.Errorf("expected no error for status %q, got: %v", status, err)
		}
	}
}

func TestKDSItem_IsPending(t *testing.T) {
	item := KDSItem{Status: "pending"}
	if !item.IsPending() {
		t.Error("expected IsPending=true")
	}
}

func TestKDSItem_IsPreparing(t *testing.T) {
	item := KDSItem{Status: "preparing"}
	if !item.IsPreparing() {
		t.Error("expected IsPreparing=true")
	}
}

func TestKDSItem_IsReady(t *testing.T) {
	item := KDSItem{Status: "ready"}
	if !item.IsReady() {
		t.Error("expected IsReady=true")
	}
}

func TestKDSOrder_AllItemsReady_WhenAll(t *testing.T) {
	order := KDSOrder{Items: []KDSItem{{Status: "ready"}, {Status: "served"}}}
	if !order.AllItemsReady() {
		t.Error("expected AllItemsReady=true when all items are ready/served")
	}
}

func TestKDSOrder_AllItemsReady_WhenNone(t *testing.T) {
	order := KDSOrder{Items: []KDSItem{{Status: "pending"}, {Status: "ready"}}}
	if order.AllItemsReady() {
		t.Error("expected AllItemsReady=false when some items are pending")
	}
}

func TestKDSOrder_AllItemsReady_Empty(t *testing.T) {
	order := KDSOrder{}
	if order.AllItemsReady() {
		t.Error("expected AllItemsReady=false for empty items")
	}
}
