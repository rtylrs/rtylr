import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace erp.product.v1;

model Product {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  @encode(string)
  category_id: int64;

  @encode(string)
  supplier_id: int64;

  name: string;
  sku: string;
  barcode: string;
  description: string;
  unit: string;
  cost_price: float64;
  sale_price: float64;
  stock_quantity: float64;
  low_stock_threshold: float64;
  reorder_quantity: float64;
  is_active: boolean;
  track_inventory: boolean;

  /** When the record was created. */
  created_at: string;

  /** When the record was last updated. */
  updated_at: string;

  @encode(string)
  business_id: int64;

  @encode(string)
  brand_id: int64;

  compare_at_price: float64;
  tax_rate: float64;
  currency: string;
  weight: float64;
  weight_unit: string;
  length: float64;
  width: float64;
  height: float64;
  dimension_unit: string;
  country_of_origin: string;
  hs_code: string;
  allow_backorder: boolean;
  lead_time_days: int32;
  image_url: string;
}

model FullProduct {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  @encode(string)
  business_id: int64;

  @encode(string)
  category_id: int64;

  @encode(string)
  supplier_id: int64;

  name: string;
  sku: string;
  barcode: string;
  description: string;
  unit: string;
  cost_price: float64;
  sale_price: float64;
  stock_quantity: float64;
  low_stock_threshold: float64;
  reorder_quantity: float64;
  is_active: boolean;
  track_inventory: boolean;

  @encode(string)
  brand_id: int64;

  compare_at_price: float64;
  tax_rate: float64;
  currency: string;
  weight: float64;
  weight_unit: string;
  length: float64;
  width: float64;
  height: float64;
  dimension_unit: string;
  country_of_origin: string;
  hs_code: string;
  allow_backorder: boolean;
  lead_time_days: int32;
  image_url: string;
  created_at: string;
  updated_at: string;
}

model ProductAttribute {
  @encode(string)
  id: int64;

  name: string;
  value: string;
  position: int32;
}

model DashboardLowStockItem {
  @encode(string)
  product_id: int64;

  name: string;
  sku: string;
  current_stock: float64;
  reorder_point: float64;
  warehouse_name: string;
}

model DashboardRecentOrder {
  @encode(string)
  id: int64;

  po_number: string;
  supplier_name: string;
  status: string;
  total: float64;
  created_at: string;
}

model DashboardTopCategory {
  name: string;
  product_count: int32;
}

model ExpiringItem {
  @encode(string)
  product_id: int64;

  name: string;
  sku: string;
  batch_id: string;
  expiry_date: string;
  quantity: float64;
  warehouse_name: string;
}

model ReorderSuggestion {
  @encode(string)
  product_id: int64;

  name: string;
  sku: string;
  current_stock: float64;
  reorder_point: float64;
  suggested_quantity: float64;
  avg_daily_sales: float64;
  lead_time_days: int32;
  confidence: string;
}

model ListProductsRequest {
  /** Filters products by name or SKU. */
  @query("search")
  search: string;

  /** Filters products by category. */
  @query("category_id")
  @encode(string)
  category_id: int64;

  /** The maximum number of products to return. */
  @query("limit")
  limit: int64;

  /** The number of records to skip. */
  @query("offset")
  offset: int64;
}

model ListProductsResponse {
  products: Product[];
  has_more: boolean;
  page: int64;
  page_size: int64;
  total: int64;
}

model CreateProductRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  category_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  supplier_id?: int64;

  @Onkt.required
  name: string;

  @Onkt.required
  sku: string;

  @Onkt.required
  barcode: string;

  @Onkt.required
  description: string;

  @Onkt.required
  unit: string;

  @minValue(0)
  cost_price: float64;

  @minValue(0)
  sale_price: float64;

  @minValue(0)
  low_stock_threshold: float64;

  @minValue(0)
  reorder_quantity: float64;

  is_active: boolean;
  track_inventory: boolean;

  /** The opening tracked quantity for this product. */
  @minValue(0)
  stock_quantity: float64;

  /** Links the product to a brand when present. */
  @encode(string)
  brand_id: int64;

  @minValue(0)
  compare_at_price: float64;

  @minValue(0)
  tax_rate: float64;

  currency: string;

  @minValue(0)
  weight: float64;

  weight_unit: string;

  @minValue(0)
  length: float64;

  @minValue(0)
  width: float64;

  @minValue(0)
  height: float64;

  dimension_unit: string;
  country_of_origin: string;
  hs_code: string;
  allow_backorder: boolean;

  @minValue(0)
  lead_time_days: int32;

  image_url: string;
}

model CreateProductResponse {
  @encode(string)
  product_id: int64;

  /** A human-readable result message. */
  message: string;
}

model GetProductByBarcodeRequest {
  @Onkt.required
  code: string;
}

model GetProductByBarcodeResponse {
  product: Product;
}

model GetProductRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model GetProductResponse {
  product: Product;
}

model GetProductFullRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model GetProductFullResponse {
  product: FullProduct;
}

model ListProductAttributesRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model ListProductAttributesResponse {
  attributes: ProductAttribute[];
}

model UpdateProductBody {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  category_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  supplier_id?: int64;

  @Onkt.required
  name: string;

  @Onkt.required
  sku: string;

  @Onkt.required
  barcode: string;

  @Onkt.required
  description: string;

  @Onkt.required
  unit: string;

  @minValue(0)
  cost_price: float64;

  @minValue(0)
  sale_price: float64;

  @minValue(0)
  low_stock_threshold: float64;

  @minValue(0)
  reorder_quantity: float64;

  is_active: boolean;
  track_inventory: boolean;

  /** The tracked quantity to store for this product. */
  @minValue(0)
  stock_quantity: float64;

  /** Links the product to a brand when present. */
  @encode(string)
  brand_id: int64;

  @minValue(0)
  compare_at_price: float64;

  @minValue(0)
  tax_rate: float64;

  currency: string;

  @minValue(0)
  weight: float64;

  weight_unit: string;

  @minValue(0)
  length: float64;

  @minValue(0)
  width: float64;

  @minValue(0)
  height: float64;

  dimension_unit: string;
  country_of_origin: string;
  hs_code: string;
  allow_backorder: boolean;

  @minValue(0)
  lead_time_days: int32;

  image_url: string;
}

model UpdateProductRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @body
  body: UpdateProductBody;
}

model UpdateProductResponse {
  message: string;
}

model DeleteProductRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteProductResponse {
  message: string;
}

model GetDashboardRequest {}

model GetDashboardResponse {
  total_products: int32;
  low_stock_count: int32;
  pending_orders: int32;
  total_suppliers: int32;
  inventory_value: float64;
  low_stock_items: DashboardLowStockItem[];
  recent_orders: DashboardRecentOrder[];
  top_categories: DashboardTopCategory[];
}

model ListReorderSuggestionsRequest {
  /** Maximum number of reorder suggestions to return. */
  @query("limit")
  limit: int64;

  /** Number of reorder suggestions to skip. */
  @query("offset")
  offset: int64;
}

model ListReorderSuggestionsResponse {
  suggestions: ReorderSuggestion[];
  has_more: boolean;
  limit: int64;
  offset: int64;
  total: int64;
}

model ListExpiringItemsRequest {
  @query("days_ahead")
  days_ahead: int32;

  /** Maximum number of expiring items to return. */
  @query("limit")
  limit: int64;

  /** Number of expiring items to skip. */
  @query("offset")
  offset: int64;
}

model ListExpiringItemsResponse {
  items: ExpiringItem[];
  has_more: boolean;
  limit: int64;
  offset: int64;
  total: int64;
}
