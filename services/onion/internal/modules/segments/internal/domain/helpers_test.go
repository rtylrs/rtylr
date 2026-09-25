package domain

import "testing"

func TestHasDescription_WhenSet(t *testing.T) {
	s := Segment{Description: "VIP customers"}
	if !s.HasDescription() {
		t.Error("expected HasDescription=true when description is set")
	}
}

func TestHasDescription_WhenEmpty(t *testing.T) {
	s := Segment{Description: ""}
	if s.HasDescription() {
		t.Error("expected HasDescription=false when description is empty")
	}
}

func TestHasCustomers_WhenPositive(t *testing.T) {
	s := Segment{CustomerCount: 5}
	if !s.HasCustomers() {
		t.Error("expected HasCustomers=true when count > 0")
	}
}

func TestHasCustomers_WhenZero(t *testing.T) {
	s := Segment{CustomerCount: 0}
	if s.HasCustomers() {
		t.Error("expected HasCustomers=false when count is 0")
	}
}

func TestHasFilters_WhenSet(t *testing.T) {
	s := Segment{Filters: []byte(`{"tier":"gold"}`)}
	if !s.HasFilters() {
		t.Error("expected HasFilters=true when filters are set")
	}
}

func TestHasFilters_WhenEmptyObject(t *testing.T) {
	s := Segment{Filters: []byte("{}")}
	if s.HasFilters() {
		t.Error("expected HasFilters=false for empty JSON object")
	}
}

func TestHasFilters_WhenNil(t *testing.T) {
	s := Segment{Filters: nil}
	if s.HasFilters() {
		t.Error("expected HasFilters=false when filters is nil")
	}
}

func TestValidateSegmentInput_Valid(t *testing.T) {
	msg := ValidateSegmentInput(SegmentInput{OwnerID: 1, Name: "Gold Members"})
	if msg != "" {
		t.Errorf("expected no error, got %q", msg)
	}
}

func TestValidateSegmentInput_ZeroOwner(t *testing.T) {
	msg := ValidateSegmentInput(SegmentInput{OwnerID: 0, Name: "Gold Members"})
	if msg == "" {
		t.Error("expected error for zero owner_id")
	}
}

func TestValidateSegmentInput_EmptyName(t *testing.T) {
	msg := ValidateSegmentInput(SegmentInput{OwnerID: 1, Name: ""})
	if msg == "" {
		t.Error("expected error for empty name")
	}
}

func TestValidateSegmentUpdateInput_Valid(t *testing.T) {
	msg := ValidateSegmentUpdateInput(SegmentUpdateInput{ID: 10, OwnerID: 1, Name: "Platinum"})
	if msg != "" {
		t.Errorf("expected no error, got %q", msg)
	}
}

func TestValidateSegmentUpdateInput_ZeroID(t *testing.T) {
	msg := ValidateSegmentUpdateInput(SegmentUpdateInput{ID: 0, OwnerID: 1, Name: "Platinum"})
	if msg == "" {
		t.Error("expected error for zero id")
	}
}

func TestValidateSegmentUpdateInput_ZeroOwner(t *testing.T) {
	msg := ValidateSegmentUpdateInput(SegmentUpdateInput{ID: 10, OwnerID: 0, Name: "Platinum"})
	if msg == "" {
		t.Error("expected error for zero owner_id")
	}
}

func TestValidateSegmentUpdateInput_EmptyName(t *testing.T) {
	msg := ValidateSegmentUpdateInput(SegmentUpdateInput{ID: 10, OwnerID: 1, Name: ""})
	if msg == "" {
		t.Error("expected error for empty name")
	}
}
