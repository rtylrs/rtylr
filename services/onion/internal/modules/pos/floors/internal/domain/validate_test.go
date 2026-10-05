package domain

import "testing"

func TestValidateFloorInput_RejectsMissingBusinessID(t *testing.T) {
	if err := ValidateFloorInput(FloorInput{BranchID: 1, Name: "Main"}); err == nil {
		t.Fatal("expected error for missing business id")
	}
}

func TestValidateFloorInput_RejectsMissingBranchID(t *testing.T) {
	if err := ValidateFloorInput(FloorInput{BusinessID: 1, Name: "Main"}); err == nil {
		t.Fatal("expected error for missing branch id")
	}
}

func TestValidateFloorInput_RejectsBlankName(t *testing.T) {
	if err := ValidateFloorInput(FloorInput{BusinessID: 1, BranchID: 2, Name: ""}); err == nil {
		t.Fatal("expected error for blank name")
	}
}

func TestValidateFloorInput_RejectsNegativeSortOrder(t *testing.T) {
	if err := ValidateFloorInput(FloorInput{BusinessID: 1, BranchID: 2, Name: "Main", SortOrder: -1}); err == nil {
		t.Fatal("expected error for negative sort order")
	}
}

func TestValidateFloorInput_AcceptsZeroSortOrder(t *testing.T) {
	if err := ValidateFloorInput(FloorInput{BusinessID: 1, BranchID: 2, Name: "Main", SortOrder: 0}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateFloorInput_AcceptsValid(t *testing.T) {
	if err := ValidateFloorInput(FloorInput{BusinessID: 1, BranchID: 2, Name: "Ground Floor", SortOrder: 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateFloorUpdateInput_RejectsMissingID(t *testing.T) {
	if err := ValidateFloorUpdateInput(FloorUpdateInput{BusinessID: 1, Name: "Updated"}); err == nil {
		t.Fatal("expected error for missing floor id")
	}
}

func TestValidateFloorUpdateInput_RejectsMissingBusinessID(t *testing.T) {
	if err := ValidateFloorUpdateInput(FloorUpdateInput{ID: 1, Name: "Updated"}); err == nil {
		t.Fatal("expected error for missing business id")
	}
}

func TestValidateFloorUpdateInput_RejectsBlankName(t *testing.T) {
	if err := ValidateFloorUpdateInput(FloorUpdateInput{ID: 1, BusinessID: 2, Name: ""}); err == nil {
		t.Fatal("expected error for blank name")
	}
}

func TestValidateFloorUpdateInput_AcceptsValid(t *testing.T) {
	if err := ValidateFloorUpdateInput(FloorUpdateInput{ID: 1, BusinessID: 2, Name: "Updated Floor"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
