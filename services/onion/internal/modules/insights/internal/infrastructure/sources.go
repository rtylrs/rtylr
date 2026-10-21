package infrastructure

import (
	"context"
	"time"

	crmcontract "github.com/voxire/onion/internal/contracts/crm"
	erpcontract "github.com/voxire/onion/internal/contracts/erp"
	financecontract "github.com/voxire/onion/internal/contracts/finance"
	hrcontract "github.com/voxire/onion/internal/contracts/hr"
	poscontract "github.com/voxire/onion/internal/contracts/pos"
)

// Sources are the owning modules Insights reports are built from. Each module
// reads its own tables; Insights only combines their results. Every source
// also reports its latest change so cached reports can be versioned.
type Sources struct {
	Finance FinanceSource
	HR      HRSource
	ERP     ERPSource
	POS     POSSource
	CRM     CRMSource
}

type lastChange interface {
	InsightsLastChange(context.Context, int64) (time.Time, error)
}

type FinanceSource interface {
	lastChange
	InsightsPaidInvoicesByDay(context.Context, int64, time.Time, time.Time) (map[string]int64, error)
	InsightsExpensesByDay(context.Context, int64, time.Time, time.Time) (map[string]int64, error)
	InsightsExpensesByCategory(context.Context, int64, time.Time, time.Time) ([]financecontract.ExpenseCategoryTotal, error)
	InsightsInvoiceAging(context.Context, int64) ([]financecontract.InvoiceAgingBucket, error)
	InsightsInvoiceRevenue(context.Context, int64, time.Time, time.Time) (int64, error)
}

type HRSource interface {
	lastChange
	InsightsActiveEmployees(context.Context, int64) (int64, error)
	InsightsShiftsThisWeek(context.Context, int64) (int64, error)
	InsightsPendingLeave(context.Context, int64) (int64, error)
	InsightsOpenLeave(context.Context, int64) (int64, error)
	InsightsHeadcountAsOf(context.Context, int64, time.Time) (int64, error)
	InsightsHires(context.Context, int64, time.Time, time.Time) (int64, error)
	InsightsDepartures(context.Context, int64, time.Time, time.Time) (int64, error)
	InsightsShiftMetrics(context.Context, int64, time.Time, time.Time) (hrcontract.ShiftMetrics, error)
	InsightsLeaveStarting(context.Context, int64, string, time.Time, time.Time) (int64, error)
	InsightsScheduledHours(context.Context, int64, time.Time, time.Time) (float64, error)
}

type ERPSource interface {
	lastChange
	InsightsProductKPIs(context.Context, int64) (erpcontract.ProductKPIs, error)
	InsightsMarginDistribution(context.Context, int64) ([]int64, error)
	InsightsProductFacts(context.Context, int64, []int64) ([]erpcontract.ProductFacts, error)
	InsightsActiveProducts(context.Context, int64) ([]erpcontract.StockedProduct, error)
}

type POSSource interface {
	lastChange
	InsightsNetRevenue(context.Context, int64, time.Time, time.Time) (int64, error)
	InsightsBuyerCount(context.Context, int64, time.Time, time.Time) (int64, error)
	InsightsLapsedBuyerCount(context.Context, int64, time.Time, time.Time) (int64, error)
	InsightsTopBuyers(context.Context, int64, poscontract.TopBuyersQuery) ([]poscontract.CustomerPurchases, error)
	InsightsBuyerSpendBuckets(context.Context, int64) ([]poscontract.SpendBucket, error)
	InsightsPaidProductSales(context.Context, int64, time.Time, time.Time) ([]poscontract.ProductSales, error)
	InsightsLifetimeProductSales(context.Context, int64) ([]poscontract.ProductSales, error)
	InsightsProductsSoldSince(context.Context, int64, time.Time) ([]int64, error)
}

type CRMSource interface {
	lastChange
	InsightsCustomerCount(context.Context, int64) (int64, error)
	InsightsCustomersAsOf(context.Context, int64, time.Time) (int64, error)
	InsightsNewCustomers(context.Context, int64, time.Time, time.Time) (int64, error)
	InsightsCustomerContacts(context.Context, int64, []int64) ([]crmcontract.CustomerContact, error)
}

// productFacts looks up the catalogue entries for the products in sales,
// keyed by product ID. Products ERP no longer has are absent.
func (r *Repository) productFacts(ctx context.Context, businessID int64, sales []poscontract.ProductSales) (map[int64]erpcontract.ProductFacts, error) {
	ids := make([]int64, 0, len(sales))
	for _, sale := range sales {
		ids = append(ids, sale.ProductID)
	}
	facts, err := r.sources.ERP.InsightsProductFacts(ctx, businessID, ids)
	if err != nil {
		return nil, err
	}
	out := make(map[int64]erpcontract.ProductFacts, len(facts))
	for _, fact := range facts {
		out[fact.ID] = fact
	}
	return out, nil
}
