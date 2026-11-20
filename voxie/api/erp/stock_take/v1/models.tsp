import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace erp.stock_take.v1;

model StockTakeItem {
  @encode(string)
  id: int64;

  @encode(string)
  stock_take_id: int64;

  @encode(string)
  product_id: int64;

  expected_qty: float64;
  counted_qty: float64;
  variance: float64;
  created_at: string;
}

model StockTake {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  @encode(string)
  warehouse_id: int64;

  status: string;
  notes: string;
  items: StockTakeItem[];
  started_at: string;
  completed_at: string;
  created_at: string;
}

model CreateStockTakeItemRequest {
  @encode(string)
  product_id: int64;

  expected_qty: float64;
  counted_qty: float64;
}

model CreateStockTakeRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  warehouse_id?: int64;

  notes: string;
  items: CreateStockTakeItemRequest[];
}

model CreateStockTakeResponse {
  @encode(string)
  stock_take_id: int64;

  message: string;
}

model ListStockTakesRequest {
  @query("page")
  page: int64;

  @query("page_size")
  page_size: int64;
}

model ListStockTakesResponse {
  stock_takes: StockTake[];
  has_more: boolean;
  page: int64;
  page_size: int64;
}

model UpdateStockTakeItemBody {
  @encode(string)
  product_id: int64;

  expected_qty: float64;
  counted_qty: float64;
}

model UpdateStockTakeBody {
  @Onkt.required
  status: string;

  notes: string;
  items: UpdateStockTakeItemBody[];
}

model UpdateStockTakeRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  body: UpdateStockTakeBody;
}

model UpdateStockTakeResponse {
  message: string;
}
