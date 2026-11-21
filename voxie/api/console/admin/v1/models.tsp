import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace console.admin.v1;

model User {
  id: int64;
  email: string;
  full_name: string;
  business_id?: int64;
  is_console_admin: boolean;
  is_banned: boolean;
  ban_reason?: string;
  banned_at?: string;
  created_at: string;
}

model Business {
  id: int64;
  owner_id: int64;
  name: string;
  slug: string;
  created_at: string;
}

model Subscription {
  id: int64;

  /** The business's current owner. */
  owner_id: int64;

  owner_name: string;
  owner_email: string;
  business_id?: int64;
  business_name: string;
  plan_id: int64;
  plan_name: string;
  status: string;
  seat_count: int32;
  active_seat_count: int32;
  included_seat_limit: int32;
  included_seat_count: int32;
  billable_seat_count: int32;
  trial_ends_at?: string;
  next_billing_date?: string;
  current_period_ends_at?: string;
  paddle_subscription_id?: string;
  paddle_customer_id?: string;
  days_remaining?: int32;
  created_at: string;
  updated_at?: string;
}

model Plan {
  id: int64;
  name: string;
  price_per_seat_cents: int32;
  annual_price_per_seat_cents: int32;
  is_active: boolean;
  created_at: string;
}

model Stats {
  total_users: int64;
  banned_users: int64;
  active_subscriptions: int64;
  total_businesses: int64;
  total_grants: int64;
  manual_addon_grants: int64;
  console_admins: int64;
}

model GetStatsRequest {}

model GetStatsResponse {
  stats: Stats;
}

model ListUsersRequest {
  @query("limit")
  limit: int32;

  @query("offset")
  offset: int32;

  @query("id")
  id?: int64;

  @query("email")
  email: string;

  @query("name")
  name: string;
}

model ListUsersResponse {
  users: User[];
  total: int64;
}

model GetUserRequest {
  @minValueExclusive(0)
  id: int64;
}

model GetUserResponse {
  user: User;
}

model CreateUserRequest {
  @format("email")
  email: string;

  password: string;
  full_name: string;
  is_console_admin: boolean;
}

model CreateUserResponse {
  user: User;
  message: string;
}

model UpdateUserRequest {
  @minValueExclusive(0)
  id: int64;

  full_name?: string;
  email?: string;
}

model UpdateUserResponse {
  message: string;
}

model DeleteUserRequest {
  @minValueExclusive(0)
  id: int64;
}

model DeleteUserResponse {
  message: string;
}

model BanUserRequest {
  @minValueExclusive(0)
  id: int64;

  reason: string;
}

model BanUserResponse {
  message: string;
}

model UnbanUserRequest {
  @minValueExclusive(0)
  id: int64;
}

model UnbanUserResponse {
  message: string;
}

model SetConsoleAdminRequest {
  @minValueExclusive(0)
  id: int64;

  is_console_admin: boolean;
}

model SetConsoleAdminResponse {
  message: string;
}

/**
 * Reads the subscription of a business the user owns. business_id may be
 * omitted while the user owns exactly one active business; an owner of
 * several must pass it, or the request is rejected.
 */
model GetUserSubscriptionRequest {
  @minValueExclusive(0)
  user_id: int64;

  @query("business_id")
  business_id?: int64;
}

model GetUserSubscriptionResponse {
  subscription?: Subscription;
}

model ListBusinessesRequest {
  @query("limit")
  limit: int32;

  @query("offset")
  offset: int32;

  @query("id")
  id?: int64;

  @query("name")
  name: string;

  @query("owner_id")
  owner_id?: int64;
}

model ListBusinessesResponse {
  businesses: Business[];
  total: int64;
}

model UpdateBusinessRequest {
  @minValueExclusive(0)
  id: int64;

  name: string;
}

model UpdateBusinessResponse {
  message: string;
}

model DeleteBusinessRequest {
  @minValueExclusive(0)
  id: int64;
}

model DeleteBusinessResponse {
  message: string;
}

model CreateBusinessOwnerRequest {
  @format("email")
  email: string;

  password: string;
  full_name: string;
  business_name: string;
  industry: string;
  country: string;
  timezone: string;
  currency: string;
  grant_plan: boolean;
  plan_id?: int64;
  subscription_status: string;
  grant_days?: int32;
}

model CreateBusinessOwnerResponse {
  owner: User;
  business: Business;
  subscription?: Subscription;
  message: string;
}

model ListSubscriptionsRequest {
  @query("limit")
  limit: int32;

  @query("offset")
  offset: int32;

  @query("status")
  status: string;
}

model ListSubscriptionsResponse {
  subscriptions: Subscription[];
  total: int64;
}

model UpsertSubscriptionRequest {
  /** The business the plan applies to. */
  business_id?: int64;

  /**
   * Names the business through its owner when business_id is unset. This is
   * accepted only while the owner has exactly one active business.
   */
  owner_id?: int64;

  @minValueExclusive(0)
  plan_id: int64;

  status: string;
  seat_count?: int32;
  trial_ends_at?: string;
  next_billing_date?: string;
  current_period_ends_at?: string;
  total_days?: int32;
  paddle_subscription_id?: string;
  paddle_customer_id?: string;
}

model UpsertSubscriptionResponse {
  message: string;
}

model DeleteSubscriptionRequest {
  @minValueExclusive(0)
  id: int64;
}

model DeleteSubscriptionResponse {
  message: string;
}

model ListPlansRequest {}

model ListPlansResponse {
  plans: Plan[];
}

model UpdatePlanRequest {
  @minValueExclusive(0)
  id: int64;

  name?: string;
  price_per_seat_cents?: int32;
  annual_price_per_seat_cents?: int32;
  is_active?: boolean;
}

model UpdatePlanResponse {
  message: string;
}

/**
 * A plan's price in one region. Visitors from the region see these prices
 * instead of the plan's own, and checkout charges the paddle_price_id.
 */
model PlanRegionPrice {
  plan_id: int64;

  /** ISO 3166-1 alpha-2, upper case. */
  region: string;

  /** ISO 4217, upper case. */
  currency: string;

  price_per_seat_cents: int32;
  annual_price_per_seat_cents: int32;
  paddle_price_id: string;
  is_active: boolean;
}

model ListPlanRegionPricesRequest {}

model ListPlanRegionPricesResponse {
  prices: PlanRegionPrice[];
}

model UpsertPlanRegionPriceRequest {
  @minValueExclusive(0)
  plan_id: int64;

  @minLength(2)
  @maxLength(2)
  region: string;

  @minLength(3)
  @maxLength(3)
  currency: string;

  @minValue(0)
  price_per_seat_cents: int32;

  @minValue(0)
  annual_price_per_seat_cents: int32;

  @minLength(0)
  @maxLength(100)
  paddle_price_id?: string;

  is_active?: boolean;
}

model UpsertPlanRegionPriceResponse {
  message: string;
}

model DeletePlanRegionPriceRequest {
  @minValueExclusive(0)
  plan_id: int64;

  @minLength(2)
  @maxLength(2)
  region: string;
}

model DeletePlanRegionPriceResponse {
  message: string;
}
