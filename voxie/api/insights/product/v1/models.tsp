import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace insights.product.v1;

model BestSeller {
  /** product_id is the product id value for this contract. */
  @encode(string)
  product_id: int64;

  /** name is the name value for this contract. */
  name: string;

  /** units_sold is the units sold value for this contract. */
  units_sold: float64;

  /** revenue_minor_units is the revenue minor units value for this contract. */
  @encode(string)
  revenue_minor_units: int64;
}

model SlowMover {
  /** product_id is the product id value for this contract. */
  @encode(string)
  product_id: int64;

  /** name is the name value for this contract. */
  name: string;

  /** stock_quantity is the stock quantity value for this contract. */
  stock_quantity: float64;
}

model MarginBucket {
  /** label is the label value for this contract. */
  label: string;

  /** lower_pct is the lower pct value for this contract. */
  lower_pct: float64;

  /** upper_pct is the upper pct value for this contract. */
  upper_pct: float64;

  /** product_count is the product count value for this contract. */
  @encode(string)
  product_count: int64;
}

model CategorySales {
  /** category_id is the category id value for this contract. */
  @encode(string)
  category_id: int64;

  /** category_name is the category name value for this contract. */
  category_name: string;

  /** units_sold is the units sold value for this contract. */
  units_sold: float64;

  /** revenue_minor_units is the revenue minor units value for this contract. */
  @encode(string)
  revenue_minor_units: int64;
}

model GetBestSellersRequest {
  /** limit is the maximum number of products to return. */
  @query("limit")
  limit: int32;
}

model GetBestSellersResponse {
  /** products is the products value for this contract. */
  products: BestSeller[];
}

model GetCategoryPerformanceRequest {}

model GetCategoryPerformanceResponse {
  /** categories is the categories value for this contract. */
  categories: CategorySales[];
}

model GetProductMarginsRequest {}

model ProductMargin {
  /** product_id is the ERP product. */
  @encode(string)
  product_id: int64;

  /** name is the product name at the time of the report. */
  name: string;

  /** units_sold is the quantity sold in the period. */
  units_sold: float64;

  /**
   * revenue_minor_units is gross sales value in the smallest unit of the
   * reporting currency.
   */
  @encode(string)
  revenue_minor_units: int64;

  /**
   * cost_minor_units is units_sold multiplied by the product's cost price,
   * in the same minor unit. Absent when the product has no cost price on
   * record, which is distinct from a cost of zero.
   */
  @encode(string)
  cost_minor_units?: int64;
}

model GetProductMarginsResponse {
  /** buckets is the buckets value for this contract. */
  buckets: MarginBucket[];

  /**
   * products is one row per product sold in the period, so margin per item
   * is answerable without pulling the ERP catalogue client-side. Absent when
   * the report is requested at distribution granularity only.
   */
  products: ProductMargin[];
}

model GetProductPerformanceRequest {}

model GetProductPerformanceResponse {
  /** total_skus is the total SKU count. */
  @encode(string)
  total_skus: int64;

  /** active_skus is the active SKU count. */
  @encode(string)
  active_skus: int64;

  /** out_of_stock is the out-of-stock SKU count. */
  @encode(string)
  out_of_stock: int64;

  /** avg_margin_pct is the average margin percentage. */
  avg_margin_pct: float64;
}

model GetSlowMoversRequest {
  /** days is the trailing sales window used to detect slow movers. */
  @query("days")
  days: int32;
}

model GetSlowMoversResponse {
  /** products is the products value for this contract. */
  products: SlowMover[];
}
