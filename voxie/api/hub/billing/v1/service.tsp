import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.billing.v1;

/** BillingService defines hub billing endpoints. */
@route("/hub/billing/v1")
interface BillingService {
  /** Returns billing summary for the active business. */
  @get
  @route("/summary")
  getBillingSummary(...GetBillingSummaryRequest): GetBillingSummaryResponse;

  /** Returns billing invoices. */
  @get
  @route("/invoices")
  listInvoices(...ListInvoicesRequest): ListInvoicesResponse;
}
