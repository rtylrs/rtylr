package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/voxire/onion/internal/modules/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{db: dbclient.Pool()}
}

// ---------------------------------------------------------------------------
// Leads
// ---------------------------------------------------------------------------

func (r *Repository) ListLeads(ctx context.Context, ownerID int64) ([]domain.Lead, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, owner_id, company_name, COALESCE(industry,''), COALESCE(size,''),
		        source, stage, assigned_to, mrr_potential, COALESCE(notes,''),
		        created_at, updated_at
		 FROM crm_leads
		 WHERE owner_id = ?
		 ORDER BY created_at DESC`,
		ownerID,
	)
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

func (r *Repository) GetLead(ctx context.Context, id, ownerID int64) (domain.Lead, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, owner_id, company_name, COALESCE(industry,''), COALESCE(size,''),
		        source, stage, assigned_to, mrr_potential, COALESCE(notes,''),
		        created_at, updated_at
		 FROM crm_leads
		 WHERE id = ? AND owner_id = ?`,
		id, ownerID,
	)
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

	result, err := r.db.ExecContext(ctx,
		`INSERT INTO crm_leads (owner_id, company_name, industry, size, source, stage, assigned_to, mrr_potential, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		ownerID, req.CompanyName, nullableString(req.Industry), nullableString(req.Size),
		source, string(stage), nullableInt64(req.AssignedTo), nullableFloat64(req.MRRPotential),
		nullableString(req.Notes),
	)
	if err != nil {
		return domain.Lead{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Lead{}, err
	}
	return r.GetLead(ctx, id, ownerID)
}

func (r *Repository) UpdateLead(ctx context.Context, id, ownerID int64, req domain.UpdateLeadRequest) (domain.Lead, error) {
	_, err := r.db.ExecContext(ctx,
		`UPDATE crm_leads
		 SET company_name = ?, industry = ?, size = ?, source = ?, stage = ?,
		     assigned_to = ?, mrr_potential = ?, notes = ?
		 WHERE id = ? AND owner_id = ?`,
		req.CompanyName, nullableString(req.Industry), nullableString(req.Size),
		req.Source, string(req.Stage), nullableInt64(req.AssignedTo),
		nullableFloat64(req.MRRPotential), nullableString(req.Notes),
		id, ownerID,
	)
	if err != nil {
		return domain.Lead{}, err
	}
	return r.GetLead(ctx, id, ownerID)
}

func (r *Repository) DeleteLead(ctx context.Context, id, ownerID int64) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM crm_leads WHERE id = ? AND owner_id = ?`,
		id, ownerID,
	)
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

func (r *Repository) GetPipeline(ctx context.Context, ownerID int64) ([]domain.PipelineSummary, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT stage, COUNT(*) as cnt, COALESCE(SUM(mrr_potential), 0) as total_mrr
		 FROM crm_leads
		 WHERE owner_id = ?
		 GROUP BY stage`,
		ownerID,
	)
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

func (r *Repository) ListContacts(ctx context.Context, ownerID int64, leadID *int64) ([]domain.Contact, error) {
	query := `SELECT id, owner_id, lead_id, name, COALESCE(role,''), COALESCE(email,''),
	                 COALESCE(phone,''), COALESCE(linkedin,''), is_primary, COALESCE(notes,''),
	                 created_at, updated_at
	          FROM crm_contacts
	          WHERE owner_id = ?`
	args := []interface{}{ownerID}
	if leadID != nil {
		query += " AND lead_id = ?"
		args = append(args, *leadID)
	}
	query += " ORDER BY created_at DESC"

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

func (r *Repository) GetContact(ctx context.Context, id, ownerID int64) (domain.Contact, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, owner_id, lead_id, name, COALESCE(role,''), COALESCE(email,''),
		        COALESCE(phone,''), COALESCE(linkedin,''), is_primary, COALESCE(notes,''),
		        created_at, updated_at
		 FROM crm_contacts
		 WHERE id = ? AND owner_id = ?`,
		id, ownerID,
	)
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
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO crm_contacts (owner_id, lead_id, name, role, email, phone, linkedin, is_primary, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		ownerID, req.LeadID, req.Name, nullableString(req.Role), nullableString(req.Email),
		nullableString(req.Phone), nullableString(req.LinkedIn), req.IsPrimary, nullableString(req.Notes),
	)
	if err != nil {
		return domain.Contact{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Contact{}, err
	}
	return r.GetContact(ctx, id, ownerID)
}

func (r *Repository) UpdateContact(ctx context.Context, id, ownerID int64, req domain.UpdateContactRequest) (domain.Contact, error) {
	_, err := r.db.ExecContext(ctx,
		`UPDATE crm_contacts
		 SET name = ?, role = ?, email = ?, phone = ?, linkedin = ?, is_primary = ?, notes = ?
		 WHERE id = ? AND owner_id = ?`,
		req.Name, nullableString(req.Role), nullableString(req.Email), nullableString(req.Phone),
		nullableString(req.LinkedIn), req.IsPrimary, nullableString(req.Notes),
		id, ownerID,
	)
	if err != nil {
		return domain.Contact{}, err
	}
	return r.GetContact(ctx, id, ownerID)
}

func (r *Repository) DeleteContact(ctx context.Context, id, ownerID int64) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM crm_contacts WHERE id = ? AND owner_id = ?`,
		id, ownerID,
	)
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

func (r *Repository) ListTouchpoints(ctx context.Context, ownerID, leadID int64) ([]domain.Touchpoint, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, owner_id, lead_id, contact_id, type, outcome,
		        COALESCE(notes,''), next_follow_up, created_by, created_at
		 FROM crm_touchpoints
		 WHERE owner_id = ? AND lead_id = ?
		 ORDER BY created_at DESC`,
		ownerID, leadID,
	)
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

func (r *Repository) GetTouchpoint(ctx context.Context, id, ownerID int64) (domain.Touchpoint, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, owner_id, lead_id, contact_id, type, outcome,
		        COALESCE(notes,''), next_follow_up, created_by, created_at
		 FROM crm_touchpoints
		 WHERE id = ? AND owner_id = ?`,
		id, ownerID,
	)
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

	result, err := r.db.ExecContext(ctx,
		`INSERT INTO crm_touchpoints (owner_id, lead_id, contact_id, type, outcome, notes, next_follow_up, created_by)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		ownerID, req.LeadID, nullableInt64(req.ContactID), tp, outcome,
		nullableString(req.Notes), nullableStringPtr(req.NextFollowUp), createdBy,
	)
	if err != nil {
		return domain.Touchpoint{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Touchpoint{}, err
	}
	return r.GetTouchpoint(ctx, id, ownerID)
}

func (r *Repository) DeleteTouchpoint(ctx context.Context, id, ownerID int64) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM crm_touchpoints WHERE id = ? AND owner_id = ?`,
		id, ownerID,
	)
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

func (r *Repository) ListContracts(ctx context.Context, ownerID int64) ([]domain.Contract, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, owner_id, lead_id, plan, mrr, seats,
		        signed_at, COALESCE(signed_by_name,''), closed_by, COALESCE(notes,''),
		        created_at, updated_at
		 FROM crm_contracts
		 WHERE owner_id = ?
		 ORDER BY created_at DESC`,
		ownerID,
	)
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

func (r *Repository) GetContract(ctx context.Context, id, ownerID int64) (domain.Contract, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, owner_id, lead_id, plan, mrr, seats,
		        signed_at, COALESCE(signed_by_name,''), closed_by, COALESCE(notes,''),
		        created_at, updated_at
		 FROM crm_contracts
		 WHERE id = ? AND owner_id = ?`,
		id, ownerID,
	)
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

	result, err := r.db.ExecContext(ctx,
		`INSERT INTO crm_contracts (owner_id, lead_id, plan, mrr, seats, signed_at, signed_by_name, closed_by, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		ownerID, req.LeadID, plan, req.MRR, seats, nullableStringPtr(req.SignedAt),
		nullableString(req.SignedByName), nullableInt64(req.ClosedBy), nullableString(req.Notes),
	)
	if err != nil {
		return domain.Contract{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Contract{}, err
	}
	return r.GetContract(ctx, id, ownerID)
}

func (r *Repository) UpdateContract(ctx context.Context, id, ownerID int64, req domain.UpdateContractRequest) (domain.Contract, error) {
	seats := req.Seats
	if seats == 0 {
		seats = 1
	}
	_, err := r.db.ExecContext(ctx,
		`UPDATE crm_contracts
		 SET plan = ?, mrr = ?, seats = ?, signed_at = ?, signed_by_name = ?, closed_by = ?, notes = ?
		 WHERE id = ? AND owner_id = ?`,
		req.Plan, req.MRR, seats, nullableStringPtr(req.SignedAt),
		nullableString(req.SignedByName), nullableInt64(req.ClosedBy), nullableString(req.Notes),
		id, ownerID,
	)
	if err != nil {
		return domain.Contract{}, err
	}
	return r.GetContract(ctx, id, ownerID)
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
		&l.ID, &l.OwnerID, &l.CompanyName, &l.Industry, &l.Size,
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
		&l.ID, &l.OwnerID, &l.CompanyName, &l.Industry, &l.Size,
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
	err := s.Scan(
		&c.ID, &c.OwnerID, &c.LeadID, &c.Name, &c.Role, &c.Email,
		&c.Phone, &c.LinkedIn, &c.IsPrimary, &c.Notes,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return domain.Contact{}, err
	}
	return c, nil
}

func scanContactRow(row *sql.Row) (domain.Contact, error) {
	var c domain.Contact
	err := row.Scan(
		&c.ID, &c.OwnerID, &c.LeadID, &c.Name, &c.Role, &c.Email,
		&c.Phone, &c.LinkedIn, &c.IsPrimary, &c.Notes,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return domain.Contact{}, err
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
		&tp.ID, &tp.OwnerID, &tp.LeadID, &contactID, &tp.Type, &tp.Outcome,
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
		&tp.ID, &tp.OwnerID, &tp.LeadID, &contactID, &tp.Type, &tp.Outcome,
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
		&c.ID, &c.OwnerID, &c.LeadID, &c.Plan, &c.MRR, &c.Seats,
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
		&c.ID, &c.OwnerID, &c.LeadID, &c.Plan, &c.MRR, &c.Seats,
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

func nullableFloat64(v *float64) sql.NullFloat64 {
	if v == nil {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *v, Valid: true}
}

// parseFloat64 is kept for future use in other scan helpers that may read
// string-encoded decimals from older schema columns.
func parseFloat64(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}
