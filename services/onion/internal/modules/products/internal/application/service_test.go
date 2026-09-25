package application

import (
	"testing"
)

func TestPaginateProducts_Defaults(t *testing.T) {
	limit, offset := paginateProducts(0, 0)
	if limit != 50 {
		t.Errorf("default limit = %d, want 50", limit)
	}
	if offset != 0 {
		t.Errorf("default offset = %d, want 0", offset)
	}
}

func TestPaginateProducts_Page2(t *testing.T) {
	limit, offset := paginateProducts(2, 10)
	if limit != 10 {
		t.Errorf("limit = %d, want 10", limit)
	}
	if offset != 10 {
		t.Errorf("offset = %d, want 10", offset)
	}
}

func TestPaginateProducts_Page3(t *testing.T) {
	limit, offset := paginateProducts(3, 20)
	if limit != 20 {
		t.Errorf("limit = %d, want 20", limit)
	}
	if offset != 40 {
		t.Errorf("offset = %d, want 40", offset)
	}
}

func TestPaginateProducts_NegativePage(t *testing.T) {
	limit, offset := paginateProducts(-5, 10)
	if offset != 0 {
		t.Errorf("negative page should give offset=0, got %d", offset)
	}
	if limit != 10 {
		t.Errorf("limit = %d, want 10", limit)
	}
}

func TestPaginateProducts_CapsPageSize(t *testing.T) {
	limit, _ := paginateProducts(1, 600)
	if limit != 500 {
		t.Errorf("limit should cap at 500, got %d", limit)
	}
}

func TestPaginateProducts_PageSizeOne(t *testing.T) {
	limit, offset := paginateProducts(1, 1)
	if limit != 1 {
		t.Errorf("limit = %d, want 1", limit)
	}
	if offset != 0 {
		t.Errorf("offset = %d, want 0", offset)
	}
}

func TestPaginateProducts_LargePageNumber(t *testing.T) {
	limit, offset := paginateProducts(100, 5)
	if limit != 5 {
		t.Errorf("limit = %d, want 5", limit)
	}
	if offset != 495 {
		t.Errorf("offset = %d, want 495", offset)
	}
}

func TestPaginateProducts_PageSizeAtMax(t *testing.T) {
	limit, _ := paginateProducts(1, 500)
	if limit != 500 {
		t.Errorf("limit = %d, want 500", limit)
	}
}

func TestPaginateProducts_NegativePageSize(t *testing.T) {
	limit, _ := paginateProducts(1, -1)
	if limit != 50 {
		t.Errorf("negative pageSize should default to 50, got %d", limit)
	}
}
