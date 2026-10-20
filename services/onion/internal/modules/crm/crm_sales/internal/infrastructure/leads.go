package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	crmcontract "github.com/voxire/onion/internal/contracts/crm"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/modules/crm/internal/outbox"
	"github.com/voxire/onion/internal/platform/eventbus"
)

// leadFilterArgs are a filter's query arguments, shared by every lead list.
type leadFilterArgs struct {
	statuses     []sqlc.CrmLeadsStatus
	assignedTo   sql.NullInt64
	source       sql.NullString
	batch        sql.NullInt64
	followFrom   sql.NullTime
	followTo     sql.NullTime
	search       sql.NullString
	searchDigits sql.NullString
}

func leadArgs(f domain.LeadFilter, today time.Time) leadFilterArgs {
	from, to := f.FollowUp.Range(today)
	return leadFilterArgs{
		statuses:     leadStatuses(f.Statuses),
		assignedTo:   nullInt64(f.AssignedTo),
		source:       nullString(f.Source),
		batch:        nullInt64(f.ImportBatchID),
		followFrom:   nullTime(from),
		followTo:     nullTime(to),
		search:       likePattern(f.Search),
		searchDigits: digitsPattern(f.Search),
	}
}

// ListLeads returns one page of leads in the filter's sort. today is the
// caller's calendar date, which follow-up windows are relative to.
func (r *Repository) ListLeads(ctx context.Context, businessID int64, f domain.LeadFilter, today time.Time) (domain.LeadPage, error) {
	cursor, err := decodeCursor(f.Cursor, string(f.Sort))
	if err != nil {
		return domain.LeadPage{}, err
	}
	a := leadArgs(f, today)
	limit := int32(f.Limit + 1) // #nosec G115 -- the service caps Limit at 200.

	var rows []sqlc.CrmLead
	switch f.Sort {
	case domain.LeadSortFollowUp:
		var at sql.NullTime
		if at, err = cursor.date(); err != nil {
			return domain.LeadPage{}, err
		}
		rows, err = r.queries.ListLeadsByFollowUp(ctx, sqlc.ListLeadsByFollowUpParams{
			BusinessID: businessID, Statuses: a.statuses, AssignedTo: a.assignedTo, Source: a.source,
			ImportBatchID: a.batch, FollowUpFrom: a.followFrom, FollowUpTo: a.followTo,
			Search: a.search, SearchDigits: a.searchDigits, CursorID: cursor.id(), CursorDate: at, Limit: limit,
		})
	case domain.LeadSortActivity:
		var at sql.NullTime
		if at, err = cursor.time(); err != nil {
			return domain.LeadPage{}, err
		}
		rows, err = r.queries.ListLeadsByActivity(ctx, sqlc.ListLeadsByActivityParams{
			BusinessID: businessID, Statuses: a.statuses, AssignedTo: a.assignedTo, Source: a.source,
			ImportBatchID: a.batch, FollowUpFrom: a.followFrom, FollowUpTo: a.followTo,
			Search: a.search, SearchDigits: a.searchDigits, CursorID: cursor.id(), CursorTime: at, Limit: limit,
		})
	case domain.LeadSortCompany:
		rows, err = r.queries.ListLeadsByCompany(ctx, sqlc.ListLeadsByCompanyParams{
			BusinessID: businessID, Statuses: a.statuses, AssignedTo: a.assignedTo, Source: a.source,
			ImportBatchID: a.batch, FollowUpFrom: a.followFrom, FollowUpTo: a.followTo,
			Search: a.search, SearchDigits: a.searchDigits, CursorID: cursor.id(), CursorText: cursor.text(), Limit: limit,
		})
	default:
		var at sql.NullTime
		if at, err = cursor.time(); err != nil {
			return domain.LeadPage{}, err
		}
		rows, err = r.queries.ListLeadsNewest(ctx, sqlc.ListLeadsNewestParams{
			BusinessID: businessID, Statuses: a.statuses, AssignedTo: a.assignedTo, Source: a.source,
			ImportBatchID: a.batch, FollowUpFrom: a.followFrom, FollowUpTo: a.followTo,
			Search: a.search, SearchDigits: a.searchDigits, CursorID: cursor.id(), CursorTime: at, Limit: limit,
		})
	}
	if err != nil {
		return domain.LeadPage{}, err
	}

	total, err := r.queries.CountLeads(ctx, sqlc.CountLeadsParams{
		BusinessID: businessID, Statuses: a.statuses, AssignedTo: a.assignedTo, Source: a.source,
		ImportBatchID: a.batch, FollowUpFrom: a.followFrom, FollowUpTo: a.followTo,
		Search: a.search, SearchDigits: a.searchDigits,
	})
	if err != nil {
		return domain.LeadPage{}, err
	}

	page := domain.LeadPage{Total: total}
	if len(rows) > f.Limit {
		rows = rows[:f.Limit]
		last := rows[len(rows)-1]
		page.NextCursor = encodeCursor(string(f.Sort), leadSortValue(f.Sort, last), last.ID)
	}
	page.Leads = leadsFromRows(rows)
	return page, nil
}

func leadSortValue(sort domain.LeadSort, row sqlc.CrmLead) string {
	switch sort {
	case domain.LeadSortFollowUp:
		if row.NextFollowUp.Valid {
			return row.NextFollowUp.Time.Format(time.DateOnly)
		}
		return noFollowUpDate
	case domain.LeadSortActivity:
		if row.LastActivityAt.Valid {
			return timeValue(row.LastActivityAt.Time)
		}
		return timeValue(neverWorked)
	case domain.LeadSortCompany:
		return row.CompanyName
	}
	return timeValue(row.CreatedAt)
}

// LeadStats counts leads by status and the work due, for one rep or everyone.
// dayStartUTC is when the caller's day began, for today's call counts.
func (r *Repository) LeadStats(ctx context.Context, businessID int64, assignedTo *int64, today, dayStartUTC, monthStartUTC time.Time) (domain.LeadStats, error) {
	assigned := nullInt64(assignedTo)
	stats := domain.LeadStats{ByStatus: map[domain.LeadStatus]int64{}}
	byStatus, err := r.queries.CountLeadsByStatus(ctx, sqlc.CountLeadsByStatusParams{BusinessID: businessID, AssignedTo: assigned})
	if err != nil {
		return stats, err
	}
	for _, row := range byStatus {
		stats.ByStatus[domain.LeadStatus(row.Status)] = row.Total
	}
	followUps, err := r.queries.CountLeadFollowUps(ctx, sqlc.CountLeadFollowUpsParams{
		Today: someTime(dayStart(today)), BusinessID: businessID, AssignedTo: assigned,
	})
	if err != nil {
		return stats, err
	}
	stats.FollowUpsOverdue, stats.FollowUpsToday = followUps.Overdue, followUps.DueToday
	if stats.CallsToday, err = r.queries.CountCallsSince(ctx, sqlc.CountCallsSinceParams{
		BusinessID: businessID, Since: dayStartUTC, CreatedBy: assigned,
	}); err != nil {
		return stats, err
	}
	if stats.ConnectedToday, err = r.queries.CountConnectedCallsSince(ctx, sqlc.CountConnectedCallsSinceParams{
		BusinessID: businessID, Since: dayStartUTC, CreatedBy: assigned,
	}); err != nil {
		return stats, err
	}
	stats.ConvertedThisMonth, err = r.queries.CountLeadsConvertedSince(ctx, sqlc.CountLeadsConvertedSinceParams{
		BusinessID: businessID, Since: someTime(monthStartUTC), AssignedTo: assigned,
	})
	return stats, err
}

func (r *Repository) ListLeadSources(ctx context.Context, businessID int64) ([]string, error) {
	return r.queries.ListLeadSources(ctx, businessID)
}

func (r *Repository) GetLead(ctx context.Context, businessID, id int64) (domain.Lead, error) {
	row, err := r.queries.GetLead(ctx, sqlc.GetLeadParams{ID: id, BusinessID: businessID})
	if err != nil {
		return domain.Lead{}, notFound(err, domain.ErrLeadNotFound)
	}
	return leadFromRow(row), nil
}

func createLeadParams(businessID, ownerID int64, in domain.LeadInput, batchID *int64) sqlc.CreateLeadParams {
	return sqlc.CreateLeadParams{
		OwnerID:       ownerID,
		BusinessID:    businessID,
		FirstName:     nullString(in.FirstName),
		LastName:      nullString(in.LastName),
		JobTitle:      nullString(in.JobTitle),
		Email:         nullString(in.Email),
		Phone:         nullString(in.Phone),
		PhoneDigits:   nullString(domain.PhoneDigits(in.Phone)),
		CompanyName:   in.CompanyName,
		Website:       nullString(in.Website),
		Linkedin:      nullString(in.LinkedIn),
		Industry:      nullString(in.Industry),
		Size:          nullString(in.Size),
		City:          nullString(in.City),
		Country:       nullString(in.Country),
		Source:        in.Source,
		Status:        sqlc.CrmLeadsStatusNew,
		AssignedTo:    nullInt64(in.AssignedTo),
		Value:         nullAmount(in.Value),
		Billing:       sqlc.CrmLeadsBilling(in.Billing),
		Notes:         nullString(in.Notes),
		ImportBatchID: nullInt64(batchID),
	}
}

// CreateLead stores a normalized lead as new.
func (r *Repository) CreateLead(ctx context.Context, businessID, ownerID int64, in domain.LeadInput) (domain.Lead, error) {
	result, err := r.queries.CreateLead(ctx, createLeadParams(businessID, ownerID, in, nil))
	if err != nil {
		return domain.Lead{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Lead{}, err
	}
	return r.GetLead(ctx, businessID, id)
}

// UpdateLead applies a full edit, checking the status change against the
// lead's status under a row lock.
func (r *Repository) UpdateLead(ctx context.Context, businessID, id int64, changes domain.LeadChanges) (domain.Lead, error) {
	err := r.inTx(ctx, func(q *sqlc.Queries, _ *sql.Tx) error {
		current, err := q.GetLeadForUpdate(ctx, sqlc.GetLeadForUpdateParams{ID: id, BusinessID: businessID})
		if err != nil {
			return notFound(err, domain.ErrLeadNotFound)
		}
		if err := changes.Normalize(domain.LeadStatus(current.Status)); err != nil {
			return err
		}
		in := changes.LeadInput
		return q.UpdateLead(ctx, sqlc.UpdateLeadParams{
			FirstName:        nullString(in.FirstName),
			LastName:         nullString(in.LastName),
			JobTitle:         nullString(in.JobTitle),
			Email:            nullString(in.Email),
			Phone:            nullString(in.Phone),
			PhoneDigits:      nullString(domain.PhoneDigits(in.Phone)),
			CompanyName:      in.CompanyName,
			Website:          nullString(in.Website),
			Linkedin:         nullString(in.LinkedIn),
			Industry:         nullString(in.Industry),
			Size:             nullString(in.Size),
			City:             nullString(in.City),
			Country:          nullString(in.Country),
			Source:           in.Source,
			Status:           sqlc.CrmLeadsStatus(changes.Status),
			DisqualifyReason: nullString(changes.DisqualifyReason),
			AssignedTo:       nullInt64(in.AssignedTo),
			Value:            nullAmount(in.Value),
			Billing:          sqlc.CrmLeadsBilling(in.Billing),
			Notes:            nullString(in.Notes),
			ID:               id,
			BusinessID:       businessID,
		})
	})
	if err != nil {
		return domain.Lead{}, err
	}
	return r.GetLead(ctx, businessID, id)
}

// DeleteLead removes an unconverted lead with its tasks and activities.
func (r *Repository) DeleteLead(ctx context.Context, businessID, id int64) error {
	return r.inTx(ctx, func(q *sqlc.Queries, _ *sql.Tx) error {
		lead, err := q.GetLeadForUpdate(ctx, sqlc.GetLeadForUpdateParams{ID: id, BusinessID: businessID})
		if err != nil {
			return notFound(err, domain.ErrLeadNotFound)
		}
		if domain.LeadStatus(lead.Status) == domain.LeadQualified {
			return domain.ErrLeadConverted
		}
		return deleteLeads(ctx, q, businessID, []int64{id})
	})
}

// DeleteLeads removes the unconverted leads among ids and reports how many.
func (r *Repository) DeleteLeads(ctx context.Context, businessID int64, ids []int64) (int64, error) {
	var deleted int64
	err := r.inTx(ctx, func(q *sqlc.Queries, _ *sql.Tx) error {
		deletable, err := q.ListDeletableLeadIDs(ctx, sqlc.ListDeletableLeadIDsParams{BusinessID: businessID, Ids: ids})
		if err != nil || len(deletable) == 0 {
			return err
		}
		deleted = int64(len(deletable))
		return deleteLeads(ctx, q, businessID, deletable)
	})
	return deleted, err
}

func deleteLeads(ctx context.Context, q *sqlc.Queries, businessID int64, ids []int64) error {
	if err := q.DeleteLeadTasks(ctx, sqlc.DeleteLeadTasksParams{BusinessID: businessID, LeadIds: nullInt64s(ids)}); err != nil {
		return err
	}
	if err := q.DeleteLeadActivities(ctx, sqlc.DeleteLeadActivitiesParams{BusinessID: businessID, LeadIds: nullInt64s(ids)}); err != nil {
		return err
	}
	_, err := q.DeleteLeads(ctx, sqlc.DeleteLeadsParams{BusinessID: businessID, Ids: ids})
	return err
}

// UpdateLeads applies a bulk assignment or status change and reports how many
// leads changed. Converted leads keep their status.
func (r *Repository) UpdateLeads(ctx context.Context, businessID int64, update domain.BulkLeadUpdate) (int64, error) {
	var (
		result sql.Result
		err    error
	)
	if update.Status != "" {
		result, err = r.queries.SetLeadsStatus(ctx, sqlc.SetLeadsStatusParams{
			Status:           sqlc.CrmLeadsStatus(update.Status),
			DisqualifyReason: nullString(update.DisqualifyReason),
			BusinessID:       businessID,
			Ids:              update.IDs,
		})
	} else {
		result, err = r.queries.AssignLeads(ctx, sqlc.AssignLeadsParams{
			AssignedTo: nullInt64(update.AssignTo),
			BusinessID: businessID,
			Ids:        update.IDs,
		})
	}
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// ConvertLead qualifies a lead: it finds or creates the company, opens a deal
// in stage when one is given, and publishes the lead's contact so the
// customers module matches or creates the person.
func (r *Repository) ConvertLead(ctx context.Context, businessID, actorID, id int64, deal *domain.DealInput, stage *domain.Stage, now time.Time) (domain.Lead, *domain.Deal, error) {
	var dealID int64
	err := r.inTx(ctx, func(q *sqlc.Queries, tx *sql.Tx) error {
		row, err := q.GetLeadForUpdate(ctx, sqlc.GetLeadForUpdateParams{ID: id, BusinessID: businessID})
		if err != nil {
			return notFound(err, domain.ErrLeadNotFound)
		}
		lead := leadFromRow(row)
		if lead.Status == domain.LeadQualified {
			return domain.ErrLeadConverted
		}

		companyID, err := findOrCreateCompany(ctx, q, businessID, actorID, lead)
		if err != nil {
			return err
		}
		if deal != nil && stage != nil {
			assigned := deal.AssignedTo
			if assigned == nil {
				assigned = lead.AssignedTo
			}
			result, err := q.CreateDeal(ctx, sqlc.CreateDealParams{
				BusinessID:        businessID,
				OwnerID:           actorID,
				PipelineID:        stage.PipelineID,
				StageID:           stage.ID,
				Title:             deal.Title,
				LeadID:            someInt64(lead.ID),
				CustomerID:        nullInt64(lead.CustomerID),
				CompanyID:         companyID,
				ContactName:       lead.PersonName(),
				CompanyName:       lead.CompanyName,
				Value:             deal.Value.String(),
				Billing:           sqlc.CrmDealsBilling(deal.Billing),
				ExpectedCloseDate: nullDate(deal.ExpectedCloseDate),
				AssignedTo:        nullInt64(assigned),
				Source:            lead.Source,
				Status:            sqlc.CrmDealsStatusOpen,
				Notes:             nullString(deal.Notes),
			})
			if err != nil {
				return err
			}
			if dealID, err = result.LastInsertId(); err != nil {
				return err
			}
			if err := q.AttachLeadTasksToDeal(ctx, sqlc.AttachLeadTasksToDealParams{
				DealID: someInt64(dealID), BusinessID: businessID, LeadID: someInt64(lead.ID),
			}); err != nil {
				return err
			}
			if err := q.SetDealNextFollowUp(ctx, sqlc.SetDealNextFollowUpParams{ID: dealID, BusinessID: businessID}); err != nil {
				return err
			}
			if lead.LastActivityAt != nil {
				if err := q.RecordDealWork(ctx, sqlc.RecordDealWorkParams{
					LastActivityAt: nullTime(lead.LastActivityAt), ID: dealID, BusinessID: businessID,
				}); err != nil {
					return err
				}
			}
		}

		var dealRef sql.NullInt64
		if dealID > 0 {
			dealRef = someInt64(dealID)
		}
		if err := q.MarkLeadConverted(ctx, sqlc.MarkLeadConvertedParams{
			CompanyID: companyID, DealID: dealRef, ConvertedAt: someTime(now), ID: lead.ID, BusinessID: businessID,
		}); err != nil {
			return err
		}
		return publishLeadContact(ctx, tx, businessID, lead)
	})
	if err != nil {
		return domain.Lead{}, nil, err
	}
	lead, err := r.GetLead(ctx, businessID, id)
	if err != nil || dealID == 0 {
		return lead, nil, err
	}
	created, err := r.GetDeal(ctx, businessID, dealID)
	if err != nil {
		return lead, nil, err
	}
	return lead, &created, nil
}

func findOrCreateCompany(ctx context.Context, q *sqlc.Queries, businessID, actorID int64, lead domain.Lead) (sql.NullInt64, error) {
	if lead.CompanyName == "" {
		return sql.NullInt64{}, nil
	}
	existing, err := q.FindCompanyByName(ctx, sqlc.FindCompanyByNameParams{BusinessID: businessID, Name: lead.CompanyName})
	if err == nil {
		return someInt64(existing.ID), nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return sql.NullInt64{}, err
	}
	result, err := q.CreateCompany(ctx, sqlc.CreateCompanyParams{
		BusinessID: businessID,
		OwnerID:    actorID,
		Name:       lead.CompanyName,
		Website:    nullString(lead.Website),
		Industry:   nullString(lead.Industry),
		Size:       nullString(lead.Size),
		City:       nullString(lead.City),
		Country:    nullString(lead.Country),
		AssignedTo: nullInt64(lead.AssignedTo),
	})
	if err != nil {
		return sql.NullInt64{}, err
	}
	companyID, err := result.LastInsertId()
	if err != nil {
		return sql.NullInt64{}, err
	}
	return someInt64(companyID), nil
}

// publishLeadContact hands a converted lead's person to the customers module,
// which links the matched or created customer back (LinkLeadCustomer).
func publishLeadContact(ctx context.Context, tx *sql.Tx, businessID int64, lead domain.Lead) error {
	if lead.CustomerID != nil {
		return nil
	}
	name := lead.PersonName()
	if name == "" && lead.Email == "" && lead.Phone == "" {
		return nil
	}
	if name == "" {
		name = lead.CompanyName
	}
	contact := crmcontract.CustomerContactCaptured{Name: name, Email: lead.Email, Phone: lead.Phone}
	return outbox.Publish(ctx, tx, eventbus.Event{
		BusinessID:  businessID,
		Type:        crmcontract.CustomerContactCapturedType,
		SubjectType: LeadSubjectType,
		SubjectID:   lead.ID,
		Payload:     contact,
		DedupeKey:   crmcontract.CustomerContactCaptureDedupeKey(LeadSubjectType, lead.ID, contact),
	})
}

// LeadSubjectType names converted leads in customer capture and link events.
const LeadSubjectType = "crm_lead"

// LinkLeadCustomer records the customer a converted lead became on the lead,
// its deals and their contracts.
func (r *Repository) LinkLeadCustomer(ctx context.Context, businessID, leadID, customerID int64) error {
	return r.inTx(ctx, func(q *sqlc.Queries, _ *sql.Tx) error {
		customer := someInt64(customerID)
		if err := q.LinkLeadCustomer(ctx, sqlc.LinkLeadCustomerParams{CustomerID: customer, ID: leadID, BusinessID: businessID}); err != nil {
			return err
		}
		if err := q.LinkLeadDealsCustomer(ctx, sqlc.LinkLeadDealsCustomerParams{
			CustomerID: customer, BusinessID: businessID, LeadID: someInt64(leadID),
		}); err != nil {
			return err
		}
		return q.LinkDealContractsCustomer(ctx, sqlc.LinkDealContractsCustomerParams{
			CustomerID: customer, BusinessID: businessID, LeadID: someInt64(leadID),
		})
	})
}
