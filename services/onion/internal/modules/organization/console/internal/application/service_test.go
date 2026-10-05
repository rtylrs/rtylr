package application

import (
	"strings"
	"testing"
	"time"
)

// ── consoleKey ────────────────────────────────────────────────────────────────

func TestConsoleKey_HasExpectedPrefix(t *testing.T) {
	key := consoleKey("users|l=10|o=0")
	if !strings.HasPrefix(key, "cache:console:") {
		t.Errorf("expected cache:console: prefix, got %q", key)
	}
}

func TestConsoleKey_AppendsDetail(t *testing.T) {
	key := consoleKey("my_detail")
	if key != "cache:console:my_detail" {
		t.Errorf("unexpected key: %q", key)
	}
}

func TestConsoleKey_EmptyDetailStillHasPrefix(t *testing.T) {
	key := consoleKey("")
	if key != "cache:console:" {
		t.Errorf("unexpected key for empty detail: %q", key)
	}
}

// ── toUserDTO ─────────────────────────────────────────────────────────────────

func TestToUserDTO_BasicFields(t *testing.T) {
	row := UserRow{ID: 1, Email: "admin@example.com", IsConsoleAdmin: true}
	dto := toUserDTO(row)
	if dto.ID != 1 || dto.Email != "admin@example.com" || !dto.IsConsoleAdmin {
		t.Errorf("basic fields mismatch: %+v", dto)
	}
}

func TestToUserDTO_FullNameWhenValid(t *testing.T) {
	row := UserRow{FullName: NullString{String: "Alice Smith", Valid: true}}
	dto := toUserDTO(row)
	if dto.FullName != "Alice Smith" {
		t.Errorf("expected FullName Alice Smith, got %q", dto.FullName)
	}
}

func TestToUserDTO_FullNameEmptyWhenNull(t *testing.T) {
	row := UserRow{FullName: NullString{Valid: false}}
	dto := toUserDTO(row)
	if dto.FullName != "" {
		t.Errorf("expected empty FullName, got %q", dto.FullName)
	}
}

func TestToUserDTO_BusinessIDWhenValid(t *testing.T) {
	biz := int64(42)
	row := UserRow{BusinessID: NullInt64{Int64: biz, Valid: true}}
	dto := toUserDTO(row)
	if dto.BusinessID == nil || *dto.BusinessID != biz {
		t.Errorf("expected BusinessID %d, got %v", biz, dto.BusinessID)
	}
}

func TestToUserDTO_BusinessIDNilWhenNull(t *testing.T) {
	row := UserRow{BusinessID: NullInt64{Valid: false}}
	dto := toUserDTO(row)
	if dto.BusinessID != nil {
		t.Errorf("expected nil BusinessID, got %v", dto.BusinessID)
	}
}

func TestToUserDTO_BanReasonWhenValid(t *testing.T) {
	reason := "spam"
	row := UserRow{
		IsBanned:  true,
		BanReason: NullString{String: reason, Valid: true},
	}
	dto := toUserDTO(row)
	if dto.BanReason == nil || *dto.BanReason != reason {
		t.Errorf("expected BanReason %q, got %v", reason, dto.BanReason)
	}
}

func TestToUserDTO_IsBanned(t *testing.T) {
	row := UserRow{IsBanned: true}
	dto := toUserDTO(row)
	if !dto.IsBanned {
		t.Error("expected IsBanned=true")
	}
}

func TestToUserDTO_BannedAtFormattedWhenValid(t *testing.T) {
	ts := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	row := UserRow{BannedAt: NullTime{Time: ts, Valid: true}}
	dto := toUserDTO(row)
	if dto.BannedAt == nil {
		t.Fatal("expected BannedAt to be set")
	}
	if !strings.Contains(*dto.BannedAt, "2026-01-15") {
		t.Errorf("unexpected BannedAt format: %q", *dto.BannedAt)
	}
}

func TestToUserDTO_CreatedAtFormattedWhenValid(t *testing.T) {
	ts := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	row := UserRow{CreatedAt: NullTime{Time: ts, Valid: true}}
	dto := toUserDTO(row)
	if !strings.Contains(dto.CreatedAt, "2025-06-01") {
		t.Errorf("unexpected CreatedAt format: %q", dto.CreatedAt)
	}
}

// ── toSubDTO ──────────────────────────────────────────────────────────────────

func TestToSubDTO_BasicFields(t *testing.T) {
	row := SubscriptionRow{ID: 5, OwnerID: 10, PlanID: 2, PlanName: "Pro", Status: "active", SeatCount: 3}
	dto := toSubDTO(row)
	if dto.ID != 5 || dto.Status != "active" || dto.SeatCount != 3 {
		t.Errorf("basic fields mismatch: %+v", dto)
	}
}

func TestToSubDTO_OwnerEmailWhenValid(t *testing.T) {
	row := SubscriptionRow{OwnerEmail: NullString{String: "owner@co.com", Valid: true}}
	dto := toSubDTO(row)
	if dto.OwnerEmail != "owner@co.com" {
		t.Errorf("expected owner@co.com, got %q", dto.OwnerEmail)
	}
}

func TestToSubDTO_TrialEndsAtFormattedWhenValid(t *testing.T) {
	ts := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	row := SubscriptionRow{TrialEndsAt: NullTime{Time: ts, Valid: true}}
	dto := toSubDTO(row)
	if dto.TrialEndsAt == nil || !strings.Contains(*dto.TrialEndsAt, "2026-03-01") {
		t.Errorf("unexpected TrialEndsAt: %v", dto.TrialEndsAt)
	}
}

// ── toGrantDTO ────────────────────────────────────────────────────────────────

func TestToGrantDTO_BasicFields(t *testing.T) {
	row := GrantRow{ID: 7, UserID: 3, PlanID: 1, PlanName: "Enterprise", GrantedBy: 99}
	dto := toGrantDTO(row)
	if dto.ID != 7 || dto.UserID != 3 || dto.PlanName != "Enterprise" || dto.GrantedBy != 99 {
		t.Errorf("basic fields mismatch: %+v", dto)
	}
}

func TestToGrantDTO_NoteWhenValid(t *testing.T) {
	row := GrantRow{Note: NullString{String: "partner deal", Valid: true}}
	dto := toGrantDTO(row)
	if dto.Note != "partner deal" {
		t.Errorf("expected note, got %q", dto.Note)
	}
}

func TestToGrantDTO_ExpiresAtFormattedWhenValid(t *testing.T) {
	ts := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	row := GrantRow{ExpiresAt: NullTime{Time: ts, Valid: true}}
	dto := toGrantDTO(row)
	if dto.ExpiresAt == nil || !strings.Contains(*dto.ExpiresAt, "2027-01-01") {
		t.Errorf("unexpected ExpiresAt: %v", dto.ExpiresAt)
	}
}

func TestToGrantDTO_RevokedAtNilWhenNull(t *testing.T) {
	row := GrantRow{RevokedAt: NullTime{Valid: false}}
	dto := toGrantDTO(row)
	if dto.RevokedAt != nil {
		t.Errorf("expected nil RevokedAt, got %v", dto.RevokedAt)
	}
}
