import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.invoices.payments.v1;

/** InvoicePaymentService records payments against invoices. */
@route("/finance/invoices/payments/v1")
interface InvoicePaymentService {
  @post
  @route("/invoices/{invoice_id}/payment")
  recordPayment(...RecordPaymentRequest): RecordPaymentResponse;
}
