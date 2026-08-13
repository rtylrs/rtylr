package realtime

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
)

func TestIssueRealtimeTicketRequiresAuthenticatedUser(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/flow/ws-ticket", http.NoBody)
	res := httptest.NewRecorder()

	IssueRealtimeTicket(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without auth context, got %d", res.Code)
	}
}

func TestIssueRealtimeTicketReturnsNoStoreOpaqueTicket(t *testing.T) {
	ctx := authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		UserID:     42,
		BusinessID: 9,
		Role:       "staff",
		Token:      "the-bearer-jwt-must-not-be-copied",
	})
	req := httptest.NewRequest(http.MethodPost, "/flow/ws-ticket", http.NoBody).WithContext(ctx)
	res := httptest.NewRecorder()

	IssueRealtimeTicket(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
	if got := res.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("expected no-store response, got %q", got)
	}

	var body struct {
		Ticket string `json:"ticket"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode ticket response: %v", err)
	}
	if body.Ticket == "" || body.Ticket == "the-bearer-jwt-must-not-be-copied" {
		t.Fatalf("unexpected ticket response: %+v", body)
	}

	user, ok := authPkg.ConsumeWebSocketTicket(context.Background(), body.Ticket)
	if !ok || user.UserID != 42 || user.BusinessID != 9 {
		t.Fatalf("ticket did not preserve authenticated claims: user=%+v ok=%t", user, ok)
	}
}
