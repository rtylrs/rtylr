package api

import (
	"context"
	"testing"
)

func TestCreateLeadHandler_Validate_RequiresCompanyName(t *testing.T) {
	h := &CreateLeadHandler{}
	if err := h.Validate(context.Background(), &CreateLeadRequest{CompanyName: ""}); err == nil {
		t.Fatal("expected error when company_name is empty, got nil")
	}
	if err := h.Validate(context.Background(), &CreateLeadRequest{CompanyName: "   "}); err == nil {
		t.Fatal("expected error when company_name is whitespace, got nil")
	}
	if err := h.Validate(context.Background(), &CreateLeadRequest{CompanyName: "Acme Corp"}); err != nil {
		t.Fatalf("expected nil for valid company_name, got %v", err)
	}
}

func TestUpdateLeadHandler_Validate_AllowsPartialUpdateWithoutCompanyName(t *testing.T) {
	h := &UpdateLeadHandler{}
	// Partial updates (e.g. stage move, notes update) should not require company_name
	if err := h.Validate(context.Background(), &UpdateLeadRequest{Stage: "contacted"}); err != nil {
		t.Fatalf("expected nil when updating stage without company_name, got %v", err)
	}
	if err := h.Validate(context.Background(), &UpdateLeadRequest{CompanyName: ""}); err != nil {
		t.Fatalf("expected nil for empty company_name in partial update, got %v", err)
	}
	if err := h.Validate(context.Background(), &UpdateLeadRequest{CompanyName: "New Name"}); err != nil {
		t.Fatalf("expected nil when updating company_name, got %v", err)
	}
}
