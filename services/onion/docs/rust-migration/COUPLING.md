# Module coupling

Reuses the `architecture_test.go` owner scan: the first directory under
`internal/modules/` (or `internal/contracts/`) is the owner;
`internal/platform/<x>` owns `platform/<x>`. A business edge exists when an
owner imports a different owner outside `platform/*`. Generated `*.gen.go`
files are excluded. Measured on `rust-migration` at Unit 0.4: **33 business
edges**, 14 owners with edges, 25 module owners.

## Edges (owner → imports)

| Owner | Imports |
| --- | --- |
| activity | audit |
| booking | crm |
| caret | organization, support |
| crm | organization |
| erp | activity, audit, crm, pos |
| finance | crm, erp, hr, organization |
| flow | hr, organization |
| hr | organization |
| insights | crm, erp, finance, flow, hr, pos |
| organization | audit, hr, notifications |
| pos | audit, crm, erp, organization |
| store | crm, erp |
| support | crm |
| webhooks | organization |

`audit` is `internal/contracts/audit` (shared contracts, not a Phase 3
module crate). Eleven modules have no business edges and are isolated
leaves: `aiquery`, `aiquota`, `automations`, `cacheops`, `emailtest`,
`hub_menu`, `integrations`, `portal`, `recruit`, `websitelead`.

Platform imports of business owners (3, test-only): `cdn/auth_test.go` →
`organization/scopes`; `eventbus/dispatcher_test.go` → `automations`,
`webhooks`. No production `platform/*` → business edges.

## Strongly connected components

- **{hr, organization}** — `hr` imports `organization`; `organization`
  imports `hr`. Port as one unit with trait seams (`AppState` injection
  per `ARCHITECTURE.md` §6 step 4), or port a trait stub first.
- **{erp, pos}** — `erp` imports `pos`; `pos` imports `erp`. Same rule.

All other components are singletons (DAG).

## Port order (dependencies first)

Within Phase 3, follow this order (isolated leaves first, sinks last).

1. **Isolated leaves, no business deps:** `aiquery`, `aiquota`,
   `automations`, `cacheops`, `emailtest`, `hub_menu`, `integrations`,
   `portal`, `recruit`, `websitelead`, plus contracts `audit`,
   then `activity`, `notifications`.
2. **SCC {hr, organization}** (together).
3. `crm`, `booking`.
4. `support`, `caret`.
5. **SCC {erp, pos}** (together).
6. `finance`, `flow`.
7. **Sinks:** `insights`, `store`, `webhooks`.
8. `compat` (aggregator, last, depends on every module it serves).

Sizing (non-test Go lines, hand-written, excl. generated sqlc): `hr` ~39k,
`organization` ~26k, `pos` ~21k, `erp` ~21k, `crm` ~21k, `caret` ~20k,
`finance` ~19k, `flow` ~18k, `recruit` ~12k, rest <6k each.
