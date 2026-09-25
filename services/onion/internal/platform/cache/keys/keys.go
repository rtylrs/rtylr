// Package keys is the canonical registry of every cache key the onion API
// uses. Every cached read in the codebase must build its key through one of
// these helpers — that gives us a single place to:
//
//  1. Know which keys actually exist in production.
//  2. Wipe a coherent slice of cache when something changes
//     (e.g. invalidating "everything cached for user 42").
//
// Cache values that should not be sent to the client live in
// `internal/platform/cache/models` — kept inside onion intentionally so we
// can include fields (audit timestamps, internal flags) that we don't want
// to leak through the public voxie contracts.
package keys

import "fmt"

// Namespace prefixes every key so we can find/clear ours without touching
// anything Redis-side that isn't ours.
const Namespace = "cache:"

// InvalidationType is the high-level scope passed to the public
// /invalidate-cache endpoint. The handler resolves a type+id into a set of
// concrete key patterns.
type InvalidationType string

const (
	InvalidateUser         InvalidationType = "user"
	InvalidateBusiness     InvalidationType = "business"
	InvalidateOrder        InvalidationType = "order"
	InvalidateProduct      InvalidationType = "product"
	InvalidateCustomer     InvalidationType = "customer"
	InvalidateMenu         InvalidationType = "menu"
	InvalidateTeam         InvalidationType = "team"
	InvalidateConfig       InvalidationType = "config"
	InvalidateSubscription InvalidationType = "subscription"
	InvalidateAll          InvalidationType = "all"
)

// IsKnownInvalidationType returns true when the value is one of the
// supported scopes.
func IsKnownInvalidationType(t string) bool {
	switch InvalidationType(t) {
	case InvalidateUser,
		InvalidateBusiness,
		InvalidateOrder,
		InvalidateProduct,
		InvalidateCustomer,
		InvalidateMenu,
		InvalidateTeam,
		InvalidateConfig,
		InvalidateSubscription,
		InvalidateAll:
		return true
	}
	return false
}

// ─── user-scoped keys ──────────────────────────────────────────────────────
//
// All keys that hold per-user data start with `cache:user:{id}:` so the
// invalidate-user handler can wipe the entire bucket in one SCAN.

func UserMe(userID int64) string {
	return fmt.Sprintf("%suser:%d:me", Namespace, userID)
}

func UserMeForBusiness(userID, businessID int64) string {
	return fmt.Sprintf("%suser:%d:business:%d:me", Namespace, userID, businessID)
}

func UserPermissions(userID int64) string {
	return fmt.Sprintf("%suser:%d:permissions", Namespace, userID)
}

func UserApps(userID int64) string {
	return fmt.Sprintf("%suser:%d:apps:v2", Namespace, userID)
}

func UserBusinesses(userID int64) string {
	return fmt.Sprintf("%suser:%d:businesses", Namespace, userID)
}

func UserBusinessID(userID int64) string {
	return fmt.Sprintf("%suser:%d:business_id", Namespace, userID)
}

func UserBusinessRole(userID, businessID int64) string {
	return fmt.Sprintf("%suser:%d:business:%d:role", Namespace, userID, businessID)
}

func UserAllowedApps(userID, businessID int64) string {
	return fmt.Sprintf("%suser:%d:business:%d:apps", Namespace, userID, businessID)
}

func UserPrefix(userID int64) string {
	return fmt.Sprintf("%suser:%d:*", Namespace, userID)
}

// ─── business-scoped keys ─────────────────────────────────────────────────

func BusinessConfig(businessID int64) string {
	return fmt.Sprintf("%sbusiness:%d:config", Namespace, businessID)
}

func BusinessProfile(businessID int64) string {
	return fmt.Sprintf("%sbusiness:%d:profile", Namespace, businessID)
}

func BusinessTheme(businessID int64) string {
	return fmt.Sprintf("%sbusiness:%d:theme", Namespace, businessID)
}

func BusinessKPIs(businessID int64) string {
	return fmt.Sprintf("%sbusiness:%d:kpis", Namespace, businessID)
}

func BusinessDashboardStats(businessID int64) string {
	return fmt.Sprintf("%sbusiness:%d:dashboard_stats", Namespace, businessID)
}

func BusinessEnabledAddons(businessID int64) string {
	return fmt.Sprintf("%sbusiness:%d:enabled_addons", Namespace, businessID)
}

func BusinessTeam(businessID int64) string {
	return fmt.Sprintf("%sbusiness:%d:team", Namespace, businessID)
}

func BusinessPrefix(businessID int64) string {
	return fmt.Sprintf("%sbusiness:%d:*", Namespace, businessID)
}

// ─── domain-scoped keys ───────────────────────────────────────────────────

func Order(ownerID, orderID int64) string {
	return fmt.Sprintf("%sorder:%d:%d", Namespace, ownerID, orderID)
}

func OrdersList(ownerID int64) string {
	return fmt.Sprintf("%sorder:%d:list", Namespace, ownerID)
}

func OrderPrefix(ownerID int64) string {
	return fmt.Sprintf("%sorder:%d:*", Namespace, ownerID)
}

func Product(ownerID, productID int64) string {
	return fmt.Sprintf("%sproduct:%d:%d", Namespace, ownerID, productID)
}

func ProductsList(ownerID int64) string {
	return fmt.Sprintf("%sproduct:%d:list", Namespace, ownerID)
}

func ProductPrefix(ownerID int64) string {
	return fmt.Sprintf("%sproduct:%d:*", Namespace, ownerID)
}

// Inventory views live under the product namespace because they are
// derived from product stock and must be invalidated whenever a product
// changes — wiping ProductPrefix wipes these too.
func InventoryReport(ownerID int64) string {
	return fmt.Sprintf("%sproduct:%d:inventory:report", Namespace, ownerID)
}

func InventoryLowStock(ownerID int64) string {
	return fmt.Sprintf("%sproduct:%d:inventory:low_stock", Namespace, ownerID)
}

func InventoryValuation(ownerID int64) string {
	return fmt.Sprintf("%sproduct:%d:inventory:valuation", Namespace, ownerID)
}

func Customer(ownerID, customerID int64) string {
	return fmt.Sprintf("%scustomer:%d:%d", Namespace, ownerID, customerID)
}

func CustomersList(ownerID int64) string {
	return fmt.Sprintf("%scustomer:%d:list", Namespace, ownerID)
}

func CustomerPrefix(ownerID int64) string {
	return fmt.Sprintf("%scustomer:%d:*", Namespace, ownerID)
}

func MenusList(ownerID int64) string {
	return fmt.Sprintf("%smenu:%d:list", Namespace, ownerID)
}

func MenuPrefix(ownerID int64) string {
	return fmt.Sprintf("%smenu:%d:*", Namespace, ownerID)
}

func Subscription(ownerID int64) string {
	return fmt.Sprintf("%ssubscription:%d", Namespace, ownerID)
}

// SubscriptionAccess is the short-lived entitlement decision used by the
// request middleware. Keep it under the subscription prefix so billing and
// console mutations invalidate it with the rest of the owner's subscription
// state.
func SubscriptionAccess(ownerID int64) string {
	return fmt.Sprintf("%ssubscription:%d:access", Namespace, ownerID)
}

// SubscriptionPrefix matches every subscription-scoped key for the owner. The
// trailing ':' is load-bearing: a bare "cache:subscription:4*" glob would also
// match owner 42's keys. The exact Subscription key carries no suffix, so
// invalidation must pair this with Subscription() — see SubscriptionPatterns.
func SubscriptionPrefix(ownerID int64) string {
	return fmt.Sprintf("%ssubscription:%d:*", Namespace, ownerID)
}

// SubscriptionPatterns returns the exact subscription object key plus the
// delimited wildcard covering its derived keys (e.g. SubscriptionAccess).
// Invalidation callers should delete both.
func SubscriptionPatterns(ownerID int64) []string {
	return []string{Subscription(ownerID), SubscriptionPrefix(ownerID)}
}

// ─── generic per-module scoping ────────────────────────────────────────────
//
// Scoped + ScopedPrefix exist so modules that don't (yet) have a dedicated
// key family in this registry can still cache through the same SCAN/UNLINK
// invalidation pattern. Picking a stable `scope` string per module (e.g.
// "pos_tables") keeps wildcard wipes contained. Modules that need
// cross-module invalidation should graduate to dedicated helpers above.

func Scoped(scope string, ownerID int64, detail string) string {
	return fmt.Sprintf("%s%s:%d:%s", Namespace, scope, ownerID, detail)
}

func ScopedPrefix(scope string, ownerID int64) string {
	return fmt.Sprintf("%s%s:%d:*", Namespace, scope, ownerID)
}
