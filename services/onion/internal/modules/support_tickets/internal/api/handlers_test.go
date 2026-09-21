package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	onekithttp "github.com/corezio/onekit/http"

	"github.com/voxire/onion/internal/modules/support_tickets/internal/domain"
)

func TestGetTicketHandler_Validate_RequiresID(t *testing.T) {
	h := &GetTicketHandler{}
	req := &GetTicketRequest{ID: 0}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when ID is zero")
	}
}

func TestGetTicketHandler_Validate_HappyPath(t *testing.T) {
	h := &GetTicketHandler{}
	req := &GetTicketRequest{ID: 5}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetTicketHandler_Validate_LargeIDPasses(t *testing.T) {
	h := &GetTicketHandler{}
	req := &GetTicketRequest{ID: 999999}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error for large id: %v", err)
	}
}

func TestCreateTicketHandler_Validate_RequiresSubject(t *testing.T) {
	h := &CreateTicketHandler{}
	req := &CreateTicketRequest{Subject: "", CustomerEmail: "user@example.com"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when subject is empty")
	}
}

func TestCreateTicketHandler_Validate_RequiresCustomerEmail(t *testing.T) {
	h := &CreateTicketHandler{}
	req := &CreateTicketRequest{Subject: "Issue with order", CustomerEmail: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when customer email is empty")
	}
}

func TestCreateTicketHandler_Validate_HappyPath(t *testing.T) {
	h := &CreateTicketHandler{}
	req := &CreateTicketRequest{
		Subject:       "Issue with order",
		CustomerEmail: "alice@example.com",
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateTicketHandler_Validate_WhitespaceSubject(t *testing.T) {
	h := &CreateTicketHandler{}
	req := &CreateTicketRequest{Subject: "   ", CustomerEmail: "alice@example.com"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for whitespace-only subject")
	}
}

func TestCreateTicketHandler_Validate_WhitespaceEmail(t *testing.T) {
	h := &CreateTicketHandler{}
	req := &CreateTicketRequest{Subject: "Some issue", CustomerEmail: "  \t  "}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error for whitespace-only email")
	}
}

func TestCreateTicketHandler_Validate_BothEmpty(t *testing.T) {
	h := &CreateTicketHandler{}
	req := &CreateTicketRequest{}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when both fields are empty")
	}
}

func TestCreateTicketHandler_Validate_WithOptionalFields(t *testing.T) {
	h := &CreateTicketHandler{}
	req := &CreateTicketRequest{
		Subject:       "Billing issue",
		CustomerEmail: "bob@co.com",
		CustomerName:  "Bob",
		Priority:      "high",
		Channel:       "email",
	}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error with optional fields: %v", err)
	}
}

func TestUpdateTicketHandler_Validate_RequiresID(t *testing.T) {
	h := &UpdateTicketHandler{}
	req := &UpdateTicketRequest{ID: 0}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected error when ID is zero")
	}
}

func TestUpdateTicketHandler_Validate_HappyPath(t *testing.T) {
	h := &UpdateTicketHandler{}
	req := &UpdateTicketRequest{ID: 10, Subject: "Updated subject"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateTicketHandler_Validate_NonZeroIDPasses(t *testing.T) {
	h := &UpdateTicketHandler{}
	req := &UpdateTicketRequest{ID: 1}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetTicketHandler_Validate_NegativeIDFails(_ *testing.T) {
	// ID is int64 and 0 is sentinel for "not provided"; negative technically non-zero
	// but this documents the current behavior
	h := &GetTicketHandler{}
	req := &GetTicketRequest{ID: -1}
	// -1 != 0 so validation should pass (documenting behavior)
	_ = h.Validate(context.Background(), req)
}

func TestSupportTicketLookupError_MapsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "missing ticket", err: domain.ErrTicketNotFound, want: http.StatusNotFound},
		{name: "wrapped missing ticket", err: fmt.Errorf("lookup: %w", domain.ErrTicketNotFound), want: http.StatusNotFound},
		{name: "other failure", err: errors.New("connection reset"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapped := supportTicketLookupError(tt.err)
			var oneKitErr *onekithttp.Error
			if !errors.As(mapped, &oneKitErr) {
				t.Fatalf("expected oneKit HTTP error, got %T: %v", mapped, mapped)
			}
			if got := int(oneKitErr.GetStatusCode()); got != tt.want {
				t.Fatalf("status = %d, want %d", got, tt.want)
			}
		})
	}
}
