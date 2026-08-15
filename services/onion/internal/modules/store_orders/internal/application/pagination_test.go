package application

import "testing"

func TestListPaginatedHasFiniteDeepOffsetLimit(t *testing.T) {
	if maxListOffset != 10000 {
		t.Fatalf("maxListOffset = %d, want 10000", maxListOffset)
	}
	limit, offset := normalizeListPagination(1000, maxListOffset+1)
	if limit != maxListLimit || offset != maxListOffset {
		t.Fatalf("normalized pagination = limit %d offset %d, want %d/%d", limit, offset, maxListLimit, maxListOffset)
	}

	limit, offset = normalizeListPagination(0, -1)
	if limit != defaultListLimit || offset != 0 {
		t.Fatalf("negative pagination = limit %d offset %d, want %d/0", limit, offset, defaultListLimit)
	}
}
