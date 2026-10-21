package api

import (
	"context"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/application"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	crmsalesv1 "github.com/voxire/voxie/gen/go/crm/sales/v1"
)

func (a *API) ListContracts(ctx context.Context, req *crmsalesv1.ListContractsRequest) (*crmsalesv1.ListContractsResponse, error) {
	c, err := a.caller(ctx, permView)
	if err != nil {
		return nil, err
	}
	page, err := a.Service.ListContracts(ctx, c.businessID, application.ContractFilter{Signed: req.Signed}, req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, fail(ctx, "list contracts", err)
	}
	out := make([]*crmsalesv1.Contract, 0, len(page.Contracts))
	for _, contract := range page.Contracts {
		out = append(out, contractToGen(contract))
	}
	return &crmsalesv1.ListContractsResponse{Contracts: out, Total: page.Total, Limit: page.Limit, Offset: page.Offset}, nil
}

func (a *API) GetContract(ctx context.Context, req *crmsalesv1.GetContractRequest) (*crmsalesv1.GetContractResponse, error) {
	c, err := a.caller(ctx, permView)
	if err != nil {
		return nil, err
	}
	contract, err := a.Service.GetContract(ctx, c.businessID, req.Id)
	if err != nil {
		return nil, fail(ctx, "get contract", err)
	}
	return &crmsalesv1.GetContractResponse{Contract: contractToGen(contract)}, nil
}

func (a *API) CreateContract(ctx context.Context, req *crmsalesv1.CreateContractRequest) (*crmsalesv1.CreateContractResponse, error) {
	c, err := a.caller(ctx, permCreate)
	if err != nil {
		return nil, err
	}
	value, err := parseAmount(req.Value, "value")
	if err != nil {
		return nil, err
	}
	in := domain.ContractInput{
		DealID:       positivePtr(req.DealId),
		CustomerID:   positivePtr(req.CustomerId),
		Plan:         req.Plan,
		Seats:        int(req.Seats),
		Value:        value,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		SignedAt:     req.SignedAt,
		SignedByName: req.SignedByName,
		Notes:        req.Notes,
	}
	if req.Billing != nil {
		in.Billing = domain.Billing(*req.Billing)
	}
	contract, err := a.Service.CreateContract(ctx, c.businessID, c.userID, in)
	if err != nil {
		return nil, fail(ctx, "create contract", err)
	}
	return &crmsalesv1.CreateContractResponse{Contract: contractToGen(contract)}, nil
}

func (a *API) UpdateContract(ctx context.Context, req *crmsalesv1.UpdateContractRequest) (*crmsalesv1.UpdateContractResponse, error) {
	c, err := a.caller(ctx, permEdit)
	if err != nil {
		return nil, err
	}
	if req.Body == nil {
		return nil, invalid("contract terms are required")
	}
	value, err := parseAmount(req.Body.Value, "value")
	if err != nil {
		return nil, err
	}
	in := domain.ContractInput{
		Plan:         req.Body.Plan,
		Seats:        int(req.Body.Seats),
		Value:        value,
		StartDate:    req.Body.StartDate,
		EndDate:      req.Body.EndDate,
		SignedAt:     req.Body.SignedAt,
		SignedByName: req.Body.SignedByName,
		Notes:        req.Body.Notes,
	}
	if req.Body.Billing != nil {
		in.Billing = domain.Billing(*req.Body.Billing)
	}
	contract, err := a.Service.UpdateContract(ctx, c.businessID, req.Id, in)
	if err != nil {
		return nil, fail(ctx, "update contract", err)
	}
	return &crmsalesv1.UpdateContractResponse{Contract: contractToGen(contract)}, nil
}
