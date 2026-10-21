package api

import (
	"context"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	crmsalesv1 "github.com/voxire/voxie/gen/go/crm/sales/v1"
)

func (a *API) ListPipelines(ctx context.Context, _ *crmsalesv1.ListPipelinesRequest) (*crmsalesv1.ListPipelinesResponse, error) {
	c, err := a.caller(ctx, permView)
	if err != nil {
		return nil, err
	}
	pipelines, err := a.Service.ListPipelines(ctx, c.businessID)
	if err != nil {
		return nil, fail(ctx, "list pipelines", err)
	}
	out := make([]*crmsalesv1.Pipeline, 0, len(pipelines))
	for _, pipeline := range pipelines {
		out = append(out, pipelineToGen(pipeline))
	}
	return &crmsalesv1.ListPipelinesResponse{Pipelines: out}, nil
}

func (a *API) CreatePipeline(ctx context.Context, req *crmsalesv1.CreatePipelineRequest) (*crmsalesv1.PipelineResponse, error) {
	c, err := a.caller(ctx, permEdit)
	if err != nil {
		return nil, err
	}
	in, err := pipelineFieldsFromGen(req.Pipeline)
	if err != nil {
		return nil, err
	}
	pipeline, err := a.Service.CreatePipeline(ctx, c.businessID, in)
	if err != nil {
		return nil, fail(ctx, "create pipeline", err)
	}
	return &crmsalesv1.PipelineResponse{Pipeline: pipelineToGen(pipeline)}, nil
}

func (a *API) UpdatePipeline(ctx context.Context, req *crmsalesv1.UpdatePipelineRequest) (*crmsalesv1.PipelineResponse, error) {
	c, err := a.caller(ctx, permEdit)
	if err != nil {
		return nil, err
	}
	in, err := pipelineFieldsFromGen(req.Body)
	if err != nil {
		return nil, err
	}
	pipeline, err := a.Service.UpdatePipeline(ctx, c.businessID, req.Id, in)
	if err != nil {
		return nil, fail(ctx, "update pipeline", err)
	}
	return &crmsalesv1.PipelineResponse{Pipeline: pipelineToGen(pipeline)}, nil
}

func (a *API) DeletePipeline(ctx context.Context, req *crmsalesv1.DeletePipelineRequest) (*crmsalesv1.DeletePipelineResponse, error) {
	c, err := a.caller(ctx, permDelete)
	if err != nil {
		return nil, err
	}
	if err := a.Service.DeletePipeline(ctx, c.businessID, req.Id); err != nil {
		return nil, fail(ctx, "delete pipeline", err)
	}
	return &crmsalesv1.DeletePipelineResponse{Deleted: true}, nil
}

func (a *API) GetDealBoard(ctx context.Context, req *crmsalesv1.GetDealBoardRequest) (*crmsalesv1.GetDealBoardResponse, error) {
	c, err := a.caller(ctx, permView)
	if err != nil {
		return nil, err
	}
	pipeline, columns, err := a.Service.DealBoard(ctx, c.businessID, int64Value(req.PipelineId), req.AssignedTo, req.Search)
	if err != nil {
		return nil, fail(ctx, "load deal board", err)
	}
	out := make([]*crmsalesv1.BoardColumn, 0, len(columns))
	for _, column := range columns {
		out = append(out, &crmsalesv1.BoardColumn{
			Stage:          stageToGen(column.Stage),
			Deals:          column.Deals,
			Amount:         column.Amount.String(),
			WeightedAmount: column.WeightedAmount.String(),
		})
	}
	return &crmsalesv1.GetDealBoardResponse{Pipeline: pipelineToGen(pipeline), Columns: out}, nil
}

func (a *API) ListStageDeals(ctx context.Context, req *crmsalesv1.ListStageDealsRequest) (*crmsalesv1.ListStageDealsResponse, error) {
	c, err := a.caller(ctx, permView)
	if err != nil {
		return nil, err
	}
	deals, next, err := a.Service.ListStageDeals(ctx, c.businessID, req.StageId, req.AssignedTo, req.Search, req.Cursor, int(req.Limit))
	if err != nil {
		return nil, fail(ctx, "list stage deals", err)
	}
	return &crmsalesv1.ListStageDealsResponse{Deals: dealsToGen(deals), NextCursor: next}, nil
}

func (a *API) GetDealStats(ctx context.Context, req *crmsalesv1.GetDealStatsRequest) (*crmsalesv1.GetDealStatsResponse, error) {
	c, err := a.caller(ctx, permView)
	if err != nil {
		return nil, err
	}
	stats, err := a.Service.DealStats(ctx, c.businessID, int64Value(req.PipelineId), req.AssignedTo, a.callerDay(req.Today, req.UtcOffsetMinutes))
	if err != nil {
		return nil, fail(ctx, "load deal stats", err)
	}
	return &crmsalesv1.GetDealStatsResponse{
		OpenDeals:          stats.OpenDeals,
		OpenAmount:         stats.OpenAmount.String(),
		WeightedAmount:     stats.WeightedAmount.String(),
		WonThisMonth:       stats.WonThisMonth,
		WonAmountThisMonth: stats.WonAmountThisMonth.String(),
		WinRate:            stats.WinRate,
		AverageDaysToWin:   stats.AverageDaysToWin,
		AverageWonAmount:   stats.AverageWonAmount.String(),
	}, nil
}

func (a *API) ListDeals(ctx context.Context, req *crmsalesv1.ListDealsRequest) (*crmsalesv1.ListDealsResponse, error) {
	c, err := a.caller(ctx, permView)
	if err != nil {
		return nil, err
	}
	filter := domain.DealFilter{
		PipelineID: positivePtr(req.PipelineId),
		StageID:    positivePtr(req.StageId),
		AssignedTo: req.AssignedTo,
		Search:     req.Search,
		Sort:       domain.DealSort(req.Sort),
		Cursor:     req.Cursor,
		Limit:      int(req.Limit),
	}
	for _, status := range splitList(req.Statuses) {
		filter.Statuses = append(filter.Statuses, domain.DealStatus(status))
	}
	page, err := a.Service.ListDeals(ctx, c.businessID, filter)
	if err != nil {
		return nil, fail(ctx, "list deals", err)
	}
	return &crmsalesv1.ListDealsResponse{Deals: dealsToGen(page.Deals), NextCursor: page.NextCursor, Total: page.Total}, nil
}

func (a *API) GetDeal(ctx context.Context, req *crmsalesv1.GetDealRequest) (*crmsalesv1.GetDealResponse, error) {
	c, err := a.caller(ctx, permView)
	if err != nil {
		return nil, err
	}
	deal, contract, err := a.Service.GetDeal(ctx, c.businessID, req.Id)
	if err != nil {
		return nil, fail(ctx, "get deal", err)
	}
	return &crmsalesv1.GetDealResponse{Deal: dealToGen(deal), Contract: contractPtrToGen(contract)}, nil
}

func (a *API) CreateDeal(ctx context.Context, req *crmsalesv1.CreateDealRequest) (*crmsalesv1.DealResponse, error) {
	c, err := a.caller(ctx, permCreate)
	if err != nil {
		return nil, err
	}
	in, err := dealFieldsFromGen(req.Deal)
	if err != nil {
		return nil, err
	}
	deal, err := a.Service.CreateDeal(ctx, c.businessID, c.userID, in)
	if err != nil {
		return nil, fail(ctx, "create deal", err)
	}
	return &crmsalesv1.DealResponse{Deal: dealToGen(deal)}, nil
}

func (a *API) UpdateDeal(ctx context.Context, req *crmsalesv1.UpdateDealRequest) (*crmsalesv1.DealResponse, error) {
	c, err := a.caller(ctx, permEdit)
	if err != nil {
		return nil, err
	}
	in, err := dealFieldsFromGen(req.Body)
	if err != nil {
		return nil, err
	}
	deal, err := a.Service.UpdateDeal(ctx, c.businessID, req.Id, in)
	if err != nil {
		return nil, fail(ctx, "update deal", err)
	}
	return &crmsalesv1.DealResponse{Deal: dealToGen(deal)}, nil
}

func (a *API) MoveDeal(ctx context.Context, req *crmsalesv1.MoveDealRequest) (*crmsalesv1.DealResponse, error) {
	c, err := a.caller(ctx, permEdit)
	if err != nil {
		return nil, err
	}
	if req.Body == nil {
		return nil, invalid("stage_id is required")
	}
	deal, err := a.Service.MoveDeal(ctx, c.businessID, c.userID, req.Id, req.Body.StageId)
	if err != nil {
		return nil, fail(ctx, "move deal", err)
	}
	return &crmsalesv1.DealResponse{Deal: dealToGen(deal)}, nil
}

func (a *API) WinDeal(ctx context.Context, req *crmsalesv1.WinDealRequest) (*crmsalesv1.WinDealResponse, error) {
	c, err := a.caller(ctx, permEdit)
	if err != nil {
		return nil, err
	}
	in := domain.WinDealInput{}
	if req.Body != nil {
		if in.Contract, err = contractFieldsFromGen(req.Body.Contract); err != nil {
			return nil, err
		}
	}
	deal, contract, err := a.Service.WinDeal(ctx, c.businessID, c.userID, req.Id, in)
	if err != nil {
		return nil, fail(ctx, "win deal", err)
	}
	return &crmsalesv1.WinDealResponse{Deal: dealToGen(deal), Contract: contractPtrToGen(contract)}, nil
}

func (a *API) LoseDeal(ctx context.Context, req *crmsalesv1.LoseDealRequest) (*crmsalesv1.DealResponse, error) {
	c, err := a.caller(ctx, permEdit)
	if err != nil {
		return nil, err
	}
	reason := ""
	if req.Body != nil {
		reason = req.Body.Reason
	}
	deal, err := a.Service.LoseDeal(ctx, c.businessID, c.userID, req.Id, reason)
	if err != nil {
		return nil, fail(ctx, "lose deal", err)
	}
	return &crmsalesv1.DealResponse{Deal: dealToGen(deal)}, nil
}

func (a *API) ReopenDeal(ctx context.Context, req *crmsalesv1.ReopenDealRequest) (*crmsalesv1.DealResponse, error) {
	c, err := a.caller(ctx, permEdit)
	if err != nil {
		return nil, err
	}
	var stageID int64
	if req.Body != nil {
		stageID = int64Value(req.Body.StageId)
	}
	deal, err := a.Service.ReopenDeal(ctx, c.businessID, c.userID, req.Id, stageID)
	if err != nil {
		return nil, fail(ctx, "reopen deal", err)
	}
	return &crmsalesv1.DealResponse{Deal: dealToGen(deal)}, nil
}

func (a *API) DeleteDeal(ctx context.Context, req *crmsalesv1.DeleteDealRequest) (*crmsalesv1.DeleteDealResponse, error) {
	c, err := a.caller(ctx, permDelete)
	if err != nil {
		return nil, err
	}
	if err := a.Service.DeleteDeal(ctx, c.businessID, req.Id); err != nil {
		return nil, fail(ctx, "delete deal", err)
	}
	return &crmsalesv1.DeleteDealResponse{Deleted: true}, nil
}
