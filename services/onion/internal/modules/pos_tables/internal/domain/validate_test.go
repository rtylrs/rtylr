package domain

import "testing"

func TestValidateTableInput_RejectsMissingOwnerID(t *testing.T) {
	if err := ValidateTableInput(TableInput{Label: "T1", Seats: 4}); err == nil {
		t.Fatal("expected error for missing owner id")
	}
}

func TestValidateTableInput_RejectsBlankLabel(t *testing.T) {
	if err := ValidateTableInput(TableInput{OwnerID: 1, Seats: 4}); err == nil {
		t.Fatal("expected error for blank label")
	}
}

func TestValidateTableInput_RejectsNegativeSeats(t *testing.T) {
	if err := ValidateTableInput(TableInput{OwnerID: 1, Label: "T1", Seats: -1}); err == nil {
		t.Fatal("expected error for negative seats")
	}
}

func TestValidateTableInput_AcceptsZeroSeats(t *testing.T) {
	if err := ValidateTableInput(TableInput{OwnerID: 1, Label: "T1", Seats: 0}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateTableInput_AcceptsValid(t *testing.T) {
	if err := ValidateTableInput(TableInput{OwnerID: 1, Label: "Table A", Seats: 4}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateTableUpdateInput_RejectsMissingID(t *testing.T) {
	if err := ValidateTableUpdateInput(TableUpdateInput{OwnerID: 1, Label: "T1"}); err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestValidateTableUpdateInput_RejectsInvalidStatus(t *testing.T) {
	if err := ValidateTableUpdateInput(TableUpdateInput{ID: 1, OwnerID: 2, Label: "T1", Status: "broken"}); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestValidateTableUpdateInput_AcceptsAllValidStatuses(t *testing.T) {
	for _, status := range ValidTableStatuses {
		err := ValidateTableUpdateInput(TableUpdateInput{ID: 1, OwnerID: 2, Label: "T1", Status: status})
		if err != nil {
			t.Errorf("expected no error for status %q, got: %v", status, err)
		}
	}
}

func TestTable_IsAvailable(t *testing.T) {
	table := Table{Status: "available"}
	if !table.IsAvailable() {
		t.Error("expected IsAvailable=true")
	}
}

func TestTable_IsAvailable_WhenOccupied(t *testing.T) {
	table := Table{Status: "occupied"}
	if table.IsAvailable() {
		t.Error("expected IsAvailable=false when occupied")
	}
}

func TestTable_IsOccupied(t *testing.T) {
	table := Table{Status: "occupied"}
	if !table.IsOccupied() {
		t.Error("expected IsOccupied=true")
	}
}

func TestTable_IsOccupied_WhenAvailable(t *testing.T) {
	table := Table{Status: "available"}
	if table.IsOccupied() {
		t.Error("expected IsOccupied=false when available")
	}
}
