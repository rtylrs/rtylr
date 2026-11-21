import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace erp.variant.v1;

model Variant {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  @encode(string)
  business_id: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  product_id?: int64;

  @Onkt.required
  name: string;

  sku: string;
  barcode: string;
  cost_price: float64;
  sale_price: float64;
  stock_quantity: float64;
  attributes: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

model CreateVariantRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  product_id?: int64;

  @Onkt.required
  name: string;

  sku: string;
  barcode: string;
  cost_price: float64;
  sale_price: float64;
  stock_quantity: float64;
  attributes: string;
  is_active: boolean;
}

model CreateVariantResponse {
  @encode(string)
  variant_id: int64;

  message: string;
}

model DeleteVariantRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteVariantResponse {
  message: string;
}

model ListVariantsRequest {
  @query("product_id")
  @encode(string)
  product_id: int64;

  /** Maximum number of variants to return. */
  @query("limit")
  limit: int64;

  /** Number of variants to skip. */
  @query("offset")
  offset: int64;
}

model ListVariantsResponse {
  variants: Variant[];
  has_more: boolean;
  limit: int64;
  offset: int64;
  total: int64;
}

model UpdateVariantBody {
  @encode(string)
  product_id: int64;

  @Onkt.required
  name: string;

  sku: string;
  barcode: string;
  cost_price: float64;
  sale_price: float64;
  stock_quantity: float64;
  attributes: string;
  is_active: boolean;
}

model UpdateVariantRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  body: UpdateVariantBody;
}

model UpdateVariantResponse {
  message: string;
}
