package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/crm_sales/internal/domain"
	customersync "github.com/voxire/onion/internal/platform/customer_sync"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ---------------------------------------------------------------------------
// Leads
// ---------------------------------------------------------------------------

func (r *Repository) ListLeads(ctx context.Context, businessID int64) ([]domain.Lead, error) {
	query, args, err := sq.
		Select("id", "owner_id", "business_id", "company_name", "COALESCE(industry,'')", "COALESCE(size,'')",
			"source", "stage", "assigned_to", "mrr_potential", "COALESCE(notes,'')",
			"created_at", "updated_at").
		From("crm_leads").
		Where(sq.Eq{"business_id": businessID}).
		OrderBy("created_at DESC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Ctx(ctx).Error("failed to close leads rows", logger.Error(closeErr))
		}
	}()

	leads := make([]domain.Lead, 0)
	for rows.Next() {
		l, scanErr := scanLead(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		leads = append(leads, l)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}
	return leads, nil
}

func (r *Repository) GetLead(ctx context.Context, id, businessID int64) (domain.Lead, error) {
	query, args, err := sq.
		Select("id", "owner_id", "business_id", "company_name", "COALESCE(industry,'')", "COALESCE(size,'')",
			"source", "stage", "assigned_to", "mrr_potential", "COALESCE(notes,'')",
			"created_at", "updated_at").
		From("crm_leads").
		Where(sq.Eq{"id": id, "business_id": businessID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return domain.Lead{}, err
	}

	row := r.db.QueryRowContext(ctx, query, args...)
	l, err := scanLeadRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Lead{}, domain.ErrLeadNotFound
		}
		return domain.Lead{}, err
	}
	return l, nil
}

func (r *Repository) CreateLead(ctx context.Context, ownerID int64, req domain.CreateLeadRequest) (domain.Lead, error) {
	source := req.Source
	if source == "" {
		source = "cold_email"
	}
	stage := req.Stage
	if stage == "" {
		stage = domain.LeadStageCold
	}

	query, args, err := sq.
		Insert("crm_leads").
		Columns("owner_id", "business_id", "company_name", "industry", "size", "source", "stage", "assigned_to", "mrr_potential", "notes").
		Values(ownerID, ownerID, req.CompanyName, nullableString(req.Industry), nullableString(req.Size),
			source, string(stage), nullableInt64(req.AssignedTo), nullableFloat64(req.MRRPotential),
			nullableString(req.Notes)).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return domain.Lead{}, err
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return domain.Lead{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Lead{}, err
	}
	return r.GetLead(ctx, id, ownerID)
}

func (r *Repository) UpdateLead(ctx context.Context, id, businessID int64, req domain.UpdateLeadRequest) (domain.Lead, error) {
	query, args, err := sq.
		Update("crm_leads").
		Set("company_name", req.CompanyName).
		Set("industry", nullableString(req.Industry)).
		Set("size", nullableString(req.Size)).
		Set("source", req.Source).
		Set("stage", string(req.Stage)).
		Set("assigned_to", nullableInt64(req.AssignedTo)).
		Set("mrr_potential", nullableFloat64(req.MRRPotential)).
		Set("notes", nullableString(req.Notes)).
		Where(sq.Eq{"id": id, "business_id": businessID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return domain.Lead{}, err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return domain.Lead{}, err
	}
	return r.GetLead(ctx, id, businessID)
}

func (r *Repository) DeleteLead(ctx context.Context, id, businessID int64) error {
	query, args, err := sq.
		Delete("crm_leads").
		Where(sq.Eq{"id": id, "business_id": businessID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrLeadNotFound
	}
	return nil
}

func (r *Repository) GetPipeline(ctx context.Context, businessID int64) ([]domain.PipelineSummary, error) {
	query, args, err := sq.
		Select("stage", "COUNT(*) as cnt", "COALESCE(SUM(mrr_potential), 0) as total_mrr").
		From("crm_leads").
		Where(sq.Eq{"business_id": businessID}).
		GroupBy("stage").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Ctx(ctx).Error("failed to close pipeline rows", logger.Error(closeErr))
		}
	}()

	summaries := make([]domain.PipelineSummary, 0)
	for rows.Next() {
		var s domain.PipelineSummary
		if scanErr := rows.Scan(&s.Stage, &s.Count, &s.TotalMRR); scanErr != nil {
			return nil, scanErr
		}
		summaries = append(summaries, s)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}
	return summaries, nil
}

// ---------------------------------------------------------------------------
// Contacts
// ---------------------------------------------------------------------------

func (r *Repository) ListContacts(ctx context.Context, businessID int64, leadID *int64) ([]domain.Contact, error) {
	builder := sq.
		Select("id", "owner_id", "business_id", "lead_id", "customer_id", "name", "COALESCE(role,'')", "COALESCE(email,'')",
			"COALESCE(phone,'')", "COALESCE(linkedin,'')", "is_primary", "COALESCE(notes,'')",
			"created_at", "updated_at").
		From("crm_contacts").
		Where(sq.Eq{"business_id": businessID})
	if leadID != nil {
		builder = builder.Where(sq.Eq{"lead_id": *leadID})
	}
	builder = builder.OrderBy("created_at DESC")

	query, args, err := builder.PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Ctx(ctx).Error("failed to close contacts rows", logger.Error(closeErr))
		}
	}()

	contacts := make([]domain.Contact, 0)
	for rows.Next() {
		c, scanErr := scanContact(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		contacts = append(contacts, c)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}
	return contacts, nil
}

func (r *Repository) GetContact(ctx context.Context, id, businessID int64) (domain.Contact, error) {
	query, args, err := sq.
		Select("id", "owner_id", "business_id", "lead_id", "customer_id", "name", "COALESCE(role,'')", "COALESCE(email,'')",
			"COALESCE(phone,'')", "COALESCE(linkedin,'')", "is_primary", "COALESCE(notes,'')",
			"created_at", "updated_at").
		From("crm_contacts").
		Where(sq.Eq{"id": id, "business_id": businessID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return domain.Contact{}, err
	}

	row := r.db.QueryRowContext(ctx, query, args...)
	c, err := scanContactRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Contact{}, domain.ErrContactNotFound
		}
		return domain.Contact{}, err
	}
	return c, nil
}

func (r *Repository) CreateContact(ctx context.Context, ownerID int64, req domain.CreateContactRequest) (domain.Contact, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Contact{}, err
	}
	defer tx.Rollback()

	customerID, err := customersync.UpsertCRMCustomer(ctx, tx, ownerID, customersync.Contact{
		Name:  req.Name,
		Email: req.Email,
		Phone: req.Phone,
	})
	if err != nil {
		return domain.Contact{}, err
	}

	query, args, err := sq.
		Insert("crm_contacts").
		Columns("owner_id", "business_id", "lead_id", "customer_id", "name", "role", "email", "phone", "linkedin", "is_primary", "notes").
		Values(ownerID, ownerID, req.LeadID, nullableInt64Value(customerID), req.Name, nullableString(req.Role), nullableString(req.Email),
			nullableString(req.Phone), nullableString(req.LinkedIn), req.IsPrimary, nullableString(req.Notes)).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return domain.Contact{}, err
	}

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return domain.Contact{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Contact{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Contact{}, err
	}
	return r.GetContact(ctx, id, ownerID)
}

func (r *Repository) UpdateContact(ctx context.Context, id, businessID int64, req domain.UpdateContactRequest) (domain.Contact, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Contact{}, err
	}
	defer tx.Rollback()

	var existingContactID int64
	if err := tx.QueryRowContext(ctx,
		`SELECT id FROM crm_contacts WHERE business_id = ? AND id = ? LIMIT 1`,
		businessID, id,
	).Scan(&existingContactID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Contact{}, domain.ErrContactNotFound
		}
		return domain.Contact{}, err
	}

	customerID, err := customersync.UpsertCRMCustomer(ctx, tx, businessID, customersync.Contact{
		Name:  req.Name,
		Email: req.Email,
		Phone: req.Phone,
	})
	if err != nil {
		return domain.Contact{}, err
	}

	query, args, err := sq.
		Update("crm_contacts").
		Set("customer_id", nullableInt64Value(customerID)).
		Set("name", req.Name).
		Set("role", nullableString(req.Role)).
		Set("email", nullableString(req.Email)).
		Set("phone", nullableString(req.Phone)).
		Set("linkedin", nullableString(req.LinkedIn)).
		Set("is_primary", req.IsPrimary).
		Set("notes", nullableString(req.Notes)).
		Where(sq.Eq{"id": id, "business_id": businessID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return domain.Contact{}, err
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return domain.Contact{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Contact{}, err
	}
	return r.GetContact(ctx, id, businessID)
}

func (r *Repository) DeleteContact(ctx context.Context, id, businessID int64) error {
	query, args, err := sq.
		Delete("crm_contacts").
		Where(sq.Eq{"id": id, "business_id": businessID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	result, err := r.db.ExecContext(ctx, query, args...)
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
	return nil
}

// ---------------------------------------------------------------------------
// Touchpoints
// ---------------------------------------------------------------------------

func (r *Repository) ListTouchpoints(ctx context.Context, businessID, leadID int64) ([]domain.Touchpoint, error) {
	query, args, err := sq.
		Select("id", "owner_id", "business_id", "lead_id", "contact_id", "type", "outcome",
			"COALESCE(notes,'')", "next_follow_up", "created_by", "created_at").
		From("crm_touchpoints").
		Where(sq.Eq{"business_id": businessID, "lead_id": leadID}).
		OrderBy("created_at DESC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Ctx(ctx).Error("failed to close touchpoints rows", logger.Error(closeErr))
		}
	}()

	tps := make([]domain.Touchpoint, 0)
	for rows.Next() {
		tp, scanErr := scanTouchpoint(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		tps = append(tps, tp)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}
	return tps, nil
}

func (r *Repository) GetTouchpoint(ctx context.Context, id, businessID int64) (domain.Touchpoint, error) {
	query, args, err := sq.
		Select("id", "owner_id", "business_id", "lead_id", "contact_id", "type", "outcome",
			"COALESCE(notes,'')", "next_follow_up", "created_by", "created_at").
		From("crm_touchpoints").
		Where(sq.Eq{"id": id, "business_id": businessID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return domain.Touchpoint{}, err
	}

	row := r.db.QueryRowContext(ctx, query, args...)
	tp, err := scanTouchpointRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Touchpoint{}, domain.ErrTouchpointNotFound
		}
		return domain.Touchpoint{}, err
	}
	return tp, nil
}

func (r *Repository) CreateTouchpoint(ctx context.Context, ownerID, createdBy int64, req domain.CreateTouchpointRequest) (domain.Touchpoint, error) {
	tp := req.Type
	if tp == "" {
		tp = "call"
	}
	outcome := req.Outcome
	if outcome == "" {
		outcome = "other"
	}

	query, args, err := sq.
		Insert("crm_touchpoints").
		Columns("owner_id", "business_id", "lead_id", "contact_id", "type", "outcome", "notes", "next_follow_up", "created_by").
		Values(ownerID, ownerID, req.LeadID, nullableInt64(req.ContactID), tp, outcome,
			nullableString(req.Notes), nullableStringPtr(req.NextFollowUp), createdBy).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return domain.Touchpoint{}, err
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return domain.Touchpoint{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Touchpoint{}, err
	}
	return r.GetTouchpoint(ctx, id, ownerID)
}

func (r *Repository) DeleteTouchpoint(ctx context.Context, id, businessID int64) error {
	query, args, err := sq.
		Delete("crm_touchpoints").
		Where(sq.Eq{"id": id, "business_id": businessID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	result, err := r.db.ExecContext(ctx, query, args...)
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

func (r *Repository) ListContracts(ctx context.Context, businessID int64) ([]domain.Contract, error) {
	query, args, err := sq.
		Select("id", "owner_id", "business_id", "lead_id", "plan", "mrr", "seats",
			"signed_at", "COALESCE(signed_by_name,'')", "closed_by", "COALESCE(notes,'')",
			"created_at", "updated_at").
		From("crm_contracts").
		Where(sq.Eq{"business_id": businessID}).
		OrderBy("created_at DESC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			logger.Ctx(ctx).Error("failed to close contracts rows", logger.Error(closeErr))
		}
	}()

	contracts := make([]domain.Contract, 0)
	for rows.Next() {
		c, scanErr := scanContract(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		contracts = append(contracts, c)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}
	return contracts, nil
}

func (r *Repository) GetContract(ctx context.Context, id, businessID int64) (domain.Contract, error) {
	query, args, err := sq.
		Select("id", "owner_id", "business_id", "lead_id", "plan", "mrr", "seats",
			"signed_at", "COALESCE(signed_by_name,'')", "closed_by", "COALESCE(notes,'')",
			"created_at", "updated_at").
		From("crm_contracts").
		Where(sq.Eq{"id": id, "business_id": businessID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return domain.Contract{}, err
	}

	row := r.db.QueryRowContext(ctx, query, args...)
	c, err := scanContractRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Contract{}, domain.ErrContractNotFound
		}
		return domain.Contract{}, err
	}
	return c, nil
}

func (r *Repository) CreateContract(ctx context.Context, ownerID int64, req domain.CreateContractRequest) (domain.Contract, error) {
	plan := req.Plan
	if plan == "" {
		plan = "pro"
	}
	seats := req.Seats
	if seats == 0 {
		seats = 1
	}

	query, args, err := sq.
		Insert("crm_contracts").
		Columns("owner_id", "business_id", "lead_id", "plan", "mrr", "seats", "signed_at", "signed_by_name", "closed_by", "notes").
		Values(ownerID, ownerID, req.LeadID, plan, req.MRR, seats, nullableStringPtr(req.SignedAt),
			nullableString(req.SignedByName), nullableInt64(req.ClosedBy), nullableString(req.Notes)).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return domain.Contract{}, err
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return domain.Contract{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Contract{}, err
	}
	return r.GetContract(ctx, id, ownerID)
}

func (r *Repository) UpdateContract(ctx context.Context, id, businessID int64, req domain.UpdateContractRequest) (domain.Contract, error) {
	seats := req.Seats
	if seats == 0 {
		seats = 1
	}
	query, args, err := sq.
		Update("crm_contracts").
		Set("plan", req.Plan).
		Set("mrr", req.MRR).
		Set("seats", seats).
		Set("signed_at", nullableStringPtr(req.SignedAt)).
		Set("signed_by_name", nullableString(req.SignedByName)).
		Set("closed_by", nullableInt64(req.ClosedBy)).
		Set("notes", nullableString(req.Notes)).
		Where(sq.Eq{"id": id, "business_id": businessID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return domain.Contract{}, err
	}

	_, err = r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return domain.Contract{}, err
	}
	return r.GetContract(ctx, id, businessID)
}

// ---------------------------------------------------------------------------
// Scan helpers
// ---------------------------------------------------------------------------

type leadScanner interface {
	Scan(dest ...interface{}) error
}

func scanLead(s leadScanner) (domain.Lead, error) {
	var l domain.Lead
	var assignedTo sql.NullInt64
	var mrrPotential sql.NullFloat64
	err := s.Scan(
		&l.ID, &l.OwnerID, &l.BusinessID, &l.CompanyName, &l.Industry, &l.Size,
		&l.Source, &l.Stage, &assignedTo, &mrrPotential, &l.Notes,
		&l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		return domain.Lead{}, err
	}
	if assignedTo.Valid {
		l.AssignedTo = &assignedTo.Int64
	}
	if mrrPotential.Valid {
		l.MRRPotential = &mrrPotential.Float64
	}
	return l, nil
}

func scanLeadRow(row *sql.Row) (domain.Lead, error) {
	var l domain.Lead
	var assignedTo sql.NullInt64
	var mrrPotential sql.NullFloat64
	err := row.Scan(
		&l.ID, &l.OwnerID, &l.BusinessID, &l.CompanyName, &l.Industry, &l.Size,
		&l.Source, &l.Stage, &assignedTo, &mrrPotential, &l.Notes,
		&l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		return domain.Lead{}, err
	}
	if assignedTo.Valid {
		l.AssignedTo = &assignedTo.Int64
	}
	if mrrPotential.Valid {
		l.MRRPotential = &mrrPotential.Float64
	}
	return l, nil
}

type contactScanner interface {
	Scan(dest ...interface{}) error
}

func scanContact(s contactScanner) (domain.Contact, error) {
	var c domain.Contact
	var customerID sql.NullInt64
	err := s.Scan(
		&c.ID, &c.OwnerID, &c.BusinessID, &c.LeadID, &customerID, &c.Name, &c.Role, &c.Email,
		&c.Phone, &c.LinkedIn, &c.IsPrimary, &c.Notes,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return domain.Contact{}, err
	}
	if customerID.Valid {
		c.CustomerID = customerID.Int64
	}
	return c, nil
}

func scanContactRow(row *sql.Row) (domain.Contact, error) {
	var c domain.Contact
	var customerID sql.NullInt64
	err := row.Scan(
		&c.ID, &c.OwnerID, &c.BusinessID, &c.LeadID, &customerID, &c.Name, &c.Role, &c.Email,
		&c.Phone, &c.LinkedIn, &c.IsPrimary, &c.Notes,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return domain.Contact{}, err
	}
	if customerID.Valid {
		c.CustomerID = customerID.Int64
	}
	return c, nil
}

type touchpointScanner interface {
	Scan(dest ...interface{}) error
}

func scanTouchpoint(s touchpointScanner) (domain.Touchpoint, error) {
	var tp domain.Touchpoint
	var contactID sql.NullInt64
	var nextFollowUp sql.NullString
	err := s.Scan(
		&tp.ID, &tp.OwnerID, &tp.BusinessID, &tp.LeadID, &contactID, &tp.Type, &tp.Outcome,
		&tp.Notes, &nextFollowUp, &tp.CreatedBy, &tp.CreatedAt,
	)
	if err != nil {
		return domain.Touchpoint{}, err
	}
	if contactID.Valid {
		tp.ContactID = &contactID.Int64
	}
	if nextFollowUp.Valid && nextFollowUp.String != "" {
		tp.NextFollowUp = &nextFollowUp.String
	}
	return tp, nil
}

func scanTouchpointRow(row *sql.Row) (domain.Touchpoint, error) {
	var tp domain.Touchpoint
	var contactID sql.NullInt64
	var nextFollowUp sql.NullString
	err := row.Scan(
		&tp.ID, &tp.OwnerID, &tp.BusinessID, &tp.LeadID, &contactID, &tp.Type, &tp.Outcome,
		&tp.Notes, &nextFollowUp, &tp.CreatedBy, &tp.CreatedAt,
	)
	if err != nil {
		return domain.Touchpoint{}, err
	}
	if contactID.Valid {
		tp.ContactID = &contactID.Int64
	}
	if nextFollowUp.Valid && nextFollowUp.String != "" {
		tp.NextFollowUp = &nextFollowUp.String
	}
	return tp, nil
}

type contractScanner interface {
	Scan(dest ...interface{}) error
}

func scanContract(s contractScanner) (domain.Contract, error) {
	var c domain.Contract
	var signedAt sql.NullString
	var closedBy sql.NullInt64
	err := s.Scan(
		&c.ID, &c.OwnerID, &c.BusinessID, &c.LeadID, &c.Plan, &c.MRR, &c.Seats,
		&signedAt, &c.SignedByName, &closedBy, &c.Notes,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return domain.Contract{}, err
	}
	if signedAt.Valid && signedAt.String != "" {
		c.SignedAt = &signedAt.String
	}
	if closedBy.Valid {
		c.ClosedBy = &closedBy.Int64
	}
	return c, nil
}

func scanContractRow(row *sql.Row) (domain.Contract, error) {
	var c domain.Contract
	var signedAt sql.NullString
	var closedBy sql.NullInt64
	err := row.Scan(
		&c.ID, &c.OwnerID, &c.BusinessID, &c.LeadID, &c.Plan, &c.MRR, &c.Seats,
		&signedAt, &c.SignedByName, &closedBy, &c.Notes,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return domain.Contract{}, err
	}
	if signedAt.Valid && signedAt.String != "" {
		c.SignedAt = &signedAt.String
	}
	if closedBy.Valid {
		c.ClosedBy = &closedBy.Int64
	}
	return c, nil
}

// ---------------------------------------------------------------------------
// SQL null helpers
// ---------------------------------------------------------------------------

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullableStringPtr(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func nullableInt64(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

func nullableInt64Value(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v != 0}
}

func nullableFloat64(v *float64) sql.NullFloat64 {
	if v == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *v, Valid: true}
}
