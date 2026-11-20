import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.gift_card.v1;

/** GiftCard is a stored-value card that can be issued to and redeemed by a customer. */
model GiftCard {
  @encode(string)
  id: int64;

  /** owner_id identifies the tenant that owns the resource. */
  @encode(string)
  owner_id: int64;

  /** business_id identifies the active business. */
  @encode(string)
  business_id: int64;

  /** customer_id identifies the linked customer when present. */
  @encode(string)
  customer_id: int64;

  code: string;
  initial_amount: float64;
  current_balance: float64;
  customer_name: string;
  customer_email: string;
  is_active: boolean;
  expires_at: string;

  /** created_at is when the record was created. */
  created_at: string;

  /** updated_at is when the record was last updated. */
  updated_at: string;
}

model CreateGiftCardRequest {
  @Onkt.required
  code: string;

  /**
   * paid_order_id must identify a paid POS order for this business. The
   * server uses it as the tender reference for the liability entry.
   */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  paid_order_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  initial_amount?: float64;

  @Onkt.required
  customer_name: string;

  @Onkt.required
  @format("email")
  customer_email: string;

  @Onkt.required
  expires_at: string;
}

model CreateGiftCardResponse {
  @encode(string)
  gift_card_id: int64;

  message: string;
}

model GetGiftCardByCodeRequest {
  @Onkt.required
  code: string;
}

model GetGiftCardByCodeResponse {
  gift_card: GiftCard;
}

model ListGiftCardsRequest {}

model ListGiftCardsResponse {
  gift_cards: GiftCard[];
}

model RedeemGiftCardRequest {
  @Onkt.required
  code: string;

  /** order_id identifies the POS order receiving this gift-card tender. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  order_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  amount?: float64;
}

model RedeemGiftCardResponse {
  gift_card: GiftCard;
  message: string;
}
