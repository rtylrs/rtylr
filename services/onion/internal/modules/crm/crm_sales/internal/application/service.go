package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	crmSalesScope = "crm_sales"
	crmSalesTTL   = 5 * time.Minute

	defaultPageSize = 50
	maxPageSize     = 200
)

// repository is the persistence the service needs; infrastructure.Repository
// implements it.
type repository interface {
	ValidateAssignee(ctx context.Context, businessID, userID int64) error
	ListReps(ctx context.Context, businessID int64) ([]domain.Rep, error)

	ListLeads(ctx context.Context, businessID int64, f domain.LeadFilter, day domain.CallerDay) (domain.LeadPage, error)
	LeadStats(ctx context.Context, businessID int64, f domain.LeadFilter, day domain.CallerDay) (domain.LeadStats, error)
	LeadFacets(ctx context.Context, businessID int64) (domain.LeadFacets, error)
	GetLead(ctx context.Context, businessID, id int64) (domain.Lead, error)
	CreateLead(ctx context.Context, businessID, ownerID int64, in domain.LeadInput) (domain.Lead, error)
	UpdateLead(ctx context.Context, businessID, id int64, changes domain.LeadChanges) (domain.Lead, error)
	DeleteLead(ctx context.Context, businessID, id int64) error
	DeleteLeads(ctx context.Context, businessID int64, ids []int64) (int64, error)
	UpdateLeads(ctx context.Context, businessID int64, update domain.BulkLeadUpdate) (int64, error)
	ConvertLead(ctx context.Context, businessID, actorID, id int64, deal *domain.DealInput, stage *domain.Stage, now time.Time) (domain.Lead, *domain.Deal, error)
	LinkLeadCustomer(ctx context.Context, businessID, leadID, customerID int64) error

	LogActivity(ctx context.Context, businessID, actorID int64, in domain.ActivityInput) (domain.Activity, *domain.Task, error)
	ListLeadActivities(ctx context.Context, businessID, leadID int64, cursor string, limit int) (domain.ActivityPage, error)
	ListDealActivities(ctx context.Context, businessID int64, deal domain.Deal, cursor string, limit int) (domain.ActivityPage, error)
	DeleteActivity(ctx context.Context, businessID, id int64) error

	CreateTask(ctx context.Context, businessID, actorID int64, in domain.TaskInput) (domain.Task, error)
	GetTask(ctx context.Context, businessID, id int64) (domain.Task, error)
	UpdateTask(ctx context.Context, businessID, id int64, in domain.TaskInput) (domain.Task, error)
	CompleteTask(ctx context.Context, businessID, id int64, done bool, now time.Time) (domain.Task, error)
	DeleteTask(ctx context.Context, businessID, id int64) error
	ListTasks(ctx context.Context, businessID int64, f domain.TaskFilter, today time.Time) (domain.TaskPage, error)
	CountTasks(ctx context.Context, businessID int64, assignedTo *int64, today time.Time) (domain.TaskCounts, error)

	EnsureDefaultPipeline(ctx context.Context, businessID int64) error
	ListPipelines(ctx context.Context, businessID int64) ([]domain.Pipeline, error)
	GetPipeline(ctx context.Context, businessID, id int64) (domain.Pipeline, error)
	GetStage(ctx context.Context, businessID, id int64) (domain.Stage, error)
	CreatePipeline(ctx context.Context, businessID int64, in domain.PipelineInput) (int64, error)
	UpdatePipeline(ctx context.Context, businessID, id int64, in domain.PipelineInput) error
	DeletePipeline(ctx context.Context, businessID, id int64) error

	ListDeals(ctx context.Context, businessID int64, f domain.DealFilter) (domain.DealPage, error)
	ListStageDeals(ctx context.Context, businessID, pipelineID, stageID int64, assignedTo *int64, search, cursor string, limit int) ([]domain.Deal, string, error)
	StageTotals(ctx context.Context, businessID, pipelineID int64, assignedTo *int64, search string) (map[int64]domain.StageTotal, error)
	DealOutcomesSince(ctx context.Context, businessID int64, pipelineID, assignedTo *int64, since time.Time) (domain.DealOutcome, error)
	// SalesActivity counts calls, new leads and won deals per caller day
	// (keyed YYYY-MM-DD, shifted by offsetMinutes) since `since`, and each
	// rep's share.
	SalesActivity(ctx context.Context, businessID int64, rep *int64, since time.Time, offsetMinutes int64) (map[string]domain.ActivityDay, []domain.RepActivity, error)
	GetDeal(ctx context.Context, businessID, id int64) (domain.Deal, error)
	CreateDeal(ctx context.Context, businessID, actorID int64, in domain.DealInput, stage domain.Stage, now time.Time) (domain.Deal, error)
	UpdateDeal(ctx context.Context, businessID, id int64, in domain.DealInput) (domain.Deal, error)
	MoveDeal(ctx context.Context, businessID, actorID, id int64, stage domain.Stage, lostReason string, contract *domain.ContractInput, now time.Time) (domain.Deal, *domain.Contract, error)
	DeleteDeal(ctx context.Context, businessID, id int64) error
	DealContract(ctx context.Context, businessID, dealID int64) (*domain.Contract, error)

	CreateImportBatch(ctx context.Context, businessID, actorID int64, fileName string, totalRows int) (domain.ImportBatch, error)
	GetImportBatch(ctx context.Context, businessID, id int64) (domain.ImportBatch, error)
	ListImportBatches(ctx context.Context, businessID int64) ([]domain.ImportBatch, error)
	ImportLeadRows(ctx context.Context, businessID, actorID, batchID int64, opts domain.ImportOptions, rows []domain.ImportRow, failed []domain.ImportRowResult) ([]domain.ImportRowResult, error)
	FinishImport(ctx context.Context, businessID, id int64, now time.Time) (domain.ImportBatch, error)
	UndoImport(ctx context.Context, businessID, id int64, now time.Time) (domain.UndoImportResult, error)

	ListContracts(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Contract, int64, error)
	ListSignedContracts(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Contract, int64, error)
	ListUnsignedContracts(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Contract, int64, error)
	GetContract(ctx context.Context, businessID, id int64) (domain.Contract, error)
	CreateContract(ctx context.Context, businessID, actorID int64, in domain.ContractInput) (domain.Contract, error)
	UpdateContract(ctx context.Context, businessID, id int64, in domain.ContractInput) (domain.Contract, error)
}

type Service struct {
	repo repository
	now  func() time.Time
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo, now: func() time.Time { return time.Now().UTC() }}
}

// WithClock replaces the service's clock, for tests.
func (s *Service) WithClock(now func() time.Time) *Service {
	s.now = now
	return s
}

// Now is the service's current time.
func (s *Service) Now() time.Time { return s.now() }

// invalidate drops cached contract reads after any write that can change a
// contract. Lists of leads and deals are read live.
func (s *Service) invalidate(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(crmSalesScope, businessID))
}

func pageSize(limit int) int {
	if limit <= 0 {
		return defaultPageSize
	}
	if limit > maxPageSize {
		return maxPageSize
	}
	return limit
}

func (s *Service) validateAssignee(ctx context.Context, businessID int64, userID *int64) error {
	if userID == nil || *userID <= 0 {
		return nil
	}
	return s.repo.ValidateAssignee(ctx, businessID, *userID)
}

func (s *Service) ListReps(ctx context.Context, businessID int64) ([]domain.Rep, error) {
	return s.repo.ListReps(ctx, businessID)
}
