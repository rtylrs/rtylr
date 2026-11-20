import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace erp.price_list.v1;

@route("/erp/price-list/v1")
interface PriceListService {
  @get
  @route("/price-lists")
  listPriceLists(...ListPriceListsRequest): ListPriceListsResponse;

  @post
  @route("/price-lists")
  createPriceList(...CreatePriceListRequest): CreatePriceListResponse;
}
