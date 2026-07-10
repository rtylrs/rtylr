package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/modules/crm_sales/internal/infrastructure/persistence/sqlc"
	customersync "github.com/voxire/onion/internal/platform/customer_sync"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	db      *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		db:      db,
	}
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
		BusinessID:   ownerID,
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
	return r.GetLead(ctx, id, ownerID)
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

func (r *Repository) DeleteLead(ctx context.Context, id, businessID int64) error {
	result, err := r.queries.DeleteLead(ctx, sqlc.DeleteLeadParams{ID: id, BusinessID: businessID})
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

func (r *Repository) CreateContact(ctx context.Context, businessID int64, req domain.CreateContactRequest) (domain.Contact, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Contact{}, err
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			logger.Ctx(ctx).Error("failed to rollback create contact tx", logger.Error(rollbackErr))
		}
	}()

	customerID, err := customersync.UpsertCRMCustomer(ctx, tx, businessID, customersync.Contact{
		Name:  req.Name,
		Email: req.Email,
		Phone: req.Phone,
	})
	if err != nil {
		return domain.Contact{}, err
	}

	qtx := r.queries.WithTx(tx)
	contactOwnerID, err := qtx.GetBusinessOwnerID(ctx, businessID)
	if err != nil {
		return domain.Contact{}, err
	}

	result, err := qtx.CreateContact(ctx, sqlc.CreateContactParams{
		OwnerID:    contactOwnerID,
		BusinessID: businessID,
		LeadID:     req.LeadID,
		CustomerID: nullableInt64Value(customerID),
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
	if _, err := qtx.GetContact(ctx, sqlc.GetContactParams{ID: id, BusinessID: businessID}); err != nil {
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

	err = qtx.UpdateContact(ctx, sqlc.UpdateContactParams{
		CustomerID: nullableInt64Value(customerID),
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
	if err := tx.Commit(); err != nil {
		return domain.Contact{}, err
	}
	return r.GetContact(ctx, id, businessID)
}

func (r *Repository) DeleteContact(ctx context.Context, id, businessID int64) error {
	result, err := r.queries.DeleteContact(ctx, sqlc.DeleteContactParams{ID: id, BusinessID: businessID})
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

func (r *Repository) CreateTouchpoint(ctx context.Context, ownerID, createdBy int64, req domain.CreateTouchpointRequest) (domain.Touchpoint, error) {
	tp := req.Type
	if tp == "" {
		tp = "call"
	}
	outcome := req.Outcome
	if outcome == "" {
		outcome = "other"
	}

	result, err := r.queries.CreateTouchpoint(ctx, sqlc.CreateTouchpointParams{
		OwnerID:      ownerID,
		BusinessID:   ownerID,
		LeadID:       req.LeadID,
		ContactID:    nullableInt64(req.ContactID),
		Type:         tp,
		Outcome:      outcome,
		Notes:        nullableString(req.Notes),
		NextFollowUp: nullableDatePtr(req.NextFollowUp),
		CreatedBy:    createdBy,
	})
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
	plan := req.Plan
	if plan == "" {
		plan = "pro"
	}
	seats := req.Seats
	if seats == 0 {
		seats = 1
	}

	result, err := r.queries.CreateContract(ctx, sqlc.CreateContractParams{
		OwnerID:       ownerID,
		BusinessID:    ownerID,
		LeadID:        req.LeadID,
		Plan:          plan,
		Mrr:           strconv.FormatFloat(req.MRR, 'g', -1, 64),
		Seats:         int32(seats),
		SignedAt:      nullableDatePtr(req.SignedAt),
		SignedByName:  nullableString(req.SignedByName),
		ClosedBy:      nullableInt64(req.ClosedBy),
		Notes:         nullableString(req.Notes),
	})
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
	err := r.queries.UpdateContract(ctx, sqlc.UpdateContractParams{
		Plan:         req.Plan,
		Mrr:          strconv.FormatFloat(req.MRR, 'g', -1, 64),
		Seats:        int32(seats),
		SignedAt:     nullableDatePtr(req.SignedAt),
		SignedByName: nullableString(req.SignedByName),
		ClosedBy:     nullableInt64(req.ClosedBy),
		Notes:        nullableString(req.Notes),
		ID:           id,
		BusinessID:   businessID,
	})
	if err != nil {
		return domain.Contract{}, err
	}
	return r.GetContract(ctx, id, businessID)
}

// ---------------------------------------------------------------------------
// Row mappers
// ---------------------------------------------------------------------------

func leadFromRow(row sqlc.ListLeadsRow) domain.Lead {
	return leadFromFields(
		row.ID, row.OwnerID, row.BusinessID, row.CompanyName, row.Industry, row.Size,
		row.Source, row.Stage, row.AssignedTo, row.MrrPotential, row.Notes,
		row.CreatedAt, row.UpdatedAt,
	)
}

func leadFromGetRow(row sqlc.GetLeadRow) domain.Lead {
	return leadFromFields(
		row.ID, row.OwnerID, row.BusinessID, row.CompanyName, row.Industry, row.Size,
		row.Source, row.Stage, row.AssignedTo, row.MrrPotential, row.Notes,
		row.CreatedAt, row.UpdatedAt,
	)
}

func leadFromFields(
	id, ownerID, businessID int64,
	companyName, industry, size, source, stage string,
	assignedTo sql.NullInt64,
	mrrPotential sql.NullString,
	notes string,
	createdAt, updatedAt time.Time,
) domain.Lead {
	l := domain.Lead{
		ID:          id,
		OwnerID:     ownerID,
		BusinessID:  businessID,
		CompanyName: companyName,
		Industry:    industry,
		Size:        size,
		Source:      source,
		Stage:       domain.LeadStage(stage),
		Notes:       notes,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
	if assignedTo.Valid {
		l.AssignedTo = &assignedTo.Int64
	}
	if mrrPotential.Valid && mrrPotential.String != "" {
		if f, err := strconv.ParseFloat(mrrPotential.String, 64); err == nil {
			l.MRRPotential = &f
		}
	}
	return l
}

func contactFromRow(row sqlc.ListContactsRow) domain.Contact {
	return contactFromFields(
		row.ID, row.OwnerID, row.BusinessID, row.LeadID, row.CustomerID,
		row.Name, row.Role, row.Email, row.Phone, row.Linkedin,
		row.IsPrimary, row.Notes, row.CreatedAt, row.UpdatedAt,
	)
}

func contactFromLeadRow(row sqlc.ListContactsByLeadRow) domain.Contact {
	return contactFromFields(
		row.ID, row.OwnerID, row.BusinessID, row.LeadID, row.CustomerID,
		row.Name, row.Role, row.Email, row.Phone, row.Linkedin,
		row.IsPrimary, row.Notes, row.CreatedAt, row.UpdatedAt,
	)
}

func contactFromGetRow(row sqlc.GetContactRow) domain.Contact {
	return contactFromFields(
		row.ID, row.OwnerID, row.BusinessID, row.LeadID, row.CustomerID,
		row.Name, row.Role, row.Email, row.Phone, row.Linkedin,
		row.IsPrimary, row.Notes, row.CreatedAt, row.UpdatedAt,
	)
}

func contactFromFields(
	id, ownerID, businessID, leadID int64,
	customerID sql.NullInt64,
	name, role, email, phone, linkedin string,
	isPrimary bool,
	notes string,
	createdAt, updatedAt time.Time,
) domain.Contact {
	c := domain.Contact{
		ID:         id,
		OwnerID:    ownerID,
		BusinessID: businessID,
		LeadID:     leadID,
		Name:       name,
		Role:       role,
		Email:      email,
		Phone:      phone,
		LinkedIn:   linkedin,
		IsPrimary:  isPrimary,
		Notes:      notes,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}
	if customerID.Valid {
		c.CustomerID = customerID.Int64
	}
	return c
}

func touchpointFromRow(row sqlc.ListTouchpointsRow) domain.Touchpoint {
	return touchpointFromFields(
		row.ID, row.OwnerID, row.BusinessID, row.LeadID, row.ContactID,
		row.Type, row.Outcome, row.Notes, row.NextFollowUp, row.CreatedBy, row.CreatedAt,
	)
}

func touchpointFromGetRow(row sqlc.GetTouchpointRow) domain.Touchpoint {
	return touchpointFromFields(
		row.ID, row.OwnerID, row.BusinessID, row.LeadID, row.ContactID,
		row.Type, row.Outcome, row.Notes, row.NextFollowUp, row.CreatedBy, row.CreatedAt,
	)
}

func touchpointFromFields(
	id, ownerID, businessID, leadID int64,
	contactID sql.NullInt64,
	tpType, outcome, notes string,
	nextFollowUp sql.NullTime,
	createdBy int64,
	createdAt time.Time,
) domain.Touchpoint {
	tp := domain.Touchpoint{
		ID:         id,
		OwnerID:    ownerID,
		BusinessID: businessID,
		LeadID:     leadID,
		Type:       tpType,
		Outcome:    outcome,
		Notes:      notes,
		CreatedBy:  createdBy,
		CreatedAt:  createdAt,
	}
	if contactID.Valid {
		tp.ContactID = &contactID.Int64
	}
	if nextFollowUp.Valid {
		s := nextFollowUp.Time.Format("2006-01-02")
		tp.NextFollowUp = &s
	}
	return tp
}

func contractFromRow(row sqlc.ListContractsRow) domain.Contract {
	return contractFromFields(
		row.ID, row.OwnerID, row.BusinessID, row.LeadID, row.Plan, row.Mrr,
		int(row.Seats), row.SignedAt, row.SignedByName, row.ClosedBy, row.Notes,
		row.CreatedAt, row.UpdatedAt,
	)
}

func contractFromGetRow(row sqlc.GetContractRow) domain.Contract {
	return contractFromFields(
		row.ID, row.OwnerID, row.BusinessID, row.LeadID, row.Plan, row.Mrr,
		int(row.Seats), row.SignedAt, row.SignedByName, row.ClosedBy, row.Notes,
		row.CreatedAt, row.UpdatedAt,
	)
}

func contractFromFields(
	id, ownerID, businessID, leadID int64,
	plan, mrr string,
	seats int,
	signedAt sql.NullTime,
	signedByName string,
	closedBy sql.NullInt64,
	notes string,
	createdAt, updatedAt time.Time,
) domain.Contract {
	c := domain.Contract{
		ID:           id,
		OwnerID:      ownerID,
		BusinessID:   businessID,
		LeadID:       leadID,
		Plan:         plan,
		MRR:          parseDecimalString(mrr),
		Seats:        seats,
		SignedByName: signedByName,
		Notes:        notes,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
	if signedAt.Valid {
		s := signedAt.Time.Format("2006-01-02")
		c.SignedAt = &s
	}
	if closedBy.Valid {
		c.ClosedBy = &closedBy.Int64
	}
	return c
}

func parsePipelineMRR(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case []byte:
		f, err := strconv.ParseFloat(string(n), 64)
		if err != nil {
			return 0
		}
		return f
	case string:
		f, err := strconv.ParseFloat(n, 64)
		if err != nil {
			return 0
		}
		return f
	default:
		return 0
	}
}

func parseDecimalString(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

// ---------------------------------------------------------------------------
// SQL null helpers
// ---------------------------------------------------------------------------

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullableDatePtr(s *string) sql.NullTime {
	if s == nil || *s == "" {
		return sql.NullTime{}
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
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

func nullableDecimalString(v *float64) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: strconv.FormatFloat(*v, 'g', -1, 64), Valid: true}
}
