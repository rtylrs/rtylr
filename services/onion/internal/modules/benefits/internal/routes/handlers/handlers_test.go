package handlers

import (
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/benefits/internal/domain"
)

func TestCreateHandler_Validate_MissingName(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: ""}
	if err := h.Validate(nil, req); err == nil {
		t.Error("expected error for missing name")
	}
}

func TestCreateHandler_Validate_WhitespaceName(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "   "}
	if err := h.Validate(nil, req); err == nil {
		t.Error("expected error for whitespace-only name")
	}
}

func TestCreateHandler_Validate_Valid(t *testing.T) {
	h := &CreateHandler{}
	req := &CreateRequest{Name: "Health Plan"}
	if err := h.Validate(nil, req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewPayload_RoundTrip(t *testing.T) {
	now := time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)
	in := domain.BenefitPlan{
		ID:           5,
		OwnerID:      1001,
		Name:         "Premium Health",
		Type:         "health",
		Provider:     "Acme Health",
		CostEmployer: 250,
		CostEmployee: 50,
		IsActive:     true,
		CreatedAt:    now,
	}
	out := newPayload(in)

	if out.ID != in.ID || out.Name != in.Name || out.Type != in.Type {
		t.Errorf("identity didn't round-trip: %+v", out)
	}
	if out.CostEmployer != in.CostEmployer || out.CostEmployee != in.CostEmployee {
		t.Errorf("cost didn't round-trip")
	}
	if !out.CreatedAt.Equal(in.CreatedAt) {
		t.Errorf("created_at didn't round-trip")
	}
	if !out.IsActive {
		t.Errorf("is_active didn't round-trip")
	}
}
