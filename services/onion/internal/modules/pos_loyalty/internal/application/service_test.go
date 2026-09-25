package application

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_loyalty/internal/domain"
)

// Service with nil repo panics on repo calls, but input guards fire first.

func TestEarnPoints_RejectsMissingBusinessID(t *testing.T) {
	s := &Service{}
	_, err := s.EarnPoints(context.Background(), domain.EarnPointsInput{
		CustomerID: 1,
		Points:     10,
	})
	if err == nil {
		t.Fatal("expected error for missing business id")
	}
}

func TestEarnPoints_RejectsMissingCustomerID(t *testing.T) {
	s := &Service{}
	_, err := s.EarnPoints(context.Background(), domain.EarnPointsInput{
		BusinessID: 1,
		Points:     10,
	})
	if err == nil {
		t.Fatal("expected error for missing customer id")
	}
}

func TestEarnPoints_RejectsZeroPoints(t *testing.T) {
	s := &Service{}
	_, err := s.EarnPoints(context.Background(), domain.EarnPointsInput{
		BusinessID: 1,
		CustomerID: 2,
		Points:     0,
	})
	if err == nil {
		t.Fatal("expected error for zero points")
	}
}

func TestRedeemPoints_RejectsMissingBusinessID(t *testing.T) {
	s := &Service{}
	_, err := s.RedeemPoints(context.Background(), domain.RedeemPointsInput{
		CustomerID: 1,
		Points:     10,
	})
	if err == nil {
		t.Fatal("expected error for missing business id")
	}
}

func TestRedeemPoints_RejectsMissingCustomerID(t *testing.T) {
	s := &Service{}
	_, err := s.RedeemPoints(context.Background(), domain.RedeemPointsInput{
		BusinessID: 1,
		Points:     10,
	})
	if err == nil {
		t.Fatal("expected error for missing customer id")
	}
}

func TestRedeemPoints_RejectsZeroPoints(t *testing.T) {
	s := &Service{}
	_, err := s.RedeemPoints(context.Background(), domain.RedeemPointsInput{
		BusinessID: 1,
		CustomerID: 2,
		Points:     0,
	})
	if err == nil {
		t.Fatal("expected error for zero points")
	}
}
