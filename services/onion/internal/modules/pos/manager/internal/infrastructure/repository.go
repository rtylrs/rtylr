package infrastructure

import (
	"context"
	"database/sql"
	"github.com/voxire/onion/internal/modules/pos/internal/outbox"

	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/modules/pos/manager/internal/domain"
	"github.com/voxire/onion/internal/modules/pos/manager/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// ErrNoManagerPINs reports that the manager credential store could not be read
// at all, as opposed to holding no usable credential. The verify path treats it
// as an infrastructure fault, not a credential verdict.
var ErrNoManagerPINs = domain.ErrNoManagerPINs

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db
	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

// ListManagerPINHashes returns the bcrypt hashes of every usable manager
// credential at a business.
//
// This is a read of hashes, not of PINs. The legacy plaintext column is gone
// (see schemas/pos/zz_pos_manager_pins_hash.sql), so a dump, backup or replica
// no longer carries a credential that can be replayed directly against the POS.
//
// The query cannot be a seek on the submitted PIN: bcrypt is not orderable by
// its input, so a slow KDF and an index lookup are mutually exclusive here
// without a peppered blind index. The caller compares each candidate.
func (r *Repository) ListManagerPINHashes(ctx context.Context, businessID int64) ([]domain.ManagerPINCandidate, error) {
	rows, err := r.queries.ListManagerPINHashes(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to read manager PIN candidates", logger.Error(err))
		return nil, err
	}

	candidates := make([]domain.ManagerPINCandidate, 0, len(rows))
	for _, row := range rows {
		candidates = append(candidates, domain.ManagerPINCandidate{
			OwnerID: row.OwnerID,
			StaffID: row.StaffID,
			PinHash: row.PinHash.String,
		})
	}
	return candidates, nil
}

// SetManagerPIN stores the bcrypt hash of a manager's override PIN, replacing
// any credential the manager already had, and appends the audit events in the
// same transaction so a credential is never written without its record.
//
// Update-then-insert rather than an upsert: pos_manager_pins carries no unique
// key on (business_id, staff_id), so ON DUPLICATE KEY UPDATE would never match
// and every set would stack another candidate row for the same manager — which
// VerifyPIN would then scan, raising the cost of every verification at the
// business. The two concurrent first-sets that could still slip past this
// sequence are bounded by MaxManagerPINCandidates and fail verification closed.
func (r *Repository) SetManagerPIN(ctx context.Context, input domain.SetPINInput, pinHash string, events ...audit.Event) (err error) {
	if input.BusinessID <= 0 || input.StaffID <= 0 || input.OwnerID <= 0 {
		return domain.ErrInvalidSetPIN
	}
	// database/sql panics on a nil pool rather than returning an error, and a
	// caller reporting success for a credential that was never stored is the one
	// outcome this must not produce.
	if r == nil || r.pool == nil {
		return ErrNoManagerPINs
	}

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin pos manager PIN transaction", logger.Error(err))
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	queries := r.queries.WithTx(tx)

	updated, err := queries.UpdateManagerPINHash(ctx, sqlc.UpdateManagerPINHashParams{
		PinHash:    sql.NullString{String: pinHash, Valid: true},
		BusinessID: input.BusinessID,
		StaffID:    input.StaffID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update manager PIN hash", logger.Error(err))
		return err
	}

	if updated == 0 {
		if err = queries.InsertManagerPIN(ctx, sqlc.InsertManagerPINParams{
			OwnerID:    input.OwnerID,
			BusinessID: input.BusinessID,
			StaffID:    input.StaffID,
			PinHash:    sql.NullString{String: pinHash, Valid: true},
		}); err != nil {
			logger.Ctx(ctx).Error("Failed to insert manager PIN hash", logger.Error(err))
			return err
		}
	}

	// The audit event carries the staff member and the action only. The PIN and
	// its hash are deliberately absent: an activity_log row is a durable,
	// widely-readable record, and neither a credential nor a credential
	// equivalent belongs in one.
	for _, event := range events {
		if err = outbox.AppendAuditTx(ctx, tx, event); err != nil {
			logger.Ctx(ctx).Error("Failed to append manager PIN audit event", logger.Error(err))
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit manager PIN transaction", logger.Error(err))
		return err
	}
	return nil
}
