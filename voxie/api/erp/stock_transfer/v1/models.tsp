import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace erp.stock_transfer.v1;

model StockTransferItem {
  @encode(string)
  id: int64;

  @encode(string)
  transfer_id: int64;

  @encode(string)
  product_id: int64;

  quantity: float64;
  created_at: string;
}

model StockTransfer {
  @encode(string)
  id: int64;

  @encode(string)
  owner_id: int64;

  @encode(string)
  from_warehouse_id: int64;

  @encode(string)
  to_warehouse_id: int64;

  status: string;
  notes: string;
  items: StockTransferItem[];
  created_at: string;
  updated_at: string;
}

model CreateStockTransferItemRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  product_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  quantity?: float64;
}

model CreateStockTransferRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  from_warehouse_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  to_warehouse_id?: int64;

  notes: string;

  @minItems(1)
  items: CreateStockTransferItemRequest[];
}

model CreateStockTransferResponse {
  @encode(string)
  transfer_id: int64;

  message: string;
}

model ListStockTransfersRequest {
  @query("page")
  page: int64;

  @query("page_size")
  page_size: int64;
}

model ListStockTransfersResponse {
  stock_transfers: StockTransfer[];
  has_more: boolean;
  page: int64;
  page_size: int64;
}

model UpdateStockTransferStatusBody {
  @Onkt.required
  status: string;
}

model UpdateStockTransferStatusRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  body: UpdateStockTransferStatusBody;
}

model UpdateStockTransferStatusResponse {
  message: string;
}
