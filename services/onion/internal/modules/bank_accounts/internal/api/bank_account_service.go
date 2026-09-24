package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/bank_accounts/internal/application"
	"github.com/voxire/onion/internal/modules/bank_accounts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	commonpaginationv1 "github.com/voxire/voxie/gen/go/common/pagination/v1"
	financebankaccountv1 "github.com/voxire/voxie/gen/go/finance/bank_account/v1"
)

type BankAccountAPI struct {
	Service *application.Service
}

func NewBankAccountAPI(service *application.Service) *BankAccountAPI {
	return &BankAccountAPI{Service: service}
}

func (a *BankAccountAPI) ListBankAccounts(ctx context.Context, req *financebankaccountv1.ListBankAccountsRequest) (*financebankaccountv1.ListBankAccountsResponse, error) {
	businessID, err := bankAccountBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	page, listErr := a.Service.ListPage(ctx, businessID, application.ListOptions{
		Page:    req.GetPage(),
		Limit:   req.GetLimit(),
		SortBy:  req.GetSortBy(),
		SortDir: req.GetSortDir(),
	})
	if listErr != nil {
		return nil, bankAccountStatusError(http.StatusInternalServerError, "Failed to list bank accounts")
	}
	out := make([]*financebankaccountv1.BankAccount, 0, len(page.Accounts))
	for _, account := range page.Accounts {
		out = append(out, newBankAccountProto(account))
	}
	return &financebankaccountv1.ListBankAccountsResponse{
		Accounts: out,
		Pagination: &commonpaginationv1.PaginationResponse{
			Page:       page.Page,
			Limit:      page.Limit,
			Total:      page.Total,
			TotalPages: totalPages(page.Limit, page.Total),
			HasNext:    page.HasNext,
			HasPrev:    page.HasPrev,
		},
	}, nil
}

func (a *BankAccountAPI) CreateBankAccount(ctx context.Context, req *financebankaccountv1.CreateBankAccountRequest) (*financebankaccountv1.CreateBankAccountResponse, error) {
	businessID, err := bankAccountBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	balance, balanceErr := money.FromFloat64(req.GetBalance())
	if balanceErr != nil {
		return nil, bankAccountStatusError(http.StatusBadRequest, "invalid balance")
	}
	input := domain.BankAccountInput{
		Name:        strings.TrimSpace(req.GetName()),
		Institution: req.GetBankName(),
		AccountType: req.GetAccountType(),
		LastFour:    req.GetLastFour(),
		Balance:     balance,
		Currency:    req.GetCurrency(),
		IsPrimary:   req.GetIsPrimary(),
		// The public create flow is the bank-feed connection flow. A false
		// is_primary value must not disable feed eligibility.
		IsConnected: true,
	}
	if validationErr := domain.ValidateBankAccountInput(input); validationErr != nil {
		return nil, bankAccountStatusError(http.StatusBadRequest, validationErr.Error())
	}
	id, createErr := a.Service.Create(ctx, businessID, input)
	if createErr != nil {
		return nil, bankAccountStatusError(http.StatusInternalServerError, "Failed to create bank account")
	}
	account, getErr := a.Service.GetByID(ctx, businessID, id)
	if getErr != nil {
		return &financebankaccountv1.CreateBankAccountResponse{Account: &financebankaccountv1.BankAccount{
			Id:          id,
			Name:        input.Name,
			BankName:    input.Institution,
			AccountType: input.AccountType,
			LastFour:    input.LastFour,
			Balance:     input.Balance.Float64(),
			Currency:    input.Currency,
			IsPrimary:   input.IsPrimary,
		}}, nil
	}
	return &financebankaccountv1.CreateBankAccountResponse{Account: newBankAccountProto(account)}, nil
}

func (a *BankAccountAPI) GetBankAccount(ctx context.Context, req *financebankaccountv1.GetBankAccountRequest) (*financebankaccountv1.GetBankAccountResponse, error) {
	businessID, err := bankAccountBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	account, getErr := a.Service.GetByID(ctx, businessID, req.GetId())
	if getErr != nil {
		if errors.Is(getErr, domain.ErrBankAccountNotFound) {
			return nil, bankAccountStatusError(http.StatusNotFound, "Bank account not found")
		}
		return nil, bankAccountStatusError(http.StatusInternalServerError, "Failed to get bank account")
	}
	return &financebankaccountv1.GetBankAccountResponse{Account: newBankAccountProto(account)}, nil
}

func newBankAccountProto(account domain.BankAccount) *financebankaccountv1.BankAccount {
	return &financebankaccountv1.BankAccount{
		Id:          account.ID,
		Name:        account.Name,
		BankName:    account.Institution,
		AccountType: account.AccountType,
		LastFour:    account.LastFour,
		Balance:     account.Balance.Float64(),
		Currency:    account.Currency,
		IsPrimary:   account.IsPrimary,
		CreatedAt:   account.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func totalPages(limit, total int32) int32 {
	if limit <= 0 || total <= 0 {
		return 0
	}
	pages := (int64(total) + int64(limit) - 1) / int64(limit)
	if pages > int64(^uint32(0)>>1) {
		return int32(^uint32(0) >> 1)
	}
	return int32(pages)
}

func bankAccountBusinessID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return 0, bankAccountStatusError(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}

func bankAccountStatusError(status int, message string) error {
	return onekiterrors.New(status, message)
}
