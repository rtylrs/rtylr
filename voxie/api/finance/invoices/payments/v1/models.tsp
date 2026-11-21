import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using finance.invoices.v1;

namespace finance.invoices.payments.v1;

model RecordPaymentRequest {
  /** invoice_id is the invoice id value for this contract. */
  @minValueExclusive(0)
  @encode(string)
  invoice_id: int64;

  /** amount is the amount value for this contract. */
  @minValue(0)
  amount: float64;

  /** currency is the currency value for this contract. */
  @Onkt.required
  currency: string;

  /** payment_date is the payment date value for this contract. */
  @Onkt.required
  payment_date: string;

  /** method is the method value for this contract. */
  @Onkt.required
  method: string;

  /** reference is the reference value for this contract. */
  @Onkt.required
  reference: string;

  /** notes is the notes value for this contract. */
  notes: string;
}

model RecordPaymentResponse {
  /** payment is the payment value for this contract. */
  payment: RecordPaymentResponsePayload;

  /** message is a human-readable result message. */
  message: string;
}
