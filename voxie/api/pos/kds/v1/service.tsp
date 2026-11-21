import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.kds.v1;

/** KDSService defines POS kitchen display endpoints. */
@route("/pos/kds/v1")
interface KDSService {
  /** Returns active kitchen display orders. */
  @get
  @route("/orders")
  listKDSOrders(...ListKDSOrdersRequest): ListKDSOrdersResponse;

  /** Updates the status of an order item on the kitchen display. */
  @put
  @route("/{order_id}/items/{item_id}")
  updateKDSItemStatus(
    ...UpdateKDSItemStatusRequest,
  ): UpdateKDSItemStatusResponse;
}
