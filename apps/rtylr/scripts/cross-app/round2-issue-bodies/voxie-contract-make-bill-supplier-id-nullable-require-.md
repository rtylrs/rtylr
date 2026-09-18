## Benefit

Bills record real suppliers instead of fabricated ids, and can reference the purchase orders they settle.

## Why

Correctness (6), double-entry elimination (2). supplier_id is @required @gt(0) yet the backend never joins erp_suppliers and the frontend slugifies typed names into the field (ERP->FIN-01). purchase_order_id enables settlement closing (FIN->ERP-02). Serves drafts 5+6 combined into one coherent bill change. Cross-links: rtylr#860, rtylr#859, rtylr#835, rtylr#851.

## ROI

| Item                   | Value                                                                  |
| ---------------------- | ---------------------------------------------------------------------- |
| Affected tenants       | every finance tenant                                                   |
| Manual work today      | vendor names retyped per bill; PO matching manual                      |
| Modelled monthly value | 3h/mo AP clerk time (reference tenant); duplicate-payment risk reduced |
| Build cost             | contract 0.25d; onion 1.5d; finance UI 2d                              |
| Payback                | < 1 month                                                              |
| Confidence             | Inferred from code                                                     |
| Cost of not doing it   | fabricated foreign keys accumulate; payables reporting unreliable      |

## What exists, what is missing

Drafted on branch: Bill.supplier_id int64? @nullable; supplier_name string @required (snapshot); Bill.purchase_order_id + CreateBillRequest fields; ListBillsRequest.supplier_id/purchase_order_id filters.

## How to build it

Voxie; onion relaxes validator, stores snapshot, verifies foreign ids same-business; finance ships supplier picker (wave-1 item 2) before PO-settlement UI.

## Coupling cost

Declared edges erp -> finance (identity) and finance -> erp (settlement reference, read by erp). Fact graph stays acyclic: settlement reference does not consume derived data.

Labelled audit:cross-app-round2-2026-08-22, area:finance, type:gap, severity:high.
