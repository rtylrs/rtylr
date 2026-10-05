package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/crm/internal/outbox"

	crmcontract "github.com/voxire/onion/internal/contracts/crm"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/eventbus"
	"github.com/voxire/onion/internal/platform/logger"
)

func (r *Repository) CountContacts(ctx context.Context, businessID int64, leadID *int64) (int64, error) {
	if leadID != nil {
		return r.queries.CountContactsByLead(ctx, sqlc.CountContactsByLeadParams{
			BusinessID: businessID,
			LeadID:     *leadID,
		})
	}
	return r.queries.CountContacts(ctx, businessID)
}

func (r *Repository) GetContact(ctx context.Context, id, businessID int64) (domain.Contact, error) {
	row, err := r.queries.GetContact(ctx, sqlc.GetContactParams{ID: id, BusinessID: businessID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Contact{}, domain.ErrContactNotFound
		}
		return domain.Contact{}, err
	}
	return contactFromGetRow(row), nil
}

func (r *Repository) CreateContact(ctx context.Context, businessID, ownerID int64, req domain.CreateContactRequest) (domain.Contact, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Contact{}, err
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			logger.Ctx(ctx).Error("failed to rollback create contact tx", logger.Error(rollbackErr))
		}
	}()
	if err := ensureLeadInBusiness(ctx, tx, businessID, req.LeadID); err != nil {
		return domain.Contact{}, err
	}

	qtx := r.queries.WithTx(tx)
	result, err := qtx.CreateContact(ctx, sqlc.CreateContactParams{
		OwnerID:    ownerID,
		BusinessID: businessID,
		LeadID:     req.LeadID,
		CustomerID: sql.NullInt64{},
		Name:       req.Name,
		Role:       nullableString(req.Role),
		Email:      nullableString(req.Email),
		Phone:      nullableString(req.Phone),
		Linkedin:   nullableString(req.LinkedIn),
		IsPrimary:  req.IsPrimary,
		Notes:      nullableString(req.Notes),
	})
	if err != nil {
		return domain.Contact{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Contact{}, err
	}
	contact := crmcontract.CustomerContactCaptured{Name: req.Name, Email: req.Email, Phone: req.Phone}
	if err = outbox.Publish(ctx, tx, eventbus.Event{
		BusinessID:  businessID,
		Type:        crmcontract.CustomerContactCapturedType,
		SubjectType: "crm_contact",
		SubjectID:   id,
		Payload:     contact,
		DedupeKey:   crmcontract.CustomerContactCaptureDedupeKey("crm_contact", id, contact),
	}); err != nil {
		return domain.Contact{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Contact{}, err
	}
	return r.GetContact(ctx, id, businessID)
}

func (r *Repository) UpdateContact(ctx context.Context, id, businessID int64, req domain.UpdateContactRequest) (domain.Contact, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Contact{}, err
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			logger.Ctx(ctx).Error("failed to rollback update contact tx", logger.Error(rollbackErr))
		}
	}()

	qtx := r.queries.WithTx(tx)
	existingContact, err := qtx.GetContact(ctx, sqlc.GetContactParams{ID: id, BusinessID: businessID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Contact{}, domain.ErrContactNotFound
		}
		return domain.Contact{}, err
	}

	customerID := existingContact.CustomerID
	if !strings.EqualFold(strings.TrimSpace(existingContact.Email), strings.TrimSpace(req.Email)) ||
		strings.TrimSpace(existingContact.Phone) != strings.TrimSpace(req.Phone) {
		customerID = sql.NullInt64{}
	}

	err = qtx.UpdateContact(ctx, sqlc.UpdateContactParams{
		CustomerID: customerID,
		Name:       req.Name,
		Role:       nullableString(req.Role),
		Email:      nullableString(req.Email),
		Phone:      nullableString(req.Phone),
		Linkedin:   nullableString(req.LinkedIn),
		IsPrimary:  req.IsPrimary,
		Notes:      nullableString(req.Notes),
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		return domain.Contact{}, err
	}
	contact := crmcontract.CustomerContactCaptured{Name: req.Name, Email: req.Email, Phone: req.Phone}
	updatedAt := time.Now().UTC().Format(time.RFC3339Nano)
	if err = outbox.Publish(ctx, tx, eventbus.Event{
		BusinessID:  businessID,
		Type:        crmcontract.CustomerContactCapturedType,
		SubjectType: "crm_contact",
		SubjectID:   id,
		Payload:     contact,
		DedupeKey:   "crm_contact_capture:" + strconv.FormatInt(id, 10) + ":" + updatedAt,
	}); err != nil {
		return domain.Contact{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Contact{}, err
	}
	return r.GetContact(ctx, id, businessID)
}

func (r *Repository) DeleteContact(ctx context.Context, id, businessID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	result, err := r.queries.WithTx(tx).DeleteContact(ctx, sqlc.DeleteContactParams{ID: id, BusinessID: businessID})
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrContactNotFound
	}

	// crm_touchpoints.contact_id is nullable: detach touchpoints instead of
	// leaving them pointing at a deleted contact.
	if _, err := tx.ExecContext(ctx,
		"UPDATE crm_touchpoints SET contact_id = NULL WHERE business_id = ? AND contact_id = ?",
		businessID, id,
	); err != nil {
		return err
	}

	return tx.Commit()
}

// ---------------------------------------------------------------------------
// Touchpoints
// ---------------------------------------------------------------------------

func (r *Repository) ListTouchpoints(ctx context.Context, businessID, leadID int64) ([]domain.Touchpoint, error) {
	rows, err := r.queries.ListTouchpoints(ctx, sqlc.ListTouchpointsParams{
		BusinessID: businessID,
		LeadID:     leadID,
	})
	if err != nil {
		return nil, err
	}

	tps := make([]domain.Touchpoint, 0, len(rows))
	for _, row := range rows {
		tps = append(tps, touchpointFromRow(row))
	}
	return tps, nil
}

func (r *Repository) GetTouchpoint(ctx context.Context, id, businessID int64) (domain.Touchpoint, error) {
	row, err := r.queries.GetTouchpoint(ctx, sqlc.GetTouchpointParams{ID: id, BusinessID: businessID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Touchpoint{}, domain.ErrTouchpointNotFound
		}
		return domain.Touchpoint{}, err
	}
	return touchpointFromGetRow(row), nil
}
