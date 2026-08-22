# 06. Proposed contract changes

Draft `.onk` for every proposal in the "needs a contract field" bucket of
`04-gaps.md`, plus the two aggregate endpoints from the "needs a query or
endpoint" bucket.

**Nothing under `packages/voxie` or `~/Development/voxie` was edited.** These
are drafts for review. Contracts land in Voxie first, then the packages
submodule pointer moves, then this repo's pointer moves (root `CLAUDE.md`,
submodules section).

## House style these drafts follow

Copied from the surrounding contracts, not invented:

- `///` doc comment on every message and every field. Optional filters are
  documented as "absent to include all X", which is the phrasing already used
  in `packages/voxie/api/hr/employee/v1/models.onk:108`.
- `int64` for ids, `@nullable` with a `?` type for optional fields,
  `@required` and `@gt(0)` where a value must be present and positive.
- Query parameters carry `@query("name")`.
- Pagination is `limit: int32 @query("limit")` and
  `offset: int32 @query("offset")`, both non-nullable, exactly as in
  `ListEmployeesRequest` (`packages/voxie/api/hr/employee/v1/models.onk:114`).

## Two wire facts every draft respects

1. **OneKit `int64` is a string on the wire.** The generated Go marshals via
   `strconv.FormatInt` (`~/Development/voxie/gen/go/finance/bill/v1/types.gen.go:32`).
   A JSON number fails to unmarshal and the string `"0"` fails `@gt(0)`. So no
   draft below uses `@gt(0)` on a field that legitimately means "unset";
   those are `int64?` with `@nullable` instead.
2. **`@nullable` optional ids, not zero sentinels.** `finance.Bill.supplier_id`
   is the cautionary example: `@required @gt(0)` on a link the product cannot
   always supply is what pushed the frontend into fabricating a value
   (`apps/finance/src/features/bills/Bills.tsx:59`).

---

## 1. Add range and dimension filters to HR time entries

Serves `FLOW->FIN-01` (unbilled billable hours by project).

`ListTimeEntriesRequest` today filters by one employee and one exact date
(`packages/voxie/api/hr/time_entry/v1/models.onk:71`), while its response
already reports `limit` and `offset` the caller cannot set.

```onk
/// ListTimeEntriesRequest selects logged time. Every filter is optional and
/// they combine with AND. The response is always a page: callers that need
/// everything must follow offset until total is reached.
message ListTimeEntriesRequest {
  /// Filter by employee; absent to include all employees.
  employee_id: int64? @query("employee_id") @nullable
  /// Filter by exact date; absent to include all dates. Ignored when
  /// from_date or to_date is supplied.
  date: string? @query("date") @nullable
  /// Include entries dated on or after this day (ISO-8601, e.g. "2026-04-01");
  /// absent to leave the range open at the start.
  from_date: string? @query("from_date") @nullable
  /// Include entries dated on or before this day; absent to leave the range
  /// open at the end.
  to_date: string? @query("to_date") @nullable
  /// Filter by the Flow project the time was logged against; absent to include
  /// all projects and entries with no project.
  project_id: int64? @query("project_id") @nullable
  /// Filter by billability; absent to include both billable and
  /// non-billable time.
  is_billable: bool? @query("is_billable") @nullable
  /// Filter by workflow status (draft | submitted | approved | rejected);
  /// absent to include all statuses.
  status: string? @query("status") @nullable
  /// Maximum number of records to return.
  limit: int32 @query("limit")
  /// Number of records to skip for pagination.
  offset: int32 @query("offset")
}
```

`date` is kept for compatibility: `apps/hr` already calls this endpoint, and
removing a filter would break it. The doc comment states the precedence rather
than leaving two overlapping filters to interact silently.

**Note there is already an uncontracted route that returns similar data.**
`POST /hr/time-entries/export-billable` returns per-entry `project_id`,
`task_id`, `employee_id`, `duration_minutes` and `is_billable`
(`~/Development/onion/internal/modules/hr_time/internal/domain/types.go:90`),
and no frontend calls it.

**Correction (#855).** This paragraph originally recommended deleting that
route, on the grounds that it "takes no parameters, returns the whole
business, and allocates a batch id as a side effect of a GET". Two of those
three are wrong:

- **It is a POST**, not a GET. `internal/platform/server/http/register.go:83` registers it through
  `RegisterPOSTWithValidation`, which calls `router.Post`. A mutation behind
  a POST is the correct verb, not a defect.
- **It is bounded.** `ListApprovedBillableEntries` carries `LIMIT 500` and a
  comment explaining that the bound stands until the contract can expose a
  page cursor. It does not return the whole business.

What survives is narrower and still worth acting on: it takes no parameters,
so it cannot be scoped to a project or a period; nothing calls it; and its
one side effect, `export_batch_id`, is written by `MarkEntryExported` and
**read by nothing but its own test**. So the batch marker accomplishes
nothing today.

The filters above do not replace it. They are a read; the export also marks
entries. Deleting it would remove a workflow, so it is left in place and the
narrower finding is recorded rather than acted on.

## 2. Add a date range to HR leave requests

Serves `HR->FLOW-02` (who is on approved leave during this cycle).

```onk
/// ListLeaveRequestsRequest selects leave requests. Filters combine with AND.
/// A request overlaps the supplied range if it starts on or before to_date and
/// ends on or after from_date.
message ListLeaveRequestsRequest {
  /// Filter by employee; absent to include all employees.
  employee_id: int64? @query("employee_id") @nullable
  /// Filter by status; absent to include all statuses.
  status: string? @query("status") @nullable
  /// Include requests overlapping this day or later (ISO-8601); absent to
  /// leave the range open at the start.
  from_date: string? @query("from_date") @nullable
  /// Include requests overlapping this day or earlier; absent to leave the
  /// range open at the end.
  to_date: string? @query("to_date") @nullable
  /// Maximum number of records to return.
  limit: int32 @query("limit")
  /// Number of records to skip for pagination.
  offset: int32 @query("offset")
}
```

The overlap semantics are spelled out because "leave between these dates" is
ambiguous for multi-day leave, and a consumer that guesses wrong silently
misses the person who is away across the boundary.

## 3. Give an invoice a customer

Serves `CRM->FIN-02`, and unblocks `FIN->CRM-01`.

The existing free-text client fields stay. An invoice is a historical document:
if a customer is later renamed, last year's invoice must still show the name it
was issued under. So this adds a link without removing the snapshot.

```onk
/// Invoice is a sales invoice issued to a customer.
///
/// customer_id links to the CRM customer this was issued to. The client_*
/// fields remain the values printed on the document at issue time and are
/// deliberately not re-derived from the customer record: renaming a customer
/// must not alter an invoice already sent.
message Invoice {
  /// customer_id is the CRM customer this invoice was issued to; absent on
  /// invoices created before customers were linked, and on one-off invoices
  /// for a party that is not a customer record.
  customer_id: int64? @nullable
}

/// CreateInvoiceRequest issues a new invoice.
message CreateInvoiceRequest {
  /// customer_id is the CRM customer to issue to. When supplied the server
  /// snapshots that customer's name, email, address and tax id onto the
  /// invoice, and the client_* fields below may be omitted. When absent the
  /// client_* fields are required, which is the existing behaviour.
  customer_id: int64? @nullable
}

/// ListInvoicesRequest selects invoices.
message ListInvoicesRequest {
  /// Filter by customer; absent to include invoices for all customers and
  /// invoices with no customer link.
  customer_id: int64? @query("customer_id") @nullable
}
```

Shown as field additions to existing messages rather than full message bodies,
because the surrounding fields are unchanged; the real diff is three fields.
`client_name` and `client_email` are currently `@required`
(`packages/voxie/api/finance/invoices/v1/models.onk:147`), so making them
conditionally optional is a validator change on the server, not just a contract
edit, and the migration note below covers it.

## 4. Link a POS menu item to its ERP product

Serves `ERP->POS-01`.

```onk
/// MenuItem is a sellable item on the POS menu.
///
/// product_id links the item to the ERP product it is sold from, which is what
/// makes cost and therefore margin knowable. It is nullable because a menu item
/// may be a service or a composite with no single stock-keeping product; when
/// it is absent, stock and cost for this item are unknown rather than zero.
message MenuItem {
  /// product_id is the ERP product this item is sold from; absent for items
  /// that are not stocked as a single product.
  product_id: int64? @nullable
  /// unit_cost is the cost of one unit at the time of the read, copied from
  /// the linked ERP product. Absent when product_id is absent. Reported for
  /// display and margin only: the authoritative value is the ERP product's.
  unit_cost: float64? @nullable
}
```

**`stock` is deliberately not touched here.** `pos.MenuItem.stock` and
`erp.Product.stock_quantity` are two independent numbers today
(`packages/voxie/api/pos/menu_item/v1/models.onk:28`,
`packages/voxie/api/erp/product/v1/models.onk:15`). Adding `product_id` makes
the disagreement _visible_, which is the necessary first step, but deciding
which one wins is a data-ownership decision with a migration behind it and it
belongs to the ERP and POS owners jointly. The dossier records that as the
follow-up rather than pretending a contract field settles it.

## 5. Link a bill to its purchase order

Serves `FIN->ERP-02`.

```onk
/// Bill is an accounts-payable bill owed to a supplier.
message Bill {
  /// purchase_order_id is the ERP purchase order this bill settles; absent for
  /// bills that did not originate from a purchase order.
  purchase_order_id: int64? @nullable
}
```

Direction matters here. ERP owns purchase order status
(`PurchaseOrderService.updatePurchaseOrderStatus`,
`packages/voxie/api/erp/purchase_order/v1/service.onk:11`), so Finance must not
write it. The bill carries the reference; ERP reads it or receives an event.

## 6. Relax `finance.Bill.supplier_id`

Serves `ERP->FIN-01`. This is a **correction, not a feature**.

```onk
/// Bill is an accounts-payable bill owed to a supplier.
message Bill {
  /// supplier_id is the ERP supplier this bill is owed to; absent on bills
  /// entered before suppliers were linked, and on one-off bills for a party
  /// with no supplier record. Callers must not invent a value: a bill with no
  /// known supplier records none.
  supplier_id: int64? @nullable
  /// supplier_name is the supplier's name as it stood when the bill was
  /// entered, so a bill list reads correctly without a second request and a
  /// renamed supplier does not rewrite history.
  supplier_name: string
}
```

`@required @gt(0)` on a link the product cannot always supply is what caused
the frontend to slugify a typed name into an integer field. Making it nullable
and adding the denormalised name removes both the fabrication and the extra
round trip. **`CreateBillRequest.supplier_id` must lose `@required @gt(0)` at
the same time** (`packages/voxie/api/finance/bill/v1/models.onk:29`), or the
frontend still cannot submit a bill without one.

## 7. Reference a customer from a support customer

Serves `SUP->CRM-01`.

```onk
/// Customer is a support-side customer record, aggregating that party's
/// ticket history.
///
/// customer_id links to the CRM customer record for the same party. It is
/// nullable because support conversations arrive by email from people who may
/// not be customers yet; matching on email alone is not reliable enough to
/// assert the link, so the server sets it only on a confident match or when a
/// human confirms one.
message Customer {
  /// customer_id is the CRM customer this support record refers to; absent
  /// when the party has not been matched to one.
  customer_id: int64? @nullable
}
```

## 8. Two new aggregate endpoints

These are the "needs a query or endpoint" items where the aggregation must
happen on the server because the inputs are more sensitive than the output.

### 8a. Labour cost, owned by HR

Serves `HR->FLOW-01` and `HR->INS-01` with **one** endpoint, so Flow and
Insights cannot report different numbers for the same question.

```onk
package hr.cost.v1

/// ProjectLabourCost is the cost of the time logged against one project over a
/// period. It reports money in aggregate only: per-person rates are never
/// returned, so a caller who may see a project's cost does not thereby gain
/// sight of any individual's pay.
message ProjectLabourCost {
  /// project_id is the Flow project the time was logged against.
  project_id: int64
  /// billable_minutes is logged time marked billable to a client.
  billable_minutes: int32
  /// non_billable_minutes is logged time not marked billable.
  non_billable_minutes: int32
  /// cost_minor_units is the labour cost of all logged time in the smallest
  /// unit of the reporting currency, so no rounding happens before the
  /// caller sees it.
  cost_minor_units: int64
  /// currency is the ISO-4217 code the cost is expressed in.
  currency: string
  /// contributor_count is how many people logged time. Reported so a caller
  /// can tell a one-person total (from which a rate is trivially derivable)
  /// from a genuine aggregate.
  contributor_count: int32
  /// rate_coverage is the fraction of logged minutes whose employee had a pay
  /// rate on record, between 0 and 1. A cost with coverage below 1 understates
  /// the true figure and must be labelled as partial where it is displayed.
  rate_coverage: float64
}

/// GetProjectLabourCostRequest asks for labour cost per project over a period.
message GetProjectLabourCostRequest {
  /// project_ids selects the projects to report on. Required, and bounded, so
  /// this cannot become a whole-business scan.
  project_ids: int64[] @required @min_items(1)
  /// from_date is the first day included (ISO-8601).
  from_date: string @required
  /// to_date is the last day included.
  to_date: string @required
}

/// GetProjectLabourCostResponse returns one row per requested project that had
/// logged time. Projects with no time in the period are omitted rather than
/// returned as zero, so the caller can tell "no time logged" from "no cost".
message GetProjectLabourCostResponse {
  /// costs is one entry per project with logged time in the period.
  costs: ProjectLabourCost[]
}
```

Three deliberate choices worth defending at review:

- **`cost_minor_units: int64`, not `float64`.** Payroll already works in
  integer cents internally
  (`~/Development/onion/internal/modules/hr_payroll/internal/application/service.go:236`)
  while Finance's contracts use `float64`. Returning minor units keeps the
  rounding decision with the caller instead of applying it twice.
- **`rate_coverage`.** An employee with no `hourly_rate` contributes hours and
  no cost. Without this field the endpoint would silently understate, which is
  exactly the class of quietly-wrong number this audit exists to remove.
- **`contributor_count`.** A single-contributor project's total divided by its
  hours _is_ that person's rate. The field does not prevent that inference, it
  makes it visible so the consuming screen can suppress the figure. The
  alternative, refusing to answer for one-person projects, breaks the feature
  for small tenants.

### 8b. Margin per product, owned by Insights

Serves `ERP->INS-01`. Insights owns it because Insights owns the revenue side;
it reads ERP cost server-side rather than the Insights app pulling the
catalogue.

```onk
/// ProductMargin pairs revenue Insights already computes with cost from the
/// ERP product record, so margin is answerable without either side shipping
/// half of it.
message ProductMargin {
  /// product_id is the ERP product.
  product_id: int64
  /// name is the product name at the time of the report.
  name: string
  /// units_sold is the quantity sold in the period.
  units_sold: float64
  /// revenue is gross sales value in the period.
  revenue: float64
  /// cost is units_sold multiplied by the product's cost price. Absent when
  /// the product has no cost price on record, which is distinct from a cost
  /// of zero.
  cost: float64? @nullable
  /// currency is the ISO-4217 code both revenue and cost are expressed in.
  currency: string
}
```

## Migration notes

Every draft above is additive except one, and additive is the point: a nullable
field on an existing message needs no backfill and cannot break an existing
caller.

| Draft                           | Migration                                                                                                                                                                                                                                                                                                                   |
| ------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1, 2 (filters)                  | None. New optional query parameters; existing callers unaffected.                                                                                                                                                                                                                                                           |
| 3 (invoice `customer_id`)       | Nullable column, no backfill. Existing invoices keep their strings and report no customer. **Not additive on the server:** `client_name` and `client_email` stop being unconditionally required, so the validator must accept either a `customer_id` or the strings. Old clients that always send the strings keep working. |
| 4 (menu item `product_id`)      | Nullable column. Backfilling by SKU match is tempting and should be a separate, reviewed migration with a report of unmatched items, not a side effect.                                                                                                                                                                     |
| 5 (bill `purchase_order_id`)    | Nullable column, no backfill.                                                                                                                                                                                                                                                                                               |
| 6 (bill `supplier_id` nullable) | **Loosening a constraint, so safe in the permissive direction.** Existing rows all have a value. Any row whose `supplier_id` does not exist in `erp_suppliers` is a fabricated id from the current UI and should be reported, not silently nulled.                                                                          |
| 7 (support `customer_id`)       | Nullable column, no backfill. An email-match backfill would assert links the product cannot stand behind.                                                                                                                                                                                                                   |
| 8a, 8b (new endpoints)          | No schema change. Both are reads over existing tables.                                                                                                                                                                                                                                                                      |

## What is deliberately not proposed

- **No `project_id` on Finance expenses.** It would make `FIN->FLOW-01`
  (project budget) possible, and that proposal was rejected: project cost
  allocation is a Finance product decision, not a plumbing gap
  (`11-rejected.md`).
- **No party or contact entity.** `IDENTITY-01` needs one, and drafting it here
  would be the audit designing the product's central data model in an appendix.
  `07-plumbing.md` specifies the shape and `08-roadmap.md` puts it in its own
  wave with the connections it unblocks listed as its return.
- **No `hr_category_id` consumer.** The field already exists and is dead
  (`02-entity-graph.md`). Proposing more contract surface around a link nothing
  reads would repeat the mistake that made it a useful warning.

## Round two: drafts landed in voxie

Round two moves these drafts from prose to a real branch. **voxie branch
`docs/cross-app-round2`** (cut from `main` at `847a34c8`) carries one commit
per draft; `gen/` is untouched and the vendored copy under `packages/voxie` was
not synced.

| Round-one draft              | Voxie state on the branch                                                                | Notes                                                                                                                                                                                                                                                                                                                                                        |
| ---------------------------- | ---------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1 time-entry filters         | **Landed upstream before round two opened**: voxie #46, onion #783                       | Excluded from the round-two PR; issue #855 gets a comment pointing at the landing commits                                                                                                                                                                                                                                                                    |
| 2 leave range + pagination   | drafted (`api/hr/leave/v1/models.onk`)                                                   | Adds from_date/to_date plus limit/offset; overlap semantics documented                                                                                                                                                                                                                                                                                       |
| 3 invoice customer_id        | drafted (`api/finance/invoices/v1/models.onk`)                                           | InvoiceDetail, CreateInvoiceRequest (client_name/client_email relaxed to optional-with-customer rule), ListInvoicesRequest                                                                                                                                                                                                                                   |
| 4 menu item product link     | drafted (`api/pos/menu_item/v1/models.onk`)                                              | product_id plus unit_cost snapshot                                                                                                                                                                                                                                                                                                                           |
| 5 bill purchase_order_id     | drafted with 6 (`api/finance/bill/v1/models.onk`)                                        |                                                                                                                                                                                                                                                                                                                                                              |
| 6 bill supplier relax        | drafted with 5                                                                           | supplier_id nullable, required supplier_name snapshot, ListBillsRequest gains supplier_id/purchase_order_id filters                                                                                                                                                                                                                                          |
| 7 support customer reference | drafted (`api/support/customer/v1/models.onk`)                                           |                                                                                                                                                                                                                                                                                                                                                              |
| 8a labour-cost aggregate     | **Superseded in part and re-drafted** (`api/flow/project/v1/service.onk` + `models.onk`) | The port exists (#743). The draft binds it to an RPC on Flow's project service rather than proposing a parallel hr.cost.v1 message set. Field names follow the live implementation (cost_minor_units, costed_minutes, uncosted_minutes, uncosted_employee_count); billable-only scope is documented rather than hidden. HR->INS-01 can consume the same RPC. |
| 8b per-product margin        | **Narrowed and re-drafted** (`api/insights/product/v1/models.onk`)                       | A margins distribution RPC already exists and is implemented; the draft adds an additive per-product rows list carrying cost_minor_units                                                                                                                                                                                                                     |

Two wire rules the round-one text established are preserved verbatim in every
new draft: int64 ids travel as strings, so no draft puts @gt(0) on a field
whose legitimate value is absence; and every foreign-id write documents the
server-side same-business verification requirement.
