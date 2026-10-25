package application

import (
	"context"
	"time"
)

type repository interface {
	ListResourceTables(ctx context.Context) ([]ResourceTableRow, error)
	DescribeResourceTable(ctx context.Context, table string) ([]ResourceColumnRow, string, error)
	ListResourceRows(ctx context.Context, table string, limit, offset int) (*ResourceRowsResult, error)
	GetResourceRow(ctx context.Context, table, id string) (map[string]any, []ResourceColumnRow, string, error)
	CreateUser(ctx context.Context, email, passwordHash, fullName string, isConsoleAdmin bool) (int64, error)
	CreateBusinessOwner(ctx context.Context, params CreateBusinessOwnerParams) (*CreateBusinessOwnerRow, error)
	DeleteCreatedBusinessOwner(ctx context.Context, ownerID, businessID int64) error
	UpdateUser(ctx context.Context, id int64, fullName, email string) error
	DeleteUser(ctx context.Context, id int64) (*DeletedUser, error)
	CreateResourceRow(ctx context.Context, table string, values map[string]any) (int64, error)
	UpdateResourceRow(ctx context.Context, table, id string, values map[string]any) error
	DeleteResourceRow(ctx context.Context, table, id string) error
	GetUserByEmail(ctx context.Context, email string) (*UserRow, error)
	ListUsers(ctx context.Context, limit, offset int, id int64, email, fullName string) ([]UserRow, int64, error)
	GetUserByID(ctx context.Context, id int64) (*UserRow, error)
	BanUser(ctx context.Context, id int64, reason string) error
	UnbanUser(ctx context.Context, id int64) error
	SetConsoleAdmin(ctx context.Context, id int64, isAdmin bool) error
	ListBusinesses(ctx context.Context, limit, offset int, id int64, name string, ownerID int64) ([]BusinessRow, int64, error)
	GetBusinessByID(ctx context.Context, id int64) (*BusinessRow, error)
	ListActiveBusinessesByOwner(ctx context.Context, ownerID int64) ([]BusinessRow, error)
	UpdateBusiness(ctx context.Context, id int64, name string) error
	DeleteBusiness(ctx context.Context, id int64) error
	CountActiveSeatsForBusiness(ctx context.Context, businessID int64) (int, error)
	ListSubscriptions(ctx context.Context, limit, offset int, status string) ([]SubscriptionRow, int64, error)
	GetSubscriptionByID(ctx context.Context, id int64) (*SubscriptionRow, error)
	GetSubscriptionByBusiness(ctx context.Context, businessID int64) (*SubscriptionRow, error)
	UpsertSubscription(ctx context.Context, params SubscriptionWriteParams) error
	UpdateSubscription(ctx context.Context, id int64, params SubscriptionWriteParams) error
	DeleteSubscription(ctx context.Context, id int64) error
	ListPlans(ctx context.Context) ([]PlanRow, error)
	UpdatePlan(ctx context.Context, id int64, name string, pricePerSeatCents, annualPricePerSeatCents int, isActive bool) error
	ListGrants(ctx context.Context, limit, offset int) ([]GrantRow, int64, error)
	CreateGrant(ctx context.Context, businessID, planID, grantedBy int64, note string, expiresAt *time.Time) (int64, error)
	CreateGrantAndActivateSubscription(ctx context.Context, businessID, planID, grantedBy int64, note string, expiresAt *time.Time, sub SubscriptionWriteParams) (int64, error)
	GetGrant(ctx context.Context, id int64) (*GrantRow, error)
	GetGrantBusinessID(ctx context.Context, id int64) (int64, error)
	RevokeGrant(ctx context.Context, id int64) error
	DeleteGrant(ctx context.Context, id int64) error
	GetStats(ctx context.Context) (*StatsRow, error)
	ListAddons(ctx context.Context) ([]AddonCatalogRow, error)
	AddonExists(ctx context.Context, addonID string) (bool, error)
	ListAddonGrants(ctx context.Context, limit, offset int) ([]AddonGrantRow, int64, error)
	CreateAddonGrant(ctx context.Context, businessID int64, addonID string, grantedBy int64, note string, expiresAt *time.Time, enabledByGrant bool) (int64, error)
	GetAddonGrant(ctx context.Context, id int64) (*AddonGrantRow, error)
	GetAddonGrantScope(ctx context.Context, id int64) (*AddonGrantScopeRow, error)
	RevokeAddonGrant(ctx context.Context, id int64) error
	DeleteAddonGrant(ctx context.Context, id int64) error
	IsBusinessAddonEnabled(ctx context.Context, businessID int64, addonID string) (bool, error)
	EnableBusinessAddon(ctx context.Context, businessID int64, addonID string) error
	DisableBusinessAddon(ctx context.Context, businessID int64, addonID string) error
	HasActiveAddonGrant(ctx context.Context, businessID int64, addonID string) (bool, error)
	AdoptAddonGrantEnablement(ctx context.Context, businessID int64, addonID string, removedID int64) error
}
