import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace erp.inventory.v1;

model InventoryReportItem {
  @encode(string)
  product_id: int64;

  product_name: string;
  sku: string;
  unit: string;
  stock_quantity: float64;
  low_stock_threshold: float64;
  reorder_quantity: float64;
  cost_price: float64;
  sale_price: float64;
  stock_value: float64;
  is_low_stock: boolean;
}

model ValuationItem {
  category_name: string;
  warehouse_name: string;

  @encode(string)
  total_items: int64;

  total_value: float64;
  total_cost: float64;
}

model AdjustInventoryRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  product_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  change?: float64;

  @Onkt.required
  type: string;

  notes: string;
}

model AdjustInventoryResponse {
  /** A human-readable result message. */
  message: string;
}

model GetInventoryReportRequest {
  /** Maximum number of inventory rows to return. */
  @query("limit")
  limit: int64;

  /** Number of inventory rows to skip. */
  @query("offset")
  offset: int64;
}

model GetInventoryReportResponse {
  items: InventoryReportItem[];
  total_value: float64;
  has_more: boolean;
  limit: int64;
  offset: int64;
  total: int64;
}

model GetLowStockRequest {
  /** Maximum number of low-stock rows to return. */
  @query("limit")
  limit: int64;

  /** Number of low-stock rows to skip. */
  @query("offset")
  offset: int64;
}

model GetLowStockResponse {
  items: InventoryReportItem[];
  count: int64;
  has_more: boolean;
  limit: int64;
  offset: int64;
  total: int64;
}

model GetValuationRequest {
  /** Maximum number of valuation rows to return. */
  @query("limit")
  limit: int64;

  /** Number of valuation rows to skip. */
  @query("offset")
  offset: int64;
}

model GetValuationResponse {
  items: ValuationItem[];
  total_value: float64;
  has_more: boolean;
  limit: int64;
  offset: int64;
  total: int64;
}
