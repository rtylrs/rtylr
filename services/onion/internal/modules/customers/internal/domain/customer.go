package domain

import (
	"errors"
	"time"
)

var (
	ErrCustomerNotFound    = errors.New("customer not found")
	ErrInsufficientPoints  = errors.New("insufficient loyalty points")
	ErrLoyaltyConfigExists = errors.New("loyalty config already exists")
	ErrSegmentNotFound     = errors.New("segment not found")
	ErrCampaignNotFound    = errors.New("campaign not found")
	ErrTierNotFound        = errors.New("loyalty tier not found")
)

type Customer struct {
	ID            int64      `json:"id"`
	OwnerID       int64      `json:"ownerId"`
	FirstName     string     `json:"firstName"`
	LastName      string     `json:"lastName"`
	Email         string     `json:"email"`
	Phone         string     `json:"phone"`
	DateOfBirth   *time.Time `json:"dateOfBirth"`
	Gender        string     `json:"gender"`
	Address       string     `json:"address"`
	City          string     `json:"city"`
	Country       string     `json:"country"`
	LoyaltyPoints int        `json:"loyaltyPoints"`
	TotalSpent    float64    `json:"totalSpent"`
	VisitCount    int        `json:"visitCount"`
	LastVisitAt   *time.Time `json:"lastVisitAt"`
	LoyaltyTier   string     `json:"loyaltyTier"`
	Tags          []string   `json:"tags"`
	Notes         string     `json:"notes"`
	IsActive      bool       `json:"isActive"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

type CustomerInput struct {
	OwnerID     int64
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
}

type CustomerUpdateInput struct {
	ID          int64
	OwnerID     int64
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
	IsActive    *bool
}

type CustomerListFilter struct {
	OwnerID  int64
	Search   string
	Tier     string
	Tag      string
	Page     int
	PageSize int
}

type CustomerListResult struct {
	Customers []Customer `json:"customers"`
	Total     int        `json:"total"`
	Page      int        `json:"page"`
	PageSize  int        `json:"pageSize"`
}

type LoyaltyTransaction struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"ownerId"`
	CustomerID    int64     `json:"customerId"`
	Type          string    `json:"type"`
	Points        int       `json:"points"`
	BalanceBefore int       `json:"balanceBefore"`
	BalanceAfter  int       `json:"balanceAfter"`
	Description   string    `json:"description"`
	ReferenceID   *int64    `json:"referenceId"`
	ReferenceType string    `json:"referenceType"`
	CreatedAt     time.Time `json:"createdAt"`
}

type LoyaltyConfig struct {
	ID                    int64     `json:"id"`
	OwnerID               int64     `json:"ownerId"`
	IsEnabled             bool      `json:"isEnabled"`
	PointsPerCurrencyUnit float64   `json:"pointsPerCurrencyUnit"`
	CurrencyPerPoint      float64   `json:"currencyPerPoint"`
	BronzeMin             int       `json:"bronzeMin"`
	SilverMin             int       `json:"silverMin"`
	GoldMin               int       `json:"goldMin"`
	PlatinumMin           int       `json:"platinumMin"`
	BronzeDiscountPct     float64   `json:"bronzeDiscountPct"`
	SilverDiscountPct     float64   `json:"silverDiscountPct"`
	GoldDiscountPct       float64   `json:"goldDiscountPct"`
	PlatinumDiscountPct   float64   `json:"platinumDiscountPct"`
	PointsExpiryDays      int       `json:"pointsExpiryDays"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

type LoyaltyConfigInput struct {
	OwnerID               int64
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
	OwnerID     int64
	CustomerID  int64
	Points      int
	Description string
	ReferenceID *int64
	ReferenceType string
}

type RedeemPointsInput struct {
	OwnerID     int64
	CustomerID  int64
	Points      int
	Description string
	ReferenceID *int64
	ReferenceType string
}

type RecordVisitInput struct {
	OwnerID    int64
	CustomerID int64
	Amount     float64
	AwardPoints bool
}

type Segment struct {
	ID               int64      `json:"id"`
	OwnerID          int64      `json:"ownerId"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	Rules            []byte     `json:"rules"`
	CustomerCount    int        `json:"customerCount"`
	LastCalculatedAt *time.Time `json:"lastCalculatedAt"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

type SegmentInput struct {
	OwnerID     int64
	Name        string
	Description string
	Rules       []byte
}

type SegmentUpdateInput struct {
	ID          int64
	OwnerID     int64
	Name        string
	Description string
	Rules       []byte
}

type Campaign struct {
	ID             int64      `json:"id"`
	OwnerID        int64      `json:"ownerId"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	Channel        string     `json:"channel"`
	Status         string     `json:"status"`
	SegmentID      *int64     `json:"segmentId"`
	Subject        string     `json:"subject"`
	Content        string     `json:"content"`
	ScheduledAt    *time.Time `json:"scheduledAt"`
	SentAt         *time.Time `json:"sentAt"`
	AudienceCount  int        `json:"audienceCount"`
	SentCount      int        `json:"sentCount"`
	DeliveredCount int        `json:"deliveredCount"`
	OpenedCount    int        `json:"openedCount"`
	ClickedCount   int        `json:"clickedCount"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type CampaignInput struct {
	OwnerID     int64
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
	Name        string
	Description string
	Channel     string
	SegmentID   *int64
	Subject     string
	Content     string
	ScheduledAt *time.Time
}

type DashboardStats struct {
	TotalCustomers       int              `json:"totalCustomers"`
	NewThisMonth         int              `json:"newThisMonth"`
	ActiveLoyaltyMembers int              `json:"activeLoyaltyMembers"`
	TopCustomers         []TopCustomer    `json:"topCustomers"`
}

type TopCustomer struct {
	ID            int64   `json:"id"`
	FirstName     string  `json:"firstName"`
	LastName      string  `json:"lastName"`
	Email         string  `json:"email"`
	TotalSpent    float64 `json:"totalSpent"`
	LoyaltyPoints int     `json:"loyaltyPoints"`
	LoyaltyTier   string  `json:"loyaltyTier"`
}

const (
	TierBronze   = "bronze"
	TierSilver   = "silver"
	TierGold     = "gold"
	TierPlatinum = "platinum"
)

// Timeline

type TimelineEntry struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"ownerId"`
	CustomerID  int64     `json:"customerId"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Metadata    []byte    `json:"metadata"`
	CreatedAt   time.Time `json:"createdAt"`
}

// CustomerDetail includes loyalty transactions and visit info

type CustomerDetail struct {
	Customer            Customer             `json:"customer"`
	LoyaltyTransactions []LoyaltyTransaction `json:"loyaltyTransactions"`
	RecentVisits        []TimelineEntry      `json:"recentVisits"`
}

// Import

type CustomerImportItem struct {
	FirstName   string   `json:"firstName"`
	LastName    string   `json:"lastName"`
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	DateOfBirth string   `json:"dateOfBirth"`
	Gender      string   `json:"gender"`
	Address     string   `json:"address"`
	City        string   `json:"city"`
	Country     string   `json:"country"`
	Tags        []string `json:"tags"`
	Notes       string   `json:"notes"`
}

type CustomerImportResult struct {
	TotalProcessed int `json:"totalProcessed"`
	Imported       int `json:"imported"`
	Failed         int `json:"failed"`
}

// Export (GDPR)

type CustomerExport struct {
	Customer     Customer             `json:"customer"`
	Transactions []LoyaltyTransaction `json:"transactions"`
	Timeline     []TimelineEntry      `json:"timeline"`
}

// Merge

type MergeCustomersInput struct {
	OwnerID     int64
	PrimaryID   int64
	SecondaryID int64
}

// Loyalty Tiers

type LoyaltyTier struct {
	ID             int64     `json:"id"`
	OwnerID        int64     `json:"ownerId"`
	Name           string    `json:"name"`
	MinPoints      int       `json:"minPoints"`
	DiscountPct    float64   `json:"discountPct"`
	Perks          []byte    `json:"perks"`
	SortOrder      int       `json:"sortOrder"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type LoyaltyTierInput struct {
	OwnerID     int64
	Name        string
	MinPoints   int
	DiscountPct float64
	Perks       []byte
	SortOrder   int
}

type LoyaltyTierUpdateInput struct {
	ID          int64
	OwnerID     int64
	Name        string
	MinPoints   int
	DiscountPct float64
	Perks       []byte
	SortOrder   int
}

// Campaign Stats

type CampaignStats struct {
	CampaignID     int64 `json:"campaignId"`
	AudienceCount  int   `json:"audienceCount"`
	SentCount      int   `json:"sentCount"`
	DeliveredCount int   `json:"deliveredCount"`
	OpenedCount    int   `json:"openedCount"`
	ClickedCount   int   `json:"clickedCount"`
	BouncedCount   int   `json:"bouncedCount"`
	FailedCount    int   `json:"failedCount"`
}

// Campaign Detail

type CampaignDetail struct {
	Campaign Campaign      `json:"campaign"`
	Stats    CampaignStats `json:"stats"`
}

// Segment Detail

type SegmentDetail struct {
	Segment   Segment    `json:"segment"`
	Customers []Customer `json:"customers"`
}
