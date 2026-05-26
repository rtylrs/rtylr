package application

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/console/internal/infrastructure"
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
	row := infrastructure.UserRow{ID: 1, Email: "admin@example.com", IsConsoleAdmin: true}
	dto := toUserDTO(row)
	if dto.ID != 1 || dto.Email != "admin@example.com" || !dto.IsConsoleAdmin {
		t.Errorf("basic fields mismatch: %+v", dto)
	}
}

func TestToUserDTO_FullNameWhenValid(t *testing.T) {
	row := infrastructure.UserRow{FullName: sql.NullString{String: "Alice Smith", Valid: true}}
	dto := toUserDTO(row)
	if dto.FullName != "Alice Smith" {
		t.Errorf("expected FullName Alice Smith, got %q", dto.FullName)
	}
}

func TestToUserDTO_FullNameEmptyWhenNull(t *testing.T) {
	row := infrastructure.UserRow{FullName: sql.NullString{Valid: false}}
	dto := toUserDTO(row)
	if dto.FullName != "" {
		t.Errorf("expected empty FullName, got %q", dto.FullName)
	}
}

func TestToUserDTO_BusinessIDWhenValid(t *testing.T) {
	biz := int64(42)
	row := infrastructure.UserRow{BusinessID: sql.NullInt64{Int64: biz, Valid: true}}
	dto := toUserDTO(row)
	if dto.BusinessID == nil || *dto.BusinessID != biz {
		t.Errorf("expected BusinessID %d, got %v", biz, dto.BusinessID)
	}
}

func TestToUserDTO_BusinessIDNilWhenNull(t *testing.T) {
	row := infrastructure.UserRow{BusinessID: sql.NullInt64{Valid: false}}
	dto := toUserDTO(row)
	if dto.BusinessID != nil {
		t.Errorf("expected nil BusinessID, got %v", dto.BusinessID)
	}
}

func TestToUserDTO_BanReasonWhenValid(t *testing.T) {
	reason := "spam"
	row := infrastructure.UserRow{
		IsBanned:  true,
		BanReason: sql.NullString{String: reason, Valid: true},
	}
	dto := toUserDTO(row)
	if dto.BanReason == nil || *dto.BanReason != reason {
		t.Errorf("expected BanReason %q, got %v", reason, dto.BanReason)
	}
}

func TestToUserDTO_IsBanned(t *testing.T) {
	row := infrastructure.UserRow{IsBanned: true}
	dto := toUserDTO(row)
	if !dto.IsBanned {
		t.Error("expected IsBanned=true")
	}
}

func TestToUserDTO_BannedAtFormattedWhenValid(t *testing.T) {
	ts := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	row := infrastructure.UserRow{BannedAt: sql.NullTime{Time: ts, Valid: true}}
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
	row := infrastructure.UserRow{CreatedAt: sql.NullTime{Time: ts, Valid: true}}
	dto := toUserDTO(row)
	if !strings.Contains(dto.CreatedAt, "2025-06-01") {
		t.Errorf("unexpected CreatedAt format: %q", dto.CreatedAt)
	}
}

// ── toSubDTO ──────────────────────────────────────────────────────────────────

func TestToSubDTO_BasicFields(t *testing.T) {
	row := infrastructure.SubscriptionRow{ID: 5, OwnerID: 10, PlanID: 2, PlanName: "Pro", Status: "active", SeatCount: 3}
	dto := toSubDTO(row)
	if dto.ID != 5 || dto.Status != "active" || dto.SeatCount != 3 {
		t.Errorf("basic fields mismatch: %+v", dto)
	}
}

func TestToSubDTO_OwnerEmailWhenValid(t *testing.T) {
	row := infrastructure.SubscriptionRow{OwnerEmail: sql.NullString{String: "owner@co.com", Valid: true}}
	dto := toSubDTO(row)
	if dto.OwnerEmail != "owner@co.com" {
		t.Errorf("expected owner@co.com, got %q", dto.OwnerEmail)
	}
}

func TestToSubDTO_TrialEndsAtFormattedWhenValid(t *testing.T) {
	ts := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	row := infrastructure.SubscriptionRow{TrialEndsAt: sql.NullTime{Time: ts, Valid: true}}
	dto := toSubDTO(row)
	if dto.TrialEndsAt == nil || !strings.Contains(*dto.TrialEndsAt, "2026-03-01") {
		t.Errorf("unexpected TrialEndsAt: %v", dto.TrialEndsAt)
	}
}

// ── toGrantDTO ────────────────────────────────────────────────────────────────

func TestToGrantDTO_BasicFields(t *testing.T) {
	row := infrastructure.GrantRow{ID: 7, UserID: 3, PlanID: 1, PlanName: "Enterprise", GrantedBy: 99}
	dto := toGrantDTO(row)
	if dto.ID != 7 || dto.UserID != 3 || dto.PlanName != "Enterprise" || dto.GrantedBy != 99 {
		t.Errorf("basic fields mismatch: %+v", dto)
	}
}

func TestToGrantDTO_NoteWhenValid(t *testing.T) {
	row := infrastructure.GrantRow{Note: sql.NullString{String: "partner deal", Valid: true}}
	dto := toGrantDTO(row)
	if dto.Note != "partner deal" {
		t.Errorf("expected note, got %q", dto.Note)
	}
}

func TestToGrantDTO_ExpiresAtFormattedWhenValid(t *testing.T) {
	ts := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	row := infrastructure.GrantRow{ExpiresAt: sql.NullTime{Time: ts, Valid: true}}
	dto := toGrantDTO(row)
	if dto.ExpiresAt == nil || !strings.Contains(*dto.ExpiresAt, "2027-01-01") {
		t.Errorf("unexpected ExpiresAt: %v", dto.ExpiresAt)
	}
}

func TestToGrantDTO_RevokedAtNilWhenNull(t *testing.T) {
	row := infrastructure.GrantRow{RevokedAt: sql.NullTime{Valid: false}}
	dto := toGrantDTO(row)
	if dto.RevokedAt != nil {
		t.Errorf("expected nil RevokedAt, got %v", dto.RevokedAt)
	}
}
