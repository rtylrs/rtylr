package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/accounts/internal/application"
	"github.com/voxire/onion/internal/modules/accounts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	financeaccountv1 "github.com/voxire/voxie/gen/go/finance/account/v1"
)

func newPayload(a domain.Account) *financeaccountv1.Account {
	return &financeaccountv1.Account{
		Id: a.ID, OwnerId: a.OwnerID, Name: a.Name, Type: a.Type,
		Currency: a.Currency, Balance: a.Balance, IsActive: a.IsActive,
		CreatedAt: a.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func ListRoutes() []string { return []string{"/finance/accounts"} }

type (
	ListHandler  struct{ Service *application.Service }
	ListRequest  struct{}
	ListResponse struct {
		Accounts []*financeaccountv1.Account `json:"accounts"`
	}
)

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}
	items, err := h.Service.List(ctx, businessID)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list accounts"}
	}
	payloads := make([]*financeaccountv1.Account, 0, len(items))
	for _, item := range items {
		payloads = append(payloads, newPayload(item))
	}
	return ListResponse{Accounts: payloads}, nil
}

func CreateRoutes() []string { return []string{"/finance/accounts"} }

type (
	CreateHandler struct{ Service *application.Service }
	CreateRequest struct {
		Name     string  `json:"name"`
		Type     string  `json:"type"`
		Currency string  `json:"currency"`
		Balance  float64 `json:"balance"`
		IsActive bool    `json:"is_active"`
	}
)

type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.Name) == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	if strings.TrimSpace(req.Type) == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "type is required"}
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}
	input := domain.AccountInput{
		Name: req.Name, Type: req.Type, Currency: req.Currency,
		Balance: req.Balance, IsActive: req.IsActive,
	}
	id, err := h.Service.Create(ctx, businessID, input)
	if err != nil {
		if errors.Is(err, domain.ErrNegativeBalance) {
			return CreateResponse{}, &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "Account balance cannot be negative"}
		}
		return CreateResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to create account"}
	}
	return CreateResponse{ID: id, Message: "Account created successfully"}, nil
}

// ---------------------------------------------------------------------------
// Chart of Accounts (tree)
// ---------------------------------------------------------------------------

func ChartOfAccountsRoutes() []string { return []string{"/finance/chart-of-accounts"} }

type (
	ChartOfAccountsHandler struct{ Service *application.Service }
	ChartOfAccountsRequest struct{}
)

type ChartOfAccountsNode struct {
	ID       int64                  `json:"id"`
	Name     string                 `json:"name"`
	Code     string                 `json:"code"`
	Type     string                 `json:"type"`
	Children []*ChartOfAccountsNode `json:"children"`
}

type ChartOfAccountsResponse struct {
	Tree []*ChartOfAccountsNode `json:"tree"`
}

func toChartNode(n *domain.AccountTreeNode) *ChartOfAccountsNode {
	children := make([]*ChartOfAccountsNode, 0, len(n.Children))
	for _, c := range n.Children {
		children = append(children, toChartNode(c))
	}
	return &ChartOfAccountsNode{
		ID:       n.ID,
		Name:     n.Name,
		Code:     n.Code,
		Type:     n.Type,
		Children: children,
	}
}

func (h *ChartOfAccountsHandler) Serve(ctx context.Context, _ *ChartOfAccountsRequest) (ChartOfAccountsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ChartOfAccountsResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ChartOfAccountsResponse{}, errorModels.PermissionDeniedServerError()
	}

	roots, err := h.Service.ChartOfAccounts(ctx, businessID)
	if err != nil {
		return ChartOfAccountsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to load chart of accounts"}
	}

	tree := make([]*ChartOfAccountsNode, 0, len(roots))
	for _, n := range roots {
		tree = append(tree, toChartNode(n))
	}
	return ChartOfAccountsResponse{Tree: tree}, nil
}
