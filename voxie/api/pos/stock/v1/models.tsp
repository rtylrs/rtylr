import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.stock.v1;

/** StockMovement is a recorded inventory movement. */
model StockMovement {
  @encode(string)
  id: int64;

  @encode(string)
  business_id: int64;

  @encode(string)
  branch_id: int64;

  @encode(string)
  product_id: int64;

  /** type is the type value for this contract. */
  type: string;

  quantity: float64;

  @encode(string)
  reference_id?: int64;

  reference_type?: string;
  note?: string;

  @encode(string)
  employee_id?: int64;

  /** created_at is when the record was created. */
  created_at: string;
}

/** BranchInventory is a per-branch inventory row for a product. */
model BranchInventory {
  @encode(string)
  id: int64;

  @encode(string)
  business_id: int64;

  @encode(string)
  branch_id: int64;

  @encode(string)
  product_id: int64;

  quantity: float64;
  reorder_point: float64;
  reorder_quantity: float64;

  /** updated_at is when the record was last updated. */
  updated_at: string;

  /** item_id is an alias for product_id used by POS catalog UIs. */
  @encode(string)
  item_id: int64;

  /** name is the ERP product display name. */
  name: string;

  /** sku is the ERP product SKU. */
  sku: string;

  unit: string;

  /** current_qty is an alias for quantity used by POS inventory UIs. */
  current_qty: float64;

  /** low_stock_threshold is the ERP product low-stock threshold. */
  low_stock_threshold: float64;
}

model AdjustStockRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  branch_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  product_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  quantity?: float64;

  note: string;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  employee_id?: int64;
}

model AdjustStockResponse {
  @encode(string)
  movement_id: int64;

  message: string;
}

model ListInventoryRequest {
  @query("branch_id")
  @encode(string)
  branch_id: int64;
}

model ListInventoryResponse {
  inventory: BranchInventory[];
}

model ListStockMovementsRequest {
  @query("branch_id")
  @encode(string)
  branch_id: int64;

  @query("product_id")
  @encode(string)
  product_id: int64;
}

model ListStockMovementsResponse {
  movements: StockMovement[];
}

model LowStockRequest {
  @Onkt.required
  @minValueExclusive(0)
  @query("branch_id")
  @encode(string)
  branch_id?: int64;
}

model LowStockResponse {
  items: BranchInventory[];
}

model ReceiveStockRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  branch_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  product_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  quantity?: float64;

  @Onkt.required
  note: string;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  employee_id?: int64;
}

model ReceiveStockResponse {
  @encode(string)
  movement_id: int64;

  message: string;
}

model TransferStockRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  from_branch_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  to_branch_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  product_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  quantity?: float64;

  @Onkt.required
  note: string;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  employee_id?: int64;
}

model TransferStockResponse {
  @encode(string)
  movement_id: int64;

  message: string;
}

model WasteStockRequest {
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  branch_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  product_id?: int64;

  @Onkt.required
  @minValueExclusive(0)
  quantity?: float64;

  @Onkt.required
  note: string;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  employee_id?: int64;
}

model WasteStockResponse {
  @encode(string)
  movement_id: int64;

  message: string;
}
