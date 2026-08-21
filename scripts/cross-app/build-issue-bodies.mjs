#!/usr/bin/env node
// Builds GitHub issue bodies for every connection dossier and every proposed
// contract change, into a scratch directory, so the issues cannot disagree with
// the documents they came from.
//
//   node scripts/cross-app/build-issue-bodies.mjs <outDir>
//
// Each connection body carries these headings in this exact order, per the
// audit brief: Benefit, Why, ROI, "What exists, what is missing", How to build
// it, Coupling cost, then a link to the dossier on the branch. An issue missing
// Benefit, Why or ROI is not filed, so this script fails rather than emitting a
// partial body.

import { readFileSync, writeFileSync, mkdirSync, readdirSync } from 'node:fs';
import { join } from 'node:path';

const ROOT = new URL('../..', import.meta.url).pathname.replace(/\/$/, '');
const outDir = process.argv[2];
if (!outDir) {
  process.stderr.write('usage: build-issue-bodies.mjs <outDir>\n');
  process.exit(1);
}
mkdirSync(outDir, { recursive: true });

const BRANCH = 'docs/cross-app-data-sharing-audit';
const REPO_URL = 'https://github.com/tryrtylr/rtylr/blob/' + BRANCH;

const data = JSON.parse(readFileSync(join(ROOT, 'docs/cross-app/connections.json'), 'utf8'));
const DIR = join(ROOT, 'docs/cross-app/05-connections');
const files = readdirSync(DIR).filter((f) => f.endsWith('.md') && f !== 'README.md');
const prefixOf = (id) => id.replace(/->/g, '-').replace(/[^A-Za-z0-9-]/g, '');

// Split a dossier into its `## ` sections.
function sections(md) {
  const out = new Map();
  let cur = null;
  const buf = [];
  for (const line of md.split('\n')) {
    const m = line.match(/^## (.+)$/);
    if (m) {
      if (cur) out.set(cur, buf.join('\n').trim());
      cur = m[1].trim();
      buf.length = 0;
    } else if (cur) buf.push(line);
  }
  if (cur) out.set(cur, buf.join('\n').trim());
  return out;
}

const APP_OF_DOMAIN = {
  hr: 'area:hr',
  flow: 'area:flow',
  finance: 'area:finance',
  crm: 'area:crm',
  erp: 'area:erp',
  hub: 'area:dash', // hub's operator surface is dash
  insights: 'area:insights',
  pos: 'area:pos',
  support: 'area:support',
  recruit: 'area:recruit',
  all: 'area:repo',
};

const manifest = [];

for (const c of data.connections.filter((x) => x.status === 'proposed')) {
  const file = files.find((f) => f.startsWith(prefixOf(c.id) + '-'));
  if (!file) {
    process.stderr.write(`no dossier for ${c.id}\n`);
    process.exit(1);
  }
  const s = sections(readFileSync(join(DIR, file), 'utf8'));

  const benefit = s.get('Benefit');
  const why = s.get('Why this pairing');
  const roi = s.get('ROI');
  const exists = s.get('What exists today');
  const missing = s.get('What is missing');
  const how = s.get('How to build it');
  const coupling = s.get('Coupling cost');
  const risks = s.get('Risks and correctness notes');
  const effort = s.get('Effort');

  for (const [name, v] of Object.entries({ Benefit: benefit, Why: why, ROI: roi })) {
    if (!v) {
      process.stderr.write(`${c.id}: missing required section ${name}; not filing\n`);
      process.exit(1);
    }
  }

  const body = [
    '## Benefit',
    '',
    benefit,
    '',
    '## Why',
    '',
    why,
    '',
    '## ROI',
    '',
    'Every figure below is **modelled**, not measured. The reference tenant is defined in [08-roadmap.md](' +
      REPO_URL +
      '/docs/cross-app/08-roadmap.md).',
    '',
    roi,
    '',
    '## What exists, what is missing',
    '',
    '### Exists today',
    '',
    exists ?? '_see dossier_',
    '',
    '### Missing',
    '',
    missing ?? '_see dossier_',
    '',
    '## How to build it',
    '',
    how ?? '_see dossier_',
    '',
    '## Coupling cost',
    '',
    coupling ?? '_see dossier_',
    '',
    ...(risks ? ['## Risks and correctness notes', '', risks, ''] : []),
    ...(effort ? ['## Effort', '', effort, ''] : []),
    '---',
    '',
    `Full dossier: [\`docs/cross-app/05-connections/${file}\`](${REPO_URL}/docs/cross-app/05-connections/${file})`,
    '',
    `Audit index: [\`docs/cross-app/README.md\`](${REPO_URL}/docs/cross-app/README.md). Produced by the cross-app data sharing audit; no code was changed.`,
  ].join('\n');

  const labels = new Set([
    'type:gap',
    `severity:${c.severity ?? 'medium'}`,
    'audit:cross-app-2026-08-21',
  ]);
  const pl = APP_OF_DOMAIN[c.producer];
  const cl = APP_OF_DOMAIN[c.consumer];
  if (pl) labels.add(pl);
  if (cl) labels.add(cl);

  const slug = prefixOf(c.id);
  writeFileSync(join(outDir, `${slug}.md`), body + '\n');
  manifest.push({
    id: c.id,
    kind: 'connection',
    // Title format from the brief: "<ID>: <what flows> from <A> to <B>".
    // shortFlows keeps it legible in a list; the full phrasing is the H1 of
    // the dossier and the first line of the body.
    title: `${c.id}: ${c.shortFlows ?? c.flows} from ${c.producer} to ${c.consumer}`,
    bodyFile: join(outDir, `${slug}.md`),
    labels: [...labels],
  });
}

// --- contract proposals -----------------------------------------------------
// One issue per drafted contract change, because these land in a different
// repository (voxie) with a different owner than the consuming app work.
const CONTRACTS = [
  {
    n: 1,
    slug: 'contract-01-time-entry-filters',
    title: 'Contract: add range, project and billability filters to hr time entries',
    serves: ['FLOW->FIN-01'],
    domain: 'hr',
    severity: 'high',
    benefit:
      'Finance can ask for billable hours on a project over a period, which is what makes unbilled time visible. Today the endpoint answers only "one employee, one exact date".',
    why: 'Every field needed already exists and is populated by Flow. `ListTimeEntriesRequest` filters by `employee_id` and a single `date` (`packages/voxie/api/hr/time_entry/v1/models.onk:71`) while its response reports `limit` and `offset` the caller cannot set. Adding a filter without pagination would make an unbounded read easier to trigger, so both go in together.',
  },
  {
    n: 2,
    slug: 'contract-02-leave-date-range',
    title: 'Contract: add a date range and pagination to hr leave requests',
    serves: ['HR->FLOW-02'],
    domain: 'hr',
    severity: 'medium',
    benefit:
      "Flow can ask who is on approved leave during a cycle, so assigning work into someone's holiday can raise a warning.",
    why: '`ListLeaveRequestsRequest` filters by `employee_id` and `status` only (`packages/voxie/api/hr/leave/v1/models.onk:87`). "Who is away between these dates" is not expressible. Overlap semantics must be stated in the doc comment, because a consumer that guesses wrong silently misses multi-day leave crossing the boundary.',
  },
  {
    n: 3,
    slug: 'contract-03-invoice-customer-id',
    title: 'Contract: give a finance invoice a crm customer_id',
    serves: ['CRM->FIN-02', 'FIN->CRM-01'],
    domain: 'finance',
    severity: 'high',
    benefit: 'Revenue per customer becomes answerable, and a customer page can show what they owe.',
    why: 'An invoice stores its customer as four free strings, two of them `@required` (`packages/voxie/api/finance/invoices/v1/models.onk:147`). The `client_*` snapshot is kept deliberately: a document that was sent must keep the name it was sent under. **This is the change whose cost grows with delay**, because every invoice issued meanwhile adds unmatchable history.',
  },
  {
    n: 4,
    slug: 'contract-04-menu-item-product-id',
    title: 'Contract: link a pos menu item to its erp product',
    serves: ['ERP->POS-01'],
    domain: 'pos',
    severity: 'high',
    benefit:
      'Item margin becomes computable, and the two independent stock numbers become comparable instead of silently divergent.',
    why: '`pos.MenuItem` carries its own `stock` and `sale_price` and no cost at all (`packages/voxie/api/pos/menu_item/v1/models.onk:28`), while `erp.Product` has `stock_quantity` and `cost_price` (`packages/voxie/api/erp/product/v1/models.onk:15`). Note `pos.StockMovement` already references `product_id` (`packages/voxie/api/pos/stock/v1/models.onk:8`), so POS stock movements know about ERP products while menu items do not.',
  },
  {
    n: 5,
    slug: 'contract-05-bill-purchase-order-id',
    title: 'Contract: link a finance bill to the erp purchase order it settles',
    serves: ['FIN->ERP-02', 'ERP->FIN-03'],
    domain: 'finance',
    severity: 'medium',
    benefit:
      'Received-but-unbilled purchase orders become computable rather than reconciled by hand.',
    why: 'A bill has no reference to the order it settles, so "no bill references this order" cannot be evaluated. Direction matters: ERP owns purchase order status (`packages/voxie/api/erp/purchase_order/v1/service.onk:11`), so Finance carries the reference and emits, never writes.',
  },
  {
    n: 6,
    slug: 'contract-06-relax-bill-supplier-id',
    title: 'Contract: make finance bill supplier_id nullable and add supplier_name',
    serves: ['ERP->FIN-01'],
    domain: 'finance',
    severity: 'high',
    benefit:
      'A bill can record a real ERP supplier, or honestly record none, instead of the UI inventing an id.',
    why: '`supplier_id` is `@required @gt(0)` (`packages/voxie/api/finance/bill/v1/models.onk:29`) on a link the product cannot always supply, and the frontend responded by slugifying the typed vendor name into the field (`apps/finance/src/features/bills/Bills.tsx:59`). A required link that cannot be guaranteed gets faked. **This is a correction, not a feature**, and `CreateBillRequest` must lose the constraint at the same time or the picker still cannot submit.',
  },
  {
    n: 7,
    slug: 'contract-07-support-customer-id',
    title: 'Contract: reference a crm customer from a support customer',
    serves: ['SUP->CRM-01'],
    domain: 'support',
    severity: 'medium',
    benefit:
      "A CRM customer page can show that customer's real ticket history instead of an email-matched guess.",
    why: '`support_customers` is a separate table from `crm_customers` and nothing joins them, so the only available key is email. Email is right until a customer writes from a second address, at which point the page shows zero tickets for someone with five. The field is nullable because support mail arrives from people who are not customers yet.',
  },
  {
    n: 8,
    slug: 'contract-08-aggregate-endpoints',
    title: 'Contract: add hr labour-cost and insights product-margin aggregate endpoints',
    serves: ['HR->FLOW-01', 'HR->INS-01', 'ERP->INS-01'],
    domain: 'hr',
    severity: 'high',
    benefit:
      'Project labour cost and product margin become answerable in the product, without any consumer gaining sight of individual pay rates or supplier costs it should not see.',
    why: "These are aggregates rather than reads **because the inputs are more sensitive than the output**. `hr.Staff` exposes `salary`, `hourly_rate` and `overtime_rate` (`packages/voxie/api/hr/staff/v1/models.onk:17`); if Flow priced hours client-side, any workspace member could read colleagues' pay. One endpoint serves both Flow and Insights on purpose, so the two cannot report different numbers for the same question. The response carries `rate_coverage` and `contributor_count` so a partial total is labelled and a one-person total can be suppressed.",
  },
];

for (const k of CONTRACTS) {
  const body = [
    '## Benefit',
    '',
    k.benefit,
    '',
    '## Why',
    '',
    k.why,
    '',
    `Serves: ${k.serves.map((s) => `\`${s}\``).join(', ')}.`,
    '',
    '## ROI',
    '',
    'This is the contract half of the connections listed above; its ROI is theirs and is not double counted. See each connection issue for the modelled figures and the reference tenant in [08-roadmap.md](' +
      REPO_URL +
      '/docs/cross-app/08-roadmap.md). Every figure in this audit is modelled, not measured.',
    '',
    '## What exists, what is missing',
    '',
    'Gap tables with citations: [`docs/cross-app/04-gaps.md`](' +
      REPO_URL +
      '/docs/cross-app/04-gaps.md).',
    '',
    '## How to build it',
    '',
    `Draft \`.onk\` in house style, with migration notes: [\`docs/cross-app/06-proposed-contracts.md\`](${REPO_URL}/docs/cross-app/06-proposed-contracts.md), item ${k.n}.`,
    '',
    'Two wire facts the draft respects. OneKit `int64` is a **string on the wire** (`~/Development/voxie/gen/go/finance/bill/v1/types.gen.go:32` marshals via `strconv.FormatInt`), so a JSON number fails to unmarshal and `"0"` fails `@gt(0)`. And optional ids are `int64?` with `@nullable`, never a zero sentinel and never `@required @gt(0)`, because that combination is what produced the fabricated `supplier_id`.',
    '',
    '**Contracts land in voxie first**, then the packages submodule pointer, then this repo (root `CLAUDE.md`, submodules section). Nothing under `packages/voxie` was edited by this audit.',
    '',
    '## Coupling cost',
    '',
    'Per-connection coupling, direction, owner and degrade behaviour are in each connection dossier. The dependency graph and its acyclicity proof: [`docs/cross-app/12-coupling.md`](' +
      REPO_URL +
      '/docs/cross-app/12-coupling.md).',
    '',
    '**Ship the consumer with the field.** `flow.WorkType.hr_category_id` is a cross-domain field that is persisted, echoed, and read by nothing (`packages/voxie/api/flow/time/v1/models.onk:20`). That is the base rate for a contract field added ahead of its reader.',
    '',
    '---',
    '',
    `Audit index: [\`docs/cross-app/README.md\`](${REPO_URL}/docs/cross-app/README.md).`,
  ].join('\n');

  const labels = new Set([
    'type:gap',
    `severity:${k.severity}`,
    'audit:cross-app-2026-08-21',
    'area:repo',
  ]);
  const dl = APP_OF_DOMAIN[k.domain];
  if (dl) labels.add(dl);

  writeFileSync(join(outDir, `${k.slug}.md`), body + '\n');
  manifest.push({
    id: `contract-${String(k.n).padStart(2, '0')}`,
    kind: 'contract',
    title: k.title,
    bodyFile: join(outDir, `${k.slug}.md`),
    labels: [...labels],
  });
}

writeFileSync(join(outDir, 'manifest.json'), JSON.stringify(manifest, null, 2) + '\n');
process.stdout.write(`${manifest.length} issue bodies written to ${outDir}\n`);
