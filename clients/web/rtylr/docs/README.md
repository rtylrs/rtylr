# rtylr docs

Architecture docs and audits. Per the root `CLAUDE.md`, this folder is the
authoritative reference for the team: update it when you change an
architectural or pattern-level thing.

## Patterns and reference

| Document                                             | What it covers                                                                                                                                                    |
| ---------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [flow-time-tracking.md](flow-time-tracking.md)       | Flow's time tracking, and the worked example of the difference between an endpoint existing and a field existing. Its gap-table format is reused by later audits. |
| [routing-search-params.md](routing-search-params.md) | TanStack Router `validateSearch` and path params, and the three rules that have each already caused a defect.                                                     |

## Audits

| Document                                                   | What it covers                                                                                                                                                                                                                                                                                                                        |
| ---------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [cross-app/](cross-app/README.md)                          | **Cross-app data sharing audit** (2026-08-21). Every piece of data shared, shareable or needing to be shared across the ten apps and twelve backend domains: a full inventory, the entity and identity graph, a 132-cell domain matrix, 21 connection proposals with 18 documented rejections, draft contract changes, and a roadmap. |
| [cross-app/flow-deep-dive.md](cross-app/flow-deep-dive.md) | **Flow cross-app follow-up** (2026-08-21). A second, Flow-centric round over the same subject: both inventory directions for Flow alone, gap classes A to D, the worked project-cost example end to end, and six findings the platform-wide round did not reach. Its section 10 reconciles the two.                                   |

### Patterns proposed by the cross-app audit

Two conventions in that audit are new and would apply repo-wide if adopted.
Neither is implemented yet, so neither is in the root `CLAUDE.md`:

- **Cross-domain query keys.** The existing convention in the root `CLAUDE.md`
  assumes a single domain. A read of a _foreign_ domain is prefixed with the
  domain that owns the data: `['erp', 'suppliers']`, not `['suppliers']`. See
  [cross-app/07-plumbing.md](cross-app/07-plumbing.md).
- **The six plumbing shapes**, ordered loosest first, with the rule that a
  cross-app connection takes the first shape that works and must justify
  reaching for a tighter one. Same document.

The audit also records a correction worth knowing before reading the root
`CLAUDE.md`: `packages/voxie` is **vendored, not a nested submodule**, and
`APP_CONFIG.allowedRoles` is **not consulted by anything**. Both are in
[cross-app/10-doc-drift.md](cross-app/10-doc-drift.md) with citations.
