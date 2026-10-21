package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
)

// ListLeads returns one page of leads. day is the caller's calendar day, which
// follow-up windows and date bounds are relative to.
func (s *Service) ListLeads(ctx context.Context, businessID int64, f domain.LeadFilter, day domain.CallerDay) (domain.LeadPage, error) {
	if err := f.Normalize(); err != nil {
		return domain.LeadPage{}, err
	}
	f.Limit = pageSize(f.Limit)
	return s.repo.ListLeads(ctx, businessID, f, day)
}

// LeadStats summarises the leads a filter selects, counting every status.
func (s *Service) LeadStats(ctx context.Context, businessID int64, f domain.LeadFilter, day domain.CallerDay) (domain.LeadStats, error) {
	f.Statuses = nil
	if err := f.Normalize(); err != nil {
		return domain.LeadStats{}, err
	}
	return s.repo.LeadStats(ctx, businessID, f, day)
}

// LeadFacets lists the values leads have, for the filter pickers.
func (s *Service) LeadFacets(ctx context.Context, businessID int64) (domain.LeadFacets, error) {
	return s.repo.LeadFacets(ctx, businessID)
}

func (s *Service) GetLead(ctx context.Context, businessID, id int64) (domain.Lead, error) {
	return s.repo.GetLead(ctx, businessID, id)
}

func (s *Service) CreateLead(ctx context.Context, businessID, actorID int64, in domain.LeadInput) (domain.Lead, error) {
	if err := in.Normalize(); err != nil {
		return domain.Lead{}, err
	}
	if err := s.validateAssignee(ctx, businessID, in.AssignedTo); err != nil {
		return domain.Lead{}, err
	}
	return s.repo.CreateLead(ctx, businessID, actorID, in)
}

// UpdateLead applies a full edit; the status rules are checked against the
// stored status.
func (s *Service) UpdateLead(ctx context.Context, businessID, id int64, changes domain.LeadChanges) (domain.Lead, error) {
	if err := s.validateAssignee(ctx, businessID, changes.AssignedTo); err != nil {
		return domain.Lead{}, err
	}
	return s.repo.UpdateLead(ctx, businessID, id, changes)
}

func (s *Service) DeleteLead(ctx context.Context, businessID, id int64) error {
	return s.repo.DeleteLead(ctx, businessID, id)
}

// UpdateLeads assigns or re-statuses many leads and reports how many changed.
func (s *Service) UpdateLeads(ctx context.Context, businessID int64, update domain.BulkLeadUpdate) (int64, error) {
	if err := update.Normalize(); err != nil {
		return 0, err
	}
	if update.Unassign {
		update.AssignTo = nil
	} else if err := s.validateAssignee(ctx, businessID, update.AssignTo); err != nil {
		return 0, err
	}
	return s.repo.UpdateLeads(ctx, businessID, update)
}

// DeleteLeads removes the unconverted leads among ids and reports how many.
func (s *Service) DeleteLeads(ctx context.Context, businessID int64, ids []int64) (int64, error) {
	if len(ids) == 0 || len(ids) > 1000 {
		return 0, &domain.ValidationError{Message: "delete between 1 and 1000 leads at a time"}
	}
	return s.repo.DeleteLeads(ctx, businessID, ids)
}

// ConvertLead qualifies a lead. With a deal, it opens in the chosen pipeline
// stage (by default the default pipeline's first open stage) and inherits the
// lead's estimated value, owner and name.
func (s *Service) ConvertLead(ctx context.Context, businessID, actorID, id int64, in domain.ConvertLeadInput) (domain.Lead, *domain.Deal, error) {
	lead, err := s.repo.GetLead(ctx, businessID, id)
	if err != nil {
		return domain.Lead{}, nil, err
	}
	if lead.Status == domain.LeadQualified {
		return domain.Lead{}, nil, domain.ErrLeadConverted
	}
	if !in.CreateDeal {
		return s.repo.ConvertLead(ctx, businessID, actorID, id, nil, nil, s.now())
	}
	deal := in.Deal
	if deal.Title == "" {
		deal.Title = lead.DisplayName()
	}
	if deal.Value == 0 && lead.Value != nil {
		deal.Value = *lead.Value
		if deal.Billing == "" {
			deal.Billing = lead.Billing
		}
	}
	if err := deal.Normalize(); err != nil {
		return domain.Lead{}, nil, err
	}
	if err := s.validateAssignee(ctx, businessID, deal.AssignedTo); err != nil {
		return domain.Lead{}, nil, err
	}
	stage, err := s.openStage(ctx, businessID, deal.PipelineID, deal.StageID)
	if err != nil {
		return domain.Lead{}, nil, err
	}
	return s.repo.ConvertLead(ctx, businessID, actorID, id, &deal, &stage, s.now())
}

// LinkLeadCustomer records the customer a converted lead's person became.
func (s *Service) LinkLeadCustomer(ctx context.Context, businessID, leadID, customerID int64) error {
	if err := s.repo.LinkLeadCustomer(ctx, businessID, leadID, customerID); err != nil {
		return err
	}
	s.invalidate(ctx, businessID)
	return nil
}
