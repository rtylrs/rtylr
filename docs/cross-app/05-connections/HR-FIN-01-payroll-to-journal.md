# HR->FIN-01: payroll run total from hr to finance

- **Class:** CAN
- **Necessity:** Essential
- **Producer:** hr.payroll.v1 (`PayrollService.listPayroll`)
- **Consumer:** finance (journal)
- **Why (taxonomy):** 2 double entry elimination, 4 trigger, 7 rollup
- **Shape:** 1, a cross-domain read plus a Finance-owned write
- **Severity:** high

## Benefit

The day after this ships, a completed payroll run can be posted to the journal
as a labour expense in one action, instead of being re-keyed from one app into
another.

## Why this pairing

**Double entry elimination, on the largest single expense most businesses have.**
Payroll computes what was paid. The journal records what the business spent.
Today a bookkeeper reads one and types the other.

Both halves are already served. HR's payroll returns `period_start`,
`period_end` and `net_pay` (`packages/voxie/api/hr/payroll/v1/models.onk:7`,
`:17`), and payroll genuinely computes the money rather than storing a typed
figure (`~/Development/onion/internal/modules/hr_payroll/internal/application/service.go:221`).
Finance's journal accepts a `date`, a `description` and `lines[]`
(`packages/voxie/api/finance/journal/v1/models.onk:25`).

**Why Finance writes and HR does not.** Rule 4: a consumer never writes another
domain's records, and by the same logic the producer should not write the
consumer's. The journal is Finance's. So the action lives in Finance: read the
payroll run, present the proposed entry, let the user post it. HR is never
asked to know what a journal is.

**Why not automatic on payroll completion.** A journal entry is an accounting
act with an account mapping behind it, and the account to debit varies by
tenant's chart of accounts. Posting automatically would require RTYLR to choose
a tenant's accounts, which the configurability rules forbid.

**What the user does instead right now.** Opens the payroll run, reads the
total, opens Finance, creates a journal entry by hand each pay period.

## ROI

| Field                  | Value                                                                                                                                                                                                                       |
| ---------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Affected tenants       | tenants running HR payroll with Finance.                                                                                                                                                                                    |
| Manual work today      | re-key the payroll total into a journal entry each period, and reconcile if it was mistyped                                                                                                                                 |
| Modelled monthly value | reference tenant pays monthly. One journal entry per period costs perhaps 15 minutes including the cross-check: **15 minutes per month.** Small in time.                                                                    |
|                        | The error side is where the value is: payroll is typically the largest expense line, so a transposed digit is a materially wrong P&L, and there is currently no link back to the payroll run to catch it.                   |
| Build cost             | backend **0 days** for the read and write paths; both exist. 1 day if a posted-marker is added to prevent double posting. Frontend 2 days (the proposed-entry screen and account mapping). Maintenance ~0.25 days per year. |
| Payback                | fast, because the build is almost entirely frontend against endpoints that already exist. **This is the best effort-to-value ratio among the Essential items.**                                                             |
| Confidence             | **Inferred from code** that both endpoints exist and payroll computes rather than stores. **Modelled from assumption** for the 15 minutes.                                                                                  |
| Cost of not doing it   | n/a, CAN not MUST. The largest expense line stays hand-keyed with no link back to its source.                                                                                                                               |

## The data

| Field                            | From            | Cited                                                 |
| -------------------------------- | --------------- | ----------------------------------------------------- |
| `period_start`, `period_end`     | hr payroll      | `packages/voxie/api/hr/payroll/v1/models.onk:7`       |
| `net_pay`                        | hr payroll      | `packages/voxie/api/hr/payroll/v1/models.onk:17`      |
| `date`, `description`, `lines[]` | finance journal | `packages/voxie/api/finance/journal/v1/models.onk:25` |

The journal entry is a mapping decision, not a formula: the payroll total
becomes one or more debit lines against tenant-chosen expense accounts and a
credit against a liability or bank account. **The mapping must be configured by
the tenant, not assumed**, and remembered between periods so it is chosen once.

## What exists today

- `PayrollService.listPayroll`, `runPayroll`, `markPayrollPaid`
  (`packages/voxie/api/hr/payroll/v1/service.onk:4`).
- A separate payroll runs capability with `listPayrollRuns`, `approvePayrollRun`
  and `exportPayroll` (`packages/voxie/api/hr/employees/payroll/runs/v1/service.onk:6`).
- `JournalService.createJournalEntry` and `listJournalEntries`
  (`packages/voxie/api/finance/journal/v1/service.onk:7`, `:5`).
- `finance.account` capability for the chart of accounts.
- **Manual workaround:** re-keying.

## What is missing

| Wanted                            | Contract reality                      | Backend reality   | What is missing                                                       |
| --------------------------------- | ------------------------------------- | ----------------- | --------------------------------------------------------------------- |
| Post a payroll run to the journal | both sides fully served               | both sides served | only the action: a Finance screen that reads one and writes the other |
| Not post twice                    | nothing records that a run was posted | nothing           | a posted marker, or reconciliation by description, which is fragile   |
| Tenant account mapping            | `finance.account` exists              | served            | somewhere to persist the chosen mapping. See the note below.          |

**Bucket: available today** for the core, with one honest caveat: there is no
obvious place to persist the account mapping. If no Finance settings surface can
hold it, the mapping should be re-chosen per posting with the previous choice
pre-selected, rather than stored per-browser. A per-browser default would be a
settings-store-as-seam situation, and the repo's own rule is that such a store
must not decide something a server would decide. An account mapping is exactly
that, so **it must not live in a Zustand persist store.**

## Precedent in the codebase

`FIN->INS-01` is the cautionary precedent for direction: the Finance app already
reads `/insights/financial/cash-flow`, which makes Finance depend on a derived
projection of its own data (`10-doc-drift.md`). This proposal deliberately goes
the other way: Finance reads a **source** fact from HR and writes its own
record. That is the correct shape for a consumer.

## How to build it

**Shape 1, a cross-domain read, plus a same-domain write.**

The Finance screen reads the payroll run, shows the proposed journal entry with
its account mapping, and on confirmation calls `createJournalEntry`. The write
crosses no boundary: Finance writes Finance.

**Why not shape 3, an event.** An event would have to carry an accounting
decision the tenant has not made. Discussed above.

**Why not shape 2.** Nothing to aggregate; payroll already totals.

Sequence: the screen first, since both endpoints exist. Add the posted-marker
before anyone uses it in anger, because double-posting payroll is a real
accounting incident.

## Coupling cost

- **New dependency:** finance depends on hr payroll. Edge `hr -> finance`, the
  same edge as `FLOW->FIN-01`, so no new node pair.
- **Direction:** consumer depends on producer. Correct. No cycle:
  `finance -> hr` is `would-invert` and unproposed.
- **Owner:** **hr** owns payroll totals. **finance** owns the journal and the
  mapping.
- **Degrade when HR is disabled:** the "post payroll" entry point is absent and
  the journal screen is unchanged. Additive.
- **Two-foreign-reads limit:** the journal screen would read hr only. Fine.

## Risks and correctness notes

- **Permissions, and this is a serious one.** A payroll total is sensitive. In a
  small team a period total plus a headcount is close to an average salary, and
  for a one-person department it _is_ a salary. Reading payroll totals must be
  gated to a finance or payroll role, not to every Finance user. The audit found
  no existing path by which a non-HR role reads payroll; **do not create one
  casually.**
- **Rounding and representation.** Payroll works in integer cents internally
  (`~/Development/onion/internal/modules/hr_payroll/internal/application/service.go:236`)
  but exposes `net_pay` as `float64`
  (`packages/voxie/api/hr/payroll/v1/models.onk:17`), and the journal will take
  a float too. A round trip through float can move a cent, and a journal that is
  out by a cent does not balance. **Compare the posted total against the payroll
  total and refuse to post on mismatch** rather than trusting the arithmetic.
- **Currency.** Payroll carries **no currency field at all**
  (`02-entity-graph.md`), while journal entries sit in Finance's currency. The
  posting must state the assumed currency explicitly rather than inheriting
  silently, and should refuse if the tenant's Finance currency and payroll
  jurisdiction plainly differ.
- **Double posting.** Covered above. Not optional.
- **Timezone.** The journal date should be the period end or the pay date, and
  the choice must be explicit; posting to the wrong period misstates a month.
- **Historical rates:** not applicable.
- **Tenancy:** standard.

## Effort

**S.**

| Part                | Estimate                                                                               |
| ------------------- | -------------------------------------------------------------------------------------- |
| Backend             | 0 to 1 day: nothing needed for the core; 1 day for a posted marker on the payroll run  |
| Frontend            | 2 days: proposed-entry screen, account mapping selection, balance check before posting |
| Ongoing maintenance | ~0.25 days per year                                                                    |

## Backend section (round two)

The backend half now exists end to end, built the day after this dossier was
filed. `POST /hr/payroll/runs/{id}/export`
(`internal/modules/hr_payroll/internal/routes/handlers/export_payroll_handler.go:18`) calls `ExportPayrollToFinance`
(`internal/modules/hr_payroll/internal/application/service.go:375`), which takes
an idempotent claim (`ClaimPayrollRunForExport`, released on failure), converts
internally-stored cents at the boundary with an explicit comment, and posts
through the Finance-owned `financejournal.PayrollJournalPoster` port
(`internal/modules/journal/module.go:32,40`), wired at `cmd/api/main.go:499`.
Tenant account selection comes from the `financeaccounts` port
(`cmd/api/main.go:498`), which answers the dossier's configurability worry directly:
RTYLR never chooses accounts, the tenant does, and posting fails loudly when
mapping is absent (`1fb0e4f1`, "require configured payroll accounts").

The design question the dossier raised ("why not automatic") was resolved
differently than assumed: invocation lives in HR behind a Finance-owned port
rather than as a Finance-hosted action, and ownership boundaries hold because
the journal contract and account config both stay Finance's.

Remaining work is the UI button plus optionally a posted-marker on the run.
No frontend caller exists today (verified in rtylr and pos).

_Round two (2026-08-22): corrected-by-progress; effort drops to S-frontend._
