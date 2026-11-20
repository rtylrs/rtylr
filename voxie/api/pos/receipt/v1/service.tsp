import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.receipt.v1;

/** ReceiptService defines POS receipt endpoints. */
@route("/pos/receipt/v1")
interface ReceiptService {
  /** Returns receipts, optionally filtered by date range. */
  @get
  @route("/receipts")
  listReceipts(...ListReceiptsRequest): ListReceiptsResponse;

  /** Emails a receipt to the given address. */
  @post
  @route("/receipts/{id}/email")
  emailReceipt(...EmailReceiptRequest): EmailReceiptResponse;
}
