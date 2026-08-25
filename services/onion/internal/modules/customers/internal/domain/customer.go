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
	ErrTierInUse           = errors.New("loyalty tier has active members")
)

type Customer struct {
	ID            int64      `json:"id"`
	OwnerID       int64      `json:"owner_id"`
	BusinessID    int64      `json:"business_id"`
	FirstName     string     `json:"first_name"`
	LastName      string     `json:"last_name"`
	Email         string     `json:"email"`
	Phone         string     `json:"phone"`
	DateOfBirth   *time.Time `json:"date_of_birth"`
	Gender        string     `json:"gender"`
	Address       string     `json:"address"`
	City          string     `json:"city"`
	Country       string     `json:"country"`
	LoyaltyPoints int        `json:"loyalty_points"`
	TotalSpent    float64    `json:"total_spent"`
	VisitCount    int        `json:"visit_count"`
	LastVisitAt   *time.Time `json:"last_visit_at"`
	LoyaltyTier   string     `json:"loyalty_tier"`
	Tags          []string   `json:"tags"`
	Notes         string     `json:"notes"`
	Consent       []byte     `json:"consent"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
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
	Customers []Customer `json:"customers"`
	Total     int        `json:"total"`
	Page      int        `json:"page"`
	PageSize  int        `json:"page_size"`
}

type LoyaltyTransaction struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"owner_id"`
	BusinessID    int64     `json:"business_id"`
	CustomerID    int64     `json:"customer_id"`
	Type          string    `json:"type"`
	Points        int       `json:"points"`
	BalanceBefore int       `json:"balance_before"`
	BalanceAfter  int       `json:"balance_after"`
	Description   string    `json:"description"`
	ReferenceID   *int64    `json:"reference_id"`
	ReferenceType string    `json:"reference_type"`
	CreatedAt     time.Time `json:"created_at"`
}

type LoyaltyConfig struct {
	ID                    int64     `json:"id"`
	OwnerID               int64     `json:"owner_id"`
	BusinessID            int64     `json:"business_id"`
	IsEnabled             bool      `json:"is_enabled"`
	PointsPerCurrencyUnit float64   `json:"points_per_currency_unit"`
	CurrencyPerPoint      float64   `json:"currency_per_point"`
	BronzeMin             int       `json:"bronze_min"`
	SilverMin             int       `json:"silver_min"`
	GoldMin               int       `json:"gold_min"`
	PlatinumMin           int       `json:"platinum_min"`
	BronzeDiscountPct     float64   `json:"bronze_discount_pct"`
	SilverDiscountPct     float64   `json:"silver_discount_pct"`
	GoldDiscountPct       float64   `json:"gold_discount_pct"`
	PlatinumDiscountPct   float64   `json:"platinum_discount_pct"`
	PointsExpiryDays      int       `json:"points_expiry_days"`
	UpdatedAt             time.Time `json:"updated_at"`
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
	ID               int64      `json:"id"`
	OwnerID          int64      `json:"owner_id"`
	BusinessID       int64      `json:"business_id"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	Rules            []byte     `json:"rules"`
	CustomerCount    int        `json:"customer_count"`
	LastCalculatedAt *time.Time `json:"last_calculated_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
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
	ID             int64      `json:"id"`
	OwnerID        int64      `json:"owner_id"`
	BusinessID     int64      `json:"business_id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	Channel        string     `json:"channel"`
	Status         string     `json:"status"`
	SegmentID      *int64     `json:"segment_id"`
	Subject        string     `json:"subject"`
	Content        string     `json:"content"`
	ScheduledAt    *time.Time `json:"scheduled_at"`
	SentAt         *time.Time `json:"sent_at"`
	AudienceCount  int        `json:"audience_count"`
	SentCount      int        `json:"sent_count"`
	DeliveredCount int        `json:"delivered_count"`
	OpenedCount    int        `json:"opened_count"`
	ClickedCount   int        `json:"clicked_count"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
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
	TotalCustomers       int           `json:"total_customers"`
	NewThisMonth         int           `json:"new_this_month"`
	ActiveLoyaltyMembers int           `json:"active_loyalty_members"`
	TopCustomers         []TopCustomer `json:"top_customers"`
}

type TopCustomer struct {
	ID            int64   `json:"id"`
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	Email         string  `json:"email"`
	TotalSpent    float64 `json:"total_spent"`
	LoyaltyPoints int     `json:"loyalty_points"`
	LoyaltyTier   string  `json:"loyalty_tier"`
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
	OwnerID     int64     `json:"owner_id"`
	BusinessID  int64     `json:"business_id"`
	CustomerID  int64     `json:"customer_id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Metadata    []byte    `json:"metadata"`
	CreatedAt   time.Time `json:"created_at"`
}

// CustomerDetail includes loyalty transactions and visit info

type CustomerDetail struct {
	Customer            Customer             `json:"customer"`
	LoyaltyTransactions []LoyaltyTransaction `json:"loyalty_transactions"`
	RecentVisits        []TimelineEntry      `json:"recent_visits"`
}

// Import

type CustomerImportItem struct {
	FirstName   string   `json:"first_name"`
	LastName    string   `json:"last_name"`
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	DateOfBirth string   `json:"date_of_birth"`
	Gender      string   `json:"gender"`
	Address     string   `json:"address"`
	City        string   `json:"city"`
	Country     string   `json:"country"`
	Tags        []string `json:"tags"`
	Notes       string   `json:"notes"`
}

type CustomerImportResult struct {
	TotalProcessed int `json:"total_processed"`
	Imported       int `json:"imported"`
	Failed         int `json:"failed"`
}

// Export (GDPR)

type CustomerExport struct {
	Customer     Customer             `json:"customer"`
	Transactions []LoyaltyTransaction `json:"transactions"`
	Timeline     []TimelineEntry      `json:"timeline"`
}

type TimelinePage struct {
	Entries    []TimelineEntry `json:"entries"`
	NextCursor string          `json:"next_cursor,omitempty"`
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
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	BusinessID  int64     `json:"business_id"`
	Name        string    `json:"name"`
	MinPoints   int       `json:"min_points"`
	DiscountPct float64   `json:"discount_pct"`
	Perks       []byte    `json:"perks"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
	CampaignID     int64 `json:"campaign_id"`
	AudienceCount  int   `json:"audience_count"`
	SentCount      int   `json:"sent_count"`
	DeliveredCount int   `json:"delivered_count"`
	OpenedCount    int   `json:"opened_count"`
	ClickedCount   int   `json:"clicked_count"`
	BouncedCount   int   `json:"bounced_count"`
	FailedCount    int   `json:"failed_count"`
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
