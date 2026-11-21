import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace erp.inventory.v1;

/** InventoryService defines ERP inventory endpoints. */
@route("/erp/inventory/v1")
interface InventoryService {
  @get
  @route("/report")
  getInventoryReport(...GetInventoryReportRequest): GetInventoryReportResponse;

  @get
  @route("/low-stock")
  getLowStock(...GetLowStockRequest): GetLowStockResponse;

  @get
  @route("/valuation")
  getValuation(...GetValuationRequest): GetValuationResponse;

  @post
  @route("/adjust")
  adjustInventory(...AdjustInventoryRequest): AdjustInventoryResponse;
}
