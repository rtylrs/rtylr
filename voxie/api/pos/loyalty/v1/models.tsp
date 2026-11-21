import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.loyalty.v1;

/** LoyaltyConfig is the loyalty program configuration for a business. */
model LoyaltyConfig {
  @encode(string)
  id: int64;

  /** business_id identifies the active business. */
  @encode(string)
  business_id: int64;

  is_enabled: boolean;
  points_per_dollar: float64;
  redemption_rate: float64;
  min_redeem_points: int32;

  /** created_at is when the record was created. */
  created_at: string;

  /** updated_at is when the record was last updated. */
  updated_at: string;
}

/** LoyaltyTransaction is a recorded loyalty points earn or redeem event. */
model LoyaltyTransaction {
  @encode(string)
  id: int64;

  /** business_id identifies the active business. */
  @encode(string)
  business_id: int64;

  /** customer_id identifies the linked customer when present. */
  @encode(string)
  customer_id: int64;

  @encode(string)
  order_id: int64;

  type: string;
  points: int32;
  balance_after: int32;
  note: string;

  /** created_at is when the record was created. */
  created_at: string;
}

model EarnPointsRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  customer_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  order_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  points?: int32;
}

model EarnPointsResponse {
  @encode(string)
  transaction_id: int64;

  message: string;
}

model GetLoyaltyBalanceRequest {
  @Onkt.required
  @minValueExclusive(0)
  @query("customer_id")
  @encode(string)
  customer_id?: int64;
}

model GetLoyaltyBalanceResponse {
  @encode(string)
  customer_id: int64;

  balance: int32;
}

model GetLoyaltyConfigRequest {}

model GetLoyaltyConfigResponse {
  config: LoyaltyConfig;
}

model ListLoyaltyTransactionsRequest {
  @query("customer_id")
  @encode(string)
  customer_id: int64;
}

model ListLoyaltyTransactionsResponse {
  transactions: LoyaltyTransaction[];
}

model RedeemPointsRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  customer_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  points?: int32;
}

model RedeemPointsResponse {
  @encode(string)
  transaction_id: int64;

  message: string;
}

model UpdateLoyaltyConfigRequest {
  is_enabled: boolean;

  @Onkt.required
  @minValueExclusive(0)
  points_per_dollar?: float64;

  @minValue(0)
  redemption_rate: float64;

  @Onkt.required
  @minValueExclusive(0)
  min_redeem_points?: int32;
}

model UpdateLoyaltyConfigResponse {
  message: string;
}
