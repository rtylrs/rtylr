package pos_cash_movements

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/pos_cash_movements/internal/domain"
)

func TestParseListCashMovementsToolInput(t *testing.T) {
	parsed, err := parseListCashMovementsToolInput(map[string]any{
		"session_id":  json.Number("33"),
		"type":        "cash out",
		"branch_id":   "4",
		"terminal_id": 6,
		"employee_id": 9,
		"page":        2,
		"limit":       10,
	})
	if err != nil {
		t.Fatalf("parseListCashMovementsToolInput: %v", err)
	}
	if parsed.SessionID != 33 {
		t.Fatalf("SessionID = %d", parsed.SessionID)
	}
	if parsed.Type != "cash_out" {
		t.Fatalf("Type = %q", parsed.Type)
	}
	if parsed.BranchID != 4 || parsed.TerminalID != 6 || parsed.EmployeeID != 9 {
		t.Fatalf("filters = branch:%d terminal:%d employee:%d", parsed.BranchID, parsed.TerminalID, parsed.EmployeeID)
	}
	if parsed.Page != 2 || parsed.Limit != 10 {
		t.Fatalf("page/limit = %d/%d", parsed.Page, parsed.Limit)
	}
}

func TestParseListCashMovementsToolInputRequiresSession(t *testing.T) {
	_, err := parseListCashMovementsToolInput(map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "session_id") {
		t.Fatalf("error = %v, want session_id validation", err)
	}
}

func TestParseListCashMovementsToolInputDefaultsAndCapsLimit(t *testing.T) {
	parsed, err := parseListCashMovementsToolInput(map[string]any{
		"session_id": 33,
		"limit":      500,
	})
	if err != nil {
		t.Fatalf("parseListCashMovementsToolInput: %v", err)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d", parsed.Page)
	}
	if parsed.Limit != maxListCashMovementsToolLimit {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
}

func TestParseListCashMovementsToolInputRejectsBadType(t *testing.T) {
	_, err := parseListCashMovementsToolInput(map[string]any{
		"session_id": 33,
		"type":       "refund",
	})
	if err == nil || !strings.Contains(err.Error(), "type") {
		t.Fatalf("error = %v, want type validation", err)
	}
}

func TestFilterCashMovementsForTool(t *testing.T) {
	terminalID := int64(8)
	movements := []domain.CashMovement{
		{ID: 1, Type: "drop", BranchID: 4, TerminalID: &terminalID, EmployeeID: 7},
		{ID: 2, Type: "pickup", BranchID: 4, TerminalID: &terminalID, EmployeeID: 7},
		{ID: 3, Type: "drop", BranchID: 5, EmployeeID: 7},
	}

	filtered := filterCashMovementsForTool(movements, listCashMovementsToolInput{
		Type:       "drop",
		BranchID:   4,
		TerminalID: terminalID,
		EmployeeID: 7,
	})
	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d", len(filtered))
	}
	if filtered[0].ID != 1 {
		t.Fatalf("filtered[0].ID = %d", filtered[0].ID)
	}
}

func TestCashMovementToolPayload(t *testing.T) {
	terminalID := int64(8)
	approvedBy := int64(11)
	reason := "End of shift"
	createdAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := cashMovementToolPayload(domain.CashMovement{
		ID:         5,
		BusinessID: 2,
		BranchID:   4,
		TerminalID: &terminalID,
		SessionID:  33,
		EmployeeID: 7,
		Type:       "drop",
		Amount:     120.5,
		Reason:     &reason,
		ApprovedBy: &approvedBy,
		CreatedAt:  createdAt,
	})

	if payload["terminal_id"] != terminalID {
		t.Fatalf("terminal_id = %#v", payload["terminal_id"])
	}
	if payload["reason"] != reason {
		t.Fatalf("reason = %#v", payload["reason"])
	}
	if payload["created_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("created_at = %#v", payload["created_at"])
	}
}
