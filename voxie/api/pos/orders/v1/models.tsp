import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.orders.v1;

// ── Shared Payloads ───────────────────────────────────────────────
model OrderItem {
  @encode(string)
  id: int64;

  @encode(string)
  order_id: int64;

  @encode(string)
  menu_item_id: int64;

  name: string;
  sku: string;
  quantity: float64;
  unit_price: float64;
  discount_pct: float64;
  discount_amount: float64;
  tax_rate: float64;
  total: float64;
  notes: string;

  /** status is the current lifecycle state. */
  status: string;

  /** created_at is when the record was created. */
  created_at: string;
}

model Order {
  @encode(string)
  id: int64;

  /** owner_id identifies the tenant that owns the resource. */
  @encode(string)
  owner_id: int64;

  @encode(string)
  session_id: int64;

  @encode(string)
  table_id: int64;

  @encode(string)
  location_id: int64;

  order_number: string;
  type: string;

  /** status is the current lifecycle state. */
  status: string;

  @encode(string)
  customer_id: int64;

  customer_name: string;

  @encode(string)
  staff_id: int64;

  subtotal: float64;
  discount_amount: float64;
  tax_amount: float64;
  total: float64;
  payment_method: string;
  payment_status: string;
  amount_paid: float64;
  change_due: float64;
  notes: string;
  kitchen_notes: string;

  /** created_at is when the record was created. */
  created_at: string;

  /** updated_at is when the record was last updated. */
  updated_at: string;

  paid_at: string;
  items: OrderItem[];
}

// ── Create Order ──────────────────────────────────────────────────
model CreateOrderItemInput {
  @encode(string)
  menu_item_id: int64;

  name: string;
  sku: string;
  quantity: float64;
  unit_price: float64;
  discount_pct: float64;
  tax_rate: float64;
  notes: string;
}

model SplitGroup {
  @Onkt.encode("string")
  item_ids: int64[];
}

model CreateOrderRequest {
  @encode(string)
  session_id: int64;

  @encode(string)
  table_id: int64;

  @encode(string)
  location_id: int64;

  type: string;

  @encode(string)
  customer_id: int64;

  customer_name: string;
  payment_method: string;
  notes: string;
  kitchen_notes: string;

  @minItems(1)
  items: CreateOrderItemInput[];
}

model CreateOrderResponse {
  @encode(string)
  order_id: int64;

  order_number: string;
  message: string;
}

// ── Create Public Order ───────────────────────────────────────────
model PublicOrderItemInput {
  @minValueExclusive(0)
  @encode(string)
  menu_item_id: int64;

  name: string;

  @minValueExclusive(0)
  quantity: float64;

  unit_price: float64;
  notes: string;
}

model CreatePublicOrderRequest {
  @minValueExclusive(0)
  @encode(string)
  owner_id: int64;

  @encode(string)
  location_id?: int64;

  @encode(string)
  table_id?: int64;

  /** dine_in, self_service, pickup, takeaway, delivery, or online. */
  fulfillment_mode?: string;

  delivery_address?: string;
  currency_code?: string;
  loyalty_phone?: string;
  crm_opt_in?: boolean;

  customer_name: string;
  customer_phone: string;
  customer_email: string;
  notes: string;

  @minItems(1)
  items: PublicOrderItemInput[];
}

model CreatePublicOrderResponse {
  @encode(string)
  order_id: int64;

  order_number: string;
  status: string;

  /** Opaque credential required to retrieve this order's public status. */
  public_token: string;

  /** Optional third-party payment gateway redirect URL configured by the branch. */
  payment_redirect_url?: string;
}

// ── Get Order Detail ──────────────────────────────────────────────
model GetOrderDetailRequest {
  @minValueExclusive(0)
  @encode(string)
  order_id: int64;
}

model GetOrderDetailResponse {
  order: Order;
}

// ── Get Public Order Status ───────────────────────────────────────
model GetPublicOrderStatusRequest {
  @minValueExclusive(0)
  @encode(string)
  order_id: int64;

  /**
   * business_id scopes the lookup to the business the order belongs to,
   * preventing cross-tenant order enumeration on this anonymous endpoint.
   */
  @Onkt.required
  @minValueExclusive(0)
  @query("business_id")
  @encode(string)
  business_id?: int64;

  /**
   * Single-use? No: this credential is scoped to this order and is safe to
   * present repeatedly for status polling. It must never be replaced by the
   * sequential order ID alone.
   */
  @Onkt.required
  @query("public_token")
  public_token: string;
}

model GetPublicOrderStatusResponse {
  @encode(string)
  order_id: int64;

  order_number: string;
  status: string;
}

// ── List Orders ───────────────────────────────────────────────────
model ListOrdersRequest {
  @Onkt.required
  @minValueExclusive(0)
  @query("session_id")
  @encode(string)
  session_id?: int64;

  /** status is the current lifecycle state. */
  @query("status")
  status: string;

  @query("date_from")
  date_from: string;

  @query("date_to")
  date_to: string;

  /** search filters orders by order number, customer, or matching text. */
  @query("search")
  search: string;

  /** page is the one-based page number to return. */
  @query("page")
  page: int64;

  /** page_size is the maximum number of orders to return. */
  @query("page_size")
  page_size: int64;
}

model ListOrdersResponse {
  orders: Order[];

  /** total is the total number of matching records. */
  total: int64;

  /** has_more indicates another page is available. */
  has_more: boolean;

  /** page is the one-based page number returned. */
  page: int64;

  /** page_size is the page size returned. */
  page_size: int64;
}

// ── Send To Kitchen ───────────────────────────────────────────────
model SendToKitchenRequest {
  @minValueExclusive(0)
  @encode(string)
  order_id: int64;
}

model SendToKitchenResponse {
  success: boolean;
}

// ── Transfer Table ────────────────────────────────────────────────
model TransferTableRequest {
  @minValueExclusive(0)
  @encode(string)
  order_id: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  new_table_id?: int64;
}

model TransferTableResponse {
  success: boolean;
}

// ── Update Order Status ───────────────────────────────────────────
model UpdateOrderStatusBody {
  /** status is the current lifecycle state. */
  @Onkt.required
  status: string;
}

model UpdateOrderStatusRequest {
  @minValueExclusive(0)
  @encode(string)
  order_id: int64;

  /** body carries the new order status bound from the HTTP payload. */
  @body
  body: UpdateOrderStatusBody;
}

model UpdateOrderStatusResponse {
  message: string;
}

// ── Void Order ────────────────────────────────────────────────────
model VoidOrderRequest {
  @minValueExclusive(0)
  @encode(string)
  order_id: int64;

  @Onkt.required
  reason: string;
}

model VoidOrderResponse {
  message: string;
}
