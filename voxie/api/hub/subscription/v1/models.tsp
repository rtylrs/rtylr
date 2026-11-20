import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.subscription.v1;

model Subscription {
  @encode(string)
  id: int64;

  plan_name: string;
  plan_slug: string;
  apps: string[];
  price_per_seat_cents: int32;
  status: string;
  is_granted: boolean;
  seat_count: int32;
  included_seat_limit: int32;
  included_seat_count: int32;
  billable_seat_count: int32;
  grant_expires_at?: string;
  trial_ends_at?: string;
  next_billing_date?: string;
  paddle_subscription_id: string;
}

model Plan {
  @encode(string)
  id: int64;

  name: string;
  slug: string;
  apps: string[];
  price_per_seat_cents: int32;
  annual_price_per_seat_cents: int32;
  trial_days: int32;
  tagline: string;
  max_seats?: int32;
  max_locations?: int32;
  ai_credits_per_month: int32;
  features: string[];
  tier_order: int32;
  is_solo: boolean;
  is_featured: boolean;
  is_active: boolean;

  /** ISO 4217 code the prices are quoted in. USD unless a regional price applied. */
  currency: string;

  /** Upper-case ISO 3166-1 region whose price override was applied, or "" for the default price. */
  region: string;
}

model BillingLineItem {
  period_start: string;
  period_end?: string;
  seat_count: int32;
  included_seat_limit: int32;
  included_seat_count: int32;
  billable_seat_count: int32;
  included_cents: int32;
  price_per_seat_cents: int32;
  amount_cents: int32;
  type: string;
  description: string;
  is_granted: boolean;
  grant_expires_at?: string;
  addon_id: string;
  quantity: int32;
  unit_price_cents: int32;
}

model BillingDetail {
  billing_month: string;
  next_billing_date?: string;
  line_items: BillingLineItem[];
  total_cents: int32;
  plan_name: string;
  price_per_seat_cents: int32;
  current_seat_count: int32;
  included_seat_limit: int32;
  included_seat_count: int32;
  billable_seat_count: int32;
  seat_discount_cents: int32;
  status: string;
  is_granted: boolean;
  grant_expires_at?: string;
  trial_ends_at?: string;
  paddle_subscription_id: string;
}

model BillingHistoryMonth {
  month: string;
  total_cents: int32;
  line_items: BillingLineItem[];
  status: string;
}

model GetMySubscriptionRequest {}

model GetMySubscriptionResponse {
  subscription: Subscription;
}

model ListPlansRequest {
  /**
   * ISO 3166-1 alpha-2 region to price for, e.g. "CA". Unknown or unpriced
   * regions get the default prices.
   */
  @query("region")
  region?: string;
}

model ListPlansResponse {
  plans: Plan[];
}

model GetBillingDetailRequest {}

model GetBillingDetailResponse {
  billing: BillingDetail;
}

model GetBillingHistoryRequest {}

model GetBillingHistoryResponse {
  history: BillingHistoryMonth[];
}

model UpsertSubscriptionRequest {
  plan_slug: string;
}

model UpsertSubscriptionResponse {
  message: string;
}

model CreateCheckoutRequest {
  plan_slug: string;
  email: string;
  billing_cycle?: string;
  seat_count?: int32;
  addons: string[];

  /**
   * ISO 3166-1 alpha-2 price list the customer was shown. A region with no
   * chargeable price for the plan is charged the plan's own price.
   */
  region?: string;
}

model CreateCheckoutResponse {
  checkout_url: string;
}
