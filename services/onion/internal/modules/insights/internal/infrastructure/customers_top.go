package infrastructure

import (
	"context"

	crmcontract "github.com/voxire/onion/internal/contracts/crm"
	poscontract "github.com/voxire/onion/internal/contracts/pos"
	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type TopCustomerRow = domain.TopCustomerRow

type TopCustomerQuery = domain.TopCustomerQuery

type TopCustomerPage = domain.TopCustomerPage

func (r *Repository) GetTopCustomers(ctx context.Context, businessID int64, query TopCustomerQuery) (TopCustomerPage, error) {
	query.From = dateOnly(query.From)
	query.To = dateOnly(query.To)
	if query.Limit <= 0 {
		query.Limit = 10
	}
	if query.Limit > 100 {
		query.Limit = 100
	}
	if query.Offset < 0 {
		query.Offset = 0
	}

	total, err := r.sources.POS.InsightsBuyerCount(ctx, businessID, query.From, query.To)
	if err != nil {
		logger.Ctx(ctx).Error("insights: count top customers failed", logger.Error(err))
		return TopCustomerPage{}, err
	}
	buyers, err := r.sources.POS.InsightsTopBuyers(ctx, businessID, poscontract.TopBuyersQuery{
		From:         query.From,
		To:           query.To,
		ByOrderCount: query.ByOrderCount,
		Descending:   query.Descending,
		Limit:        query.Limit,
		Offset:       query.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: top customers query failed", logger.Error(err))
		return TopCustomerPage{}, err
	}
	ids := make([]int64, 0, len(buyers))
	for _, buyer := range buyers {
		ids = append(ids, buyer.CustomerID)
	}
	contacts, err := r.sources.CRM.InsightsCustomerContacts(ctx, businessID, ids)
	if err != nil {
		logger.Ctx(ctx).Error("insights: top customer contacts failed", logger.Error(err))
		return TopCustomerPage{}, err
	}
	byID := make(map[int64]crmcontract.CustomerContact, len(contacts))
	for _, contact := range contacts {
		byID[contact.ID] = contact
	}
	out := make([]TopCustomerRow, 0, len(buyers))
	for _, buyer := range buyers {
		row := TopCustomerRow{
			CustomerID:         buyer.CustomerID,
			LifetimeSpendMinor: buyer.SpendMinorUnits,
			OrderCount:         buyer.Orders,
			LastOrderAt:        buyer.LastOrderAt,
		}
		// A customer deleted from CRM keeps the name recorded on the order.
		if contact, ok := byID[buyer.CustomerID]; ok {
			row.FirstName, row.LastName, row.Email = contact.FirstName, contact.LastName, contact.Email
		} else {
			row.FirstName = buyer.CustomerName
		}
		out = append(out, row)
	}
	return TopCustomerPage{Rows: out, Total: total}, nil
}
