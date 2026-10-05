package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/application"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	crmsalesv1 "github.com/voxire/voxie/gen/go/crm/sales/v1"
)

// ContractAPI is the generated crm.sales.v1 surface over the existing sales
// application service. It is a thin translation layer: tenancy, validation
// caps and cache behaviour live in the service, exactly as the legacy chi
// handlers had them.
type ContractAPI struct {
	Service *application.Service
}

func NewContractAPI(service *application.Service) *ContractAPI {
	return &ContractAPI{Service: service}
}

func (a *ContractAPI) businessID(ctx context.Context) (int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return 0, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}

func contractToGen(c domain.Contract) *crmsalesv1.Contract {
	var leadID *int64
	if c.LeadID > 0 {
		v := c.LeadID
		leadID = &v
	}
	return &crmsalesv1.Contract{
		Id:           c.ID,
		LeadId:       leadID,
		Plan:         c.Plan,
		Mrr:          c.MRR.String(),
		Seats:        int32(c.Seats), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		SignedAt:     c.SignedAt,
		SignedByName: c.SignedByName,
		Notes:        c.Notes,
		CreatedAt:    c.CreatedAt.Unix(),
		UpdatedAt:    c.UpdatedAt.Unix(),
	}
}

func (a *ContractAPI) ListContracts(ctx context.Context, req *crmsalesv1.ListContractsRequest) (*crmsalesv1.ListContractsResponse, error) {
	businessID, err := a.businessID(ctx)
	if err != nil {
		return nil, err
	}
	limit, offset := req.GetLimit(), req.GetOffset()
	var page domain.ContractPage
	switch {
	case req.Signed != nil && *req.Signed:
		page, err = a.Service.ListSignedContracts(ctx, businessID, limit, offset)
	case req.Signed != nil && !*req.Signed:
		page, err = a.Service.ListUnsignedContracts(ctx, businessID, limit, offset)
	default:
		page, err = a.Service.ListContracts(ctx, businessID, limit, offset)
	}
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list contracts")
	}
	out := make([]*crmsalesv1.Contract, 0, len(page.Contracts))
	for _, c := range page.Contracts {
		out = append(out, contractToGen(c))
	}
	return &crmsalesv1.ListContractsResponse{
		Contracts: out,
		Total:     page.Total,
		Limit:     page.Limit,
		Offset:    page.Offset,
	}, nil
}

func (a *ContractAPI) GetContract(ctx context.Context, req *crmsalesv1.GetContractRequest) (*crmsalesv1.GetContractResponse, error) {
	businessID, err := a.businessID(ctx)
	if err != nil {
		return nil, err
	}
	c, err := a.Service.GetContract(ctx, req.GetId(), businessID)
	if err != nil {
		if errors.Is(err, domain.ErrContractNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Contract not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get contract")
	}
	return &crmsalesv1.GetContractResponse{Contract: contractToGen(c)}, nil
}

func (a *ContractAPI) CreateContract(ctx context.Context, req *crmsalesv1.CreateContractRequest) (*crmsalesv1.CreateContractResponse, error) {
	businessID, err := a.businessID(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetLeadId() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "lead_id is required")
	}
	if req.GetMrr() == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "mrr is required")
	}
	mrr, parseErr := money.FromString(req.GetMrr())
	if parseErr != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "mrr must be a decimal number")
	}
	if mrr < 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "mrr must not be negative")
	}
	c, err := a.Service.CreateContract(ctx, businessID, domain.CreateContractRequest{
		LeadID:       req.GetLeadId(),
		Plan:         req.GetPlan(),
		MRR:          mrr,
		Seats:        int(req.GetSeats()),
		SignedAt:     req.SignedAt,
		SignedByName: req.GetSignedByName(),
		Notes:        req.GetNotes(),
	})
	if err != nil {
		if errors.Is(err, domain.ErrLeadNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Lead not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create contract")
	}
	return &crmsalesv1.CreateContractResponse{Contract: contractToGen(c)}, nil
}

func (a *ContractAPI) UpdateContract(ctx context.Context, req *crmsalesv1.UpdateContractRequest) (*crmsalesv1.UpdateContractResponse, error) {
	businessID, err := a.businessID(ctx)
	if err != nil {
		return nil, err
	}
	// The generated body exposes plan/mrr/seats/signed_by_name/notes as plain
	// values, so an omitted field arrives as its zero value. Merge onto the
	// stored contract instead of writing zero values over fields the caller
	// did not send, as the contract documents.
	stored, err := a.Service.GetContract(ctx, req.GetId(), businessID)
	if err != nil {
		if errors.Is(err, domain.ErrContractNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Contract not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update contract")
	}
	domainReq, err := contractUpdateFromBody(req.GetBody(), stored)
	if err != nil {
		return nil, err
	}
	c, err := a.Service.UpdateContract(ctx, req.GetId(), businessID, domainReq)
	if err != nil {
		if errors.Is(err, domain.ErrContractNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Contract not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update contract")
	}
	return &crmsalesv1.UpdateContractResponse{Contract: contractToGen(c)}, nil
}

// contractUpdateFromBody merges an amend body onto the stored contract,
// mirroring the partial-update semantics the generated contract documents.
// Absent fields are indistinguishable from their zero value on this generated
// shape, so for those a zero value keeps the stored one: a contract always has
// a non-empty plan and at least one seat, and mrr=0 is expressed as "0"
// (an empty decimal string is not a valid amount). signed_at is optional on
// the wire: nil keeps the stored signature, an empty string clears it, and a
// value replaces it.
func contractUpdateFromBody(body *crmsalesv1.UpdateContractBody, stored domain.Contract) (domain.UpdateContractRequest, error) {
	domainReq := domain.UpdateContractRequest{
		Plan:         stored.Plan,
		MRR:          stored.MRR,
		Seats:        stored.Seats,
		SignedAt:     stored.SignedAt,
		SignedByName: stored.SignedByName,
		ClosedBy:     stored.ClosedBy,
		Notes:        stored.Notes,
	}
	if body == nil {
		return domainReq, nil
	}
	if plan := body.GetPlan(); strings.TrimSpace(plan) != "" {
		domainReq.Plan = plan
	}
	if body.GetMrr() != "" {
		mrr, parseErr := money.FromString(body.GetMrr())
		if parseErr != nil {
			return domain.UpdateContractRequest{}, onekiterrors.New(http.StatusBadRequest, "mrr must be a decimal number")
		}
		if mrr < 0 {
			return domain.UpdateContractRequest{}, onekiterrors.New(http.StatusBadRequest, "mrr must not be negative")
		}
		domainReq.MRR = mrr
	}
	if seats := int(body.GetSeats()); seats != 0 {
		// The repository stores seats as INT; larger values wrap negative (#836).
		if seats < 0 {
			return domain.UpdateContractRequest{}, onekiterrors.New(http.StatusBadRequest, "seats must be between 0 and 2147483647")
		}
		domainReq.Seats = seats
	}
	if body.SignedAt != nil {
		v := *body.SignedAt
		if v == "" {
			domainReq.SignedAt = nil
		} else {
			domainReq.SignedAt = &v
		}
	}
	if body.GetSignedByName() != "" {
		domainReq.SignedByName = body.GetSignedByName()
	}
	if body.GetNotes() != "" {
		domainReq.Notes = body.GetNotes()
	}
	return domainReq, nil
}
