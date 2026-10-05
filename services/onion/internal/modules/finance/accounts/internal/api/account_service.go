package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/finance/accounts/internal/application"
	"github.com/voxire/onion/internal/modules/finance/accounts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	financeaccountv1 "github.com/voxire/voxie/gen/go/finance/account/v1"
)

type AccountAPI struct {
	Service *application.Service
}

func NewAccountAPI(service *application.Service) *AccountAPI {
	return &AccountAPI{Service: service}
}

func (a *AccountAPI) ListAccounts(ctx context.Context, req *financeaccountv1.ListAccountsRequest) (*financeaccountv1.ListAccountsResponse, error) {
	businessID, err := accountBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	page, pageSize := req.GetPage(), req.GetPageSize()
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 500 {
		pageSize = 500
	}
	items, hasMore, listErr := a.Service.List(ctx, businessID, page, pageSize)
	if listErr != nil {
		return nil, accountStatusError(http.StatusInternalServerError, "Failed to list accounts")
	}
	accounts := make([]*financeaccountv1.Account, 0, len(items))
	for _, item := range items {
		accounts = append(accounts, newAccountProto(item))
	}
	return &financeaccountv1.ListAccountsResponse{Accounts: accounts, Page: page, PageSize: pageSize, HasMore: hasMore}, nil
}

func (a *AccountAPI) CreateAccount(ctx context.Context, req *financeaccountv1.CreateAccountRequest) (*financeaccountv1.CreateAccountResponse, error) {
	businessID, err := accountBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	id, createErr := a.Service.Create(ctx, businessID, domain.AccountInput{
		Name:     strings.TrimSpace(req.GetName()),
		Code:     strings.TrimSpace(req.GetCode()),
		Type:     strings.TrimSpace(req.GetType()),
		Currency: req.GetCurrency(),
		Balance:  req.GetBalance(),
		IsActive: req.GetIsActive(),
	})
	if createErr != nil {
		if errors.Is(createErr, domain.ErrNegativeBalance) {
			return nil, accountStatusError(http.StatusBadRequest, "Account balance cannot be negative")
		}
		return nil, accountStatusError(http.StatusInternalServerError, "Failed to create account")
	}
	return &financeaccountv1.CreateAccountResponse{Id: id, Message: "Account created successfully"}, nil
}

func (a *AccountAPI) UpdateAccount(ctx context.Context, req *financeaccountv1.UpdateAccountRequest) (*financeaccountv1.UpdateAccountResponse, error) {
	businessID, err := accountBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	body := req.GetBody()
	if body == nil {
		return nil, accountStatusError(http.StatusBadRequest, "Account body is required")
	}
	updateErr := a.Service.Update(ctx, businessID, req.GetId(), domain.AccountInput{
		Name:     strings.TrimSpace(body.GetName()),
		Code:     strings.TrimSpace(body.GetCode()),
		Type:     strings.TrimSpace(body.GetType()),
		Currency: body.GetCurrency(),
		Balance:  body.GetBalance(),
		IsActive: body.GetIsActive(),
	})
	if updateErr != nil {
		if errors.Is(updateErr, domain.ErrNegativeBalance) {
			return nil, accountStatusError(http.StatusBadRequest, "Account balance cannot be negative")
		}
		if errors.Is(updateErr, domain.ErrAccountNotFound) {
			return nil, accountStatusError(http.StatusNotFound, "Account not found")
		}
		logger.Ctx(ctx).Error("failed to update account", logger.Any("account_id", req.GetId()), logger.Error(updateErr))
		return nil, accountStatusError(http.StatusInternalServerError, "Failed to update account")
	}
	return &financeaccountv1.UpdateAccountResponse{Message: "Account updated successfully"}, nil
}

func (a *AccountAPI) DeleteAccount(ctx context.Context, req *financeaccountv1.DeleteAccountRequest) (*financeaccountv1.DeleteAccountResponse, error) {
	businessID, err := accountBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	deleteErr := a.Service.Delete(ctx, businessID, req.GetId())
	if deleteErr != nil {
		if errors.Is(deleteErr, domain.ErrAccountNotFound) {
			return nil, accountStatusError(http.StatusNotFound, "Account not found")
		}
		logger.Ctx(ctx).Error("failed to delete account", logger.Any("account_id", req.GetId()), logger.Error(deleteErr))
		return nil, accountStatusError(http.StatusInternalServerError, "Failed to delete account")
	}
	return &financeaccountv1.DeleteAccountResponse{Message: "Account deleted successfully"}, nil
}

func (a *AccountAPI) GetChartOfAccounts(ctx context.Context, _ *financeaccountv1.GetChartOfAccountsRequest) (*financeaccountv1.GetChartOfAccountsResponse, error) {
	businessID, err := accountBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	roots, chartErr := a.Service.ChartOfAccounts(ctx, businessID)
	if chartErr != nil {
		return nil, accountStatusError(http.StatusInternalServerError, "Failed to load chart of accounts")
	}
	tree := make([]*financeaccountv1.ChartOfAccountsNode, 0, len(roots))
	for _, root := range roots {
		tree = append(tree, newChartNodeProto(root))
	}
	return &financeaccountv1.GetChartOfAccountsResponse{Tree: tree}, nil
}

func newAccountProto(a domain.Account) *financeaccountv1.Account {
	return &financeaccountv1.Account{
		Id:        a.ID,
		OwnerId:   a.OwnerID,
		Name:      a.Name,
		Code:      a.Code,
		Type:      a.Type,
		Currency:  a.Currency,
		Balance:   a.Balance,
		IsActive:  a.IsActive,
		CreatedAt: a.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func newChartNodeProto(n *domain.AccountTreeNode) *financeaccountv1.ChartOfAccountsNode {
	children := make([]*financeaccountv1.ChartOfAccountsNode, 0, len(n.Children))
	for _, child := range n.Children {
		children = append(children, newChartNodeProto(child))
	}
	return &financeaccountv1.ChartOfAccountsNode{
		Id:       n.ID,
		Name:     n.Name,
		Code:     n.Code,
		Type:     n.Type,
		Children: children,
	}
}

func accountBusinessID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return 0, accountStatusError(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}

func accountStatusError(status int, message string) error {
	return onekiterrors.New(status, message)
}
