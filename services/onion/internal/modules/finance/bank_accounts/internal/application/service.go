package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/finance/bank_accounts/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/money"
)

const (
	bankAccountsScope = "bank_accounts"
	bankAccountsTTL   = 5 * time.Minute

	defaultBankAccountsPageSize int32 = 100
	maxBankAccountsPageSize     int32 = 500
)

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.BankAccount, error)
	Create(ctx context.Context, businessID int64, input domain.BankAccountInput) (int64, error)
	TotalBalance(ctx context.Context, businessID int64) (money.Amount, error)
	CountConnected(ctx context.Context, businessID int64) (int64, error)
}

type pagedRepository interface {
	ListPage(ctx context.Context, businessID int64, page, limit int32, sortBy, sortDir string) ([]domain.BankAccount, int64, error)
}

type directRepository interface {
	GetByID(ctx context.Context, businessID, id int64) (domain.BankAccount, error)
}

type ListOptions struct {
	Page    int32
	Limit   int32
	SortBy  string
	SortDir string
}

type BankAccountPage struct {
	Accounts []domain.BankAccount
	Page     int32
	Limit    int32
	Total    int32
	HasNext  bool
	HasPrev  bool
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func invalidateBankAccountsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(bankAccountsScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.BankAccount, error) {
	return cacher.Of[[]domain.BankAccount]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(bankAccountsScope, businessID, "list")+"|domain=v2",
		func(ctx context.Context) ([]domain.BankAccount, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, bankAccountsTTL),
	)
}

func (s *Service) ListPage(ctx context.Context, businessID int64, options ListOptions) (BankAccountPage, error) {
	page, limit, sortBy, sortDir := normalizeListOptions(options)
	key := cacheKeys.Scoped(bankAccountsScope, businessID, fmt.Sprintf("page=%d&limit=%d&sort=%s&dir=%s", page, limit, sortBy, sortDir))
	return cacher.Of[BankAccountPage]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) (BankAccountPage, error) {
			var (
				accounts []domain.BankAccount
				total    int64
				err      error
			)
			if paged, ok := s.repo.(pagedRepository); ok {
				accounts, total, err = paged.ListPage(ctx, businessID, page, limit, sortBy, sortDir)
			} else {
				accounts, err = s.repo.List(ctx, businessID)
				total = int64(len(accounts))
				if int64(len(accounts)) > int64(limit) {
					accounts = accounts[:limit]
				}
			}
			if err != nil {
				return BankAccountPage{}, err
			}
			totalPage := saturatingInt32(total)
			return BankAccountPage{
				Accounts: accounts,
				Page:     page,
				Limit:    limit,
				Total:    totalPage,
				HasNext:  int64(page)*int64(limit) < total,
				HasPrev:  page > 1,
			}, nil
		},
		cacher.WithExpiry(ctx, bankAccountsTTL),
	)
}

func normalizeListOptions(options ListOptions) (int32, int32, string, string) {
	page := options.Page
	if page < 1 {
		page = 1
	}
	limit := options.Limit
	if limit < 1 {
		limit = defaultBankAccountsPageSize
	}
	if limit > maxBankAccountsPageSize {
		limit = maxBankAccountsPageSize
	}
	sortBy := strings.ToLower(strings.TrimSpace(options.SortBy))
	if sortBy == "" {
		sortBy = "name"
	}
	sortDir := strings.ToLower(strings.TrimSpace(options.SortDir))
	if sortDir != "desc" {
		sortDir = "asc"
	}
	return page, limit, sortBy, sortDir
}

func saturatingInt32(value int64) int32 {
	if value > int64(^uint32(0)>>1) {
		return int32(^uint32(0) >> 1)
	}
	if value < 0 {
		return 0
	}
	return int32(value)
}

func (s *Service) GetByID(ctx context.Context, businessID, bankAccountID int64) (domain.BankAccount, error) {
	if direct, ok := s.repo.(directRepository); ok {
		return direct.GetByID(ctx, businessID, bankAccountID)
	}
	accounts, err := s.List(ctx, businessID)
	if err != nil {
		return domain.BankAccount{}, err
	}
	for _, account := range accounts {
		if account.ID == bankAccountID {
			return account, nil
		}
	}
	return domain.BankAccount{}, domain.ErrBankAccountNotFound
}

func (s *Service) Create(ctx context.Context, businessID int64, input domain.BankAccountInput) (int64, error) {
	id, err := s.repo.Create(ctx, businessID, input)
	if err != nil {
		return 0, err
	}
	invalidateBankAccountsScope(ctx, businessID)
	return id, nil
}

// TotalBalance returns the aggregate balance across all bank accounts for the
// owner. It queries the database directly (no cache) because balance data is
// time-sensitive and should never be stale.
func (s *Service) TotalBalance(ctx context.Context, businessID int64) (money.Amount, error) {
	return s.repo.TotalBalance(ctx, businessID)
}

// ConnectedCount returns how many accounts have is_connected=TRUE for the
// owner. Useful for showing integration status in the dashboard.
func (s *Service) ConnectedCount(ctx context.Context, businessID int64) (int64, error) {
	return s.repo.CountConnected(ctx, businessID)
}
