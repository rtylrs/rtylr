package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/infrastructure/persistence/sqlc"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
)

func (r *Repository) CreateTouchpoint(ctx context.Context, ownerID, createdBy int64, req domain.CreateTouchpointRequest) (domain.Touchpoint, error) {
	businessID := ownerID
	ownerID = ownerIDFromContext(ctx, ownerID)
	tp := req.Type
	if tp == "" {
		tp = "call"
	}
	outcome := req.Outcome
	if outcome == "" {
		outcome = "other"
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Touchpoint{}, err
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			logger.Ctx(ctx).Error("failed to rollback create touchpoint tx", logger.Error(rollbackErr))
		}
	}()
	if err := ensureLeadInBusiness(ctx, tx, businessID, req.LeadID); err != nil {
		return domain.Touchpoint{}, err
	}
	if req.ContactID != nil {
		if err := ensureContactForLead(ctx, tx, businessID, req.LeadID, *req.ContactID); err != nil {
			return domain.Touchpoint{}, err
		}
	}

	nextFollowUp, err := nullableDatePtr(req.NextFollowUp)
	if err != nil {
		return domain.Touchpoint{}, err
	}

	result, err := r.queries.WithTx(tx).CreateTouchpoint(ctx, sqlc.CreateTouchpointParams{
		OwnerID:      ownerID,
		BusinessID:   businessID,
		LeadID:       req.LeadID,
		ContactID:    nullableInt64(req.ContactID),
		Type:         tp,
		Outcome:      outcome,
		Notes:        nullableString(req.Notes),
		NextFollowUp: nextFollowUp,
		CreatedBy:    createdBy,
	})
	if err != nil {
		return domain.Touchpoint{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Touchpoint{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Touchpoint{}, err
	}
	return r.GetTouchpoint(ctx, id, businessID)
}

func (r *Repository) DeleteTouchpoint(ctx context.Context, id, businessID int64) error {
	result, err := r.queries.DeleteTouchpoint(ctx, sqlc.DeleteTouchpointParams{ID: id, BusinessID: businessID})
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrTouchpointNotFound
	}
	return nil
}

// ---------------------------------------------------------------------------
// Contracts
// ---------------------------------------------------------------------------

func (r *Repository) ListContracts(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Contract, error) {
	rows, err := r.queries.ListContracts(ctx, sqlc.ListContractsParams{
		BusinessID: businessID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, err
	}

	contracts := make([]domain.Contract, 0, len(rows))
	for _, row := range rows {
		contracts = append(contracts, contractFromRow(row))
	}
	return contracts, nil
}

func (r *Repository) CountContracts(ctx context.Context, businessID int64) (int64, error) {
	return r.queries.CountContracts(ctx, businessID)
}

func (r *Repository) GetContract(ctx context.Context, id, businessID int64) (domain.Contract, error) {
	row, err := r.queries.GetContract(ctx, sqlc.GetContractParams{ID: id, BusinessID: businessID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Contract{}, domain.ErrContractNotFound
		}
		return domain.Contract{}, err
	}
	return contractFromGetRow(row), nil
}

func (r *Repository) CreateContract(ctx context.Context, ownerID int64, req domain.CreateContractRequest) (domain.Contract, error) {
	businessID := ownerID
	ownerID = ownerIDFromContext(ctx, ownerID)
	plan := req.Plan
	if plan == "" {
		plan = "pro"
	}
	seats := req.Seats
	if seats == 0 {
		seats = 1
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Contract{}, err
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			logger.Ctx(ctx).Error("failed to rollback create contract tx", logger.Error(rollbackErr))
		}
	}()
	if err := ensureLeadInBusiness(ctx, tx, businessID, req.LeadID); err != nil {
		return domain.Contract{}, err
	}
	if err := ensureLeadCanContract(ctx, tx, businessID, req.LeadID); err != nil {
		return domain.Contract{}, err
	}

	signedAt, err := nullableDatePtr(req.SignedAt)
	if err != nil {
		return domain.Contract{}, err
	}

	result, err := r.queries.WithTx(tx).CreateContract(ctx, sqlc.CreateContractParams{
		OwnerID:      ownerID,
		BusinessID:   businessID,
		LeadID:       req.LeadID,
		Plan:         plan,
		Mrr:          req.MRR.String(),
		Seats:        int32(seats),
		SignedAt:     signedAt,
		SignedByName: nullableString(req.SignedByName),
		ClosedBy:     nullableInt64(req.ClosedBy),
		Notes:        nullableString(req.Notes),
	})
	if err != nil {
		return domain.Contract{}, err
	}
	if _, err := tx.ExecContext(ctx,
		"UPDATE crm_leads SET stage = ? WHERE id = ? AND business_id = ?",
		string(domain.LeadStageContracted), req.LeadID, businessID,
	); err != nil {
		return domain.Contract{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Contract{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Contract{}, err
	}
	return r.GetContract(ctx, id, businessID)
}

func ownerIDFromContext(ctx context.Context, fallback int64) int64 {
	if device := authPkg.DeviceCtx(ctx); device.User != nil && device.User.UserID > 0 {
		return device.User.UserID
	}
	return fallback
}

// ensureLeadInBusiness prevents callers from attaching sales records to a
// lead belonging to another tenant. The API always resolves businessID from
// the authenticated user, so this check is the final tenant-boundary guard at
// the persistence layer.
func ensureLeadInBusiness(ctx context.Context, tx *sql.Tx, businessID, leadID int64) error {
	// Lock the lead row so a concurrent DeleteLead cannot interleave between
	// this check and the child insert (child tables have no FKs).
	var id int64
	err := tx.QueryRowContext(ctx,
		"SELECT id FROM crm_leads WHERE id = ? AND business_id = ? FOR UPDATE",
		leadID, businessID,
	).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrLeadNotFound
	}
	return err
}
