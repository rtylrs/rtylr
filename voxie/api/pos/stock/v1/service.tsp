import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.stock.v1;

/** StockService defines POS inventory and stock movement endpoints. */
@route("/pos/stock/v1")
interface StockService {
  /** Returns stock movements, optionally filtered by branch and product. */
  @get
  @route("/movements")
  listStockMovements(...ListStockMovementsRequest): ListStockMovementsResponse;

  /** Records a stock receipt. */
  @post
  @route("/receive")
  receiveStock(...ReceiveStockRequest): ReceiveStockResponse;

  /** Records a manual stock adjustment. */
  @post
  @route("/adjust")
  adjustStock(...AdjustStockRequest): AdjustStockResponse;

  /** Records wasted stock. */
  @post
  @route("/waste")
  wasteStock(...WasteStockRequest): WasteStockResponse;

  /** Transfers stock between branches. */
  @post
  @route("/transfer")
  transferStock(...TransferStockRequest): TransferStockResponse;

  /** Returns branch inventory levels. */
  @get
  @route("/inventory")
  listInventory(...ListInventoryRequest): ListInventoryResponse;

  /** Returns branch inventory items at or below their reorder point. */
  @get
  @route("/low-stock")
  lowStock(...LowStockRequest): LowStockResponse;
}
