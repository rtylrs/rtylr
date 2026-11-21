import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace erp.purchase_order.v1;

model PurchaseOrderItem {
  @encode(string)
  id: int64;

  @encode(string)
  purchase_order_id: int64;

  @encode(string)
  product_id: int64;

  quantity_ordered: float64;
  quantity_received: float64;
  unit_cost: float64;
  total_cost: float64;
}

model PurchaseOrder {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  @encode(string)
  supplier_id: int64;

  po_number: string;

  /** Current lifecycle state. */
  status: string;

  subtotal: float64;
  tax_amount: float64;
  total: float64;
  expected_delivery?: string;
  received_at?: string;
  notes: string;

  @encode(string)
  created_by: int64;

  /** When the record was created. */
  created_at: string;

  /** When the record was last updated. */
  updated_at: string;

  items: PurchaseOrderItem[];
}

model CreatePurchaseOrderItemRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  product_id?: int64;

  @minValue(0)
  quantity_ordered: float64;

  @minValue(0)
  unit_cost: float64;
}

model ListPurchaseOrdersRequest {
  /** Filters purchase orders by lifecycle state. */
  @query("status")
  status: string;

  /** The 1-based page number. */
  @query("page")
  page: int32;

  /** The number of rows to return. */
  @query("page_size")
  page_size: int32;
}

model ListPurchaseOrdersResponse {
  purchase_orders: PurchaseOrder[];

  /** True when another page is available. */
  has_more: boolean;

  /** The returned 1-based page number. */
  page: int32;

  /** The returned page size. */
  page_size: int32;
}

model CreatePurchaseOrderRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  supplier_id?: int64;

  @Onkt.required
  po_number: string;

  @minValue(0)
  tax_amount: float64;

  @Onkt.required
  expected_delivery: string;

  notes: string;

  @minItems(1)
  items: CreatePurchaseOrderItemRequest[];
}

model CreatePurchaseOrderResponse {
  @encode(string)
  purchase_order_id: int64;

  /** A human-readable result message. */
  message: string;
}

model GetPurchaseOrderRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model GetPurchaseOrderResponse {
  purchase_order: PurchaseOrder;
}

model UpdatePurchaseOrderStatusBody {
  /** Current lifecycle state. */
  @Onkt.required
  status: string;

  /** Per-line quantities submitted by the receiving operator. */
  items: PurchaseOrderItem[];
}

model UpdatePurchaseOrderStatusRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @body
  body: UpdatePurchaseOrderStatusBody;
}

model UpdatePurchaseOrderStatusResponse {
  message: string;
}

model ApprovePurchaseOrderRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model ApprovePurchaseOrderResponse {
  message: string;
}
