import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.reconciliation.v1;

model UnmatchedBankTransaction {
  @encode(string)
  id: int64;

  @encode(string)
  bank_account_id: int64;

  @encode(string)
  amount_minor_units: int64;

  currency: string;
  occurred_at: string;
  description: string;
  external_id: string;
}

model UnmatchedInvoice {
  @encode(string)
  id: int64;

  invoice_number: string;
  client_name: string;

  @encode(string)
  amount_minor_units: int64;

  currency: string;
  issue_date: string;
  due_date: string;
}

model MatchSuggestion {
  @encode(string)
  bank_feed_id: int64;

  target_type: string;

  @encode(string)
  target_id: int64;

  @encode(string)
  amount_minor_units: int64;

  days_apart: int32;
  score: float64;
}

model GetOverviewRequest {}

model GetOverviewResponse {
  unmatched_bank_txns: UnmatchedBankTransaction[];
  unmatched_invoices: UnmatchedInvoice[];
  suggestions: MatchSuggestion[];

  /** Overview lists are bounded previews; true means more matching records exist. */
  bank_txns_truncated: boolean;

  invoices_truncated: boolean;
  expenses_truncated: boolean;
  preview_limit: int32;
}

model MatchRequest {
  @Onkt.required
  @encode(string)
  bank_feed_id?: int64;

  @Onkt.required
  target_type: string;

  @Onkt.required
  @encode(string)
  target_id?: int64;
}

model MatchResponse {
  @encode(string)
  id: int64;

  message: string;
}
