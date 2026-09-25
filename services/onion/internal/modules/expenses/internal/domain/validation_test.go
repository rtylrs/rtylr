package domain

import (
	"testing"
	"time"
)

func TestValidatExpenseInput_HappyPath(t *testing.T) {
	err := ValidatExpenseInput(100, "USD", "Office supplies", time.Now().AddDate(0, 0, -1))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidatExpenseInput_ZeroAmount(t *testing.T) {
	err := ValidatExpenseInput(0, "USD", "Supplies", time.Now())
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestValidatExpenseInput_NegativeAmount(t *testing.T) {
	err := ValidatExpenseInput(-50, "USD", "Supplies", time.Now())
	if err == nil {
		t.Fatal("expected error for negative amount")
	}
}

func TestValidatExpenseInput_ZeroDate(t *testing.T) {
	err := ValidatExpenseInput(100, "USD", "Supplies", time.Time{})
	if err == nil {
		t.Fatal("expected error for zero date")
	}
}

func TestValidatExpenseInput_FutureDate(t *testing.T) {
	future := time.Now().AddDate(0, 0, 5)
	err := ValidatExpenseInput(100, "USD", "Supplies", future)
	if err == nil {
		t.Fatal("expected error for future date")
	}
}

func TestValidatExpenseInput_EmptyDescription(t *testing.T) {
	err := ValidatExpenseInput(100, "USD", "  ", time.Now())
	if err == nil {
		t.Fatal("expected error for empty description")
	}
}

func TestValidatExpenseInput_BadCurrencyLength(t *testing.T) {
	err := ValidatExpenseInput(100, "US", "Supplies", time.Now())
	if err == nil {
		t.Fatal("expected error for 2-letter currency")
	}
}

func TestValidatExpenseInput_EmptyCurrencyOK(t *testing.T) {
	err := ValidatExpenseInput(100, "", "Supplies", time.Now())
	if err != nil {
		t.Errorf("unexpected error for empty currency: %v", err)
	}
}

func TestValidatCategoryInput_HappyPath(t *testing.T) {
	err := ValidateCategoryInput("Travel")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidatCategoryInput_EmptyName(t *testing.T) {
	err := ValidateCategoryInput("  ")
	if err == nil {
		t.Fatal("expected error for empty category name")
	}
}

func TestValidateExpenseStatusTransition(t *testing.T) {
	allowed := [][2]string{{"pending", "approved"}, {"pending", "rejected"}, {"approved", "approved"}}
	for _, transition := range allowed {
		if err := ValidateExpenseStatusTransition(transition[0], transition[1]); err != nil {
			t.Errorf("%s -> %s: unexpected error: %v", transition[0], transition[1], err)
		}
	}
	if err := ValidateExpenseStatusTransition("approved", "rejected"); err == nil {
		t.Fatal("expected approved -> rejected to be rejected")
	}
	if err := ValidateExpenseStatusTransition("rejected", "approved"); err == nil {
		t.Fatal("expected rejected -> approved to be rejected")
	}
}
