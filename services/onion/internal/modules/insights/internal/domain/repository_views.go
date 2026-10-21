package domain

import (
	"errors"
	"time"
)

var (
	ErrSettingsNotFound        = errors.New("insights settings not found")
	ErrScheduledReportNotFound = errors.New("scheduled report not found")
)

type TopCustomerRow struct {
	CustomerID         int64
	FirstName          string
	LastName           string
	Email              string
	LifetimeSpendMinor int64
	OrderCount         int64
	LastOrderAt        time.Time
}

type TopCustomerQuery struct {
	From         time.Time
	To           time.Time
	Limit        int32
	Offset       int32
	ByOrderCount bool
	Descending   bool
}

type TopCustomerPage struct {
	Rows  []TopCustomerRow
	Total int64
}

type InvoiceAgingBucketRow struct {
	Bucket           string
	InvoiceCount     int64
	AmountMinorUnits int64
}

type LTVBucketRow struct {
	Label           string
	LowerMinorUnits int64
	UpperMinorUnits int64
	CustomerCount   int64
}

type HeadcountSnapshot struct {
	ActiveHeadcount    int64
	HeadcountAtStart   int64
	HiredThisPeriod    int64
	DepartedThisPeriod int64
}

type UtilizationRow struct {
	ScheduledHours float64
	AvailableHours float64
}

type ScheduledReportRow struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	TemplateID string
	Cadence    string
	Channel    string
	Recipient  string
	IsActive   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type ScheduledReportWrite struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	TemplateID string
	Cadence    string
	Channel    string
	Recipient  string
	IsActive   *bool
}

type ProductKPIs struct {
	TotalSKUs    int64
	ActiveSKUs   int64
	OutOfStock   int64
	AvgMarginPct float64
}

type SlowMoverRow struct {
	ProductID        int64
	Name             string
	StockQuantity    float64
	LastSoldAtExists bool
}

type ShiftMetricsRow struct {
	Scheduled int64
	Completed int64
	NoShows   int64
}

type RevenueByAppRow struct {
	App              string
	AmountMinorUnits int64
}

type ChurnRateRow struct {
	WindowDays   int32
	ChurnedCount int64
	ActiveCount  int64
	ChurnPct     float64
}

type TimeOffSummaryRow struct {
	Pending  int64
	Approved int64
}

type SettingsRow struct {
	OwnerID          int64
	BusinessID       int64
	DefaultCurrency  string
	FiscalYearStart  string
	Timezone         string
	DefaultPeriod    string
	ComparePrevious  bool
	DashboardWidgets []string
	FavoriteReports  []string
}

type MarginBucketRow struct {
	Label        string
	LowerPct     float64
	UpperPct     float64
	ProductCount int64
}

type CashFlowDayRow struct {
	Date              string
	InflowMinorUnits  int64
	OutflowMinorUnits int64
}

type OperationsKPIs struct {
	ActiveStaff       int64
	ShiftsThisWeek    int64
	PendingTimeOff    int64
	OpenLeaveRequests int64
}

type BestSellerRow struct {
	ProductID         int64
	Name              string
	UnitsSold         float64
	RevenueMinorUnits int64
}

type CategorySalesRow struct {
	CategoryID        int64
	CategoryName      string
	UnitsSold         float64
	RevenueMinorUnits int64
}

type ExpenseByCategoryRow struct {
	CategoryID       int64
	CategoryName     string
	AmountMinorUnits int64
}

type CustomerKPIs struct {
	Total         int64
	NewThisPeriod int64
	Active        int64
	Churned       int64
}
