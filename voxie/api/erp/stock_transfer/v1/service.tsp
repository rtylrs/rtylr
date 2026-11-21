import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace erp.stock_transfer.v1;

/** StockTransferService defines ERP stock transfer endpoints. */
@route("/erp/stock-transfer/v1")
interface StockTransferService {
  @get
  @route("/stock-transfers")
  listStockTransfers(...ListStockTransfersRequest): ListStockTransfersResponse;

  @post
  @route("/stock-transfers")
  createStockTransfer(
    ...CreateStockTransferRequest,
  ): CreateStockTransferResponse;

  @patch
  @route("/stock-transfers/{id}/status")
  updateStockTransferStatus(
    ...UpdateStockTransferStatusRequest,
  ): UpdateStockTransferStatusResponse;
}
