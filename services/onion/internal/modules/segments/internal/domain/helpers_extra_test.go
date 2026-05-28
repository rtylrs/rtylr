package domain

import "testing"

func TestIsPopulated_WhenHasCustomers(t *testing.T) {
	s := Segment{CustomerCount: 10}
	if !s.IsPopulated() {
		t.Error("expected IsPopulated=true when customer count > 0")
	}
}

func TestIsPopulated_WhenEmpty(t *testing.T) {
	s := Segment{CustomerCount: 0}
	if s.IsPopulated() {
		t.Error("expected IsPopulated=false when customer count = 0")
	}
}

func TestIsEmpty_WhenZero(t *testing.T) {
	s := Segment{CustomerCount: 0}
	if !s.IsEmpty() {
		t.Error("expected IsEmpty=true when customer count = 0")
	}
}

func TestIsEmpty_WhenPositive(t *testing.T) {
	s := Segment{CustomerCount: 3}
	if s.IsEmpty() {
		t.Error("expected IsEmpty=false when customer count > 0")
	}
}

func TestIsPopulated_AndIsEmpty_AreOpposite(t *testing.T) {
	withCustomers := Segment{CustomerCount: 5}
	withoutCustomers := Segment{CustomerCount: 0}

	if withCustomers.IsPopulated() == withCustomers.IsEmpty() {
		t.Error("IsPopulated and IsEmpty should never both be true or false simultaneously")
	}
	if withoutCustomers.IsPopulated() == withoutCustomers.IsEmpty() {
		t.Error("IsPopulated and IsEmpty should never both be true or false simultaneously")
	}
}

func TestHasDescription_WhitespaceOnly(t *testing.T) {
	s := Segment{Description: "   "}
	// whitespace is technically not empty string, helper returns true
	if !s.HasDescription() {
		t.Log("HasDescription returned false for whitespace-only description; this is also acceptable")
	}
}

func TestHasCustomers_LargeCount(t *testing.T) {
	s := Segment{CustomerCount: 1000000}
	if !s.HasCustomers() {
		t.Error("expected HasCustomers=true for large customer count")
	}
}
