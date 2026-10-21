package domain

import "testing"

func TestPurchaseOrder_IsTerminal_Received(t *testing.T) {
	po := PurchaseOrder{Status: POStatusReceived}
	if !po.IsTerminal() {
		t.Error("received order should be terminal")
	}
}

func TestPurchaseOrder_IsTerminal_Cancelled(t *testing.T) {
	po := PurchaseOrder{Status: POStatusCancelled}
	if !po.IsTerminal() {
		t.Error("cancelled order should be terminal")
	}
}

func TestPurchaseOrder_IsTerminal_Draft(t *testing.T) {
	po := PurchaseOrder{Status: POStatusDraft}
	if po.IsTerminal() {
		t.Error("draft order should not be terminal")
	}
}

func TestPurchaseOrder_IsTerminal_Approved(t *testing.T) {
	po := PurchaseOrder{Status: POStatusApproved}
	if po.IsTerminal() {
		t.Error("approved order should not be terminal")
	}
}

func TestPurchaseOrder_IsOpen_Draft(t *testing.T) {
	po := PurchaseOrder{Status: POStatusDraft}
	if !po.IsOpen() {
		t.Error("draft order should be open")
	}
}

func TestPurchaseOrder_IsOpen_Sent(t *testing.T) {
	po := PurchaseOrder{Status: POStatusSent}
	if !po.IsOpen() {
		t.Error("sent order should be open")
	}
}

func TestPurchaseOrder_IsOpen_Received(t *testing.T) {
	po := PurchaseOrder{Status: POStatusReceived}
	if po.IsOpen() {
		t.Error("received order should not be open")
	}
}

func TestPurchaseOrder_StatusLabel_Draft(t *testing.T) {
	po := PurchaseOrder{Status: POStatusDraft}
	if got := po.StatusLabel(); got != "Draft" {
		t.Errorf("expected 'Draft', got %q", got)
	}
}

func TestPurchaseOrder_StatusLabel_Sent(t *testing.T) {
	po := PurchaseOrder{Status: POStatusSent}
	if got := po.StatusLabel(); got != "Sent to Supplier" {
		t.Errorf("expected 'Sent to Supplier', got %q", got)
	}
}

func TestPurchaseOrder_StatusLabel_Approved(t *testing.T) {
	po := PurchaseOrder{Status: POStatusApproved}
	if got := po.StatusLabel(); got != "Approved" {
		t.Errorf("expected 'Approved', got %q", got)
	}
}

func TestPurchaseOrder_StatusLabel_Partial(t *testing.T) {
	po := PurchaseOrder{Status: POStatusPartial}
	if got := po.StatusLabel(); got != "Partially Received" {
		t.Errorf("expected 'Partially Received', got %q", got)
	}
}

func TestPurchaseOrder_StatusLabel_Received(t *testing.T) {
	po := PurchaseOrder{Status: POStatusReceived}
	if got := po.StatusLabel(); got != "Fully Received" {
		t.Errorf("expected 'Fully Received', got %q", got)
	}
}

func TestPurchaseOrder_StatusLabel_Cancelled(t *testing.T) {
	po := PurchaseOrder{Status: POStatusCancelled}
	if got := po.StatusLabel(); got != "Cancelled" {
		t.Errorf("expected 'Cancelled', got %q", got)
	}
}

func TestValidatePOStatusTransition(t *testing.T) {
	cases := []struct {
		current PurchaseOrderStatus
		next    PurchaseOrderStatus
		wantErr bool
	}{
		{POStatusDraft, POStatusDraft, false},
		{POStatusDraft, POStatusSent, false},
		{POStatusDraft, POStatusApproved, false},
		{POStatusDraft, POStatusCancelled, false},
		{POStatusDraft, POStatusReceived, true},
		{POStatusSent, POStatusApproved, false},
		{POStatusSent, POStatusCancelled, false},
		{POStatusSent, POStatusDraft, true},
		{POStatusApproved, POStatusPartial, false},
		{POStatusApproved, POStatusReceived, false},
		{POStatusApproved, POStatusCancelled, false},
		{POStatusPartial, POStatusReceived, false},
		{POStatusPartial, POStatusCancelled, false},
		{POStatusReceived, POStatusDraft, true},
		{POStatusReceived, POStatusCancelled, true},
		{POStatusCancelled, POStatusDraft, true},
		{POStatusCancelled, POStatusApproved, true},
	}

	for _, tc := range cases {
		err := ValidatePOStatusTransition(tc.current, tc.next)
		if (err != nil) != tc.wantErr {
			t.Errorf("transition %s -> %s: got err=%v, wantErr=%v", tc.current, tc.next, err, tc.wantErr)
		}
	}
}
