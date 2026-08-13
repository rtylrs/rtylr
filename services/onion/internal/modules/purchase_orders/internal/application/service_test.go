package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/purchase_orders/internal/domain"
)

func TestPaginate_Defaults(t *testing.T) {
	limit, offset := paginate(0, 0)
	if limit != 50 {
		t.Errorf("default limit = %d, want 50", limit)
	}
	if offset != 0 {
		t.Errorf("default offset = %d, want 0", offset)
	}
}

func TestPaginate_Page1(t *testing.T) {
	limit, offset := paginate(1, 10)
	if limit != 10 {
		t.Errorf("limit = %d, want 10", limit)
	}
	if offset != 0 {
		t.Errorf("offset = %d, want 0", offset)
	}
}

func TestPaginate_Page2(t *testing.T) {
	limit, offset := paginate(2, 10)
	if limit != 10 {
		t.Errorf("limit = %d, want 10", limit)
	}
	if offset != 10 {
		t.Errorf("offset = %d, want 10", offset)
	}
}

func TestPaginate_Page3(t *testing.T) {
	limit, offset := paginate(3, 25)
	if limit != 25 {
		t.Errorf("limit = %d, want 25", limit)
	}
	if offset != 50 {
		t.Errorf("offset = %d, want 50", offset)
	}
}

func TestPaginate_NegativePage(t *testing.T) {
	limit, offset := paginate(-1, 10)
	if offset != 0 {
		t.Errorf("negative page should give offset=0, got %d", offset)
	}
	if limit != 10 {
		t.Errorf("limit = %d, want 10", limit)
	}
}

func TestPaginate_CapsPageSize(t *testing.T) {
	limit, _ := paginate(1, 600)
	if limit != 500 {
		t.Errorf("limit should cap at 500, got %d", limit)
	}
}

func TestPaginate_NegativePageSize(t *testing.T) {
	limit, _ := paginate(1, -1)
	if limit != 50 {
		t.Errorf("negative pageSize should default to 50, got %d", limit)
	}
}

func TestPaginate_ExactMax(t *testing.T) {
	limit, _ := paginate(1, 500)
	if limit != 500 {
		t.Errorf("limit at max = %d, want 500", limit)
	}
}

func TestPaginate_PageSizeOne(t *testing.T) {
	limit, offset := paginate(5, 1)
	if limit != 1 {
		t.Errorf("limit = %d, want 1", limit)
	}
	if offset != 4 {
		t.Errorf("offset = %d, want 4", offset)
	}
}

func TestPaginate_LargePageNumber(t *testing.T) {
	limit, offset := paginate(100, 5)
	if limit != 5 {
		t.Errorf("limit = %d, want 5", limit)
	}
	if offset != 495 {
		t.Errorf("offset = %d, want 495", offset)
	}
}

func TestCreatePurchaseOrder_ValidatesNestedItemsBeforeRepository(t *testing.T) {
	s := &Service{}
	_, err := s.CreatePurchaseOrder(context.Background(), 1, domain.PurchaseOrderInput{
		SupplierID: 1,
		PONumber:   "PO-INVALID-ITEM",
		Items: []domain.PurchaseOrderItemInput{
			{ProductID: 1, QuantityOrdered: 0, UnitCost: -1},
		},
	}, 1)
	if !errors.Is(err, domain.ErrInvalidPurchaseOrder) {
		t.Fatalf("expected ErrInvalidPurchaseOrder, got %v", err)
	}
}

func TestCanApprovePurchaseOrder(t *testing.T) {
	for _, role := range []string{"owner", "admin", "operator", " ADMIN "} {
		if !canApprovePurchaseOrder(role) {
			t.Errorf("canApprovePurchaseOrder(%q) = false, want true", role)
		}
	}
	for _, role := range []string{"billing", "staff", "viewer", ""} {
		if canApprovePurchaseOrder(role) {
			t.Errorf("canApprovePurchaseOrder(%q) = true, want false", role)
		}
	}
}
