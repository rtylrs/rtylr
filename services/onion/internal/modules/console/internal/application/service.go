package application

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/console/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	consoleTTL            = 1 * time.Minute
	freeIncludedSeatLimit = 10
)

var (
	ErrNotOrganizationOwner   = errors.New("access can only be granted to organization owners")
	ErrBusinessNotOwnedByUser = errors.New("business does not belong to the requested owner")
	ErrAddonNotFound          = errors.New("addon not found")
	ErrInvalidSubscription    = errors.New("invalid subscription input")
)

// Console is a global admin surface — caches are keyed on the admin namespace
// rather than per-owner because they aggregate across all tenants.
func consoleKey(detail string) string {
	return "cache:console:" + detail
}

func invalidateConsoleScope(ctx context.Context) {
	cacher.InvalidateScope(ctx, "cache:console:*")
}

func invalidateOwnerSubscriptionScope(ctx context.Context, ownerID int64) {
	patterns := append(cacheKeys.SubscriptionPatterns(ownerID), cacheKeys.UserPrefix(ownerID))
	cacher.InvalidateScope(ctx, patterns...)
}

func invalidateBusinessAddonScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx,
		cacheKeys.ScopedPrefix("addons", businessID),
		cacheKeys.BusinessPrefix(businessID),
	)
}

type Service struct {
	repo *infrastructure.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

// ── Generic Resources ────────────────────────────────────────────────────────

type ResourceTableDTO struct {
	Name    string `json:"name"`
	Comment string `json:"comment"`
}

type ResourceColumnDTO struct {
	Name       string  `json:"name"`
	DataType   string  `json:"data_type"`
	IsNullable bool    `json:"is_nullable"`
	IsPrimary  bool    `json:"is_primary"`
	Default    *string `json:"default,omitempty"`
	Extra      string  `json:"extra"`
	Comment    string  `json:"comment"`
	Writable   bool    `json:"writable"`
}

type ListResourceTablesResult struct {
	Tables []ResourceTableDTO `json:"tables"`
}

type DescribeResourceResult struct {
	Table      string              `json:"table"`
	PrimaryKey string              `json:"primary_key"`
	Columns    []ResourceColumnDTO `json:"columns"`
}

type ListResourceRowsResult struct {
	Table      string              `json:"table"`
	PrimaryKey string              `json:"primary_key"`
	Columns    []ResourceColumnDTO `json:"columns"`
	Rows       []map[string]any    `json:"rows"`
	Total      int64               `json:"total"`
	Limit      int                 `json:"limit"`
	Offset     int                 `json:"offset"`
}

type GetResourceRowResult struct {
	Table      string              `json:"table"`
	PrimaryKey string              `json:"primary_key"`
	Columns    []ResourceColumnDTO `json:"columns"`
	Row        map[string]any      `json:"row"`
}

func toResourceColumnDTO(c infrastructure.ResourceColumnRow) ResourceColumnDTO {
	dto := ResourceColumnDTO{
		Name:       c.Name,
		DataType:   c.DataType,
		IsNullable: c.IsNullable,
		IsPrimary:  c.IsPrimary,
		Extra:      c.Extra,
		Writable:   !c.IsPrimary && !strings.Contains(strings.ToLower(c.Extra), "auto_increment") && !strings.Contains(strings.ToLower(c.Extra), "generated"),
	}
	if c.Default.Valid {
		dto.Default = &c.Default.String
	}
	if c.Comment.Valid {
		dto.Comment = c.Comment.String
	}
	return dto
}

func toResourceColumnDTOs(columns []infrastructure.ResourceColumnRow) []ResourceColumnDTO {
	dtos := make([]ResourceColumnDTO, 0, len(columns))
	for _, col := range columns {
		dtos = append(dtos, toResourceColumnDTO(col))
	}
	return dtos
}

func (s *Service) ListResources(ctx context.Context) (*ListResourceTablesResult, error) {
	rows, err := s.repo.ListResourceTables(ctx)
	if err != nil {
		return nil, err
	}
	tables := make([]ResourceTableDTO, 0, len(rows))
	for _, row := range rows {
		dto := ResourceTableDTO{Name: row.Name}
		if row.Comment.Valid {
			dto.Comment = row.Comment.String
		}
		tables = append(tables, dto)
	}
	return &ListResourceTablesResult{Tables: tables}, nil
}

func (s *Service) DescribeResource(ctx context.Context, table string) (*DescribeResourceResult, error) {
	columns, primaryKey, err := s.repo.DescribeResourceTable(ctx, table)
	if err != nil {
		return nil, err
	}
	return &DescribeResourceResult{
		Table:      table,
		PrimaryKey: primaryKey,
		Columns:    toResourceColumnDTOs(columns),
	}, nil
}

func (s *Service) ListResourceRows(ctx context.Context, table string, limit, offset int) (*ListResourceRowsResult, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	result, err := s.repo.ListResourceRows(ctx, table, limit, offset)
	if err != nil {
		return nil, err
	}
	return &ListResourceRowsResult{
		Table:      table,
		PrimaryKey: result.PrimaryKey,
		Columns:    toResourceColumnDTOs(result.Columns),
		Rows:       result.Rows,
		Total:      result.Total,
		Limit:      limit,
		Offset:     offset,
	}, nil
}

func (s *Service) GetResourceRow(ctx context.Context, table, id string) (*GetResourceRowResult, error) {
	row, columns, primaryKey, err := s.repo.GetResourceRow(ctx, table, id)
	if err != nil {
		return nil, err
	}
	return &GetResourceRowResult{
		Table:      table,
		PrimaryKey: primaryKey,
		Columns:    toResourceColumnDTOs(columns),
		Row:        row,
	}, nil
}

func (s *Service) CreateResourceRow(ctx context.Context, table string, values map[string]any) (int64, error) {
	defer invalidateConsoleScope(ctx)
	return s.repo.CreateResourceRow(ctx, table, values)
}

func (s *Service) UpdateResourceRow(ctx context.Context, table, id string, values map[string]any) error {
	defer invalidateConsoleScope(ctx)
	return s.repo.UpdateResourceRow(ctx, table, id, values)
}

func (s *Service) DeleteResourceRow(ctx context.Context, table, id string) error {
	defer invalidateConsoleScope(ctx)
	return s.repo.DeleteResourceRow(ctx, table, id)
}

// ── Users ─────────────────────────────────────────────────────────────────────
