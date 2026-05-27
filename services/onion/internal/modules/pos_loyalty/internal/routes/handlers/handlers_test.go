package handlers

import (
	"context"
	"testing"
)

// ─── UpdateConfigHandler.Validate ────────────────────────────────────────────

func TestUpdateConfig_Validate_AcceptsZeroValues(t *testing.T) {
	h := &UpdateConfigHandler{}
	req := &UpdateConfigRequest{}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("zero value should be valid: %v", err.Message)
	}
}

func TestUpdateConfig_Validate_NegativePointsPerDollar(t *testing.T) {
	h := &UpdateConfigHandler{}
	if err := h.Validate(context.Background(), &UpdateConfigRequest{PointsPerDollar: -1}); err == nil {
		t.Error("expected error for negative points_per_dollar")
	}
}

func TestUpdateConfig_Validate_NegativeRedemptionRate(t *testing.T) {
	h := &UpdateConfigHandler{}
	if err := h.Validate(context.Background(), &UpdateConfigRequest{RedemptionRate: -0.5}); err == nil {
		t.Error("expected error for negative redemption_rate")
	}
}

func TestUpdateConfig_Validate_NegativeMinRedeemPoints(t *testing.T) {
	h := &UpdateConfigHandler{}
	if err := h.Validate(context.Background(), &UpdateConfigRequest{MinRedeemPoints: -10}); err == nil {
		t.Error("expected error for negative min_redeem_points")
	}
}

func TestUpdateConfig_Validate_ValidPositiveValues(t *testing.T) {
	h := &UpdateConfigHandler{}
	req := &UpdateConfigRequest{
		IsEnabled:       true,
		PointsPerDollar: 1.5,
		RedemptionRate:  0.01,
		MinRedeemPoints: 100,
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("valid config should pass: %v", err.Message)
	}
}

func TestUpdateConfig_Validate_ZeroIsEnabled(t *testing.T) {
	h := &UpdateConfigHandler{}
	req := &UpdateConfigRequest{IsEnabled: false, PointsPerDollar: 0, RedemptionRate: 0, MinRedeemPoints: 0}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("disabled config with zero values should be valid: %v", err.Message)
	}
}

// ─── Request struct default values ───────────────────────────────────────────

func TestEarnPointsRequest_ZeroCustomerID(t *testing.T) {
	req := &EarnPointsRequest{}
	if req.CustomerID != 0 {
		t.Error("expected zero customer id")
	}
}

func TestRedeemPointsRequest_ZeroPoints(t *testing.T) {
	req := &RedeemPointsRequest{}
	if req.Points != 0 {
		t.Error("expected zero points")
	}
}

func TestGetBalanceRequest_ZeroCustomerID(t *testing.T) {
	req := &GetBalanceRequest{}
	if req.CustomerID != 0 {
		t.Error("expected zero customer id")
	}
}

func TestListTransactionsRequest_NonZeroCustomerID(t *testing.T) {
	req := &ListTransactionsRequest{CustomerID: 42}
	if req.CustomerID == 0 {
		t.Error("expected non-zero customer id")
	}
}

// ─── Handler struct creation ──────────────────────────────────────────────────

func TestHandlers_CanBeInstantiated(_ *testing.T) {
	_ = &GetConfigHandler{}
	_ = &UpdateConfigHandler{}
	_ = &ListTransactionsHandler{}
	_ = &EarnPointsHandler{}
	_ = &RedeemPointsHandler{}
	_ = &GetBalanceHandler{}
}

// ─── Routes helpers ───────────────────────────────────────────────────────────

func TestRoutes_NonEmpty(t *testing.T) {
	for name, fn := range map[string]func() []string{
		"GetConfig":        GetConfigRoutes,
		"UpdateConfig":     UpdateConfigRoutes,
		"ListTransactions": ListTransactionsRoutes,
		"EarnPoints":       EarnPointsRoutes,
		"RedeemPoints":     RedeemPointsRoutes,
		"GetBalance":       GetBalanceRoutes,
	} {
		if len(fn()) == 0 {
			t.Errorf("%s: expected at least one route", name)
		}
	}
}
