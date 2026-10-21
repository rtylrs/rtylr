package api

import (
	"context"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	crmsalesv1 "github.com/voxire/voxie/gen/go/crm/sales/v1"
)

func (a *API) ListLeads(ctx context.Context, req *crmsalesv1.ListLeadsRequest) (*crmsalesv1.ListLeadsResponse, error) {
	c, err := a.caller(ctx, permView)
	if err != nil {
		return nil, err
	}
	filter, err := listQuery(req).filter()
	if err != nil {
		return nil, err
	}
	filter.Sort, filter.Cursor, filter.Limit = domain.LeadSort(req.Sort), req.Cursor, int(req.Limit)
	for _, status := range splitList(req.Statuses) {
		filter.Statuses = append(filter.Statuses, domain.LeadStatus(status))
	}
	page, err := a.Service.ListLeads(ctx, c.businessID, filter, a.callerDay(req.Today, req.UtcOffsetMinutes))
	if err != nil {
		return nil, fail(ctx, "list leads", err)
	}
	leads := make([]*crmsalesv1.Lead, 0, len(page.Leads))
	for _, lead := range page.Leads {
		leads = append(leads, leadToGen(lead))
	}
	return &crmsalesv1.ListLeadsResponse{Leads: leads, NextCursor: page.NextCursor, Total: page.Total}, nil
}

func (a *API) GetLeadStats(ctx context.Context, req *crmsalesv1.GetLeadStatsRequest) (*crmsalesv1.GetLeadStatsResponse, error) {
	c, err := a.caller(ctx, permView)
	if err != nil {
		return nil, err
	}
	filter, err := leadQuery(*req).filter()
	if err != nil {
		return nil, err
	}
	stats, err := a.Service.LeadStats(ctx, c.businessID, filter, a.callerDay(req.Today, req.UtcOffsetMinutes))
	if err != nil {
		return nil, fail(ctx, "load lead stats", err)
	}
	byStatus := make([]*crmsalesv1.LeadStatusCount, 0, len(domain.LeadStatuses()))
	for _, status := range domain.LeadStatuses() {
		byStatus = append(byStatus, &crmsalesv1.LeadStatusCount{Status: string(status), Count: stats.ByStatus[status]})
	}
	return &crmsalesv1.GetLeadStatsResponse{
		ByStatus:           byStatus,
		FollowUpsOverdue:   stats.FollowUpsOverdue,
		FollowUpsToday:     stats.FollowUpsToday,
		CallsToday:         stats.CallsToday,
		ConnectedToday:     stats.ConnectedToday,
		ConvertedThisMonth: stats.ConvertedThisMonth,
	}, nil
}

func (a *API) ListLeadFacets(ctx context.Context, _ *crmsalesv1.ListLeadFacetsRequest) (*crmsalesv1.ListLeadFacetsResponse, error) {
	c, err := a.caller(ctx, permView)
	if err != nil {
		return nil, err
	}
	facets, err := a.Service.LeadFacets(ctx, c.businessID)
	if err != nil {
		return nil, fail(ctx, "list lead facets", err)
	}
	return &crmsalesv1.ListLeadFacetsResponse{
		Sources:    nonNil(facets.Sources),
		Industries: nonNil(facets.Industries),
		Sizes:      nonNil(facets.Sizes),
		Countries:  nonNil(facets.Countries),
		Cities:     nonNil(facets.Cities),
	}, nil
}

// nonNil keeps empty lists as [] on the wire.
func nonNil(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func (a *API) GetLead(ctx context.Context, req *crmsalesv1.GetLeadRequest) (*crmsalesv1.GetLeadResponse, error) {
	c, err := a.caller(ctx, permView)
	if err != nil {
		return nil, err
	}
	lead, err := a.Service.GetLead(ctx, c.businessID, req.Id)
	if err != nil {
		return nil, fail(ctx, "get lead", err)
	}
	return &crmsalesv1.GetLeadResponse{Lead: leadToGen(lead)}, nil
}

func (a *API) CreateLead(ctx context.Context, req *crmsalesv1.CreateLeadRequest) (*crmsalesv1.CreateLeadResponse, error) {
	c, err := a.caller(ctx, permCreate)
	if err != nil {
		return nil, err
	}
	in, err := leadFieldsFromGen(req.Lead)
	if err != nil {
		return nil, err
	}
	lead, err := a.Service.CreateLead(ctx, c.businessID, c.userID, in)
	if err != nil {
		return nil, fail(ctx, "create lead", err)
	}
	return &crmsalesv1.CreateLeadResponse{Lead: leadToGen(lead)}, nil
}

func (a *API) UpdateLead(ctx context.Context, req *crmsalesv1.UpdateLeadRequest) (*crmsalesv1.UpdateLeadResponse, error) {
	c, err := a.caller(ctx, permEdit)
	if err != nil {
		return nil, err
	}
	if req.Body == nil {
		return nil, invalid("lead is required")
	}
	in, err := leadFieldsFromGen(req.Body.Lead)
	if err != nil {
		return nil, err
	}
	lead, err := a.Service.UpdateLead(ctx, c.businessID, req.Id, domain.LeadChanges{
		LeadInput:        in,
		Status:           domain.LeadStatus(req.Body.Status),
		DisqualifyReason: req.Body.DisqualifyReason,
	})
	if err != nil {
		return nil, fail(ctx, "update lead", err)
	}
	return &crmsalesv1.UpdateLeadResponse{Lead: leadToGen(lead)}, nil
}

func (a *API) DeleteLead(ctx context.Context, req *crmsalesv1.DeleteLeadRequest) (*crmsalesv1.DeleteLeadResponse, error) {
	c, err := a.caller(ctx, permDelete)
	if err != nil {
		return nil, err
	}
	if err := a.Service.DeleteLead(ctx, c.businessID, req.Id); err != nil {
		return nil, fail(ctx, "delete lead", err)
	}
	return &crmsalesv1.DeleteLeadResponse{Deleted: true}, nil
}

func (a *API) UpdateLeads(ctx context.Context, req *crmsalesv1.UpdateLeadsRequest) (*crmsalesv1.UpdateLeadsResponse, error) {
	c, err := a.caller(ctx, permEdit)
	if err != nil {
		return nil, err
	}
	updated, err := a.Service.UpdateLeads(ctx, c.businessID, domain.BulkLeadUpdate{
		IDs:              req.Ids,
		AssignTo:         positivePtr(req.AssignTo),
		Unassign:         req.Unassign,
		Status:           domain.LeadStatus(req.Status),
		DisqualifyReason: req.DisqualifyReason,
	})
	if err != nil {
		return nil, fail(ctx, "update leads", err)
	}
	return &crmsalesv1.UpdateLeadsResponse{Updated: updated}, nil
}

func (a *API) DeleteLeads(ctx context.Context, req *crmsalesv1.DeleteLeadsRequest) (*crmsalesv1.DeleteLeadsResponse, error) {
	c, err := a.caller(ctx, permDelete)
	if err != nil {
		return nil, err
	}
	deleted, err := a.Service.DeleteLeads(ctx, c.businessID, req.Ids)
	if err != nil {
		return nil, fail(ctx, "delete leads", err)
	}
	return &crmsalesv1.DeleteLeadsResponse{Deleted: deleted}, nil
}

func (a *API) ConvertLead(ctx context.Context, req *crmsalesv1.ConvertLeadRequest) (*crmsalesv1.ConvertLeadResponse, error) {
	c, err := a.caller(ctx, permEdit)
	if err != nil {
		return nil, err
	}
	in := domain.ConvertLeadInput{}
	if req.Body != nil {
		in.CreateDeal = req.Body.CreateDeal
		if in.Deal, err = dealFieldsFromGen(req.Body.Deal); err != nil {
			return nil, err
		}
	}
	lead, deal, err := a.Service.ConvertLead(ctx, c.businessID, c.userID, req.Id, in)
	if err != nil {
		return nil, fail(ctx, "convert lead", err)
	}
	response := &crmsalesv1.ConvertLeadResponse{Lead: leadToGen(lead)}
	if deal != nil {
		response.Deal = dealToGen(*deal)
	}
	return response, nil
}

func (a *API) ListReps(ctx context.Context, _ *crmsalesv1.ListRepsRequest) (*crmsalesv1.ListRepsResponse, error) {
	c, err := a.caller(ctx, permView)
	if err != nil {
		return nil, err
	}
	reps, err := a.Service.ListReps(ctx, c.businessID)
	if err != nil {
		return nil, fail(ctx, "list reps", err)
	}
	out := make([]*crmsalesv1.Rep, 0, len(reps))
	for _, rep := range reps {
		out = append(out, &crmsalesv1.Rep{Id: rep.ID, Name: rep.Name})
	}
	return &crmsalesv1.ListRepsResponse{Reps: out}, nil
}
