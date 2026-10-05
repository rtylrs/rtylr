package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/pos/terminals/internal/domain"
	"github.com/voxire/onion/internal/modules/pos/terminals/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

const (
	terminalsBusinessID int64 = 1
	terminalsBranchID   int64 = 3001
)

func TestListTerminals_RunsWithoutError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.List(ctx, terminalsBusinessID)
	if err != nil {
		t.Fatalf("unexpected error listing terminals for business %d: %v", terminalsBusinessID, err)
	}
}

func TestListTerminals_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	terminals, err := repo.List(ctx, 0)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(terminals) != 0 {
		t.Errorf("expected empty slice for businessID=0, got %d rows", len(terminals))
	}
}

func TestListTerminals_IsNotNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	terminals, err := repo.List(ctx, terminalsBusinessID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if terminals == nil {
		t.Error("expected non-nil slice, got nil")
	}
}

func TestCreateTerminal_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.TerminalInput{
		BusinessID: terminalsBusinessID,
		BranchID:   terminalsBranchID,
		Name:       "Register 1",
	}

	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestCreateTerminal_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.TerminalInput{
		BusinessID: terminalsBusinessID,
		BranchID:   terminalsBranchID,
		Name:       "StoredTerminal",
	}

	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	terminals, err := repo.List(ctx, terminalsBusinessID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	for _, term := range terminals {
		if term.ID == id {
			if term.Name != "StoredTerminal" {
				t.Errorf("expected name='StoredTerminal', got %q", term.Name)
			}
			return
		}
	}
	t.Fatalf("created terminal id=%d not found in list", id)
}

func TestCreateTerminal_LocationIDIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.TerminalInput{
		BusinessID: terminalsBusinessID,
		BranchID:   terminalsBranchID,
		Name:       "BranchTerminal",
	}

	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	terminals, err := repo.List(ctx, terminalsBusinessID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	for _, term := range terminals {
		if term.ID == id {
			if term.BranchID != terminalsBranchID {
				t.Errorf("expected BranchID=%d, got %d", terminalsBranchID, term.BranchID)
			}
			return
		}
	}
	t.Fatalf("created terminal id=%d not found in list", id)
}

func TestCreateTerminal_TypeIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	deviceID := "DEV-001"
	input := domain.TerminalInput{
		BusinessID: terminalsBusinessID,
		BranchID:   terminalsBranchID,
		Name:       "TypeTerminal",
		DeviceID:   &deviceID,
	}

	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	terminals, err := repo.List(ctx, terminalsBusinessID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	for _, term := range terminals {
		if term.ID == id {
			if term.DeviceID == nil || *term.DeviceID != "DEV-001" {
				t.Errorf("expected DeviceID='DEV-001', got %v", term.DeviceID)
			}
			return
		}
	}
	t.Fatalf("created terminal id=%d not found in list", id)
}

func TestCreateTerminal_IsActiveDefault(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.TerminalInput{
		BusinessID: terminalsBusinessID,
		BranchID:   terminalsBranchID,
		Name:       "ActiveDefaultTerminal",
	}

	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	terminals, err := repo.List(ctx, terminalsBusinessID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	for _, term := range terminals {
		if term.ID == id {
			// New terminals should be active by default.
			if !term.IsActive {
				t.Errorf("expected IsActive=true for newly created terminal, got false")
			}
			return
		}
	}
	t.Fatalf("created terminal id=%d not found in list", id)
}

func TestUpdateTerminal_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.TerminalInput{
		BusinessID: terminalsBusinessID,
		BranchID:   terminalsBranchID,
		Name:       "OldTerminalName",
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	if err := repo.Update(ctx, domain.TerminalUpdateInput{
		ID:         id,
		BusinessID: terminalsBusinessID,
		BranchID:   terminalsBranchID,
		Name:       "NewTerminalName",
		IsActive:   true,
	}); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	terminals, err := repo.List(ctx, terminalsBusinessID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	for _, term := range terminals {
		if term.ID == id {
			if term.Name != "NewTerminalName" {
				t.Errorf("expected name='NewTerminalName', got %q", term.Name)
			}
			return
		}
	}
	t.Fatalf("updated terminal id=%d not found in list", id)
}

func TestHeartbeat_ExistingTerminal_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.TerminalInput{
		BusinessID: terminalsBusinessID,
		BranchID:   terminalsBranchID,
		Name:       "HeartbeatTerminal",
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	if err := repo.Heartbeat(ctx, id, terminalsBusinessID); err != nil {
		t.Errorf("Heartbeat on existing terminal: unexpected error: %v", err)
	}
}

func TestHeartbeat_NonExistentTerminal_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.Heartbeat(ctx, 999999999, terminalsBusinessID)
	if err == nil {
		t.Error("expected error for Heartbeat on non-existent terminal, got nil")
	}
}

func TestCreate_ThenList_FindsTerminal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.TerminalInput{
		BusinessID: terminalsBusinessID,
		BranchID:   terminalsBranchID,
		Name:       "RoundtripTerminal",
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	terminals, err := repo.List(ctx, terminalsBusinessID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	found := false
	for _, term := range terminals {
		if term.ID == id && term.Name == "RoundtripTerminal" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("roundtrip terminal id=%d with name='RoundtripTerminal' not found", id)
	}
}
