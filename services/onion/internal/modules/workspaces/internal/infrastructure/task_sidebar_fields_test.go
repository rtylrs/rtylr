package infrastructure_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// readEffortMinutes reads the column straight out of the database rather than
// back through the repository, so these tests prove the write reached MySQL and
// are not just asserting that a struct round-tripped.
func readEffortMinutes(t *testing.T, db *sql.DB, taskID int64) *int32 {
	t.Helper()
	var got sql.NullInt32
	if err := db.QueryRowContext(context.Background(),
		"SELECT effort_minutes FROM flow_tasks WHERE id = ?", taskID,
	).Scan(&got); err != nil {
		t.Fatalf("scan effort_minutes: %v", err)
	}
	if !got.Valid {
		return nil
	}
	v := got.Int32
	return &v
}

func TestWorkspaces_UpdateTask_PersistsEffortMinutes(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	taskID, businessID := seedArchivedTestTask(t, db, repo, "EffortPersist")

	effort := int32(30)
	if err := repo.UpdateTask(ctx, domain.UpdateTaskInput{
		ID: taskID, BusinessID: businessID, EffortMinutes: &effort,
	}); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	got := readEffortMinutes(t, db, taskID)
	if got == nil || *got != 30 {
		t.Fatalf("effort_minutes = %v, want 30", got)
	}
}

func TestWorkspaces_UpdateTask_OmittedEffortMinutesLeavesValue(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	taskID, businessID := seedArchivedTestTask(t, db, repo, "EffortPreserve")

	effort := int32(45)
	if err := repo.UpdateTask(ctx, domain.UpdateTaskInput{
		ID: taskID, BusinessID: businessID, EffortMinutes: &effort,
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	// A title-only patch must not disturb effort_minutes. This is the exact
	// class of bug COALESCE(?, col) exists to prevent.
	if err := repo.UpdateTask(ctx, domain.UpdateTaskInput{
		ID: taskID, BusinessID: businessID, Title: "renamed",
	}); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	got := readEffortMinutes(t, db, taskID)
	if got == nil || *got != 45 {
		t.Fatalf("effort_minutes = %v, want 45 preserved", got)
	}
}

func TestWorkspaces_UpdateTask_ZeroEffortMinutesClears(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	taskID, businessID := seedArchivedTestTask(t, db, repo, "EffortClear")

	seed := int32(60)
	if err := repo.UpdateTask(ctx, domain.UpdateTaskInput{
		ID: taskID, BusinessID: businessID, EffortMinutes: &seed,
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	// 0 is the clear sentinel, not a value. Without this the "no effort" chip
	// in the sidebar would silently no-op.
	zero := int32(0)
	if err := repo.UpdateTask(ctx, domain.UpdateTaskInput{
		ID: taskID, BusinessID: businessID, EffortMinutes: &zero,
	}); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	if got := readEffortMinutes(t, db, taskID); got != nil {
		t.Fatalf("effort_minutes = %v, want NULL after clear", *got)
	}
}

// readWaitingOn reads the column straight out of the database rather than
// back through the repository, so these tests prove the write reached MySQL.
func readWaitingOn(t *testing.T, db *sql.DB, taskID int64) *string {
	t.Helper()
	var got sql.NullString
	if err := db.QueryRowContext(context.Background(),
		"SELECT waiting_on FROM flow_tasks WHERE id = ?", taskID,
	).Scan(&got); err != nil {
		t.Fatalf("scan waiting_on: %v", err)
	}
	if !got.Valid {
		return nil
	}
	v := got.String
	return &v
}

func TestWorkspaces_UpdateTask_PersistsWaitingOn(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	taskID, businessID := seedArchivedTestTask(t, db, repo, "WaitingOnPersist")

	waitingOn := "Legal approval"
	if err := repo.UpdateTask(ctx, domain.UpdateTaskInput{
		ID: taskID, BusinessID: businessID, WaitingOn: &waitingOn,
	}); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	got := readWaitingOn(t, db, taskID)
	if got == nil || *got != "Legal approval" {
		t.Fatalf("waiting_on = %v, want %q", got, "Legal approval")
	}
}

func TestWorkspaces_UpdateTask_OmittedWaitingOnLeavesValue(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	taskID, businessID := seedArchivedTestTask(t, db, repo, "WaitingOnPreserve")

	waitingOn := "Finance sign-off"
	if err := repo.UpdateTask(ctx, domain.UpdateTaskInput{
		ID: taskID, BusinessID: businessID, WaitingOn: &waitingOn,
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	// A title-only patch must not disturb waiting_on. This is the exact class
	// of bug COALESCE(?, col) exists to prevent.
	if err := repo.UpdateTask(ctx, domain.UpdateTaskInput{
		ID: taskID, BusinessID: businessID, Title: "renamed",
	}); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	got := readWaitingOn(t, db, taskID)
	if got == nil || *got != "Finance sign-off" {
		t.Fatalf("waiting_on = %v, want %q preserved", got, "Finance sign-off")
	}
}

func TestWorkspaces_UpdateTask_EmptyWaitingOnClears(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	taskID, businessID := seedArchivedTestTask(t, db, repo, "WaitingOnClear")

	seed := "Ops review"
	if err := repo.UpdateTask(ctx, domain.UpdateTaskInput{
		ID: taskID, BusinessID: businessID, WaitingOn: &seed,
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	// "" is the clear sentinel, not a value. Without this the "clear blocker"
	// action in the sidebar would silently no-op.
	empty := ""
	if err := repo.UpdateTask(ctx, domain.UpdateTaskInput{
		ID: taskID, BusinessID: businessID, WaitingOn: &empty,
	}); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	if got := readWaitingOn(t, db, taskID); got != nil {
		t.Fatalf("waiting_on = %v, want NULL after clear", *got)
	}
}

// readDecisionNote reads the column straight out of the database rather than
// back through the repository, so these tests prove the write reached MySQL.
func readDecisionNote(t *testing.T, db *sql.DB, taskID int64) *string {
	t.Helper()
	var got sql.NullString
	if err := db.QueryRowContext(context.Background(),
		"SELECT decision_note FROM flow_tasks WHERE id = ?", taskID,
	).Scan(&got); err != nil {
		t.Fatalf("scan decision_note: %v", err)
	}
	if !got.Valid {
		return nil
	}
	v := got.String
	return &v
}

func TestWorkspaces_UpdateTask_PersistsDecisionNote(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	taskID, businessID := seedArchivedTestTask(t, db, repo, "DecisionNotePersist")

	note := "Chose vendor B for cost reasons"
	if err := repo.UpdateTask(ctx, domain.UpdateTaskInput{
		ID: taskID, BusinessID: businessID, DecisionNote: &note,
	}); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	got := readDecisionNote(t, db, taskID)
	if got == nil || *got != "Chose vendor B for cost reasons" {
		t.Fatalf("decision_note = %v, want %q", got, "Chose vendor B for cost reasons")
	}
}

func TestWorkspaces_UpdateTask_OmittedDecisionNoteLeavesValue(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	taskID, businessID := seedArchivedTestTask(t, db, repo, "DecisionNotePreserve")

	note := "Deferred to next sprint"
	if err := repo.UpdateTask(ctx, domain.UpdateTaskInput{
		ID: taskID, BusinessID: businessID, DecisionNote: &note,
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	// A title-only patch must not disturb decision_note. This is the exact
	// class of bug COALESCE(?, col) exists to prevent.
	if err := repo.UpdateTask(ctx, domain.UpdateTaskInput{
		ID: taskID, BusinessID: businessID, Title: "renamed",
	}); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	got := readDecisionNote(t, db, taskID)
	if got == nil || *got != "Deferred to next sprint" {
		t.Fatalf("decision_note = %v, want %q preserved", got, "Deferred to next sprint")
	}
}

func TestWorkspaces_UpdateTask_EmptyDecisionNoteClears(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	taskID, businessID := seedArchivedTestTask(t, db, repo, "DecisionNoteClear")

	seed := "Picked plan A"
	if err := repo.UpdateTask(ctx, domain.UpdateTaskInput{
		ID: taskID, BusinessID: businessID, DecisionNote: &seed,
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	// "" is the clear sentinel, not a value. Without this the "clear decision"
	// action in the sidebar would silently no-op.
	empty := ""
	if err := repo.UpdateTask(ctx, domain.UpdateTaskInput{
		ID: taskID, BusinessID: businessID, DecisionNote: &empty,
	}); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	if got := readDecisionNote(t, db, taskID); got != nil {
		t.Fatalf("decision_note = %v, want NULL after clear", *got)
	}
}

// readEstimatedHours reads the column straight out of the database rather
// than back through the repository, so these tests prove the write reached
// MySQL and are not just asserting that a struct round-tripped.
func readEstimatedHours(t *testing.T, db *sql.DB, taskID int64) *float64 {
	t.Helper()
	var got sql.NullFloat64
	if err := db.QueryRowContext(context.Background(),
		"SELECT estimated_hours FROM flow_tasks WHERE id = ?", taskID,
	).Scan(&got); err != nil {
		t.Fatalf("scan estimated_hours: %v", err)
	}
	if !got.Valid {
		return nil
	}
	v := got.Float64
	return &v
}

func TestWorkspaces_UpdateTask_PersistsEstimatedHours(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	taskID, businessID := seedArchivedTestTask(t, db, repo, "EstimatedHoursPersist")

	hours := 3.5
	if err := repo.UpdateTask(ctx, domain.UpdateTaskInput{
		ID: taskID, BusinessID: businessID, EstimatedHours: &hours,
	}); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	got := readEstimatedHours(t, db, taskID)
	if got == nil || *got != 3.5 {
		t.Fatalf("estimated_hours = %v, want 3.5", got)
	}

	// estimated_hours already has a Task read field, so also assert it
	// through the read model, not only the raw column.
	task, err := repo.GetTaskByID(ctx, taskID)
	if err != nil {
		t.Fatalf("GetTaskByID: %v", err)
	}
	if task.EstimatedHours == nil || *task.EstimatedHours != 3.5 {
		t.Fatalf("GetTaskByID EstimatedHours = %v, want 3.5", task.EstimatedHours)
	}
}

func TestWorkspaces_UpdateTask_OmittedEstimatedHoursLeavesValue(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	taskID, businessID := seedArchivedTestTask(t, db, repo, "EstimatedHoursPreserve")

	hours := 5.25
	if err := repo.UpdateTask(ctx, domain.UpdateTaskInput{
		ID: taskID, BusinessID: businessID, EstimatedHours: &hours,
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	// A title-only patch must not disturb estimated_hours. This is the exact
	// class of bug COALESCE(?, col) exists to prevent.
	if err := repo.UpdateTask(ctx, domain.UpdateTaskInput{
		ID: taskID, BusinessID: businessID, Title: "renamed",
	}); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	got := readEstimatedHours(t, db, taskID)
	if got == nil || *got != 5.25 {
		t.Fatalf("estimated_hours = %v, want 5.25 preserved", got)
	}
}

func TestWorkspaces_UpdateTask_ZeroEstimatedHoursClears(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	taskID, businessID := seedArchivedTestTask(t, db, repo, "EstimatedHoursClear")

	seed := 8.0
	if err := repo.UpdateTask(ctx, domain.UpdateTaskInput{
		ID: taskID, BusinessID: businessID, EstimatedHours: &seed,
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	// 0 is the clear sentinel, not a value. Without this the "no estimate"
	// chip in the sidebar would silently no-op.
	zero := 0.0
	if err := repo.UpdateTask(ctx, domain.UpdateTaskInput{
		ID: taskID, BusinessID: businessID, EstimatedHours: &zero,
	}); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	if got := readEstimatedHours(t, db, taskID); got != nil {
		t.Fatalf("estimated_hours = %v, want NULL after clear", *got)
	}
}

// countFocusPins reads the join table directly. The per-user guarantee is the
// whole point of this table, so it has to be asserted at the storage layer.
func countFocusPins(t *testing.T, db *sql.DB, taskID, userID int64) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM flow_task_focus_pins WHERE task_id = ? AND user_id = ?",
		taskID, userID,
	).Scan(&n); err != nil {
		t.Fatalf("count focus pins: %v", err)
	}
	return n
}

func TestWorkspaces_SetTaskFocusPin_IsPerUser(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	taskID, businessID := seedArchivedTestTask(t, db, repo, "PinPerUser")
	const alice, bob = int64(90101), int64(90102)
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_task_focus_pins WHERE task_id = ?", taskID)
	})

	// Two users must be able to hold independent pins on the same task, and
	// one user unpinning must not disturb another's. A test that only checks
	// the pinning user's own row would pass even if the pin were stored per
	// task instead of per (task, user): bob's row would just never exist, so
	// his count would trivially read 0 regardless of whether his write ever
	// happened. Exercising both writes is what actually exercises the unique
	// key on (task_id, user_id).
	if err := repo.SetTaskFocusPin(ctx, taskID, alice, businessID, true); err != nil {
		t.Fatalf("pin for alice: %v", err)
	}
	if err := repo.SetTaskFocusPin(ctx, taskID, bob, businessID, true); err != nil {
		t.Fatalf("pin for bob: %v", err)
	}

	if got := countFocusPins(t, db, taskID, alice); got != 1 {
		t.Fatalf("alice pin count = %d, want 1", got)
	}
	if got := countFocusPins(t, db, taskID, bob); got != 1 {
		t.Fatalf("bob pin count = %d, want 1: bob's pin must not be dropped as a duplicate of alice's", got)
	}
	var total int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM flow_task_focus_pins WHERE task_id = ?", taskID,
	).Scan(&total); err != nil {
		t.Fatalf("count total pins: %v", err)
	}
	if total != 2 {
		t.Fatalf("total pin rows for task = %d, want 2 (one per user)", total)
	}

	if err := repo.SetTaskFocusPin(ctx, taskID, alice, businessID, false); err != nil {
		t.Fatalf("unpin alice: %v", err)
	}
	if got := countFocusPins(t, db, taskID, alice); got != 0 {
		t.Fatalf("alice pin count after unpin = %d, want 0", got)
	}
	if got := countFocusPins(t, db, taskID, bob); got != 1 {
		t.Fatalf("bob pin count = %d, want 1: alice unpinning must not disturb bob's pin", got)
	}
}

func TestWorkspaces_SetTaskFocusPin_IsIdempotentAndUnpins(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	taskID, businessID := seedArchivedTestTask(t, db, repo, "PinIdempotent")
	const alice = int64(90101)
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_task_focus_pins WHERE task_id = ?", taskID)
	})

	for i := 0; i < 2; i++ {
		if err := repo.SetTaskFocusPin(ctx, taskID, alice, businessID, true); err != nil {
			t.Fatalf("pin %d: %v", i, err)
		}
	}
	if got := countFocusPins(t, db, taskID, alice); got != 1 {
		t.Fatalf("pin count after double pin = %d, want 1", got)
	}

	if err := repo.SetTaskFocusPin(ctx, taskID, alice, businessID, false); err != nil {
		t.Fatalf("unpin: %v", err)
	}
	if got := countFocusPins(t, db, taskID, alice); got != 0 {
		t.Fatalf("pin count after unpin = %d, want 0", got)
	}
}

func TestWorkspaces_GetFocusPinsIsBusinessAndUserScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	taskID, businessID := seedArchivedTestTask(t, db, repo, "PinReadScope")
	const alice = int64(90103)
	const otherBusiness = int64(9010301)
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_task_focus_pins WHERE task_id = ?", taskID)
	})

	if err := repo.SetTaskFocusPin(ctx, taskID, alice, businessID, true); err != nil {
		t.Fatalf("pin: %v", err)
	}
	pins, err := repo.GetFocusPins(ctx, alice, businessID, []int64{taskID})
	if err != nil {
		t.Fatalf("GetFocusPins: %v", err)
	}
	if !pins[taskID] {
		t.Fatal("caller's pin was not returned")
	}
	pins, err = repo.GetFocusPins(ctx, alice, otherBusiness, []int64{taskID})
	if err != nil {
		t.Fatalf("GetFocusPins other business: %v", err)
	}
	if pins[taskID] {
		t.Fatal("pin leaked across business scope")
	}
}
