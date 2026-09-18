## Benefit

Support history attaches to real customers, so account views show ticket health without email guesswork.

## Why

Identity resolution (8), context in place (5). support_customers is an unrelated table joining to crm_customers only by loose email strings. Explicit nullable reference with confident-match-or-human-confirm semantics beats automatic merging. Cross-links: rtylr#861, rtylr#842.

## ROI

| Item                   | Value                                                 |
| ---------------------- | ----------------------------------------------------- |
| Affected tenants       | support users; operator surface still absent in rtylr |
| Manual work today      | agents search CRM by email                            |
| Modelled monthly value | low today; rises when a support app ships             |
| Build cost             | contract 0.1d; onion column+match UX 1d               |
| Payback                | contingent on support surface shipping                |
| Confidence             | Modelled from assumption                              |

## What exists, what is missing

Drafted on branch: Customer.customer_id int64? @nullable with match-policy doc comment.

## How to build it

Voxie; onion adds column + set-on-confident-match logic; consumer lands with SUP->CRM-01. Keep email matching as suggestion, never auto-write.

## Coupling cost

Declared edge support -> crm. Nullable preserves current behaviour entirely.

Labelled audit:cross-app-round2-2026-08-22, area:support, type:gap, severity:medium.
