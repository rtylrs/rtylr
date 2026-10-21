package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/infrastructure/persistence/sqlc"
)

// EnsureDefaultPipeline creates the business's first pipeline and its stages
// unless it exists. The insert is a no-op for a concurrent second caller,
// which then waits on the unique key until the first commits its stages.
func (r *Repository) EnsureDefaultPipeline(ctx context.Context, businessID int64) error {
	return r.inTx(ctx, func(q *sqlc.Queries, _ *sql.Tx) error {
		result, err := q.CreateDefaultPipeline(ctx, sqlc.CreateDefaultPipelineParams{
			BusinessID: businessID, Name: domain.DefaultPipelineName,
		})
		if err != nil {
			return err
		}
		created, err := result.RowsAffected()
		if err != nil || created == 0 {
			return err
		}
		pipelineID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		return createStages(ctx, q, businessID, pipelineID, domain.DefaultStages(), 0)
	})
}

func createStages(ctx context.Context, q *sqlc.Queries, businessID, pipelineID int64, stages []domain.StageInput, from int) error {
	for i, stage := range stages {
		if _, err := q.CreateStage(ctx, sqlc.CreateStageParams{
			BusinessID:  businessID,
			PipelineID:  pipelineID,
			Name:        stage.Name,
			Position:    int32(from + i),          // #nosec G115 -- a pipeline has at most 20 stages.
			Probability: uint8(stage.Probability), // #nosec G115 -- validated to 0-100.
			Kind:        sqlc.CrmPipelineStagesKind(stage.Kind),
		}); err != nil {
			return err
		}
	}
	return nil
}

// ListPipelines returns the business's pipelines with their stages in order.
func (r *Repository) ListPipelines(ctx context.Context, businessID int64) ([]domain.Pipeline, error) {
	pipelineRows, err := r.queries.ListPipelines(ctx, businessID)
	if err != nil {
		return nil, err
	}
	stageRows, err := r.queries.ListPipelineStages(ctx, businessID)
	if err != nil {
		return nil, err
	}
	stages := map[int64][]domain.Stage{}
	for _, row := range stageRows {
		stages[row.PipelineID] = append(stages[row.PipelineID], stageFromRow(row))
	}
	pipelines := make([]domain.Pipeline, 0, len(pipelineRows))
	for _, row := range pipelineRows {
		pipelines = append(pipelines, domain.Pipeline{
			ID:        row.ID,
			Name:      row.Name,
			Position:  int(row.Position),
			IsDefault: row.DefaultMarker.Valid,
			Stages:    stages[row.ID],
		})
	}
	return pipelines, nil
}

func (r *Repository) GetPipeline(ctx context.Context, businessID, id int64) (domain.Pipeline, error) {
	pipelines, err := r.ListPipelines(ctx, businessID)
	if err != nil {
		return domain.Pipeline{}, err
	}
	for _, pipeline := range pipelines {
		if pipeline.ID == id {
			return pipeline, nil
		}
	}
	return domain.Pipeline{}, domain.ErrPipelineNotFound
}

func (r *Repository) GetStage(ctx context.Context, businessID, id int64) (domain.Stage, error) {
	row, err := r.queries.GetStage(ctx, sqlc.GetStageParams{ID: id, BusinessID: businessID})
	if err != nil {
		return domain.Stage{}, notFound(err, domain.ErrStageNotFound)
	}
	return stageFromRow(row), nil
}

// CreatePipeline adds a pipeline after the existing ones.
func (r *Repository) CreatePipeline(ctx context.Context, businessID int64, in domain.PipelineInput) (int64, error) {
	var id int64
	err := r.inTx(ctx, func(q *sqlc.Queries, _ *sql.Tx) error {
		count, err := q.CountPipelines(ctx, businessID)
		if err != nil {
			return err
		}
		result, err := q.CreatePipeline(ctx, sqlc.CreatePipelineParams{
			BusinessID: businessID, Name: in.Name, Position: int32(count), // #nosec G115 -- pipelines per business are few.
		})
		if err != nil {
			return err
		}
		if id, err = result.LastInsertId(); err != nil {
			return err
		}
		for i := range in.Stages {
			in.Stages[i].ID = nil
		}
		return createStages(ctx, q, businessID, id, in.Stages, 0)
	})
	return id, err
}

// UpdatePipeline renames a pipeline and replaces its stages in order. A
// stage's kind is fixed once created; a removed stage must hold no deals.
func (r *Repository) UpdatePipeline(ctx context.Context, businessID, id int64, in domain.PipelineInput) error {
	current, err := r.GetPipeline(ctx, businessID, id)
	if err != nil {
		return err
	}
	return r.inTx(ctx, func(q *sqlc.Queries, _ *sql.Tx) error {
		if err := q.RenamePipeline(ctx, sqlc.RenamePipelineParams{Name: in.Name, ID: id, BusinessID: businessID}); err != nil {
			return err
		}
		kept := map[int64]bool{}
		for position, stage := range in.Stages {
			if stage.ID == nil {
				if err := createStages(ctx, q, businessID, id, []domain.StageInput{stage}, position); err != nil {
					return err
				}
				continue
			}
			existing, ok := current.Stage(*stage.ID)
			if !ok {
				return domain.ErrStageNotFound
			}
			if existing.Kind != stage.Kind {
				return &domain.ValidationError{Message: "a stage's kind cannot change"}
			}
			kept[existing.ID] = true
			if err := q.UpdateStage(ctx, sqlc.UpdateStageParams{
				Name:        stage.Name,
				Position:    int32(position),          // #nosec G115 -- at most 20 stages.
				Probability: uint8(stage.Probability), // #nosec G115 -- validated to 0-100.
				ID:          existing.ID,
				BusinessID:  businessID,
				PipelineID:  id,
			}); err != nil {
				return err
			}
		}
		for _, stage := range current.Stages {
			if kept[stage.ID] {
				continue
			}
			deals, err := q.CountStageDeals(ctx, sqlc.CountStageDealsParams{BusinessID: businessID, StageID: stage.ID})
			if err != nil {
				return err
			}
			if deals > 0 {
				return domain.ErrStageInUse
			}
			if _, err := q.DeleteStage(ctx, sqlc.DeleteStageParams{ID: stage.ID, BusinessID: businessID, PipelineID: id}); err != nil {
				return err
			}
		}
		return nil
	})
}

// DeletePipeline removes an empty pipeline other than the default.
func (r *Repository) DeletePipeline(ctx context.Context, businessID, id int64) error {
	pipeline, err := r.GetPipeline(ctx, businessID, id)
	if err != nil {
		return err
	}
	if pipeline.IsDefault {
		return &domain.ValidationError{Message: "the default pipeline cannot be deleted"}
	}
	return r.inTx(ctx, func(q *sqlc.Queries, _ *sql.Tx) error {
		deals, err := q.CountPipelineDeals(ctx, sqlc.CountPipelineDealsParams{BusinessID: businessID, PipelineID: id})
		if err != nil {
			return err
		}
		if deals > 0 {
			return domain.ErrPipelineInUse
		}
		if err := q.DeletePipelineStages(ctx, sqlc.DeletePipelineStagesParams{BusinessID: businessID, PipelineID: id}); err != nil {
			return err
		}
		_, err = q.DeletePipeline(ctx, sqlc.DeletePipelineParams{ID: id, BusinessID: businessID})
		return err
	})
}
