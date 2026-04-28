package application

import (
	"testing"
	"time"
)

func TestComputeLineItem_FullMonth(t *testing.T) {
	billing := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 30, 23, 59, 59, 0, time.UTC)
	li := computeLineItem(10, 1000, billing, &end, billing)

	// April has 30 days. 10 seats × 30 days × $10 / 30 = $100 = 10000 cents.
	if li.AmountCents != 10000 {
		t.Errorf("expected 10000 cents, got %d", li.AmountCents)
	}
	if li.SeatCount != 10 {
		t.Errorf("seat count round trip: %d", li.SeatCount)
	}
}

func TestComputeLineItem_PartialMonth(t *testing.T) {
	billing := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	start := time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)

	li := computeLineItem(10, 3000, start, &end, billing)
	// 10 seats × 15 days (16→30 inclusive) × $30 / 30 days = $150 = 15000 cents.
	if li.AmountCents != 15000 {
		t.Errorf("expected 15000 cents for half-month, got %d", li.AmountCents)
	}
}

func TestComputeLineItem_OpenEndedUsesNow(t *testing.T) {
	now := time.Now().UTC()
	billing := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	li := computeLineItem(5, 1000, billing, nil, billing)
	if li.PeriodEnd != nil {
		t.Errorf("nil end should be preserved on the line item, got %v", li.PeriodEnd)
	}
	if li.AmountCents <= 0 {
		t.Errorf("open-ended period should still bill, got %d", li.AmountCents)
	}
}

func TestComputeLineItem_NeverNegativeOrTooSmall(t *testing.T) {
	billing := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	// Start AFTER end -> daysUsed bottoms out at 1.
	start := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC)

	li := computeLineItem(10, 3000, start, &end, billing)
	// Bottoms out at 1 day: 10 × 1 × 3000 / 30 = 1000.
	if li.AmountCents != 1000 {
		t.Errorf("expected 1000 cents (1-day floor), got %d", li.AmountCents)
	}
}
