import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace erp.price_list.v1;

model PriceList {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  name: string;
  currency: string;
  is_default: boolean;
  created_at: string;
  updated_at: string;
}

model CreatePriceListRequest {
  @Onkt.required
  name: string;

  currency: string;
  is_default: boolean;
}

model CreatePriceListResponse {
  @encode(string)
  price_list_id: int64;

  message: string;
}

model ListPriceListsRequest {
  /** Maximum number of price lists to return. */
  @query("limit")
  limit: int64;

  /** Number of price lists to skip. */
  @query("offset")
  offset: int64;
}

model ListPriceListsResponse {
  price_lists: PriceList[];
  has_more: boolean;
  limit: int64;
  offset: int64;
  total: int64;
}
