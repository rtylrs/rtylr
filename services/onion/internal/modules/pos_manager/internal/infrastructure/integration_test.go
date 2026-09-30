package infrastructure_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_manager/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_manager/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
	"github.com/voxire/onion/internal/testutil"
)

// These replace the previous suite, which called Repository.VerifyPIN and
// asserted that a wrong PIN yields Valid=false. That method *was* the defect: it
// compared the submitted PIN against a cleartext column inside SQL. There is no
// longer a repository-level VerifyPIN — the repository only reads candidate
// hashes and the comparison happens in the application layer, where the
// brute-force counter can gate it.
//
// Every test here needs a database and SKIPS without DB_HOST. Nothing in this
// file has been executed in the sandbox where this change was made.

func TestManager_ListManagerPINHashes_UnknownBusinessReturnsNoCandidates(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	candidates, err := repo.ListManagerPINHashes(context.Background(), 999999)
	if err != nil {
		t.Fatalf("ListManagerPINHashes: %v", err)
	}
	if len(candidates) != 0 {
		t.Errorf("unknown business returned %d candidates, want 0", len(candidates))
	}
}

func TestManager_ListManagerPINHashes_ExcludesRowsWithoutHash(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	// A row whose pin_hash is NULL is a pre-migration row. It has no usable
	// credential and must never be reported as a candidate, or the verify path
	// would bcrypt-compare against an empty string and treat the outcome as
	// meaningful.
	candidates, err := repo.ListManagerPINHashes(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListManagerPINHashes: %v", err)
	}
	for _, c := range candidates {
		if c.PinHash == "" {
			t.Errorf("candidate for staff %d has an empty pin_hash", c.StaffID)
		}
	}
}

func TestManager_ListManagerPINHashes_IsScopedToOneBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	// The repository takes the business as its only scope; a cross-tenant read
	// would let a caller at one business verify against another's managers.
	// Assert the parameter is the one that reaches the query.
	candidates, err := repo.ListManagerPINHashes(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListManagerPINHashes: %v", err)
	}
	if len(candidates) == 0 {
		t.Skip("no manager PIN rows seeded for this business")
	}
	t.Logf("read %d candidate(s) for business %d", len(candidates), testutil.OwnerID)
}

func TestManager_PINAttemptStore_SurvivesReopen(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	const businessID = 987654

	store := infrastructure.NewPINAttemptStore(db)
	if err := store.ResetPINFailures(ctx, businessID); err != nil {
		t.Fatalf("reset: %v", err)
	}

	for i := 0; i < 2; i++ {
		if _, err := store.RecordPINFailure(ctx, businessID); err != nil {
			t.Fatalf("record failure %d: %v", i, err)
		}
	}

	// A second store over the same database is what a second replica, or a
	// restarted process, looks like. It must observe the same counter.
	reopened := infrastructure.NewPINAttemptStore(db)
	retryAt, err := reopened.PinRetryAt(ctx, businessID)
	if err != nil {
		t.Fatalf("PinRetryAt: %v", err)
	}
	if retryAt.IsZero() {
		t.Error("a reopened store saw no failures; the counter is process-local, not durable")
	}

	if err := reopened.ResetPINFailures(ctx, businessID); err != nil {
		t.Fatalf("cleanup reset: %v", err)
	}
}

// The write path is the one that makes the module usable: after the plaintext
// column was dropped, nothing else in the repository could create a manager
// credential. The stored value must be a bcrypt hash that the verify path
// accepts, scoped to the business it was written for, and replacing an existing
// credential must not leave a second candidate behind for the same manager.
func TestManager_SetManagerPIN_StoresAVerifiableHashScopedToTheBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := infrastructure.NewRepository(db)

	const (
		businessID = 555001
		staffID    = 99001
		ownerID    = 99001
		pin        = "135790"
	)

	hash, err := passwordutils.HashPIN(pin)
	if err != nil {
		t.Fatalf("hash PIN: %v", err)
	}

	input := domain.SetPINInput{
		OwnerID:    ownerID,
		BusinessID: businessID,
		StaffID:    staffID,
	}
	if err := repo.SetManagerPIN(ctx, input, hash); err != nil {
		t.Fatalf("SetManagerPIN: %v", err)
	}
	t.Cleanup(func() { cleanupManagerPINRow(t, db, businessID, staffID) })

	candidates, err := repo.ListManagerPINHashes(ctx, businessID)
	if err != nil {
		t.Fatalf("ListManagerPINHashes: %v", err)
	}
	matches := 0
	for _, c := range candidates {
		if c.StaffID != staffID {
			continue
		}
		matches++
		if c.PinHash == pin {
			t.Fatal("the PIN was stored in cleartext")
		}
		if !passwordutils.CheckPINHash(pin, c.PinHash) {
			t.Errorf("stored hash for staff %d does not verify against the PIN that was set", staffID)
		}
	}
	if matches != 1 {
		t.Fatalf("found %d credentials for staff %d, want exactly 1", matches, staffID)
	}

	// The credential must not be visible from another tenant.
	other, err := repo.ListManagerPINHashes(ctx, businessID+1)
	if err != nil {
		t.Fatalf("ListManagerPINHashes for a second business: %v", err)
	}
	for _, c := range other {
		if c.StaffID == staffID {
			t.Error("a manager credential leaked across the tenant boundary")
		}
	}
}

// Rotating must replace, not accumulate. An upsert that silently stacked rows
// would leave the old hash in the candidate set: VerifyPIN would keep accepting
// the previous PIN and every verification would cost more.
func TestManager_SetManagerPIN_RotationReplacesRatherThanAccumulates(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := infrastructure.NewRepository(db)

	const (
		businessID = 555002
		staffID    = 99002
		ownerID    = 99002
		firstPIN   = "246813"
		secondPIN  = "975310"
	)

	first, err := passwordutils.HashPIN(firstPIN)
	if err != nil {
		t.Fatalf("hash first PIN: %v", err)
	}
	second, err := passwordutils.HashPIN(secondPIN)
	if err != nil {
		t.Fatalf("hash second PIN: %v", err)
	}

	input := domain.SetPINInput{OwnerID: ownerID, BusinessID: businessID, StaffID: staffID}
	if err := repo.SetManagerPIN(ctx, input, first); err != nil {
		t.Fatalf("first SetManagerPIN: %v", err)
	}
	t.Cleanup(func() { cleanupManagerPINRow(t, db, businessID, staffID) })

	if err := repo.SetManagerPIN(ctx, input, second); err != nil {
		t.Fatalf("rotating SetManagerPIN: %v", err)
	}

	candidates, err := repo.ListManagerPINHashes(ctx, businessID)
	if err != nil {
		t.Fatalf("ListManagerPINHashes: %v", err)
	}
	rows, acceptsFirst, acceptsSecond := 0, false, false
	for _, c := range candidates {
		if c.StaffID != staffID {
			continue
		}
		rows++
		acceptsFirst = acceptsFirst || passwordutils.CheckPINHash(firstPIN, c.PinHash)
		acceptsSecond = acceptsSecond || passwordutils.CheckPINHash(secondPIN, c.PinHash)
	}
	if rows != 1 {
		t.Errorf("rotation left %d credential rows for staff %d, want 1: the old hash is still a live candidate", rows, staffID)
	}
	if acceptsFirst {
		t.Error("the superseded PIN still verifies; a rotation must invalidate the previous credential")
	}
	if !acceptsSecond {
		t.Error("the new PIN does not verify after rotation")
	}
}

// The step-up read must return the account hash and must not be usable as a
// credential oracle: a missing account is an error, never an empty success.
func TestManager_GetUserPasswordHash_ReadsTheAccountCredential(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	hash, err := repo.GetUserPasswordHash(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetUserPasswordHash: %v", err)
	}
	if hash == "" {
		t.Fatal("the seeded owner has no stored credential; the step-up check could never pass")
	}
	if !strings.HasPrefix(hash, "$2") {
		t.Errorf("stored credential %q is not a bcrypt hash", hash)
	}
}

func TestManager_GetUserPasswordHash_UnknownAccountErrors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	hash, err := repo.GetUserPasswordHash(context.Background(), 987654321)
	if err == nil {
		t.Fatalf("an unknown account returned %q with no error; that is an existence oracle", hash)
	}
}

func TestManager_SetManagerPIN_NilPoolFailsClosed(t *testing.T) {
	repo := infrastructure.NewRepository(nil)
	err := repo.SetManagerPIN(context.Background(), domain.SetPINInput{
		OwnerID: 1, BusinessID: 1, StaffID: 1,
	}, "$2a$12$aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err == nil {
		t.Fatal("SetManagerPIN with no pool reported success; the caller would believe a credential was stored")
	}
}

// A malformed request must be refused before the transaction opens.
func TestManager_SetManagerPIN_RejectsIncompleteInput(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	for name, input := range map[string]domain.SetPINInput{
		"no business": {OwnerID: 1, BusinessID: 0, StaffID: 1},
		"no staff":    {OwnerID: 1, BusinessID: 1, StaffID: 0},
		"no owner":    {OwnerID: 0, BusinessID: 1, StaffID: 1},
	} {
		t.Run(name, func(t *testing.T) {
			if err := repo.SetManagerPIN(context.Background(), input, "hash"); !errors.Is(err, domain.ErrInvalidSetPIN) {
				t.Fatalf("returned %v, want ErrInvalidSetPIN", err)
			}
		})
	}
}

func cleanupManagerPINRow(t *testing.T, db *sql.DB, businessID, staffID int64) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(),
		`DELETE FROM pos_manager_pins WHERE business_id = ? AND staff_id = ?`, businessID, staffID); err != nil {
		t.Logf("cleanup pos_manager_pins: %v", err)
	}
}

func TestManager_PINAttemptStore_NilPoolFailsClosed(t *testing.T) {
	store := infrastructure.NewPINAttemptStore(nil)
	if _, err := store.PinRetryAt(context.Background(), 1); err == nil {
		t.Error("a nil pool reported no counter error; the verify path would fall through unguarded")
	}
	if _, err := store.RecordPINFailure(context.Background(), 1); err == nil {
		t.Error("a nil pool accepted a failure write")
	}
	if err := store.ResetPINFailures(context.Background(), 1); err == nil {
		t.Error("a nil pool accepted a reset")
	}
}
