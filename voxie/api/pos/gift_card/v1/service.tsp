import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.gift_card.v1;

/** GiftCardService defines POS gift card endpoints. */
@route("/pos/gift-card/v1")
interface GiftCardService {
  /** Returns gift cards for the active business. */
  @get
  @route("/gift-cards")
  listGiftCards(...ListGiftCardsRequest): ListGiftCardsResponse;

  /** Issues a new gift card. */
  @post
  @route("/gift-cards")
  createGiftCard(...CreateGiftCardRequest): CreateGiftCardResponse;

  /** Redeems an amount from a gift card by code. */
  @post
  @route("/gift-cards/{code}/redeem")
  redeemGiftCard(...RedeemGiftCardRequest): RedeemGiftCardResponse;

  /** Returns a gift card by code. */
  @get
  @route("/gift-cards/{code}")
  getGiftCardByCode(...GetGiftCardByCodeRequest): GetGiftCardByCodeResponse;
}
