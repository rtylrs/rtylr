package application

import "testing"

func TestPaginateSuppliers_Defaults(t *testing.T) {
	limit, offset := paginateSuppliers(0, 0)
	if limit != 50 {
		t.Errorf("default limit = %d, want 50", limit)
	}
	if offset != 0 {
		t.Errorf("default offset = %d, want 0", offset)
	}
}

func TestPaginateSuppliers_Page1(t *testing.T) {
	limit, offset := paginateSuppliers(1, 10)
	if limit != 10 {
		t.Errorf("limit = %d, want 10", limit)
	}
	if offset != 0 {
		t.Errorf("offset = %d, want 0", offset)
	}
}

func TestPaginateSuppliers_Page2(t *testing.T) {
	limit, offset := paginateSuppliers(2, 10)
	if limit != 10 {
		t.Errorf("limit = %d, want 10", limit)
	}
	if offset != 10 {
		t.Errorf("offset = %d, want 10", offset)
	}
}

func TestPaginateSuppliers_Page3(t *testing.T) {
	limit, offset := paginateSuppliers(3, 20)
	if limit != 20 {
		t.Errorf("limit = %d, want 20", limit)
	}
	if offset != 40 {
		t.Errorf("offset = %d, want 40", offset)
	}
}

func TestPaginateSuppliers_NegativePage(t *testing.T) {
	_, offset := paginateSuppliers(-1, 10)
	if offset != 0 {
		t.Errorf("expected offset=0 for negative page, got %d", offset)
	}
}

func TestPaginateSuppliers_CapsPageSize(t *testing.T) {
	limit, _ := paginateSuppliers(1, 999)
	if limit != 500 {
		t.Errorf("expected limit capped at 500, got %d", limit)
	}
}

func TestPaginateSuppliers_NegativePageSize(t *testing.T) {
	limit, _ := paginateSuppliers(1, -10)
	if limit != 50 {
		t.Errorf("expected default limit=50 for negative pageSize, got %d", limit)
	}
}

func TestPaginateSuppliers_ExactMax(t *testing.T) {
	limit, _ := paginateSuppliers(1, 500)
	if limit != 500 {
		t.Errorf("expected limit=500 at max, got %d", limit)
	}
}

func TestPaginateSuppliers_LargePage(t *testing.T) {
	limit, offset := paginateSuppliers(10, 5)
	if limit != 5 {
		t.Errorf("limit = %d, want 5", limit)
	}
	if offset != 45 {
		t.Errorf("offset = %d, want 45", offset)
	}
}
