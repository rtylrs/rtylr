import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.invoices.v1;

/** InvoicesService defines finance invoice endpoints. */
@route("/finance/invoices/v1")
interface InvoicesService {
  @get
  @route("/invoices")
  listInvoices(...ListInvoicesRequest): ListInvoicesResponse;

  @get
  @route("/invoices/billable-time")
  listBillableTimeEntries(
    ...ListBillableTimeEntriesRequest,
  ): ListBillableTimeEntriesResponse;

  @post
  @route("/invoices")
  createInvoice(...CreateInvoiceRequest): CreateInvoiceResponse;

  @get
  @route("/invoices/{id}")
  getInvoice(...GetInvoiceRequest): GetInvoiceResponse;

  @put
  @route("/invoices/{id}")
  updateInvoice(...UpdateInvoiceRequest): UpdateInvoiceResponse;

  @patch
  @route("/invoices/{id}/status")
  updateInvoiceStatus(
    ...UpdateInvoiceStatusRequest,
  ): UpdateInvoiceStatusResponse;

  /**
   * setInvoiceProject sets or clears (project_id 0) the Work project an
   * invoice bills, in any status except cancelled and void. Only the tag
   * changes; totals and status are never touched.
   */
  @put
  @route("/invoices/{id}/project")
  setInvoiceProject(...SetInvoiceProjectRequest): SetInvoiceProjectResponse;

  @delete
  @route("/invoices/{id}")
  deleteInvoice(...DeleteInvoiceRequest): DeleteInvoiceResponse;

  @post
  @route("/invoices/{id}/send")
  sendInvoice(...SendInvoiceRequest): SendInvoiceResponse;
}
