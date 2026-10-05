package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries      *sqlc.Queries
	memberReader func(context.Context, int64, int64) (bool, error)
	db           *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		db:      db,
	}
}

func (r *Repository) LinkCustomerToContact(ctx context.Context, businessID, contactID, customerID int64) error {
	return r.queries.LinkCustomerToContact(ctx, sqlc.LinkCustomerToContactParams{
		CustomerID: sql.NullInt64{Int64: customerID, Valid: true},
		BusinessID: businessID,
		ID:         contactID,
	})
}

// ---------------------------------------------------------------------------
// Leads
// ---------------------------------------------------------------------------

func (r *Repository) ListLeads(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Lead, error) {
	rows, err := r.queries.ListLeads(ctx, sqlc.ListLeadsParams{
		BusinessID: businessID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, err
	}

	leads := make([]domain.Lead, 0, len(rows))
	for _, row := range rows {
		leads = append(leads, leadFromRow(row))
	}
	return leads, nil
}

func (r *Repository) CountLeads(ctx context.Context, businessID int64) (int64, error) {
	return r.queries.CountLeads(ctx, businessID)
}

func (r *Repository) GetLead(ctx context.Context, id, businessID int64) (domain.Lead, error) {
	row, err := r.queries.GetLead(ctx, sqlc.GetLeadParams{ID: id, BusinessID: businessID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Lead{}, domain.ErrLeadNotFound
		}
		return domain.Lead{}, err
	}
	return leadFromGetRow(row), nil
}

func (r *Repository) CreateLead(ctx context.Context, ownerID int64, req domain.CreateLeadRequest) (domain.Lead, error) {
	businessID := ownerID
	ownerID = ownerIDFromContext(ctx, ownerID)
	source := req.Source
	if source == "" {
		source = "cold_email"
	}
	stage := req.Stage
	if stage == "" {
		stage = domain.LeadStageCold
	}

	result, err := r.queries.CreateLead(ctx, sqlc.CreateLeadParams{
		OwnerID:      ownerID,
		BusinessID:   businessID,
		CompanyName:  req.CompanyName,
		Industry:     nullableString(req.Industry),
		Size:         nullableString(req.Size),
		Source:       source,
		Stage:        string(stage),
		AssignedTo:   nullableInt64(req.AssignedTo),
		MrrPotential: nullableDecimalString(req.MRRPotential),
		Notes:        nullableString(req.Notes),
	})
	if err != nil {
		return domain.Lead{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Lead{}, err
	}
	return r.GetLead(ctx, id, businessID)
}

func (r *Repository) UpdateLead(ctx context.Context, id, businessID int64, req domain.UpdateLeadRequest) (domain.Lead, error) {
	err := r.queries.UpdateLead(ctx, sqlc.UpdateLeadParams{
		CompanyName:  req.CompanyName,
		Industry:     nullableString(req.Industry),
		Size:         nullableString(req.Size),
		Source:       req.Source,
		Stage:        string(req.Stage),
		AssignedTo:   nullableInt64(req.AssignedTo),
		MrrPotential: nullableDecimalString(req.MRRPotential),
		Notes:        nullableString(req.Notes),
		ID:           id,
		BusinessID:   businessID,
	})
	if err != nil {
		return domain.Lead{}, err
	}
	return r.GetLead(ctx, id, businessID)
}

func (r *Repository) ValidateLeadAssignee(ctx context.Context, businessID, userID int64) error {
	if r.memberReader == nil {
		return errors.New("lead assignee reader is not configured")
	}
	exists, err := r.memberReader(ctx, businessID, userID)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrInvalidLeadAssignee
	}
	return nil
}

func (r *Repository) DeleteLead(ctx context.Context, id, businessID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			logger.Ctx(ctx).Error("failed to rollback delete lead tx", logger.Error(rollbackErr))
		}
	}()

	// The sales tables intentionally do not use database foreign keys because
	// they are shared with legacy installations. Delete dependent records in a
	// transaction so the API's documented cascade is atomic and cannot leave
	// contacts, touchpoints, or contracts orphaned when a lead is removed.
	if err := ensureLeadInBusiness(ctx, tx, businessID, id); err != nil {
		return err
	}
	for _, table := range []string{"crm_touchpoints", "crm_contacts", "crm_contracts"} {
		// #nosec G202 -- This table list is a package-fixed set of tenant-owned dependent tables; values remain bound.
		if _, err := tx.ExecContext(ctx, "DELETE FROM "+table+" WHERE lead_id = ? AND business_id = ?", id, businessID); err != nil {
			return err
		}
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM crm_leads WHERE id = ? AND business_id = ?", id, businessID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) GetPipeline(ctx context.Context, businessID int64) ([]domain.PipelineSummary, error) {
	rows, err := r.queries.GetPipeline(ctx, businessID)
	if err != nil {
		return nil, err
	}

	summaries := make([]domain.PipelineSummary, 0, len(rows))
	for _, row := range rows {
		summaries = append(summaries, domain.PipelineSummary{
			Stage:    row.Stage,
			Count:    int(row.Cnt),
			TotalMRR: parsePipelineMRR(row.TotalMrr),
		})
	}
	return summaries, nil
}

// ---------------------------------------------------------------------------
// Contacts
// ---------------------------------------------------------------------------

func (r *Repository) ListContacts(ctx context.Context, businessID int64, leadID *int64, limit, offset int32) ([]domain.Contact, error) {
	if leadID != nil {
		rows, err := r.queries.ListContactsByLead(ctx, sqlc.ListContactsByLeadParams{
			BusinessID: businessID,
			LeadID:     *leadID,
			Limit:      limit,
			Offset:     offset,
		})
		if err != nil {
			return nil, err
		}
		contacts := make([]domain.Contact, 0, len(rows))
		for _, row := range rows {
			contacts = append(contacts, contactFromLeadRow(row))
		}
		return contacts, nil
	}

	rows, err := r.queries.ListContacts(ctx, sqlc.ListContactsParams{
		BusinessID: businessID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, err
	}
	contacts := make([]domain.Contact, 0, len(rows))
	for _, row := range rows {
		contacts = append(contacts, contactFromRow(row))
	}
	return contacts, nil
}

func (r *Repository) WithMembershipReader(reader func(context.Context, int64, int64) (bool, error)) *Repository {
	r.memberReader = reader
	return r
}
