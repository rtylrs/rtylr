package application

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/finance/bills/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/pagination"
)

const (
	billsScope = "bills"
	billsTTL   = 5 * time.Minute

	defaultListLimit int32 = 50
	maxListLimit     int32 = 500
)

var billsListPolicy = pagination.Policy{
	DefaultPageSize: int64(defaultListLimit),
	MaxPageSize:     int64(maxListLimit),
	MaxOffset:       1<<31 - 1,
}

type BillPage struct {
	Bills    []domain.Bill
	Page     int64
	PageSize int64
	HasMore  bool
}

type repository interface {
	List(ctx context.Context, businessID int64) ([]domain.Bill, error)
	ListPaginated(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Bill, error)
	ListByPurchaseOrder(ctx context.Context, businessID, purchaseOrderID int64) ([]domain.Bill, error)
	CountByPurchaseOrder(ctx context.Context, businessID, purchaseOrderID int64) (int64, error)
	Get(ctx context.Context, businessID, id int64) (domain.Bill, error)
	Delete(ctx context.Context, businessID, id int64) error
	Create(ctx context.Context, businessID int64, input domain.BillInput) (int64, error)
	Update(ctx context.Context, businessID, id int64, input domain.BillInput) error
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func invalidateBillsScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(billsScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID int64) ([]domain.Bill, error) {
	return cacher.Of[[]domain.Bill]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(billsScope, businessID, "list")+"|domain=v2",
		func(ctx context.Context) ([]domain.Bill, error) {
			return s.repo.List(ctx, businessID)
		},
		cacher.WithExpiry(ctx, billsTTL),
	)
}

// ListPaginated returns a single page of bills. A non-positive limit defaults
// to 50 and limit is capped at 500; a negative offset is clamped to zero.
func (s *Service) ListPaginated(ctx context.Context, businessID int64, limit, offset int32) ([]domain.Bill, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListPaginated(ctx, businessID, limit, offset)
}

func (s *Service) ListPage(ctx context.Context, businessID, page, pageSize int64) (BillPage, error) {
	page, pageSize, offset := normalizeListPage(page, pageSize)
	bills, err := s.repo.ListPaginated(ctx, businessID, int32(pageSize+1), offset) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	if err != nil {
		return BillPage{}, err
	}

	hasMore := int64(len(bills)) > pageSize
	if hasMore {
		bills = bills[:int(pageSize)]
	}
	return BillPage{Bills: bills, Page: page, PageSize: pageSize, HasMore: hasMore}, nil
}

// ListFilteredPage applies the optional supplier and purchase-order filters
// before applying the requested page. The legacy repository exposes a
// business-wide paginated query, so scan it in bounded chunks rather than
// filtering the first page (or the legacy 500-row preview) in memory.
func (s *Service) ListFilteredPage(ctx context.Context, businessID int64, supplierID, purchaseOrderID *int64, page, pageSize int64) (BillPage, error) {
	if supplierID == nil && purchaseOrderID == nil {
		return s.ListPage(ctx, businessID, page, pageSize)
	}

	page, pageSize, offset := normalizeListPage(page, pageSize)
	filtered := make([]domain.Bill, 0)
	var scanOffset int32
	for {
		rows, err := s.repo.ListPaginated(ctx, businessID, maxListLimit, scanOffset)
		if err != nil {
			return BillPage{}, err
		}
		for _, bill := range rows {
			if supplierID != nil && bill.SupplierID != *supplierID {
				continue
			}
			if purchaseOrderID != nil && (bill.PurchaseOrderID == nil || *bill.PurchaseOrderID != *purchaseOrderID) {
				continue
			}
			filtered = append(filtered, bill)
		}

		if len(rows) < int(maxListLimit) {
			break
		}
		// Do not wrap the scan cursor if a database contains an unexpectedly
		// large number of rows.
		if scanOffset > int32(math.MaxInt32)-maxListLimit {
			break
		}
		scanOffset += maxListLimit
	}

	total := int64(len(filtered))
	start := int64(offset)
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	selected := filtered[int(start):int(end)]
	return BillPage{
		Bills:    selected,
		Page:     page,
		PageSize: pageSize,
		HasMore:  end < total,
	}, nil
}

func normalizeListPage(page, pageSize int64) (int64, int64, int32) {
	normalized := billsListPolicy.Normalize(page, pageSize)
	return normalized.Number, normalized.Size, int32(normalized.Offset) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
}

func (s *Service) GetBill(ctx context.Context, businessID, id int64) (domain.Bill, error) {
	key := cacheKeys.Scoped(billsScope, businessID, "id="+strconv.FormatInt(id, 10))
	return cacher.Of[domain.Bill]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) (domain.Bill, error) {
			return s.repo.Get(ctx, businessID, id)
		},
		cacher.WithExpiry(ctx, billsTTL),
	)
}

func (s *Service) DeleteBill(ctx context.Context, businessID, id int64) error {
	if err := s.repo.Delete(ctx, businessID, id); err != nil {
		return err
	}
	invalidateBillsScope(ctx, businessID)
	return nil
}

func (s *Service) Create(ctx context.Context, businessID int64, input domain.BillInput) (int64, error) {
	id, err := s.repo.Create(ctx, businessID, input)
	if err != nil {
		return 0, err
	}
	invalidateBillsScope(ctx, businessID)
	return id, nil
}

func (s *Service) Update(ctx context.Context, businessID, id int64, input domain.BillInput) error {
	if err := s.repo.Update(ctx, businessID, id, input); err != nil {
		return err
	}
	invalidateBillsScope(ctx, businessID)
	return nil
}

// ListByPurchaseOrder returns every bill settling one ERP purchase order
// (FIN->ERP-02's read surface). Served off the (business_id, purchase_order_id)
// index added with the column.
func (s *Service) ListByPurchaseOrder(ctx context.Context, businessID, purchaseOrderID int64) ([]domain.Bill, error) {
	return s.repo.ListByPurchaseOrder(ctx, businessID, purchaseOrderID)
}
