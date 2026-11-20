import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.cash.v1;

/** CashService defines POS cash drawer movement endpoints. */
@route("/pos/cash/v1")
interface CashService {
  /** Returns cash movements for a given session. */
  @get
  @route("/movements")
  listCashMovements(...ListCashMovementsRequest): ListCashMovementsResponse;

  /** Records a cash drop. */
  @post
  @route("/drop")
  cashDrop(...CashDropRequest): CashDropResponse;

  /** Records a cash pickup. */
  @post
  @route("/pickup")
  cashPickup(...CashPickupRequest): CashPickupResponse;

  /** Records a cash drawer adjustment. */
  @post
  @route("/adjust")
  cashAdjust(...CashAdjustRequest): CashAdjustResponse;
}
