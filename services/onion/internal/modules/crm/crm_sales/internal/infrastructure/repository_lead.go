package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/money"
)

func ensureLeadCanContract(ctx context.Context, tx *sql.Tx, businessID, leadID int64) error {
	var stage string
	if err := tx.QueryRowContext(ctx,
		"SELECT stage FROM crm_leads WHERE id = ? AND business_id = ? FOR UPDATE",
		leadID, businessID,
	).Scan(&stage); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrLeadNotFound
		}
		return err
	}
	if !domain.LeadStage(stage).CanTransitionTo(domain.LeadStageContracted) {
		return domain.ErrInvalidLeadTransition
	}
	return nil
}

func ensureContactForLead(ctx context.Context, tx *sql.Tx, businessID, leadID, contactID int64) error {
	var exists bool
	if err := tx.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM crm_contacts WHERE id = ? AND lead_id = ? AND business_id = ?)",
		contactID, leadID, businessID,
	).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return domain.ErrContactNotFound
	}
	return nil
}

func (r *Repository) UpdateContract(ctx context.Context, id, businessID int64, req domain.UpdateContractRequest) (domain.Contract, error) {
	seats := req.Seats
	if seats == 0 {
		seats = 1
	}
	signedAt, err := nullableDatePtr(req.SignedAt)
	if err != nil {
		return domain.Contract{}, err
	}
	terms, err := contractTermParams(req.Value, req.Billing, req.StartDate, req.EndDate)
	if err != nil {
		return domain.Contract{}, err
	}
	// An update that names no customer keeps the linked one, or links the
	// customer the lead's contact has since become.
	existing, err := r.GetContract(ctx, id, businessID)
	if err != nil {
		return domain.Contract{}, err
	}
	customerID := nullableInt64(req.CustomerID)
	if !customerID.Valid {
		customerID = nullableInt64(existing.CustomerID)
	}
	if !customerID.Valid {
		customerID, err = leadCustomerID(ctx, r.queries, businessID, existing.LeadID)
		if err != nil {
			return domain.Contract{}, err
		}
	}
	err = r.queries.UpdateContract(ctx, sqlc.UpdateContractParams{
		Plan:         req.Plan,
		Mrr:          req.MRR.String(),
		Seats:        int32(seats),
		Value:        terms.value,
		Billing:      terms.billing,
		StartDate:    terms.startDate,
		EndDate:      terms.endDate,
		CustomerID:   customerID,
		SignedAt:     signedAt,
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
		row.CreatedAt, row.UpdatedAt, row.NextFollowUp,
	)
}

func leadFromGetRow(row sqlc.GetLeadRow) domain.Lead {
	return leadFromFields(
		row.ID, row.OwnerID, row.BusinessID, row.CompanyName, row.Industry, row.Size,
		row.Source, row.Stage, row.AssignedTo, row.MrrPotential, row.Notes,
		row.CreatedAt, row.UpdatedAt, row.NextFollowUp,
	)
}

func leadFromFields(
	id, ownerID, businessID int64,
	companyName, industry, size, source, stage string,
	assignedTo sql.NullInt64,
	mrrPotential sql.NullString,
	notes string,
	createdAt, updatedAt time.Time,
	nextFollowUp interface{},
) domain.Lead {
	l := domain.Lead{
		ID:           id,
		OwnerID:      ownerID,
		BusinessID:   businessID,
		CompanyName:  companyName,
		Industry:     industry,
		Size:         size,
		Source:       source,
		Stage:        domain.LeadStage(stage),
		Notes:        notes,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		NextFollowUp: followUpDate(nextFollowUp),
	}
	if assignedTo.Valid {
		l.AssignedTo = &assignedTo.Int64
	}
	if mrrPotential.Valid && mrrPotential.String != "" {
		if amount, err := money.FromString(mrrPotential.String); err == nil {
			l.MRRPotential = &amount
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

// followUpDate reads the next_follow_up column, which the query formats as
// YYYY-MM-DD ("" when the latest touchpoint set none). sqlc cannot type the
// subquery, so the driver hands it over as bytes.
func followUpDate(v interface{}) string {
	switch value := v.(type) {
	case []byte:
		return string(value)
	case string:
		return value
	case time.Time:
		return value.Format("2006-01-02")
	}
	return ""
}

// contractParams are a contract's value, billing and term as the queries
// take them.
type contractParams struct {
	value     string
	billing   sqlc.CrmContractsBilling
	startDate sql.NullTime
	endDate   sql.NullTime
}

func contractTermParams(value money.Amount, billing domain.Billing, startDate, endDate *string) (contractParams, error) {
	start, err := nullableDatePtr(startDate)
	if err != nil {
		return contractParams{}, err
	}
	end, err := nullableDatePtr(endDate)
	if err != nil {
		return contractParams{}, err
	}
	if billing == "" {
		billing = domain.BillingMonthly
	}
	return contractParams{
		value:     value.String(),
		billing:   sqlc.CrmContractsBilling(billing),
		startDate: start,
		endDate:   end,
	}, nil
}

// leadCustomerID is the CRM customer the lead's primary (or first linked)
// contact became, or NULL while no contact has been captured yet.
func leadCustomerID(ctx context.Context, queries *sqlc.Queries, businessID, leadID int64) (sql.NullInt64, error) {
	customerID, err := queries.GetLeadCustomerID(ctx, sqlc.GetLeadCustomerIDParams{LeadID: leadID, BusinessID: businessID})
	if errors.Is(err, sql.ErrNoRows) {
		return sql.NullInt64{}, nil
	}
	return customerID, err
}
