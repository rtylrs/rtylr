package domain

import (
	"errors"
	"time"
)

var (
	ErrCustomerNotFound   = errors.New("customer not found")
	ErrInsufficientPoints = errors.New("insufficient loyalty points")
	ErrSegmentNotFound    = errors.New("segment not found")
	ErrCampaignNotFound   = errors.New("campaign not found")
	ErrTierNotFound       = errors.New("loyalty tier not found")
	ErrTierInUse          = errors.New("loyalty tier has active members")
)

type Customer struct {
	ID            int64
	OwnerID       int64
	BusinessID    int64
	FirstName     string
	LastName      string
	Email         string
	Phone         string
	DateOfBirth   *time.Time
	Gender        string
	Address       string
	City          string
	Country       string
	LoyaltyPoints int
	TotalSpent    float64
	VisitCount    int
	LastVisitAt   *time.Time
	LoyaltyTier   string
	Tags          []string
	Notes         string
	Consent       []byte
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CustomerInput struct {
	OwnerID     int64
	BusinessID  int64
	FirstName   string
	LastName    string
	Email       string
	Phone       string
	DateOfBirth *time.Time
	Gender      string
	Address     string
	City        string
	Country     string
	Tags        []string
	Notes       string
	Consent     []byte
}

type CustomerUpdateInput struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	FirstName   string
	LastName    string
	Email       string
	Phone       string
	DateOfBirth *time.Time
	Gender      string
	Address     string
	City        string
	Country     string
	LoyaltyTier string
	Tags        []string
	Notes       string
	Consent     []byte
	IsActive    *bool
}

type CustomerListFilter struct {
	OwnerID    int64
	BusinessID int64
	Search     string
	Tier       string
	Tag        string
	Page       int
	PageSize   int
}

type CustomerListResult struct {
	Customers []Customer
	Total     int
	Page      int
	PageSize  int
}

type LoyaltyTransaction struct {
	ID            int64
	OwnerID       int64
	BusinessID    int64
	CustomerID    int64
	Type          string
	Points        int
	BalanceBefore int
	BalanceAfter  int
	Description   string
	ReferenceID   *int64
	ReferenceType string
	CreatedAt     time.Time
}

type LoyaltyConfig struct {
	ID                    int64
	OwnerID               int64
	BusinessID            int64
	IsEnabled             bool
	PointsPerCurrencyUnit float64
	CurrencyPerPoint      float64
	BronzeMin             int
	SilverMin             int
	GoldMin               int
	PlatinumMin           int
	BronzeDiscountPct     float64
	SilverDiscountPct     float64
	GoldDiscountPct       float64
	PlatinumDiscountPct   float64
	PointsExpiryDays      int
	UpdatedAt             time.Time
}

type LoyaltyConfigInput struct {
	OwnerID               int64
	BusinessID            int64
	IsEnabled             bool
	PointsPerCurrencyUnit float64
	CurrencyPerPoint      float64
	BronzeMin             int
	SilverMin             int
	GoldMin               int
	PlatinumMin           int
	BronzeDiscountPct     float64
	SilverDiscountPct     float64
	GoldDiscountPct       float64
	PlatinumDiscountPct   float64
	PointsExpiryDays      int
}

type AwardPointsInput struct {
	OwnerID       int64
	BusinessID    int64
	CustomerID    int64
	Points        int
	Description   string
	ReferenceID   *int64
	ReferenceType string
}

type RedeemPointsInput struct {
	OwnerID       int64
	BusinessID    int64
	CustomerID    int64
	Points        int
	Description   string
	ReferenceID   *int64
	ReferenceType string
}

type RecordVisitInput struct {
	OwnerID     int64
	BusinessID  int64
	CustomerID  int64
	Amount      float64
	AwardPoints bool
}

type Segment struct {
	ID               int64
	OwnerID          int64
	BusinessID       int64
	Name             string
	Description      string
	Rules            []byte
	CustomerCount    int
	LastCalculatedAt *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type SegmentInput struct {
	OwnerID     int64
	BusinessID  int64
	Name        string
	Description string
	Rules       []byte
}

type SegmentUpdateInput struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	Name        string
	Description string
	Rules       []byte
}

type Campaign struct {
	ID             int64
	OwnerID        int64
	BusinessID     int64
	Name           string
	Description    string
	Channel        string
	Status         string
	SegmentID      *int64
	Subject        string
	Content        string
	ScheduledAt    *time.Time
	SentAt         *time.Time
	AudienceCount  int
	SentCount      int
	DeliveredCount int
	OpenedCount    int
	ClickedCount   int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CampaignInput struct {
	OwnerID     int64
	BusinessID  int64
	Name        string
	Description string
	Channel     string
	SegmentID   *int64
	Subject     string
	Content     string
	ScheduledAt *time.Time
}

type CampaignUpdateInput struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	Name        string
	Description string
	Channel     string
	SegmentID   *int64
	Subject     string
	Content     string
	ScheduledAt *time.Time
}

type DashboardStats struct {
	TotalCustomers       int
	NewThisMonth         int
	ActiveLoyaltyMembers int
	TopCustomers         []TopCustomer
}

type TopCustomer struct {
	ID            int64
	FirstName     string
	LastName      string
	Email         string
	TotalSpent    float64
	LoyaltyPoints int
	LoyaltyTier   string
}

const (
	TierBronze   = "bronze"
	TierSilver   = "silver"
	TierGold     = "gold"
	TierPlatinum = "platinum"
)

// Timeline

type TimelineEntry struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	CustomerID  int64
	Type        string
	Description string
	Metadata    []byte
	CreatedAt   time.Time
}

// CustomerDetail includes loyalty transactions and visit info

type CustomerDetail struct {
	Customer            Customer
	LoyaltyTransactions []LoyaltyTransaction
	RecentVisits        []TimelineEntry
}

// Import

type CustomerImportItem struct {
	FirstName   string
	LastName    string
	Email       string
	Phone       string
	DateOfBirth string
	Gender      string
	Address     string
	City        string
	Country     string
	Tags        []string
	Notes       string
}

type CustomerImportResult struct {
	TotalProcessed int
	Imported       int
	Failed         int
}

// Export (GDPR)

type CustomerExport struct {
	Customer     Customer
	Transactions []LoyaltyTransaction
	Timeline     []TimelineEntry
}

type TimelinePage struct {
	Entries    []TimelineEntry
	NextCursor string
}

// Merge

type MergeCustomersInput struct {
	OwnerID     int64
	BusinessID  int64
	PrimaryID   int64
	SecondaryID int64
}

// Loyalty Tiers

type LoyaltyTier struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	Name        string
	MinPoints   int
	DiscountPct float64
	Perks       []byte
	SortOrder   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type LoyaltyTierInput struct {
	OwnerID     int64
	BusinessID  int64
	Name        string
	MinPoints   int
	DiscountPct float64
	Perks       []byte
	SortOrder   int
}

type LoyaltyTierUpdateInput struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	Name        string
	MinPoints   int
	DiscountPct float64
	Perks       []byte
	SortOrder   int
}

// Campaign Stats

type CampaignStats struct {
	CampaignID     int64
	AudienceCount  int
	SentCount      int
	DeliveredCount int
	OpenedCount    int
	ClickedCount   int
	BouncedCount   int
	FailedCount    int
}

// Campaign Detail

type CampaignDetail struct {
	Campaign Campaign
	Stats    CampaignStats
}

// Segment Detail

type SegmentDetail struct {
	Segment   Segment
	Customers []Customer
}
