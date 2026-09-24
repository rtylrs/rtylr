package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

type TopCustomerPayload struct {
	CustomerID              int64     `json:"customer_id"`
	Name                    string    `json:"name"`
	Email                   string    `json:"email"`
	LifetimeSpendMinorUnits int64     `json:"lifetime_spend_minor_units"`
	OrderCount              int64     `json:"order_count"`
	LastOrderAt             time.Time `json:"last_order_at"`
}

type TopCustomerQuery struct {
	From    time.Time
	To      time.Time
	Page    int32
	Limit   int32
	SortBy  string
	SortDir string
}

type TopCustomerPage struct {
	Rows  []TopCustomerPayload
	Total int64
}

func NormalizeTopCustomerSort(sortBy, sortDir string) (string, string, error) {
	normalizedSort := strings.ToLower(strings.TrimSpace(sortBy))
	switch normalizedSort {
	case "", "total_spent", "totalspent", "total spent":
		normalizedSort = "total_spent"
	case "order_count", "ordercount", "order count":
		normalizedSort = "order_count"
	default:
		return "", "", fmt.Errorf("unsupported sort_by %q", sortBy)
	}

	normalizedDir := strings.ToLower(strings.TrimSpace(sortDir))
	switch normalizedDir {
	case "":
		normalizedDir = "desc"
	case "asc", "desc":
	default:
		return "", "", fmt.Errorf("unsupported sort_dir %q", sortDir)
	}
	return normalizedSort, normalizedDir, nil
}

func (s *Service) GetTopCustomers(ctx context.Context, businessID int64, query TopCustomerQuery) (TopCustomerPage, error) {
	if query.Limit <= 0 {
		query.Limit = 10
	}
	if query.Limit > 100 {
		query.Limit = 100
	}
	query.Page = normalizePage(query.Page)
	sortBy, sortDir, err := NormalizeTopCustomerSort(query.SortBy, query.SortDir)
	if err != nil {
		return TopCustomerPage{}, err
	}
	query.SortBy, query.SortDir = sortBy, sortDir

	key, err := s.insightsCacheKey(ctx, businessID, fmt.Sprintf(
		"customers:top:%s:%s:%d:%d:%s:%s",
		query.From.Format("2006-01-02"), query.To.Format("2006-01-02"),
		query.Page, query.Limit, sortBy, sortDir,
	))
	if err != nil {
		return TopCustomerPage{}, err
	}
	type result struct {
		Rows  []TopCustomerPayload
		Total int64
	}
	out, err := cacher.Of[result]().GetOrComputeSingle(ctx, key, func(ctx context.Context) (result, error) {
		page, rerr := s.repo.GetTopCustomers(ctx, businessID, infrastructure.TopCustomerQuery{
			From:         query.From,
			To:           query.To,
			Limit:        query.Limit,
			Offset:       pageOffset(query.Page, query.Limit),
			ByOrderCount: sortBy == "order_count",
			Descending:   sortDir == "desc",
		})
		if rerr != nil {
			return result{}, rerr
		}
		payloads := make([]TopCustomerPayload, 0, len(page.Rows))
		for _, row := range page.Rows {
			name := strings.TrimSpace(row.FirstName + " " + row.LastName)
			payloads = append(payloads, TopCustomerPayload{
				CustomerID:              row.CustomerID,
				Name:                    name,
				Email:                   row.Email,
				LifetimeSpendMinorUnits: row.LifetimeSpendMinor,
				OrderCount:              row.OrderCount,
				LastOrderAt:             row.LastOrderAt,
			})
		}
		return result{Rows: payloads, Total: page.Total}, nil
	}, cacher.WithExpiry(ctx, 5*time.Minute))
	if err != nil {
		return TopCustomerPage{}, err
	}
	return TopCustomerPage{Rows: out.Rows, Total: out.Total}, nil
}
