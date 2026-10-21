package pos

import "time"

// CustomerPurchases is one customer's paid POS orders over a period.
// CustomerName is a name recorded on those orders.
type CustomerPurchases struct {
	CustomerID      int64
	CustomerName    string
	SpendMinorUnits int64
	Orders          int64
	LastOrderAt     time.Time
}

// TopBuyersQuery ranks customers by spend, or by order count when
// ByOrderCount is set, over the calendar days From to To.
type TopBuyersQuery struct {
	From, To     time.Time
	ByOrderCount bool
	Descending   bool
	Limit        int32
	Offset       int32
}

// SpendBucket counts customers whose lifetime paid spend falls in
// [LowerMinorUnits, UpperMinorUnits). An UpperMinorUnits of zero is unbounded.
type SpendBucket struct {
	Label           string
	LowerMinorUnits int64
	UpperMinorUnits int64
	CustomerCount   int64
}

// ProductSales is the quantity and revenue a product sold.
type ProductSales struct {
	ProductID         int64
	Units             float64
	RevenueMinorUnits int64
}
