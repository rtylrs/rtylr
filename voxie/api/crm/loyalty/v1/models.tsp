import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace crm.loyalty.v1;

/** LoyaltyConfig is a business's loyalty program configuration. */
model LoyaltyConfig {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  is_enabled: boolean;
  points_per_currency_unit: float64;
  currency_per_point: float64;

  @encode(string)
  bronze_min: int64;

  @encode(string)
  silver_min: int64;

  @encode(string)
  gold_min: int64;

  @encode(string)
  platinum_min: int64;

  bronze_discount_pct: float64;
  silver_discount_pct: float64;
  gold_discount_pct: float64;
  platinum_discount_pct: float64;

  @encode(string)
  points_expiry_days: int64;

  updated_at: string;
}

/** LoyaltyTransaction is a single points award or redemption event. */
model LoyaltyTransaction {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  @encode(string)
  customer_id: int64;

  type: string;

  @encode(string)
  points: int64;

  @encode(string)
  balance_before: int64;

  @encode(string)
  balance_after: int64;

  description: string;

  @encode(string)
  reference_id?: int64;

  reference_type: string;
  created_at: string;
}

/** LoyaltyTier is a spend-based tier with an associated discount. */
model LoyaltyTier {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  name: string;

  @encode(string)
  min_points: int64;

  discount_pct: float64;
  perks: string;

  @encode(string)
  sort_order: int64;

  created_at: string;
  updated_at: string;
}

model AwardPointsRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  customer_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  points?: int64;

  @Onkt.required
  description: string;

  @encode(string)
  reference_id?: int64;

  @Onkt.required
  reference_type: string;
}

model AwardPointsResponse {
  message: string;
}

model CreateLoyaltyTierRequest {
  @Onkt.required
  name: string;

  /**
   * Points threshold at which this tier starts applying. Zero is valid and
   * means an entry tier every customer qualifies for from their first point.
   */
  @Onkt.required
  @minValue(0)
  @encode(string)
  min_points?: int64;

  /** Percentage discount for the tier. Zero is valid and means perks only. */
  @Onkt.required
  @minValue(0)
  discount_pct?: float64;

  @Onkt.required
  perks: string;

  /** Display order. Zero is valid and means first. */
  @Onkt.required
  @minValue(0)
  @encode(string)
  sort_order?: int64;
}

model CreateLoyaltyTierResponse {
  @encode(string)
  tier_id: int64;

  message: string;
}

model GetLoyaltyConfigRequest {}

model GetLoyaltyConfigResponse {
  config: LoyaltyConfig;
}

model ListLoyaltyTiersRequest {}

model ListLoyaltyTiersResponse {
  tiers: LoyaltyTier[];
}

model ListLoyaltyTransactionsRequest {
  /** Filters transactions for a single customer. */
  @query("customer_id")
  @encode(string)
  customer_id: int64;

  @query("page")
  @encode(string)
  page: int64;

  @query("page_size")
  @encode(string)
  page_size: int64;
}

model ListLoyaltyTransactionsResponse {
  transactions: LoyaltyTransaction[];

  @encode(string)
  total: int64;

  @encode(string)
  page: int64;

  @encode(string)
  page_size: int64;
}

model RedeemPointsRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  customer_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  points?: int64;

  @Onkt.required
  description: string;

  @encode(string)
  reference_id?: int64;

  @Onkt.required
  reference_type: string;
}

model RedeemPointsResponse {
  message: string;
}

model UpdateLoyaltyConfigRequest {
  is_enabled: boolean;

  @Onkt.required
  @minValueExclusive(0)
  points_per_currency_unit?: float64;

  @Onkt.required
  @minValueExclusive(0)
  currency_per_point?: float64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  bronze_min?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  silver_min?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  gold_min?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  platinum_min?: int64;

  @Onkt.required
  @minValueExclusive(0)
  bronze_discount_pct?: float64;

  @Onkt.required
  @minValueExclusive(0)
  silver_discount_pct?: float64;

  @Onkt.required
  @minValueExclusive(0)
  gold_discount_pct?: float64;

  @Onkt.required
  @minValueExclusive(0)
  platinum_discount_pct?: float64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  points_expiry_days?: int64;
}

model UpdateLoyaltyConfigResponse {
  message: string;
}

model UpdateLoyaltyTierBody {
  @Onkt.required
  name: string;

  /**
   * Points threshold at which this tier starts applying. Zero is valid and
   * means an entry tier every customer qualifies for from their first point.
   */
  @Onkt.required
  @minValue(0)
  @encode(string)
  min_points?: int64;

  /** Percentage discount for the tier. Zero is valid and means perks only. */
  @Onkt.required
  @minValue(0)
  discount_pct?: float64;

  @Onkt.required
  perks: string;

  /** Display order. Zero is valid and means first. */
  @Onkt.required
  @minValue(0)
  @encode(string)
  sort_order?: int64;
}

model UpdateLoyaltyTierRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @body
  body: UpdateLoyaltyTierBody;
}

model UpdateLoyaltyTierResponse {
  message: string;
}
