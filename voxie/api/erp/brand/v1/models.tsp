import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace erp.brand.v1;

model Brand {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  name: string;
  logo_url: string;
  description: string;

  @encode(string)
  product_count: int64;

  created_at: string;
  updated_at: string;
}

model CreateBrandRequest {
  @Onkt.required
  name: string;

  logo_url: string;
  description: string;
}

model CreateBrandResponse {
  @encode(string)
  brand_id: int64;

  message: string;
}

model UpdateBrandBody {
  @Onkt.required
  name: string;

  logo_url: string;
  description: string;
}

model UpdateBrandRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @body
  body: UpdateBrandBody;
}

model UpdateBrandResponse {
  message: string;
}

model DeleteBrandRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteBrandResponse {
  message: string;
}

model ListBrandsRequest {
  /** Maximum number of brands to return. */
  @query("limit")
  limit: int64;

  /** Number of brands to skip. */
  @query("offset")
  offset: int64;
}

model ListBrandsResponse {
  brands: Brand[];
  has_more: boolean;
  limit: int64;
  offset: int64;
  total: int64;
}
