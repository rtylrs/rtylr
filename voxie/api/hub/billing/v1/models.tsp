import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.billing.v1;

model BillingSummary {
  plan: string;
  status: string;
  current_usage: float64;
  usage_limit: float64;
  next_billing_date: string;
}

model Invoice {
  @encode(string)
  id: int64;

  amount: float64;
  currency: string;
  status: string;
  issued_at: string;
  paid_at: string;
}

model GetBillingSummaryRequest {}

model GetBillingSummaryResponse {
  summary: BillingSummary;
}

model ListInvoicesRequest {}

model ListInvoicesResponse {
  invoices: Invoice[];
}
