package erp

// ProductKPIs summarizes a business's catalogue. AvgMarginPct averages the
// margin of active products with a sale price, rounded to two places.
type ProductKPIs struct {
	TotalSKUs    int64
	ActiveSKUs   int64
	OutOfStock   int64
	AvgMarginPct float64
}

// ProductFacts is what Insights shows next to a product's sales figures.
// CategoryID is zero and CategoryName empty for uncategorised products.
type ProductFacts struct {
	ID             int64
	Name           string
	CategoryID     int64
	CategoryName   string
	CostMinorUnits int64
}

// StockedProduct is an active product and its stock on hand.
type StockedProduct struct {
	ID            int64
	Name          string
	StockQuantity float64
}
