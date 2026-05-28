package domain

import "testing"

func TestValidateLocationInput_RejectsMissingOwnerID(t *testing.T) {
	if err := ValidateLocationInput(LocationInput{Name: "Branch"}); err == nil {
		t.Fatal("expected error for missing owner id")
	}
}

func TestValidateLocationInput_RejectsBlankName(t *testing.T) {
	if err := ValidateLocationInput(LocationInput{OwnerID: 1}); err == nil {
		t.Fatal("expected error for blank name")
	}
}

func TestValidateLocationInput_AcceptsValid(t *testing.T) {
	if err := ValidateLocationInput(LocationInput{OwnerID: 1, Name: "Downtown Branch"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateLocationInput_AcceptsWithAddress(t *testing.T) {
	if err := ValidateLocationInput(LocationInput{OwnerID: 1, Name: "Branch", Address: "123 Main St"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateLocationUpdateInput_RejectsMissingID(t *testing.T) {
	if err := ValidateLocationUpdateInput(LocationUpdateInput{OwnerID: 1, Name: "Branch"}); err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestValidateLocationUpdateInput_RejectsMissingOwnerID(t *testing.T) {
	if err := ValidateLocationUpdateInput(LocationUpdateInput{ID: 1, Name: "Branch"}); err == nil {
		t.Fatal("expected error for missing owner id")
	}
}

func TestValidateLocationUpdateInput_RejectsBlankName(t *testing.T) {
	if err := ValidateLocationUpdateInput(LocationUpdateInput{ID: 1, OwnerID: 2}); err == nil {
		t.Fatal("expected error for blank name")
	}
}

func TestValidateLocationUpdateInput_AcceptsValid(t *testing.T) {
	if err := ValidateLocationUpdateInput(LocationUpdateInput{ID: 1, OwnerID: 2, Name: "Updated Branch"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
