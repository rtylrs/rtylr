package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/platform/money"
)

// openStage resolves where a new deal starts: the given stage, or the first
// open stage of the given pipeline (the default pipeline when 0).
func (s *Service) openStage(ctx context.Context, businessID, pipelineID, stageID int64) (domain.Stage, error) {
	if stageID > 0 {
		stage, err := s.repo.GetStage(ctx, businessID, stageID)
		if err != nil {
			return domain.Stage{}, err
		}
		if pipelineID > 0 && stage.PipelineID != pipelineID {
			return domain.Stage{}, domain.ErrStageNotFound
		}
		return stage, nil
	}
	pipeline, err := s.pipeline(ctx, businessID, pipelineID)
	if err != nil {
		return domain.Stage{}, err
	}
	stage, ok := pipeline.StageOfKind(domain.StageOpen)
	if !ok {
		return domain.Stage{}, domain.ErrStageNotFound
	}
	return stage, nil
}

// pipeline returns a pipeline by id, or the default one when id is 0.
func (s *Service) pipeline(ctx context.Context, businessID, id int64) (domain.Pipeline, error) {
	if id > 0 {
		return s.repo.GetPipeline(ctx, businessID, id)
	}
	pipelines, err := s.ListPipelines(ctx, businessID)
	if err != nil {
		return domain.Pipeline{}, err
	}
	for _, pipeline := range pipelines {
		if pipeline.IsDefault {
			return pipeline, nil
		}
	}
	if len(pipelines) > 0 {
		return pipelines[0], nil
	}
	return domain.Pipeline{}, domain.ErrPipelineNotFound
}

func (s *Service) ListDeals(ctx context.Context, businessID int64, f domain.DealFilter) (domain.DealPage, error) {
	if f.Sort == "" {
		f.Sort = domain.DealSortNewest
	}
	if !f.Sort.IsValid() {
		return domain.DealPage{}, &domain.ValidationError{Message: "unknown sort"}
	}
	for _, status := range f.Statuses {
		if !status.IsValid() {
			return domain.DealPage{}, &domain.ValidationError{Message: "unknown deal status"}
		}
	}
	f.Limit = pageSize(f.Limit)
	return s.repo.ListDeals(ctx, businessID, f)
}

// DealBoard returns a pipeline (the default when id is 0) with each stage's
// deal count, amount and weighted amount.
func (s *Service) DealBoard(ctx context.Context, businessID, pipelineID int64, assignedTo *int64, search string) (domain.Pipeline, []domain.StageTotal, error) {
	pipeline, err := s.pipeline(ctx, businessID, pipelineID)
	if err != nil {
		return domain.Pipeline{}, nil, err
	}
	totals, err := s.repo.StageTotals(ctx, businessID, pipeline.ID, assignedTo, search)
	if err != nil {
		return domain.Pipeline{}, nil, err
	}
	columns := make([]domain.StageTotal, 0, len(pipeline.Stages))
	for _, stage := range pipeline.Stages {
		total := totals[stage.ID]
		total.Stage = stage
		total.WeightedAmount = weighted(total.Amount, stage.Probability)
		columns = append(columns, total)
	}
	return pipeline, columns, nil
}

func weighted(amount money.Amount, probability int) money.Amount {
	return amount * money.Amount(probability) / 100
}

// ListStageDeals returns one page of a board column.
func (s *Service) ListStageDeals(ctx context.Context, businessID, stageID int64, assignedTo *int64, search, cursor string, limit int) ([]domain.Deal, string, error) {
	stage, err := s.repo.GetStage(ctx, businessID, stageID)
	if err != nil {
		return nil, "", err
	}
	return s.repo.ListStageDeals(ctx, businessID, stage.PipelineID, stage.ID, assignedTo, search, cursor, pageSize(limit))
}

// DealStats summarise a pipeline (the default when id is 0): open value now,
// wins this month in the caller's calendar, and the last 90 days' closes.
func (s *Service) DealStats(ctx context.Context, businessID, pipelineID int64, assignedTo *int64, day domain.CallerDay) (domain.DealStats, error) {
	pipeline, columns, err := s.DealBoard(ctx, businessID, pipelineID, assignedTo, "")
	if err != nil {
		return domain.DealStats{}, err
	}
	var stats domain.DealStats
	for _, column := range columns {
		if column.Stage.Kind != domain.StageOpen {
			continue
		}
		stats.OpenDeals += column.Deals
		stats.OpenAmount += column.Amount
		stats.WeightedAmount += column.WeightedAmount
	}
	month, err := s.repo.DealOutcomesSince(ctx, businessID, &pipeline.ID, assignedTo, day.MonthStartUTC())
	if err != nil {
		return domain.DealStats{}, err
	}
	stats.WonThisMonth, stats.WonAmountThisMonth = month.Won, month.WonAmount
	quarter, err := s.repo.DealOutcomesSince(ctx, businessID, &pipeline.ID, assignedTo, day.StartUTC.AddDate(0, 0, -90))
	if err != nil {
		return domain.DealStats{}, err
	}
	if closed := quarter.Won + quarter.Lost; closed > 0 {
		stats.WinRate = float64(quarter.Won) * 100 / float64(closed)
	}
	if quarter.Won > 0 {
		stats.AverageWonAmount = quarter.WonAmount / money.Amount(quarter.Won)
	}
	stats.AverageDaysToWin = quarter.AverageDaysToWin
	return stats, nil
}

// GetDeal returns a deal with the contract drafted for it, if any.
func (s *Service) GetDeal(ctx context.Context, businessID, id int64) (domain.Deal, *domain.Contract, error) {
	deal, err := s.repo.GetDeal(ctx, businessID, id)
	if err != nil {
		return domain.Deal{}, nil, err
	}
	contract, err := s.repo.DealContract(ctx, businessID, id)
	return deal, contract, err
}

// CreateDeal opens a deal directly, without a lead.
func (s *Service) CreateDeal(ctx context.Context, businessID, actorID int64, in domain.DealInput) (domain.Deal, error) {
	if err := in.Normalize(); err != nil {
		return domain.Deal{}, err
	}
	if err := s.validateAssignee(ctx, businessID, in.AssignedTo); err != nil {
		return domain.Deal{}, err
	}
	stage, err := s.openStage(ctx, businessID, in.PipelineID, in.StageID)
	if err != nil {
		return domain.Deal{}, err
	}
	return s.repo.CreateDeal(ctx, businessID, actorID, in, stage, s.now())
}

func (s *Service) UpdateDeal(ctx context.Context, businessID, id int64, in domain.DealInput) (domain.Deal, error) {
	if err := in.Normalize(); err != nil {
		return domain.Deal{}, err
	}
	if err := s.validateAssignee(ctx, businessID, in.AssignedTo); err != nil {
		return domain.Deal{}, err
	}
	return s.repo.UpdateDeal(ctx, businessID, id, in)
}

// MoveDeal puts a deal in another open stage, in its pipeline or another.
// Closing goes through WinDeal and LoseDeal.
func (s *Service) MoveDeal(ctx context.Context, businessID, actorID, id, stageID int64) (domain.Deal, error) {
	stage, err := s.repo.GetStage(ctx, businessID, stageID)
	if err != nil {
		return domain.Deal{}, err
	}
	if stage.Kind != domain.StageOpen {
		return domain.Deal{}, &domain.ValidationError{Message: "use win or lose to close a deal"}
	}
	deal, _, err := s.repo.MoveDeal(ctx, businessID, actorID, id, stage, "", nil, s.now())
	return deal, err
}

// WinDeal closes a deal as won in its pipeline's won stage and, when asked,
// drafts its contract from the deal's terms.
func (s *Service) WinDeal(ctx context.Context, businessID, actorID, id int64, in domain.WinDealInput) (domain.Deal, *domain.Contract, error) {
	deal, err := s.repo.GetDeal(ctx, businessID, id)
	if err != nil {
		return domain.Deal{}, nil, err
	}
	stage, err := s.closingStage(ctx, businessID, deal.PipelineID, domain.StageWon)
	if err != nil {
		return domain.Deal{}, nil, err
	}
	if in.Contract != nil {
		contract := in.Contract
		if contract.Plan == "" {
			contract.Plan = deal.Title
		}
		if contract.Value == 0 {
			contract.Value = deal.Value
			if contract.Billing == "" {
				contract.Billing = deal.Billing
			}
		}
		if contract.Billing == "" {
			contract.Billing = deal.Billing
		}
		if contract.StartDate == nil {
			today := s.now().Format(time.DateOnly)
			contract.StartDate = &today
		}
		if err := contract.Normalize(); err != nil {
			return domain.Deal{}, nil, err
		}
	}
	won, contract, err := s.repo.MoveDeal(ctx, businessID, actorID, id, stage, "", in.Contract, s.now())
	if err != nil {
		return domain.Deal{}, nil, err
	}
	s.invalidate(ctx, businessID)
	return won, contract, nil
}

// LoseDeal closes a deal as lost; a reason is required.
func (s *Service) LoseDeal(ctx context.Context, businessID, actorID, id int64, reason string) (domain.Deal, error) {
	if len([]rune(reason)) > 255 {
		reason = string([]rune(reason)[:255])
	}
	if reason == "" {
		return domain.Deal{}, &domain.ValidationError{Message: "a lost deal needs a reason"}
	}
	deal, err := s.repo.GetDeal(ctx, businessID, id)
	if err != nil {
		return domain.Deal{}, err
	}
	stage, err := s.closingStage(ctx, businessID, deal.PipelineID, domain.StageLost)
	if err != nil {
		return domain.Deal{}, err
	}
	lost, _, err := s.repo.MoveDeal(ctx, businessID, actorID, id, stage, reason, nil, s.now())
	return lost, err
}

// ReopenDeal returns a closed deal to an open stage (by default its
// pipeline's first).
func (s *Service) ReopenDeal(ctx context.Context, businessID, actorID, id, stageID int64) (domain.Deal, error) {
	deal, err := s.repo.GetDeal(ctx, businessID, id)
	if err != nil {
		return domain.Deal{}, err
	}
	stage, err := s.openStage(ctx, businessID, deal.PipelineID, stageID)
	if err != nil {
		return domain.Deal{}, err
	}
	if stage.Kind != domain.StageOpen {
		return domain.Deal{}, &domain.ValidationError{Message: "reopen a deal into an open stage"}
	}
	reopened, _, err := s.repo.MoveDeal(ctx, businessID, actorID, id, stage, "", nil, s.now())
	return reopened, err
}

func (s *Service) closingStage(ctx context.Context, businessID, pipelineID int64, kind domain.StageKind) (domain.Stage, error) {
	pipeline, err := s.repo.GetPipeline(ctx, businessID, pipelineID)
	if err != nil {
		return domain.Stage{}, err
	}
	stage, ok := pipeline.StageOfKind(kind)
	if !ok {
		return domain.Stage{}, domain.ErrStageNotFound
	}
	return stage, nil
}

func (s *Service) DeleteDeal(ctx context.Context, businessID, id int64) error {
	return s.repo.DeleteDeal(ctx, businessID, id)
}
