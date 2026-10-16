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

func TestValidateUpdateItemStatusInput_AcceptsAliases(t *testing.T) {
	aliases := []string{"cooking", "new", "COOKING", "  ready  "}
	for _, alias := range aliases {
		err := ValidateUpdateItemStatusInput(UpdateItemStatusInput{OwnerID: 1, OrderID: 2, ItemID: 3, Status: alias})
		if err != nil {
			t.Errorf("expected no error for alias %q, got: %v", alias, err)
		}
	}
}

func TestNormalizeItemStatus(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"cooking", "preparing", false},
		{"COOKING", "preparing", false},
		{"new", "pending", false},
		{"pending", "pending", false},
		{"preparing", "preparing", false},
		{"ready", "ready", false},
		{"served", "served", false},
		{"cancelled", "cancelled", false},
		{"bogus", "", true},
		{"", "", true},
	}
	for _, tt := range tests {
		got, err := NormalizeItemStatus(tt.in)
		if (err != nil) != tt.wantErr {
			t.Errorf("NormalizeItemStatus(%q) err = %v, wantErr = %v", tt.in, err, tt.wantErr)
		}
		if got != tt.want {
			t.Errorf("NormalizeItemStatus(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestDeriveOrderStatusFromItemStatuses(t *testing.T) {
	tests := []struct {
		name     string
		statuses []string
		want     string
	}{
		{"empty returns confirmed", nil, "confirmed"},
		{"all pending returns confirmed", []string{"pending", "pending"}, "confirmed"},
		{"one cooking returns preparing", []string{"pending", "cooking"}, "preparing"},
		{"one preparing returns preparing", []string{"pending", "preparing"}, "preparing"},
		{"one ready and one pending returns preparing", []string{"pending", "ready"}, "preparing"},
		{"all ready returns ready", []string{"ready", "ready"}, "ready"},
		{"ready and served returns ready", []string{"ready", "served"}, "ready"},
		{"all served returns served", []string{"served", "served"}, "served"},
		{"served and cancelled returns served", []string{"served", "cancelled"}, "served"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeriveOrderStatusFromItemStatuses(tt.statuses)
			if got != tt.want {
				t.Errorf("DeriveOrderStatusFromItemStatuses(%v) = %q, want %q", tt.statuses, got, tt.want)
			}
		})
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
	cookingItem := KDSItem{Status: "cooking"}
	if !cookingItem.IsPreparing() {
		t.Error("expected IsPreparing=true for cooking")
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
