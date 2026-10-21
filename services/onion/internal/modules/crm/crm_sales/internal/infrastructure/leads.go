package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"time"

	crmcontract "github.com/voxire/onion/internal/contracts/crm"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/modules/crm/internal/outbox"
	"github.com/voxire/onion/internal/platform/eventbus"
)

// leadFilterArgs are a filter's query arguments, shared by every lead list
// and the per-status counts. Flags are 0 or 1.
type leadFilterArgs struct {
	filterReps       int64
	reps             []int64
	unassigned       int64
	filterSources    int64
	sources          []string
	filterIndustries int64
	industries       []sql.NullString
	filterSizes      int64
	sizes            []sql.NullString
	filterCountries  int64
	countries        []sql.NullString
	filterCities     int64
	cities           []sql.NullString
	filterBilling    int64
	billing          []sqlc.CrmLeadsBilling
	batch            sql.NullInt64
	followFrom       sql.NullTime
	followTo         sql.NullTime
	noFollowUp       int64
	hasPhone         sql.NullString
	hasEmail         sql.NullString
	hasWebsite       sql.NullString
	hasLinkedIn      sql.NullString
	hasCompany       sql.NullString
	hasValue         sql.NullString
	neverContacted   int64
	attemptsMin      sql.NullInt32
	attemptsMax      sql.NullInt32
	valueMin         sql.NullString
	valueMax         sql.NullString
	createdFrom      sql.NullTime
	createdBefore    sql.NullTime
	activeFrom       sql.NullTime
	activeBefore     sql.NullTime
	search           sql.NullString
	searchDigits     sql.NullString
}

// everWorked is before any activity, so last_activity_at >= it means "worked".
var everWorked = time.Unix(1, 0).UTC()

func leadArgs(f domain.LeadFilter, day domain.CallerDay) leadFilterArgs {
	from, to := f.FollowUp.Range(day.Date)
	a := leadFilterArgs{
		filterReps:       flag(len(f.Reps) > 0),
		filterSources:    flag(len(f.Sources) > 0),
		sources:          f.Sources,
		filterIndustries: flag(len(f.Industries) > 0),
		industries:       nullStrings(f.Industries),
		filterSizes:      flag(len(f.Sizes) > 0),
		sizes:            nullStrings(f.Sizes),
		filterCountries:  flag(len(f.Countries) > 0),
		countries:        nullStrings(f.Countries),
		filterCities:     flag(len(f.Cities) > 0),
		cities:           nullStrings(f.Cities),
		filterBilling:    flag(len(f.Billing) > 0),
		batch:            nullInt64(f.ImportBatchID),
		followFrom:       nullTime(from),
		followTo:         nullTime(to),
		noFollowUp:       flag(f.FollowUp == domain.FollowUpNone),
		hasPhone:         nullString(string(f.HasPhone)),
		hasEmail:         nullString(string(f.HasEmail)),
		hasWebsite:       nullString(string(f.HasWebsite)),
		hasLinkedIn:      nullString(string(f.HasLinkedIn)),
		hasCompany:       nullString(string(f.HasCompany)),
		hasValue:         nullString(string(f.HasValue)),
		neverContacted:   flag(f.Contacted == domain.PresenceNo),
		attemptsMin:      nullInt32(f.AttemptsMin),
		attemptsMax:      nullInt32(f.AttemptsMax),
		valueMin:         nullAmount(f.ValueMin),
		valueMax:         nullAmount(f.ValueMax),
		createdFrom:      dayBound(day, f.CreatedFrom, 0),
		createdBefore:    dayBound(day, f.CreatedTo, 1),
		activeFrom:       dayBound(day, f.ActiveFrom, 0),
		activeBefore:     dayBound(day, f.ActiveTo, 1),
		search:           likePattern(f.Search),
		searchDigits:     digitsPattern(f.Search),
	}
	for _, id := range f.Reps {
		if id == 0 {
			a.unassigned = 1
		} else {
			a.reps = append(a.reps, id)
		}
	}
	for _, billing := range f.Billing {
		a.billing = append(a.billing, sqlc.CrmLeadsBilling(billing))
	}
	if f.Contacted == domain.PresenceYes && !a.activeFrom.Valid {
		a.activeFrom = someTime(everWorked)
	}
	return a
}

// dayBound is when the caller's day began, plus days; null without a day.
func dayBound(day domain.CallerDay, date *time.Time, days int) sql.NullTime {
	if date == nil {
		return sql.NullTime{}
	}
	return someTime(day.StartOf(date.AddDate(0, 0, days)))
}

func flag(on bool) int64 {
	if on {
		return 1
	}
	return 0
}

func (a leadFilterArgs) newest(businessID int64, statuses []sqlc.CrmLeadsStatus, cursorID sql.NullInt64, cursorTime sql.NullTime, limit int32) sqlc.ListLeadsNewestParams {
	return sqlc.ListLeadsNewestParams{
		BusinessID: businessID, Statuses: statuses,
		FilterReps: a.filterReps, Reps: nullInt64s(a.reps), Unassigned: a.unassigned,
		FilterSources: a.filterSources, Sources: a.sources,
		FilterIndustries: a.filterIndustries, Industries: a.industries,
		FilterSizes: a.filterSizes, Sizes: a.sizes,
		FilterCountries: a.filterCountries, Countries: a.countries,
		FilterCities: a.filterCities, Cities: a.cities,
		FilterBilling: a.filterBilling, Billing: a.billing,
		ImportBatchID: a.batch, FollowUpFrom: a.followFrom, FollowUpTo: a.followTo, NoFollowUp: a.noFollowUp,
		HasPhone: a.hasPhone, HasEmail: a.hasEmail, HasWebsite: a.hasWebsite, HasLinkedin: a.hasLinkedIn,
		HasCompany: a.hasCompany, HasValue: a.hasValue, NeverContacted: a.neverContacted,
		AttemptsMin: a.attemptsMin, AttemptsMax: a.attemptsMax, ValueMin: a.valueMin, ValueMax: a.valueMax,
		CreatedFrom: a.createdFrom, CreatedBefore: a.createdBefore, ActiveFrom: a.activeFrom, ActiveBefore: a.activeBefore,
		Search: a.search, SearchDigits: a.searchDigits,
		CursorID: cursorID, CursorTime: cursorTime, Limit: limit,
	}
}

func (a leadFilterArgs) byCompany(businessID int64, statuses []sqlc.CrmLeadsStatus, cursorID sql.NullInt64, cursorText sql.NullString, limit int32) sqlc.ListLeadsByCompanyParams {
	return sqlc.ListLeadsByCompanyParams{
		BusinessID: businessID, Statuses: statuses,
		FilterReps: a.filterReps, Reps: nullInt64s(a.reps), Unassigned: a.unassigned,
		FilterSources: a.filterSources, Sources: a.sources,
		FilterIndustries: a.filterIndustries, Industries: a.industries,
		FilterSizes: a.filterSizes, Sizes: a.sizes,
		FilterCountries: a.filterCountries, Countries: a.countries,
		FilterCities: a.filterCities, Cities: a.cities,
		FilterBilling: a.filterBilling, Billing: a.billing,
		ImportBatchID: a.batch, FollowUpFrom: a.followFrom, FollowUpTo: a.followTo, NoFollowUp: a.noFollowUp,
		HasPhone: a.hasPhone, HasEmail: a.hasEmail, HasWebsite: a.hasWebsite, HasLinkedin: a.hasLinkedIn,
		HasCompany: a.hasCompany, HasValue: a.hasValue, NeverContacted: a.neverContacted,
		AttemptsMin: a.attemptsMin, AttemptsMax: a.attemptsMax, ValueMin: a.valueMin, ValueMax: a.valueMax,
		CreatedFrom: a.createdFrom, CreatedBefore: a.createdBefore, ActiveFrom: a.activeFrom, ActiveBefore: a.activeBefore,
		Search: a.search, SearchDigits: a.searchDigits,
		CursorID: cursorID, CursorText: cursorText, Limit: limit,
	}
}

func (a leadFilterArgs) counts(businessID int64, day domain.CallerDay) sqlc.CountLeadsByStatusParams {
	return sqlc.CountLeadsByStatusParams{
		Today: someTime(dayStart(day.Date)), MonthStart: someTime(day.MonthStartUTC()), BusinessID: businessID,
		FilterReps: a.filterReps, Reps: nullInt64s(a.reps), Unassigned: a.unassigned,
		FilterSources: a.filterSources, Sources: a.sources,
		FilterIndustries: a.filterIndustries, Industries: a.industries,
		FilterSizes: a.filterSizes, Sizes: a.sizes,
		FilterCountries: a.filterCountries, Countries: a.countries,
		FilterCities: a.filterCities, Cities: a.cities,
		FilterBilling: a.filterBilling, Billing: a.billing,
		ImportBatchID: a.batch, FollowUpFrom: a.followFrom, FollowUpTo: a.followTo, NoFollowUp: a.noFollowUp,
		HasPhone: a.hasPhone, HasEmail: a.hasEmail, HasWebsite: a.hasWebsite, HasLinkedin: a.hasLinkedIn,
		HasCompany: a.hasCompany, HasValue: a.hasValue, NeverContacted: a.neverContacted,
		AttemptsMin: a.attemptsMin, AttemptsMax: a.attemptsMax, ValueMin: a.valueMin, ValueMax: a.valueMax,
		CreatedFrom: a.createdFrom, CreatedBefore: a.createdBefore, ActiveFrom: a.activeFrom, ActiveBefore: a.activeBefore,
		Search: a.search, SearchDigits: a.searchDigits,
	}
}

// ListLeads returns one page of leads in the filter's sort. day is the
// caller's calendar day, which follow-up windows and date bounds use.
func (r *Repository) ListLeads(ctx context.Context, businessID int64, f domain.LeadFilter, day domain.CallerDay) (domain.LeadPage, error) {
	cursor, err := decodeCursor(f.Cursor, string(f.Sort))
	if err != nil {
		return domain.LeadPage{}, err
	}
	a := leadArgs(f, day)
	statuses := leadStatuses(f.Statuses)
	limit := int32(f.Limit + 1) // #nosec G115 -- the service caps Limit at 200.

	var rows []sqlc.CrmLead
	if f.Sort == domain.LeadSortCompany {
		rows, err = r.queries.ListLeadsByCompany(ctx, a.byCompany(businessID, statuses, cursor.id(), cursor.text(), limit))
	} else {
		var at sql.NullTime
		if f.Sort == domain.LeadSortFollowUp {
			at, err = cursor.date()
		} else {
			at, err = cursor.time()
		}
		if err != nil {
			return domain.LeadPage{}, err
		}
		// The keyset sorts on a time share one parameter shape.
		params := a.newest(businessID, statuses, cursor.id(), at, limit)
		switch f.Sort {
		case domain.LeadSortFollowUp:
			rows, err = r.queries.ListLeadsByFollowUp(ctx, sqlc.ListLeadsByFollowUpParams(params))
		case domain.LeadSortActivity:
			rows, err = r.queries.ListLeadsByActivity(ctx, sqlc.ListLeadsByActivityParams(params))
		default:
			rows, err = r.queries.ListLeadsNewest(ctx, params)
		}
	}
	if err != nil {
		return domain.LeadPage{}, err
	}

	counts, err := r.queries.CountLeadsByStatus(ctx, a.counts(businessID, day))
	if err != nil {
		return domain.LeadPage{}, err
	}
	page := domain.LeadPage{}
	for _, row := range counts {
		if slices.Contains(statuses, row.Status) {
			page.Total += row.Total
		}
	}
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

// LeadStats counts the filter's leads by status, the follow-ups due among the
// ones being worked, and today's calls by the filter's reps.
func (r *Repository) LeadStats(ctx context.Context, businessID int64, f domain.LeadFilter, day domain.CallerDay) (domain.LeadStats, error) {
	a := leadArgs(f, day)
	stats := domain.LeadStats{ByStatus: map[domain.LeadStatus]int64{}}
	rows, err := r.queries.CountLeadsByStatus(ctx, a.counts(businessID, day))
	if err != nil {
		return stats, err
	}
	for _, row := range rows {
		status := domain.LeadStatus(row.Status)
		stats.ByStatus[status] = row.Total
		stats.ConvertedThisMonth += row.Converted
		if status.IsWorking() {
			stats.FollowUpsOverdue += row.Overdue
			stats.FollowUpsToday += row.DueToday
		}
	}
	if stats.CallsToday, err = r.queries.CountCallsSince(ctx, sqlc.CountCallsSinceParams{
		BusinessID: businessID, Since: day.StartUTC, FilterReps: a.filterReps, Reps: a.reps,
	}); err != nil {
		return stats, err
	}
	stats.ConnectedToday, err = r.queries.CountConnectedCallsSince(ctx, sqlc.CountConnectedCallsSinceParams{
		BusinessID: businessID, Since: day.StartUTC, FilterReps: a.filterReps, Reps: a.reps,
	})
	return stats, err
}

// LeadFacets lists the distinct values leads have, for the filter pickers.
func (r *Repository) LeadFacets(ctx context.Context, businessID int64) (domain.LeadFacets, error) {
	var facets domain.LeadFacets
	var err error
	if facets.Sources, err = r.queries.ListLeadSources(ctx, businessID); err != nil {
		return facets, err
	}
	for _, facet := range []struct {
		list func(context.Context, int64) ([]sql.NullString, error)
		into *[]string
	}{
		{r.queries.ListLeadIndustries, &facets.Industries},
		{r.queries.ListLeadSizes, &facets.Sizes},
		{r.queries.ListLeadCountries, &facets.Countries},
		{r.queries.ListLeadCities, &facets.Cities},
	} {
		values, err := facet.list(ctx, businessID)
		if err != nil {
			return facets, err
		}
		for _, value := range values {
			*facet.into = append(*facet.into, value.String)
		}
	}
	return facets, nil
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
