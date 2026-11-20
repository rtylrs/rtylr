import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.invoices.v1;

/** InvoiceItem is a single line item on an invoice. */
model InvoiceItem {
  /** description is the description value for this contract. */
  description: string;

  /** quantity is the quantity value for this contract. */
  quantity: float64;

  /** unit is the unit value for this contract. */
  unit: string;

  /** unit_price is the unit price value for this contract. */
  unit_price: float64;

  /** tax_rate is the tax rate value for this contract. */
  tax_rate: float64;

  /** discount_pct is the discount pct value for this contract. */
  discount_pct: float64;

  /** sort_order is the sort order value for this contract. */
  @encode(string)
  sort_order: int64;

  /** total is the total number of matching records. */
  total: float64;
}

/** InvoiceSummary is a compact invoice row for list views. */
model InvoiceSummary {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** invoice_number is the invoice number value for this contract. */
  invoice_number: string;

  /** client_name is the client name value for this contract. */
  client_name: string;

  /** status is the current lifecycle state. */
  status: string;

  /** issue_date is the issue date value for this contract. */
  issue_date: string;

  /** due_date is the due date value for this contract. */
  due_date: string;

  /** paid_date is the paid date value for this contract. */
  paid_date: string;

  /** currency is the currency value for this contract. */
  currency: string;

  /** total is the total number of matching records. */
  total: float64;

  /** amount_paid is the amount paid value for this contract. */
  amount_paid: float64;

  /** amount_due is the amount due value for this contract. */
  amount_due: float64;

  /** created_at is when the record was created. */
  created_at: string;

  /**
   * project_id is the Work project this invoice bills; absent when the
   * invoice is not tagged with a project.
   */
  @encode(string)
  project_id?: int64;
}

/**
 * BillableTimeEntry is approved billable HR time that has not been linked to
 * a finance invoice yet.
 */
model BillableTimeEntry {
  /** id is the unique identifier for the time entry. */
  @encode(string)
  id: int64;

  /** employee_id identifies the employee who logged the time. */
  @encode(string)
  employee_id: int64;

  /** project_id identifies the optional Flow project. */
  @encode(string)
  project_id?: int64;

  /** task_id identifies the optional Flow task. */
  @encode(string)
  task_id?: int64;

  /** date is the day on which the time was logged. */
  date: string;

  /** duration_minutes is the approved billable duration. */
  @encode(string)
  duration_minutes: int64;

  /** description is the time entry description. */
  description: string;

  /** is_billable is always true for this response. */
  is_billable: boolean;

  /** status is always approved for this response. */
  status: string;

  /** source identifies the system that recorded the time. */
  source: string;
}

/** InvoiceDetail is the full invoice record. */
model InvoiceDetail {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** owner_id identifies the tenant that owns the resource. */
  @encode(string)
  owner_id: int64;

  /**
   * customer_id links to the CRM customer this invoice was issued to; absent
   * on invoices created before customers were linked, and on one-off invoices
   * for a party that is not a customer record. The client_* fields stay the
   * values printed on the document at issue time and are deliberately not
   * re-derived from the customer record.
   */
  @encode(string)
  customer_id?: int64;

  /**
   * project_id is the Work project this invoice bills, so per-project
   * invoiced, paid and overdue totals count it; absent when the invoice is
   * not tagged with a project.
   */
  @encode(string)
  project_id?: int64;

  /** invoice_number is the invoice number value for this contract. */
  invoice_number: string;

  /** client_name is the client name value for this contract. */
  client_name: string;

  /** client_email is the client email value for this contract. */
  client_email: string;

  /** client_address is the client address value for this contract. */
  client_address: string;

  /** client_tax_id is the client tax id value for this contract. */
  client_tax_id: string;

  /** status is the current lifecycle state. */
  status: string;

  /** issue_date is the issue date value for this contract. */
  issue_date: string;

  /** due_date is the due date value for this contract. */
  due_date: string;

  /** paid_date is the paid date value for this contract. */
  paid_date: string;

  /** currency is the currency value for this contract. */
  currency: string;

  /** subtotal is the subtotal value for this contract. */
  subtotal: float64;

  /** discount_type is the discount type value for this contract. */
  discount_type: string;

  /** discount_value is the discount value value for this contract. */
  discount_value: float64;

  /** discount_amount is the discount amount value for this contract. */
  discount_amount: float64;

  /** tax_rate is the tax rate value for this contract. */
  tax_rate: float64;

  /** tax_amount is the tax amount value for this contract. */
  tax_amount: float64;

  /** total is the total number of matching records. */
  total: float64;

  /** amount_paid is the amount paid value for this contract. */
  amount_paid: float64;

  /** amount_due is the amount due value for this contract. */
  amount_due: float64;

  /** notes is the notes value for this contract. */
  notes: string;

  /** terms is the terms value for this contract. */
  terms: string;

  /** created_at is when the record was created. */
  created_at: string;

  /** updated_at is when the record was last updated. */
  updated_at: string;

  /** items is the items value for this contract. */
  items: InvoiceItem[];
}

/** Payment is a recorded payment applied to an invoice. */
model Payment {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** amount is the amount value for this contract. */
  amount: float64;

  /** currency is the currency value for this contract. */
  currency: string;

  /** payment_date is the payment date value for this contract. */
  payment_date: string;

  /** method is the method value for this contract. */
  method: string;

  /** reference is the reference value for this contract. */
  reference: string;

  /** notes is the notes value for this contract. */
  notes: string;
}

/** RecordPaymentResponsePayload describes a payment just recorded against an invoice. */
model RecordPaymentResponsePayload {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** invoice_id is the invoice id value for this contract. */
  @encode(string)
  invoice_id: int64;

  /** amount is the amount value for this contract. */
  amount: float64;

  /** currency is the currency value for this contract. */
  currency: string;

  /** payment_date is the payment date value for this contract. */
  payment_date: string;

  /** method is the method value for this contract. */
  method: string;

  /** reference is the reference value for this contract. */
  reference: string;

  /** notes is the notes value for this contract. */
  notes: string;
}

model CreateInvoiceRequest {
  /** client_name is the client name value for this contract. */
  client_name: string;

  /** client_email is the client email value for this contract. */
  @format("email")
  client_email: string;

  /** client_address is the client address value for this contract. */
  client_address: string;

  /** client_tax_id is the client tax id value for this contract. */
  client_tax_id: string;

  /**
   * customer_id is the CRM customer to issue to. When supplied the server
   * snapshots that customer's name, email, address and tax id onto the
   * invoice and the client_* fields above may be omitted. When absent the
   * server requires client_name and client_email, which is the existing
   * behaviour. The server must verify the referenced customer belongs to
   * the caller's business before accepting the link.
   */
  @encode(string)
  customer_id?: int64;

  /**
   * project_id tags the invoice with the Work project it bills. Absent or 0
   * leaves the invoice untagged. The server rejects a project that does not
   * belong to the caller's business.
   */
  @encode(string)
  project_id?: int64;

  /** issue_date is the issue date value for this contract. */
  @Onkt.required
  issue_date: string;

  /** due_date is the due date value for this contract. */
  @Onkt.required
  due_date: string;

  /** currency is the currency value for this contract. */
  @Onkt.required
  currency: string;

  /** discount_type is the discount type value for this contract. */
  @Onkt.required
  discount_type: string;

  /** discount_value is the discount value value for this contract. */
  @minValue(0)
  discount_value: float64;

  /** tax_rate is the tax rate value for this contract. */
  @minValue(0)
  tax_rate: float64;

  /** notes is the notes value for this contract. */
  notes: string;

  /** terms is the terms value for this contract. */
  terms: string;

  /** items is the items value for this contract. */
  @minItems(1)
  items: InvoiceItem[];
}

model CreateInvoiceResponse {
  /** invoice is the invoice value for this contract. */
  invoice: InvoiceDetail;
}

model DeleteInvoiceRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteInvoiceResponse {
  /** message is a human-readable result message. */
  message: string;
}

model GetInvoiceRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model GetInvoiceResponse {
  /** invoice is the invoice value for this contract. */
  invoice: InvoiceDetail;

  /** payments is the payments value for this contract. */
  payments: Payment[];
}

model ListInvoicesRequest {
  /** status filters invoices by lifecycle state. */
  @query("status")
  status: string;

  /** search filters by customer or invoice number. */
  @query("search")
  search: string;

  /**
   * customer_id filters by the linked CRM customer; absent to include
   * invoices for all customers and invoices with no customer link.
   */
  @query("customer_id")
  @encode(string)
  customer_id?: int64;

  /**
   * project_id filters by the tagged Work project; absent to include
   * invoices for every project and untagged invoices.
   */
  @query("project_id")
  @encode(string)
  project_id?: int64;

  /** date_from is the inclusive start date filter. */
  @query("date_from")
  date_from: string;

  /** date_to is the inclusive end date filter. */
  @query("date_to")
  date_to: string;

  /** page is the one-based page number to return. */
  @query("page")
  page: int64;

  /** page_size is the maximum number of invoices to return. */
  @query("page_size")
  page_size: int64;
}

model ListInvoicesResponse {
  /** invoices is the invoices value for this contract. */
  invoices: InvoiceSummary[];

  /** page is the current result page. */
  page: int64;

  /** page_size is the number of records requested per page. */
  page_size: int64;

  /** has_more indicates whether additional pages are available. */
  has_more: boolean;
}

model ListBillableTimeEntriesRequest {
  /** page is the one-based page number to return. */
  @query("page")
  page: int64;

  /** page_size is the maximum number of time entries to return. */
  @query("page_size")
  page_size: int64;
}

model ListBillableTimeEntriesResponse {
  /** entries contains approved billable time not yet linked to an invoice. */
  entries: BillableTimeEntry[];

  /** total is the number of matching time entries. */
  total: int64;

  /** total_minutes is the sum of matching entry durations. */
  total_minutes: int64;

  /** page is the current result page. */
  page: int64;

  /** page_size is the number of records requested per page. */
  page_size: int64;

  /** has_more indicates whether additional pages are available. */
  has_more: boolean;
}

model SendInvoiceRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model SendInvoiceResponse {
  /** message is a human-readable result message. */
  message: string;
}

model UpdateInvoiceBody {
  /** client_name is the client name value for this contract. */
  @Onkt.required
  client_name: string;

  /** client_email is the client email value for this contract. */
  @Onkt.required
  @format("email")
  client_email: string;

  /** client_address is the client address value for this contract. */
  client_address: string;

  /** client_tax_id is the client tax id value for this contract. */
  client_tax_id: string;

  /**
   * project_id re-tags the draft with a Work project. Absent keeps the
   * current tag, 0 clears it, and any other value must be a project in the
   * caller's business. Optional so an omitted field can never be read as a
   * clear, and 0 is safe as the clear value because no project has id 0.
   */
  @encode(string)
  project_id?: int64;

  /** issue_date is the issue date value for this contract. */
  @Onkt.required
  issue_date: string;

  /** due_date is the due date value for this contract. */
  @Onkt.required
  due_date: string;

  /** currency is the currency value for this contract. */
  @Onkt.required
  currency: string;

  /** discount_type is the discount type value for this contract. */
  @Onkt.required
  discount_type: string;

  /** discount_value is the discount value value for this contract. */
  @minValue(0)
  discount_value: float64;

  /** tax_rate is the tax rate value for this contract. */
  @minValue(0)
  tax_rate: float64;

  /** notes is the notes value for this contract. */
  notes: string;

  /** terms is the terms value for this contract. */
  terms: string;

  /** items is the items value for this contract. */
  @minItems(1)
  items: InvoiceItem[];
}

model UpdateInvoiceRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries updatable invoice fields bound from the HTTP payload. */
  @body
  body: UpdateInvoiceBody;
}

model UpdateInvoiceResponse {
  /** invoice is the invoice value for this contract. */
  invoice: InvoiceDetail;
}

model UpdateInvoiceStatusBody {
  /** status is the current lifecycle state. */
  @Onkt.required
  status: string;
}

model UpdateInvoiceStatusRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries status update fields bound from the HTTP payload. */
  @body
  body: UpdateInvoiceStatusBody;
}

model UpdateInvoiceStatusResponse {
  /** message is a human-readable result message. */
  message: string;
}

model SetInvoiceProjectBody {
  /**
   * project_id is the Work project to tag the invoice with, or 0 to clear
   * the tag. Required: an absent field is refused rather than read as a
   * clear. Any other value must be a project in the caller's business.
   */
  @minValue(0)
  @encode(string)
  project_id?: int64;
}

model SetInvoiceProjectRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries the project tag bound from the HTTP payload. */
  @body
  body: SetInvoiceProjectBody;
}

model SetInvoiceProjectResponse {
  /**
   * invoice is the invoice after the tag change. Only project_id (and
   * updated_at) differ from before; totals and status are never touched.
   */
  invoice: InvoiceDetail;
}
