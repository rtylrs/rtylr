package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/voxire/onion/internal/modules/support/tickets/internal/domain"
	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"
	supportticketsv1 "github.com/voxire/voxie/gen/go/support/tickets/v1"
)

func statusOf(t *testing.T, err error) int {
	t.Helper()
	var oneKitErr *onekithttp.Error
	if !errors.As(err, &oneKitErr) {
		t.Fatalf("expected oneKit HTTP error, got %T: %v", err, err)
	}
	return int(oneKitErr.GetStatusCode())
}

// Invalid input is rejected with 400 before the business lookup; valid input
// gets past validation and, with no signed-in user, stops at 403.
func TestTicketAPIValidatesBeforeAuth(t *testing.T) {
	api := &TicketAPI{}
	ctx := context.Background()
	assignee := int64(5)

	cases := []struct {
		name string
		call func() error
		want int
	}{
		{"get without id", func() error { _, err := api.GetTicket(ctx, &supportticketsv1.GetTicketRequest{}); return err }, http.StatusBadRequest},
		{"get with id", func() error {
			_, err := api.GetTicket(ctx, &supportticketsv1.GetTicketRequest{Id: 9_007_199_254_740_993})
			return err
		}, http.StatusForbidden},
		{"create without subject", func() error {
			_, err := api.CreateTicket(ctx, &supportticketsv1.CreateTicketRequest{CustomerEmail: "a@b.co"})
			return err
		}, http.StatusBadRequest},
		{"create with whitespace subject", func() error {
			_, err := api.CreateTicket(ctx, &supportticketsv1.CreateTicketRequest{Subject: "  ", CustomerEmail: "a@b.co"})
			return err
		}, http.StatusBadRequest},
		{"create without email", func() error {
			_, err := api.CreateTicket(ctx, &supportticketsv1.CreateTicketRequest{Subject: "Help"})
			return err
		}, http.StatusBadRequest},
		{"create with whitespace email", func() error {
			_, err := api.CreateTicket(ctx, &supportticketsv1.CreateTicketRequest{Subject: "Help", CustomerEmail: "   "})
			return err
		}, http.StatusBadRequest},
		{"create valid with optional fields", func() error {
			_, err := api.CreateTicket(ctx, &supportticketsv1.CreateTicketRequest{
				Subject: "Help", CustomerEmail: "a@b.co", Priority: "high", Channel: "email", AssignedTo: &assignee, Body: "hi",
			})
			return err
		}, http.StatusForbidden},
		{"update without id", func() error { _, err := api.UpdateTicket(ctx, &supportticketsv1.UpdateTicketRequest{}); return err }, http.StatusBadRequest},
		{"update with id", func() error {
			_, err := api.UpdateTicket(ctx, &supportticketsv1.UpdateTicketRequest{Id: 10, Subject: "Updated"})
			return err
		}, http.StatusForbidden},
		{"list without user", func() error { _, err := api.ListTickets(ctx, &supportticketsv1.ListTicketsRequest{}); return err }, http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := statusOf(t, tc.call()); got != tc.want {
				t.Fatalf("status = %d, want %d", got, tc.want)
			}
		})
	}
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
			if got := statusOf(t, supportTicketLookupError(tt.err)); got != tt.want {
				t.Fatalf("status = %d, want %d", got, tt.want)
			}
		})
	}
}
