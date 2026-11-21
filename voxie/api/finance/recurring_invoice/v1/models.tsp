import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.recurring_invoice.v1;

// ── Recurring Invoices ──────────────────────────────────────────────
model RecurringInvoice {
  @encode(string)
  id: int64;

  /** The business owner for this recurring invoice. */
  @encode(string)
  owner_id: int64;

  /** The invoice used as the recurring invoice template. */
  @encode(string)
  template_invoice_id: int64;

  /** How often the recurring invoice should be generated. */
  frequency: string;

  /** The next generation date. */
  next_invoice_date: string;

  /** When recurring generation should stop; absent when open-ended. */
  end_date?: string;

  is_active: boolean;

  /** When the record was created. */
  created_at: string;
}

model CreateRecurringInvoiceRequest {
  /** The invoice used as the recurring invoice template. */
  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  template_invoice_id?: int64;

  /** How often the recurring invoice should be generated. */
  @Onkt.required
  frequency: string;

  /** The next generation date, formatted as YYYY-MM-DD. */
  @Onkt.required
  @pattern("^\\d{4}-\\d{2}-\\d{2}$")
  next_invoice_date: string;

  /** When recurring generation should stop; absent when open-ended. */
  end_date?: string;

  /** Whether the recurring invoice should start active. */
  is_active?: boolean;
}

model CreateRecurringInvoiceResponse {
  /** The recurring invoice that was created. */
  @encode(string)
  id: int64;

  /** The backend confirmation message. */
  message: string;
}

model GetRecurringInvoiceRequest {
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model GetRecurringInvoiceResponse {
  invoice: RecurringInvoice;
}

model ListRecurringInvoicesRequest {
  @query("page")
  page: int64;

  @query("page_size")
  page_size: int64;
}

model ListRecurringInvoicesResponse {
  /** The full set of recurring invoices for the owner. */
  recurring_invoices: RecurringInvoice[];

  page: int64;
  page_size: int64;
  has_more: boolean;
}

model PauseRecurringInvoiceRequest {
  /** The recurring invoice to pause. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model PauseRecurringInvoiceResponse {
  /** The backend confirmation message. */
  message: string;
}

model ResumeRecurringInvoiceRequest {
  /** The recurring invoice to resume. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model ResumeRecurringInvoiceResponse {
  /** The backend confirmation message. */
  message: string;
}

model UpdateRecurringInvoiceBody {
  /** The invoice used as the recurring invoice template. */
  @encode(string)
  template_invoice_id?: int64;

  /** How often the recurring invoice should be generated. */
  @Onkt.required
  frequency: string;

  /** The next generation date, formatted as YYYY-MM-DD. */
  @Onkt.required
  @pattern("^\\d{4}-\\d{2}-\\d{2}$")
  next_invoice_date: string;

  /** When recurring generation should stop; absent when open-ended. */
  end_date?: string;

  /** Whether the recurring invoice is active. */
  is_active?: boolean;
}

model UpdateRecurringInvoiceRequest {
  /** The recurring invoice to update. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** Carries updatable recurring invoice fields bound from the HTTP payload. */
  body: UpdateRecurringInvoiceBody;
}

model UpdateRecurringInvoiceResponse {
  /** The backend confirmation message. */
  message: string;
}
