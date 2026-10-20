package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/infrastructure/persistence/sqlc"
)

// ListDeals returns one page of deals in the filter's sort.
func (r *Repository) ListDeals(ctx context.Context, businessID int64, f domain.DealFilter) (domain.DealPage, error) {
	cursor, err := decodeCursor(f.Cursor, string(f.Sort))
	if err != nil {
		return domain.DealPage{}, err
	}
	pipeline, stage := nullInt64(f.PipelineID), nullInt64(f.StageID)
	statuses, assigned, search := dealStatuses(f.Statuses), nullInt64(f.AssignedTo), likePattern(f.Search)
	limit := int32(f.Limit + 1) // #nosec G115 -- the service caps Limit at 200.

	var rows []sqlc.CrmDeal
	switch f.Sort {
	case domain.DealSortCloseDate:
		var at sql.NullTime
		if at, err = cursor.date(); err != nil {
			return domain.DealPage{}, err
		}
		rows, err = r.queries.ListDealsByCloseDate(ctx, sqlc.ListDealsByCloseDateParams{
			BusinessID: businessID, PipelineID: pipeline, StageID: stage, Statuses: statuses,
			AssignedTo: assigned, Search: search, CursorID: cursor.id(), CursorDate: at, Limit: limit,
		})
	case domain.DealSortValue:
		rows, err = r.queries.ListDealsByValue(ctx, sqlc.ListDealsByValueParams{
			BusinessID: businessID, PipelineID: pipeline, StageID: stage, Statuses: statuses,
			AssignedTo: assigned, Search: search, CursorID: cursor.id(), CursorAmount: cursor.text(), Limit: limit,
		})
	default:
		var at sql.NullTime
		if at, err = cursor.time(); err != nil {
			return domain.DealPage{}, err
		}
		rows, err = r.queries.ListDealsNewest(ctx, sqlc.ListDealsNewestParams{
			BusinessID: businessID, PipelineID: pipeline, StageID: stage, Statuses: statuses,
			AssignedTo: assigned, Search: search, CursorID: cursor.id(), CursorTime: at, Limit: limit,
		})
	}
	if err != nil {
		return domain.DealPage{}, err
	}
	total, err := r.queries.CountDeals(ctx, sqlc.CountDealsParams{
		BusinessID: businessID, PipelineID: pipeline, StageID: stage, Statuses: statuses,
		AssignedTo: assigned, Search: search,
	})
	if err != nil {
		return domain.DealPage{}, err
	}
	page := domain.DealPage{Total: total}
	if len(rows) > f.Limit {
		rows = rows[:f.Limit]
		last := dealFromRow(rows[len(rows)-1])
		page.NextCursor = encodeCursor(string(f.Sort), dealSortValue(f.Sort, last), last.ID)
	}
	page.Deals = dealsFromRows(rows)
	return page, nil
}

func dealSortValue(sort domain.DealSort, deal domain.Deal) string {
	switch sort {
	case domain.DealSortCloseDate:
		if deal.ExpectedCloseDate != "" {
			return deal.ExpectedCloseDate
		}
		return noCloseDate
	case domain.DealSortValue:
		return deal.Amount().String()
	}
	return timeValue(deal.CreatedAt)
}

// ListStageDeals returns one page of a board column, newest first.
func (r *Repository) ListStageDeals(ctx context.Context, businessID, pipelineID, stageID int64, assignedTo *int64, search, token string, limit int) ([]domain.Deal, string, error) {
	const sort = "stage"
	cursor, err := decodeCursor(token, sort)
	if err != nil {
		return nil, "", err
	}
	at, err := cursor.time()
	if err != nil {
		return nil, "", err
	}
	rows, err := r.queries.ListStageDeals(ctx, sqlc.ListStageDealsParams{
		BusinessID: businessID, PipelineID: pipelineID, StageID: stageID,
		AssignedTo: nullInt64(assignedTo), Search: likePattern(search),
		CursorID: cursor.id(), CursorTime: at, Limit: int32(limit + 1), // #nosec G115 -- capped by the service.
	})
	if err != nil {
		return nil, "", err
	}
	next := ""
	if len(rows) > limit {
		rows = rows[:limit]
		last := rows[len(rows)-1]
		next = encodeCursor(sort, timeValue(last.CreatedAt), last.ID)
	}
	return dealsFromRows(rows), next, nil
}

// StageTotals counts each stage's deals and first-year amount.
func (r *Repository) StageTotals(ctx context.Context, businessID, pipelineID int64, assignedTo *int64, search string) (map[int64]domain.StageTotal, error) {
	rows, err := r.queries.DealStageTotals(ctx, sqlc.DealStageTotalsParams{
		BusinessID: businessID, PipelineID: pipelineID, AssignedTo: nullInt64(assignedTo), Search: likePattern(search),
	})
	if err != nil {
		return nil, err
	}
	totals := make(map[int64]domain.StageTotal, len(rows))
	for _, row := range rows {
		totals[row.StageID] = domain.StageTotal{Deals: row.Deals, Amount: amount(row.Amount)}
	}
	return totals, nil
}

func (r *Repository) DealOutcomesSince(ctx context.Context, businessID int64, pipelineID, assignedTo *int64, since time.Time) (domain.DealOutcome, error) {
	row, err := r.queries.DealOutcomesSince(ctx, sqlc.DealOutcomesSinceParams{
		BusinessID: businessID, PipelineID: nullInt64(pipelineID), Since: someTime(since), AssignedTo: nullInt64(assignedTo),
	})
	if err != nil {
		return domain.DealOutcome{}, err
	}
	days, err := strconv.ParseFloat(row.AverageDaysToWin, 64)
	if err != nil {
		days = 0
	}
	return domain.DealOutcome{Won: row.Won, Lost: row.Lost, WonAmount: amount(row.WonAmount), AverageDaysToWin: days}, nil
}

func (r *Repository) GetDeal(ctx context.Context, businessID, id int64) (domain.Deal, error) {
	row, err := r.queries.GetDeal(ctx, sqlc.GetDealParams{ID: id, BusinessID: businessID})
	if err != nil {
		return domain.Deal{}, notFound(err, domain.ErrDealNotFound)
	}
	return dealFromRow(row), nil
}

// CreateDeal opens a deal in stage, closed at once when stage closes deals.
func (r *Repository) CreateDeal(ctx context.Context, businessID, actorID int64, in domain.DealInput, stage domain.Stage, now time.Time) (domain.Deal, error) {
	status := domain.StatusForStage(stage.Kind)
	var closedAt sql.NullTime
	if status != domain.DealOpen {
		closedAt = someTime(now)
	}
	result, err := r.queries.CreateDeal(ctx, sqlc.CreateDealParams{
		BusinessID:        businessID,
		OwnerID:           actorID,
		PipelineID:        stage.PipelineID,
		StageID:           stage.ID,
		Title:             in.Title,
		ContactName:       in.ContactName,
		CompanyName:       in.CompanyName,
		Value:             in.Value.String(),
		Billing:           sqlc.CrmDealsBilling(in.Billing),
		ExpectedCloseDate: nullDate(in.ExpectedCloseDate),
		AssignedTo:        nullInt64(in.AssignedTo),
		Source:            in.Source,
		Status:            sqlc.CrmDealsStatus(status),
		ClosedAt:          closedAt,
		Notes:             nullString(in.Notes),
	})
	if err != nil {
		return domain.Deal{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Deal{}, err
	}
	return r.GetDeal(ctx, businessID, id)
}

func (r *Repository) UpdateDeal(ctx context.Context, businessID, id int64, in domain.DealInput) (domain.Deal, error) {
	if _, err := r.GetDeal(ctx, businessID, id); err != nil {
		return domain.Deal{}, err
	}
	if err := r.queries.UpdateDeal(ctx, sqlc.UpdateDealParams{
		Title:             in.Title,
		ContactName:       in.ContactName,
		CompanyName:       in.CompanyName,
		Value:             in.Value.String(),
		Billing:           sqlc.CrmDealsBilling(in.Billing),
		ExpectedCloseDate: nullDate(in.ExpectedCloseDate),
		AssignedTo:        nullInt64(in.AssignedTo),
		Source:            in.Source,
		Notes:             nullString(in.Notes),
		ID:                id,
		BusinessID:        businessID,
	}); err != nil {
		return domain.Deal{}, err
	}
	return r.GetDeal(ctx, businessID, id)
}

// MoveDeal puts a deal in stage, which may close or reopen it. With contract,
// it also drafts the won deal's contract in the same transaction.
func (r *Repository) MoveDeal(ctx context.Context, businessID, actorID, id int64, stage domain.Stage, lostReason string, contract *domain.ContractInput, now time.Time) (domain.Deal, *domain.Contract, error) {
	var contractID int64
	err := r.inTx(ctx, func(q *sqlc.Queries, _ *sql.Tx) error {
		row, err := q.GetDealForUpdate(ctx, sqlc.GetDealForUpdateParams{ID: id, BusinessID: businessID})
		if err != nil {
			return notFound(err, domain.ErrDealNotFound)
		}
		deal := dealFromRow(row)
		status := domain.StatusForStage(stage.Kind)
		closedAt := nullTime(deal.ClosedAt)
		switch {
		case status == domain.DealOpen:
			closedAt = sql.NullTime{}
		case deal.Status != status:
			closedAt = someTime(now)
		}
		var reason sql.NullString
		if status == domain.DealLost {
			reason = nullString(lostReason)
		}
		changedAt := deal.StageChangedAt
		if deal.StageID != stage.ID {
			changedAt = now
		}
		if err := q.MoveDeal(ctx, sqlc.MoveDealParams{
			PipelineID: stage.PipelineID, StageID: stage.ID, Status: sqlc.CrmDealsStatus(status),
			ClosedAt: closedAt, LostReason: reason, StageChangedAt: changedAt, ID: id, BusinessID: businessID,
		}); err != nil {
			return err
		}
		if contract == nil {
			return nil
		}
		contract.DealID = &deal.ID
		if contract.CustomerID == nil {
			contract.CustomerID = deal.CustomerID
		}
		contractID, err = insertContract(ctx, q, businessID, actorID, deal.LeadID, *contract)
		return err
	})
	if err != nil {
		return domain.Deal{}, nil, err
	}
	deal, err := r.GetDeal(ctx, businessID, id)
	if err != nil || contractID == 0 {
		return deal, nil, err
	}
	created, err := r.GetContract(ctx, businessID, contractID)
	if err != nil {
		return deal, nil, err
	}
	return deal, &created, nil
}

// DeleteDeal removes a deal without a contract, with its own tasks and
// activities. A lead it came from goes back to being worked.
func (r *Repository) DeleteDeal(ctx context.Context, businessID, id int64) error {
	return r.inTx(ctx, func(q *sqlc.Queries, _ *sql.Tx) error {
		if _, err := q.GetDealForUpdate(ctx, sqlc.GetDealForUpdateParams{ID: id, BusinessID: businessID}); err != nil {
			return notFound(err, domain.ErrDealNotFound)
		}
		contracts, err := q.CountDealContracts(ctx, sqlc.CountDealContractsParams{BusinessID: businessID, DealID: someInt64(id)})
		if err != nil {
			return err
		}
		if contracts > 0 {
			return domain.ErrDealHasContract
		}
		if err := q.DeleteDealTasks(ctx, sqlc.DeleteDealTasksParams{BusinessID: businessID, DealID: someInt64(id)}); err != nil {
			return err
		}
		if err := q.DeleteDealActivities(ctx, sqlc.DeleteDealActivitiesParams{BusinessID: businessID, DealID: someInt64(id)}); err != nil {
			return err
		}
		if err := q.DetachLeadFromDeal(ctx, sqlc.DetachLeadFromDealParams{BusinessID: businessID, DealID: someInt64(id)}); err != nil {
			return err
		}
		_, err = q.DeleteDeal(ctx, sqlc.DeleteDealParams{ID: id, BusinessID: businessID})
		return err
	})
}

// DealContract returns the latest contract drafted for a deal, if any.
func (r *Repository) DealContract(ctx context.Context, businessID, dealID int64) (*domain.Contract, error) {
	row, err := r.queries.GetDealContract(ctx, sqlc.GetDealContractParams{BusinessID: businessID, DealID: someInt64(dealID)})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	contract := contractFromRow(row)
	return &contract, nil
}
