import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.recurring_invoice.v1;

/** RecurringInvoiceService defines recurring invoice endpoints. */
@route("/finance/recurring-invoice/v1")
interface RecurringInvoiceService {
  @get
  @route("/recurring-invoices")
  listRecurringInvoices(
    ...ListRecurringInvoicesRequest,
  ): ListRecurringInvoicesResponse;

  @post
  @route("/recurring-invoices")
  createRecurringInvoice(
    ...CreateRecurringInvoiceRequest,
  ): CreateRecurringInvoiceResponse;

  @get
  @route("/recurring-invoices/{id}")
  getRecurringInvoice(
    ...GetRecurringInvoiceRequest,
  ): GetRecurringInvoiceResponse;

  @put
  @route("/recurring-invoices/{id}")
  updateRecurringInvoice(
    ...UpdateRecurringInvoiceRequest,
  ): UpdateRecurringInvoiceResponse;

  @patch
  @route("/recurring-invoices/{id}/pause")
  pauseRecurringInvoice(
    ...PauseRecurringInvoiceRequest,
  ): PauseRecurringInvoiceResponse;

  @patch
  @route("/recurring-invoices/{id}/resume")
  resumeRecurringInvoice(
    ...ResumeRecurringInvoiceRequest,
  ): ResumeRecurringInvoiceResponse;
}
