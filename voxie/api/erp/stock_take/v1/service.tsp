import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace erp.stock_take.v1;

/** StockTakeService defines ERP stock take endpoints. */
@route("/erp/stock-take/v1")
interface StockTakeService {
  @get
  @route("/stock-takes")
  listStockTakes(...ListStockTakesRequest): ListStockTakesResponse;

  @post
  @route("/stock-takes")
  createStockTake(...CreateStockTakeRequest): CreateStockTakeResponse;

  @put
  @route("/stock-takes/{id}")
  updateStockTake(...UpdateStockTakeRequest): UpdateStockTakeResponse;
}
