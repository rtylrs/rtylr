package domain

import (
	"errors"
	"testing"
	"time"
)

func TestValidateBillStatusTransition(t *testing.T) {
	allowed := [][2]string{{"pending", "overdue"}, {"pending", "paid"}, {"overdue", "paid"}, {"paid", "paid"}}
	for _, transition := range allowed {
		if err := ValidateBillStatusTransition(transition[0], transition[1]); err != nil {
			t.Errorf("%s -> %s: unexpected error: %v", transition[0], transition[1], err)
		}
	}

	if err := ValidateBillStatusTransition("paid", "pending"); !errors.Is(err, ErrInvalidStatusTransition) {
		t.Fatalf("paid -> pending error = %v, want ErrInvalidStatusTransition", err)
	}
}

func TestIsPaid_TrueWhenStatusPaid(t *testing.T) {
	b := &Bill{Status: "paid"}
	if !b.IsPaid() {
		t.Error("expected IsPaid=true for status=paid")
	}
}

func TestIsPaid_FalseWhenStatusPending(t *testing.T) {
	b := &Bill{Status: "pending"}
	if b.IsPaid() {
		t.Error("expected IsPaid=false for status=pending")
	}
}

func TestIsPaid_FalseWhenStatusOverdue(t *testing.T) {
	b := &Bill{Status: "overdue"}
	if b.IsPaid() {
		t.Error("expected IsPaid=false for status=overdue")
	}
}

func TestIsOverdue_TrueWhenPastDueAndNotPaid(t *testing.T) {
	b := &Bill{
		Status:  "pending",
		DueDate: time.Now().AddDate(0, 0, -1),
	}
	if !b.IsOverdue() {
		t.Error("expected IsOverdue=true for past-due unpaid bill")
	}
}

func TestIsOverdue_FalseWhenPaid(t *testing.T) {
	b := &Bill{
		Status:  "paid",
		DueDate: time.Now().AddDate(0, 0, -5),
	}
	if b.IsOverdue() {
		t.Error("expected IsOverdue=false for paid bill even if past due date")
	}
}

func TestIsOverdue_FalseWhenFutureDue(t *testing.T) {
	b := &Bill{
		Status:  "pending",
		DueDate: time.Now().AddDate(0, 0, 10),
	}
	if b.IsOverdue() {
		t.Error("expected IsOverdue=false when due date is in the future")
	}
}

func TestDaysUntilDue_PositiveWhenFuture(t *testing.T) {
	b := &Bill{DueDate: time.Now().AddDate(0, 0, 5)}
	days := b.DaysUntilDue()
	if days < 4 || days > 5 {
		t.Errorf("expected approximately 5 days, got %d", days)
	}
}

func TestDaysUntilDue_NegativeWhenPast(t *testing.T) {
	b := &Bill{DueDate: time.Now().AddDate(0, 0, -3)}
	days := b.DaysUntilDue()
	if days > -2 {
		t.Errorf("expected negative days for past-due bill, got %d", days)
	}
}
