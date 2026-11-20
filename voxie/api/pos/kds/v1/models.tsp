import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.kds.v1;

/** KDSItem is a single line item on a kitchen display order. */
model KDSItem {
  @encode(string)
  id: int64;

  @encode(string)
  order_id: int64;

  name: string;
  quantity: float64;
  notes: string;

  /** status is the current lifecycle state. */
  status: string;
}

/** KDSOrder is an order shown on the kitchen display. */
model KDSOrder {
  @encode(string)
  id: int64;

  order_number: string;
  type: string;

  /** status is the current lifecycle state. */
  status: string;

  @encode(string)
  table_id: int64;

  customer_name: string;
  kitchen_notes: string;

  /** created_at is when the record was created. */
  created_at: string;

  items: KDSItem[];
}

model ListKDSOrdersRequest {}

model ListKDSOrdersResponse {
  orders: KDSOrder[];
}

model UpdateKDSItemByIDBody {
  /** status is the current lifecycle state. */
  @Onkt.required
  status: string;
}

model UpdateKDSItemByIDRequest {
  /** item_id is the unique identifier for the order item. */
  @minValueExclusive(0)
  @encode(string)
  item_id: int64;

  /** body carries the new KDS item status bound from the HTTP payload. */
  body: UpdateKDSItemByIDBody;
}

model UpdateKDSItemByIDResponse {
  message: string;
}

model UpdateKDSItemStatusBody {
  /** status is the current lifecycle state. */
  @Onkt.required
  status: string;
}

model UpdateKDSItemStatusRequest {
  /** order_id is the unique identifier for the order. */
  @minValueExclusive(0)
  @encode(string)
  order_id: int64;

  /** item_id is the unique identifier for the order item. */
  @minValueExclusive(0)
  @encode(string)
  item_id: int64;

  /** body carries the new KDS item status bound from the HTTP payload. */
  @body
  body: UpdateKDSItemStatusBody;
}

model UpdateKDSItemStatusResponse {
  message: string;
}
