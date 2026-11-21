import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.orders.v1;

// No @route on purpose: the base path comes from this directory
// (/pos/orders/v1). pos/orders/compat/v1 re-declares GET /pos/orders/v1/orders
// for Dash, and onkt only accepts that shared route while this service's base
// path is inferred rather than declared. See voxie/README.md.
/** OrderService defines the POS order endpoints. */
interface OrderService {
  /** ListOrders returns POS orders for the active business and filters. */
  @get
  @route("/orders")
  listOrders(...ListOrdersRequest): ListOrdersResponse;

  /** CreateOrder creates a POS order. */
  @post
  @route("/orders")
  createOrder(...CreateOrderRequest): CreateOrderResponse;

  /** CreatePublicOrder places an anonymous online order from the public menu app. */
  @post
  @route("/orders/public")
  createPublicOrder(...CreatePublicOrderRequest): CreatePublicOrderResponse;

  /** GetPublicOrderStatus returns lifecycle status for a public online order. */
  @get
  @route("/orders/{order_id}/public-status")
  getPublicOrderStatus(
    ...GetPublicOrderStatusRequest,
  ): GetPublicOrderStatusResponse;

  /** GetOrderDetail returns an order and its purchased items. */
  @get
  @route("/orders/{order_id}")
  getOrderDetail(...GetOrderDetailRequest): GetOrderDetailResponse;

  /** UpdateOrderStatus changes the order lifecycle status. */
  @put
  @route("/orders/{order_id}/status")
  updateOrderStatus(...UpdateOrderStatusRequest): UpdateOrderStatusResponse;

  /** VoidOrder voids an existing order. */
  @post
  @route("/orders/{order_id}/void")
  voidOrder(...VoidOrderRequest): VoidOrderResponse;
}
