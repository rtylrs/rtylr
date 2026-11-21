import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.quote.v1;

// ── Quotes ──────────────────────────────────────────────────────────
model Quote {
  @encode(string)
  id: int64;

  /** Identifies the tenant that owns the resource. */
  @encode(string)
  owner_id: int64;

  customer_name: string;
  customer_email: string;

  /** draft, sent, accepted, rejected, expired, converted */
  status: string;

  lines: QuoteLine[];
  subtotal: float64;
  tax: float64;
  discount: float64;
  total: float64;
  valid_until: string;
  notes: string;

  /** When the record was created. */
  created_at: string;

  /** When the record was last updated. */
  updated_at: string;
}

model QuoteLine {
  description: string;
  quantity: int32;
  unit_price: float64;
  tax_rate: float64;
  total: float64;
}

model CreateQuoteRequest {
  @Onkt.required
  customer_name: string;

  customer_email: string;

  /** The initial lifecycle state - defaults to "draft" if empty. */
  status: string;

  @minItems(1)
  lines: QuoteLine[];

  @minValue(0)
  subtotal: float64;

  @minValue(0)
  tax: float64;

  @Onkt.required
  valid_until: string;

  @Onkt.required
  @minValue(0)
  discount?: float64;

  @minValue(0)
  total: float64;

  notes: string;
}

model CreateQuoteResponse {
  quote: Quote;
}

model GetQuoteRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model GetQuoteResponse {
  quote: Quote;
}

model ListQuotesRequest {
  /** The one-based page number to return. */
  @query("page")
  page: int64;

  /** The maximum number of quotes to return. */
  @query("page_size")
  page_size: int64;
}

model ListQuotesResponse {
  quotes: Quote[];

  /** The current result page. */
  page: int64;

  /** The number of records requested per page. */
  page_size: int64;

  /** Whether additional pages are available. */
  has_more: boolean;
}

model UpdateQuoteStatusBody {
  @Onkt.required
  status: string;
}

model UpdateQuoteStatusRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** Carries status update fields bound from the HTTP payload. */
  body: UpdateQuoteStatusBody;
}

model UpdateQuoteStatusResponse {
  message: string;
}

model ConvertQuoteRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model ConvertQuoteResponse {
  /** The identifier of the invoice created from this quote. */
  @encode(string)
  invoice_id: int64;

  message: string;
}
