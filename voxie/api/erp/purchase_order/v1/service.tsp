import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace erp.purchase_order.v1;

/** PurchaseOrderService defines ERP purchase order endpoints. */
@route("/erp/purchase-order/v1")
interface PurchaseOrderService {
  @get
  @route("/purchase-orders")
  listPurchaseOrders(...ListPurchaseOrdersRequest): ListPurchaseOrdersResponse;

  @post
  @route("/purchase-orders")
  createPurchaseOrder(
    ...CreatePurchaseOrderRequest,
  ): CreatePurchaseOrderResponse;

  @get
  @route("/purchase-orders/{id}")
  getPurchaseOrder(...GetPurchaseOrderRequest): GetPurchaseOrderResponse;

  @patch
  @route("/purchase-orders/{id}/status")
  updatePurchaseOrderStatus(
    ...UpdatePurchaseOrderStatusRequest,
  ): UpdatePurchaseOrderStatusResponse;

  @post
  @route("/purchase-orders/{id}/approve")
  approvePurchaseOrder(
    ...ApprovePurchaseOrderRequest,
  ): ApprovePurchaseOrderResponse;
}
