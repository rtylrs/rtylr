package api

import "testing"

// A delivery listing is scanned for failures, so the default page is small and
// the cap stops a client asking for the whole history at once.
func TestResolveDeliveryPaginationNormalisesAndCaps(t *testing.T) {
	cases := []struct {
		name        string
		page, limit int32
		wantPage    int32
		wantLimit   int32
		wantOffset  int64
	}{
		{"defaults", 0, 0, 1, defaultDeliveryPageLimit, 0},
		{"negative page", -5, 10, 1, 10, 0},
		{"second page", 2, 10, 2, 10, 10},
		{"limit above the cap", 1, 10000, 1, maxDeliveryPageLimit, 0},
		{"negative limit", 1, -1, 1, defaultDeliveryPageLimit, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			page, limit, offset := resolveDeliveryPagination(tc.page, tc.limit)
			if page != tc.wantPage || limit != tc.wantLimit || offset != tc.wantOffset {
				t.Errorf("resolveDeliveryPagination(%d, %d) = %d, %d, %d; want %d, %d, %d",
					tc.page, tc.limit, page, limit, offset, tc.wantPage, tc.wantLimit, tc.wantOffset)
			}
		})
	}
}

// The offset is narrowed to int32 by the repository, so an extreme page must
// clamp rather than wrap negative and alias to a wrong page.
func TestResolveDeliveryPaginationClampsAnExtremeOffset(t *testing.T) {
	_, _, offset := resolveDeliveryPagination(1<<30, maxDeliveryPageLimit)
	if offset < 0 {
		t.Fatalf("offset = %d, want a non-negative clamp: a negative offset aliases to the wrong page", offset)
	}
	if offset > 1<<31 {
		t.Errorf("offset = %d, want it clamped to fit int32", offset)
	}
}

// total is the full match count, not the page size, or a client cannot tell
// there are more failures than it can see.
func TestBuildDeliveryPaginationReportsTheFullTotal(t *testing.T) {
	got := buildDeliveryPagination(1, 10, 25, 10)
	if got.GetTotal() != 25 {
		t.Errorf("total = %d, want 25", got.GetTotal())
	}
	if got.GetTotalPages() != 3 {
		t.Errorf("total_pages = %d, want 3 for 25 over pages of 10", got.GetTotalPages())
	}
	if !got.GetHasNext() {
		t.Error("has_next is false on the first of three pages")
	}
	if got.GetHasPrev() {
		t.Error("has_prev is true on the first page")
	}

	last := buildDeliveryPagination(3, 10, 25, 5)
	if last.GetHasNext() {
		t.Error("has_next is true on the last page")
	}
	if !last.GetHasPrev() {
		t.Error("has_prev is false on the last page")
	}
}
