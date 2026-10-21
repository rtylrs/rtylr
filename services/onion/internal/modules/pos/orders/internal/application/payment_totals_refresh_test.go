package application

import (
	"context"
	"errors"
	"testing"
)

type paymentTotalsRefresherStub struct {
	sessionID      int64
	sessionErr     error
	totals         [4]float64
	totalOrders    int
	totalsErr      error
	updateErr      error
	updatedSession int64
	updatedTotals  [4]float64
	updatedOrders  int
	updateCalls    int
}

func (s *paymentTotalsRefresherStub) GetOrderSessionID(context.Context, int64) (int64, error) {
	return s.sessionID, s.sessionErr
}

func (s *paymentTotalsRefresherStub) GetSessionTotals(context.Context, int64) (float64, float64, float64, float64, int, error) {
	return s.totals[0], s.totals[1], s.totals[2], s.totals[3], s.totalOrders, s.totalsErr
}

func (s *paymentTotalsRefresherStub) UpdateSessionTotals(_ context.Context, sessionID int64, sales, cash, card, online float64, orders int) error {
	s.updateCalls++
	s.updatedSession = sessionID
	s.updatedTotals = [4]float64{sales, cash, card, online}
	s.updatedOrders = orders
	return s.updateErr
}

func TestRefreshSessionTotalsAfterPayment_UpdateFailureIsNonFatal(t *testing.T) {
	refreshErr := errors.New("totals update unavailable")
	repo := &paymentTotalsRefresherStub{
		sessionID:   42,
		totals:      [4]float64{100, 60, 40, 0},
		totalOrders: 3,
		updateErr:   refreshErr,
	}

	refreshSessionTotalsAfterPayment(context.Background(), repo, 17)

	if repo.updateCalls != 1 {
		t.Fatalf("UpdateSessionTotals calls = %d, want 1", repo.updateCalls)
	}
	if repo.updatedSession != 42 || repo.updatedOrders != 3 {
		t.Fatalf("updated session/order count = (%d, %d), want (42, 3)", repo.updatedSession, repo.updatedOrders)
	}
}

func TestRefreshSessionTotalsAfterPayment_ReadFailureSkipsUpdate(t *testing.T) {
	readErr := errors.New("totals read unavailable")
	repo := &paymentTotalsRefresherStub{
		sessionID: 42,
		totalsErr: readErr,
	}

	refreshSessionTotalsAfterPayment(context.Background(), repo, 17)

	if repo.updateCalls != 0 {
		t.Fatalf("UpdateSessionTotals calls = %d, want 0", repo.updateCalls)
	}
}

func TestRefreshSessionTotalsAfterPayment_LookupFailureSkipsRefresh(t *testing.T) {
	lookupErr := errors.New("session lookup unavailable")
	repo := &paymentTotalsRefresherStub{sessionErr: lookupErr}

	refreshSessionTotalsAfterPayment(context.Background(), repo, 17)

	if repo.updateCalls != 0 {
		t.Fatalf("UpdateSessionTotals calls = %d, want 0", repo.updateCalls)
	}
}
