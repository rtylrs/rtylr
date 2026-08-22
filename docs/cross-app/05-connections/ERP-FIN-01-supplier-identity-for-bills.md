# ERP->FIN-01: supplier identity from erp to finance

- **Class:** MUST
- **Necessity:** Essential
- **Producer:** erp.supplier.v1 (`SupplierService.listSuppliers`)
- **Consumer:** finance (the bills screens)
- **Why (taxonomy):** 2 double entry elimination, 3 single source of truth,
  8 identity resolution
- **Shape:** 1, a cross-domain read from the Finance app
- **Severity:** high

## Benefit

The day after this ships, entering a bill means picking the supplier you
already have in ERP, and the bill list shows who the bill is owed to instead of
a description field standing in for a vendor name.

## Why this pairing

**Double entry elimination and single source of truth.** ERP owns suppliers.
Finance owes them money. Today Finance neither reads nor stores the supplier: it
asks the user to retype a vendor name, and then fabricates an integer id from
it.

This is the most concrete finding in the audit, so it is worth quoting exactly.
The bill list gives up on the supplier and substitutes the description:

```
// finance.bill.v1.Bill has no vendor name or category - only a
// supplier_id FK and a free-text description (which the create form
// already uses to carry the vendor name when the field is left blank).
```

`apps/finance/src/features/bills/useBills.ts:13`

And the create form manufactures the foreign key:

```
const derivedSupplierId =
  form.vendor.trim().toLowerCase().replace(/[^a-z0-9]/g, '-') || '1';
```

`apps/finance/src/features/bills/Bills.tsx:59`

That string is sent as `supplier_id`, which the contract declares as
`int64 @required @gt(0)` (`packages/voxie/api/finance/bill/v1/models.onk:29`).
OneKit encodes `int64` as a string on the wire and parses it with `strconv`
(`~/Development/voxie/gen/go/finance/bill/v1/types.gen.go:32`), and the server
rejects anything not greater than zero
(`~/Development/voxie/gen/go/finance/bill/v1/validate.gen.go:39`).

**Inferred, not verified:** a vendor name like `Acme Corp` becomes `acme-corp`,
which cannot parse as an int64, so the request fails with a 400; and a blank
vendor falls back to `'1'`, filing the bill against supplier id 1 whoever that
is. Confirming which of those two happens needs one request against a real
backend, which this audit deliberately does not make. **Either branch is a
defect**, and both are caused by the same missing connection.

**Why ERP and not a Finance-local supplier list.** Adding suppliers to Finance
would be the third place a vendor is named, after ERP suppliers and purchase
orders (`packages/voxie/api/erp/purchase_order/v1/models.onk:16` already
carries `supplier_id`). ERP is where purchasing happens, so ERP is where the
supplier record belongs.

**What the user does instead right now.** Types the vendor name into a free
text field on the bill form, having already created that supplier in ERP to
raise the purchase order.

## ROI

| Field                  | Value                                                                                                                                                                                                                                                                                                                                                        |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Affected tenants       | any tenant using Finance with ERP. The bug in the id fabrication affects **every** Finance tenant that enters bills, ERP or not.                                                                                                                                                                                                                             |
| Manual work today      | retype the vendor name per bill; read the description column to work out who a bill is owed to                                                                                                                                                                                                                                                               |
| Modelled monthly value | reference tenant enters roughly 40 bills a month (a third of its ~120 invoices). Retyping and re-reading costs about 30 seconds per bill: **40 x 0.5 min = 20 minutes per month**. Small.                                                                                                                                                                    |
|                        | The real value is correctness: today bills are either unfilable or misattributed to supplier 1. Accounts payable that cannot be grouped by supplier means no supplier balance, no payment run, and no answer to "how much do we owe Acme". That is a **modelled** rather than measured claim about tenant workflow, but the missing grouping is a code fact. |
| Build cost             | backend 0 days for the read path; **1 day** to relax the contract and add `supplier_name` (`06-proposed-contracts.md` item 6); frontend 1.5 days (supplier picker, list column, create form); maintenance ~0.25 days per year                                                                                                                                |
| Payback                | under a month if the fabricated-id defect is counted as a bug fix, which it should be.                                                                                                                                                                                                                                                                       |
| Confidence             | **Inferred from code** for everything about the current behaviour: the comment, the slugify, the contract constraint and the generated validator were all read. **Modelled from assumption** for bill volume.                                                                                                                                                |
| Cost of not doing it   | Accounts payable cannot be reported by supplier. A bill either fails to save or is attributed to an arbitrary supplier, and because `supplier_id` is never joined the mistake is invisible on every screen.                                                                                                                                                  |

## The data

| Field                    | From         | Cited                                                                |
| ------------------------ | ------------ | -------------------------------------------------------------------- |
| supplier `id`            | erp supplier | `packages/voxie/api/erp/supplier/v1/service.onk:5` (`listSuppliers`) |
| supplier name            | erp supplier | same service                                                         |
| `supplier_id` written    | finance bill | `packages/voxie/api/finance/bill/v1/models.onk:10`                   |
| `supplier_name` snapshot | proposed     | `06-proposed-contracts.md` item 6                                    |

No formula.

## What exists today

- `SupplierService.listSuppliers` is served
  (`packages/voxie/api/erp/supplier/v1/service.onk:5`), backed by
  `erp_suppliers`.
- `finance.Bill.supplier_id` exists and is required
  (`packages/voxie/api/finance/bill/v1/models.onk:10`).
- The `bills` backend module touches only `finance_bills` and never joins
  `erp_suppliers`.
- **Manual workaround:** the vendor name is carried in `description`, by
  deliberate design, and the code says so
  (`apps/finance/src/features/bills/useBills.ts:13`).

## What is missing

| Wanted                                    | Contract reality                             | Backend reality           | What is missing                                                                                      |
| ----------------------------------------- | -------------------------------------------- | ------------------------- | ---------------------------------------------------------------------------------------------------- |
| Pick a real supplier when entering a bill | `listSuppliers` exists; `supplier_id` exists | both served, never joined | a call from the Finance app. Nothing on the wire.                                                    |
| See the supplier on a bill list           | `Bill` has no name field                     | not joined                | either a `supplier_name` snapshot on the bill, or a second read the Finance app resolves client-side |
| File a bill with no known supplier        | `@required @gt(0)`                           | validator enforces it     | relaxing the constraint, which is what pushed the frontend into fabricating a value                  |

**Bucket: available today** for the read; the accompanying contract relaxation
is a correction rather than a new capability.

## Precedent in the codebase

`apps/dash` already reads foreign domains directly for display, including
`pos` (9 requests) and `insights` (6) per `01-inventory.md`. A Finance app
calling an ERP list endpoint is the same shape as something the product already
does in five places. That is the argument for shape 1 being sufficient.

## How to build it

**Shape 1, a cross-domain read.** The Finance app calls
`erp.supplier.v1 listSuppliers` for the picker and sends the real id.

**Why no looser shape exists.** Shape 1 is the loosest.

**Why not shape 2, a Finance-owned aggregate.** Not needed: the supplier list
is small, not sensitive relative to Finance's own data, and one read does not
breach the two-foreign-domain limit. The bill screens read ERP only.

**Why the denormalised `supplier_name` anyway.** Without it, every bill list
render needs a second request and a client-side join, and a bill's history
changes when a supplier is renamed. The snapshot is the same reasoning applied
to invoices in `CRM->FIN-02`.

Order of work: relax the contract first, or the picker cannot submit a bill for
a supplier that does not exist yet.

## Coupling cost

- **New dependency:** finance depends on erp suppliers. Edge `erp -> finance`.
- **Direction:** consumer depends on producer. Correct.
- **Cycle:** `erp <-> finance` is one of the two bidirectional pairs in the
  audit. `FIN->ERP-01` and `FIN->ERP-02` point back. `12-coupling.md` shows the
  facts flow in opposite directions and no connection consumes the output of a
  connection pointing back at it. **`ERP->FIN-01` must ship before
  `FIN->ERP-02`**, because a bill cannot reference a purchase order usefully
  while its supplier is fabricated.
- **Owner:** the supplier contract is owned by **erp**. The bill's reference is
  owned by **finance**.
- **Degrade when ERP is disabled:** the supplier picker becomes a free-text
  field again, and `supplier_id` is sent absent rather than fabricated (which
  the contract relaxation permits). The bill screen keeps working. This is
  strictly better than today, because today the fabrication happens whether ERP
  is licensed or not. Check the tenant's app entitlement through the existing
  guard rather than probing the endpoint: `packages/config/src/appGuards.tsx`.

## Risks and correctness notes

- **Tenancy.** `listSuppliers` is already tenant-scoped server-side; the
  Finance app must not filter client-side and assume completeness.
- **Permissions.** A Finance user who can enter bills would gain sight of the
  supplier list. That is a widening, and it is the right one: you cannot file a
  payable without knowing who to pay. It exposes names, not costs. **No pricing
  or purchase history crosses**, and the dossier for that (`FIN->ERP-01`) runs
  in the other direction.
- **Currency.** Bills already carry their own `currency`
  (`packages/voxie/api/finance/bill/v1/models.onk:14`) and suppliers do not
  impose one, so no conversion is introduced. Worth noting because there is no
  single currency authority in the product (`02-entity-graph.md`).
- **Migration.** Existing rows carry fabricated ids. Any `supplier_id` with no
  matching `erp_suppliers` row should be **reported, not silently nulled**: it
  is evidence of how many bills are misattributed, which is the only
  measurement of this defect anyone will get.
- **Rounding, timezone, historical rates:** not applicable.

## Effort

**S.**

| Part                | Estimate                                                                                                        |
| ------------------- | --------------------------------------------------------------------------------------------------------------- |
| Backend             | 1 day: relax `supplier_id` to nullable, add `supplier_name`, adjust the validator, write the orphaned-id report |
| Frontend            | 1.5 days: supplier picker on the bill form, supplier column on the list, remove the slugify                     |
| Ongoing maintenance | ~0.25 days per year. One read, stable shape.                                                                    |

## Backend section (round two)

Verified: every statement in `internal/modules/bills/sql/queries/bills.sql`
touches only `finance_bills`; none references `erp_suppliers`.
`CreateBillRequest.supplier_id` remains `@required @gt(0)`
(`packages/voxie/api/finance/bill/v1/models.onk:29`). The frontend fabrication
risk described above is therefore still live end to end.

_Round two (2026-08-22): confirmed unchanged._
