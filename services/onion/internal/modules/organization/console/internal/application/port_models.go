package application

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type NullString struct {
	String string
	Valid  bool
}

func (v *NullString) Scan(src any) error {
	if src == nil {
		*v = NullString{}
		return nil
	}
	switch value := src.(type) {
	case string:
		v.String, v.Valid = value, true
	case []byte:
		v.String, v.Valid = string(value), true
	default:
		return fmt.Errorf("scan optional string from %T", src)
	}
	return nil
}

type NullInt64 struct {
	Int64 int64
	Valid bool
}

func (v *NullInt64) Scan(src any) error {
	if src == nil {
		*v = NullInt64{}
		return nil
	}
	var parsed int64
	switch value := src.(type) {
	case int64:
		parsed = value
	case []byte:
		var err error
		parsed, err = strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			return err
		}
	case string:
		var err error
		parsed, err = strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("scan optional integer from %T", src)
	}
	v.Int64, v.Valid = parsed, true
	return nil
}

type NullBool struct {
	Bool  bool
	Valid bool
}

func (v *NullBool) Scan(src any) error {
	if src == nil {
		*v = NullBool{}
		return nil
	}
	var parsed bool
	switch value := src.(type) {
	case bool:
		parsed = value
	case int64:
		parsed = value != 0
	case []byte:
		var err error
		parsed, err = strconv.ParseBool(string(value))
		if err != nil {
			return err
		}
	case string:
		var err error
		parsed, err = strconv.ParseBool(value)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("scan optional boolean from %T", src)
	}
	v.Bool, v.Valid = parsed, true
	return nil
}

type NullTime struct {
	Time  time.Time
	Valid bool
}

func (v *NullTime) Scan(src any) error {
	if src == nil {
		*v = NullTime{}
		return nil
	}
	value, ok := src.(time.Time)
	if !ok {
		return errors.New("scan optional time requires time.Time")
	}
	v.Time, v.Valid = value, true
	return nil
}

type UserRow struct {
	ID             int64
	Email          string
	FullName       NullString
	BusinessID     NullInt64
	IsConsoleAdmin bool
	IsBanned       bool
	BanReason      NullString
	BannedAt       NullTime
	CreatedAt      NullTime
	PasswordHash   string
}

type BusinessRow struct {
	ID        int64
	Name      string
	OwnerID   int64
	CreatedAt NullTime
}

type CreateBusinessOwnerParams struct {
	Email              string
	PasswordHash       string
	FullName           string
	BusinessName       string
	Industry           string
	Country            string
	Timezone           string
	Currency           string
	GrantPlan          bool
	PlanID             int64
	SubscriptionStatus string
	GrantDays          int
}

type CreateBusinessOwnerRow struct {
	OwnerID        int64
	BusinessID     int64
	SubscriptionID NullInt64
}

type SubscriptionRow struct {
	ID                   int64
	OwnerID              int64
	OwnerName            NullString
	PlanID               int64
	PlanName             string
	Status               string
	SeatCount            int
	ActiveSeatCount      int
	IncludedSeatCount    int
	BillableSeatCount    int
	TrialEndsAt          NullTime
	NextBillingDate      NullTime
	CurrentPeriodEndsAt  NullTime
	PaddleSubscriptionID NullString
	PaddleCustomerID     NullString
	CreatedAt            NullTime
	UpdatedAt            NullTime
	OwnerEmail           NullString
	BusinessID           NullInt64
	BusinessName         NullString
}

type SubscriptionWriteParams struct {
	OwnerID              int64
	PlanID               int64
	Status               string
	SeatCount            int
	TrialEndsAt          NullTime
	NextBillingDate      NullTime
	CurrentPeriodEndsAt  NullTime
	PaddleSubscriptionID NullString
	PaddleCustomerID     NullString
}

type PlanRow struct {
	ID                      int64
	Name                    string
	Slug                    string
	PricePerSeatCents       int
	AnnualPricePerSeatCents int
	TrialDays               int
	MaxSeats                NullInt64
	MaxLocations            NullInt64
	IsActive                bool
	IsSolo                  bool
	IsFeatured              bool
	TierOrder               int
	Tagline                 string
	CreatedAt               NullTime
}

type GrantRow struct {
	ID              int64
	UserID          int64
	UserName        NullString
	PlanID          int64
	PlanName        string
	GrantedBy       int64
	GranterName     NullString
	Note            NullString
	ExpiresAt       NullTime
	RevokedAt       NullTime
	CreatedAt       NullTime
	UserEmail       NullString
	BusinessID      NullInt64
	BusinessName    NullString
	ActiveSeatCount int
}

type AddonCatalogRow struct {
	ID          string
	Name        string
	Description NullString
	Icon        NullString
	Price       NullString
	IsFree      NullBool
	SortOrder   NullInt64
}

type AddonGrantRow struct {
	ID             int64
	UserID         int64
	UserName       NullString
	UserEmail      NullString
	BusinessID     int64
	BusinessName   NullString
	AddonID        string
	AddonName      string
	AddonPrice     NullString
	GrantedBy      int64
	GranterName    NullString
	Note           NullString
	ExpiresAt      NullTime
	RevokedAt      NullTime
	EnabledByGrant bool
	CreatedAt      NullTime
}

type AddonGrantScopeRow struct {
	UserID         int64
	BusinessID     int64
	AddonID        string
	EnabledByGrant bool
}

type StatsRow struct {
	TotalUsers        int64
	BannedUsers       int64
	TotalBusinesses   int64
	ActiveSubs        int64
	TrialSubs         int64
	ManualGrants      int64
	ManualAddonGrants int64
	ConsoleAdmins     int64
}

type ResourceTableRow struct {
	Name    string
	Comment NullString
}

type ResourceColumnRow struct {
	Name       string
	DataType   string
	IsNullable bool
	IsPrimary  bool
	Default    NullString
	Extra      string
	Comment    NullString
}

type ResourceRowsResult struct {
	Columns    []ResourceColumnRow
	PrimaryKey string
	Rows       []map[string]any
	Total      int64
}

// BusinessRef names a business in an error an admin acts on.
type BusinessRef struct {
	ID   int64
	Name string
}

// OwnedBusinessesInUseError blocks deleting a user while other members still
// use businesses they own; ownership has to move to someone else first.
type OwnedBusinessesInUseError struct {
	Businesses []BusinessRef
}

func (e *OwnedBusinessesInUseError) Error() string {
	return "user owns businesses other members still use: " + e.BusinessList()
}

// BusinessList renders the businesses as "Name (#id), ...".
func (e *OwnedBusinessesInUseError) BusinessList() string {
	names := make([]string, 0, len(e.Businesses))
	for _, business := range e.Businesses {
		names = append(names, fmt.Sprintf("%s (#%d)", business.Name, business.ID))
	}
	return strings.Join(names, ", ")
}

// DeletedUser reports what a user deletion changed beyond the user itself.
type DeletedUser struct {
	DeactivatedBusinessIDs []int64
	// MemberBusinessIDs are businesses the user belonged to without owning.
	MemberBusinessIDs []int64
	// SeatOwnerIDs own businesses where the user held an active seat.
	SeatOwnerIDs []int64
}
