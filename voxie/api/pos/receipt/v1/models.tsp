import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.receipt.v1;

/** Receipt is a completed order receipt. */
model Receipt {
  @encode(string)
  id: int64;

  order_number: string;
  customer_name: string;
  total: float64;
  payment_method: string;
  paid_at: string;
}

model EmailReceiptRequest {
  /** id is the unique identifier for the receipt. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** Recipient email address for the receipt. */
  @Onkt.required
  email: string;
}

model EmailReceiptResponse {
  message: string;
}

model ListReceiptsRequest {
  @query("from")
  from_date: string;

  @query("to")
  to_date: string;
}

model ListReceiptsResponse {
  receipts: Receipt[];
}
