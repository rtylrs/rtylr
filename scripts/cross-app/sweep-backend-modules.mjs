#!/usr/bin/env node
// Maps every onion backend module to the .onk domains it serves and the
// frontend apps that consume it, emitting
// docs/cross-app/inventory/backend_modules.csv. Requires a local onion
// checkout (read-only); path defaults to ~/Development/onion, override with
// ONION_DIR. Run AFTER sweep-endpoints.mjs:
//   node scripts/cross-app/sweep-backend-modules.mjs
//
// Domain evidence, per module:
//   gen/go imports    gen/go/<domain>/<capability>/<version> import paths
//   legacy routes     XRoutes() []string { return []string{"/path"} }
// Consumer evidence: endpoints.csv rows whose path matches either a
// contract path served by one of the module's imported capabilities, or a
// legacy route path (params normalised on both sides).
//
// Columns: module,onk_domains,capabilities,legacy_routes,consumer_apps,
//          has_frontend_consumer,module_path

import { readdirSync, readFileSync, writeFileSync, existsSync } from 'node:fs';
import { join, relative } from 'node:path';
import { homedir } from 'node:os';

const ROOT = new URL('../..', import.meta.url).pathname;
const ONION = process.env.ONION_DIR ?? join(homedir(), 'Development/onion');
const MODULES = join(ONION, 'internal/modules');
const OUT = join(ROOT, 'docs/cross-app/inventory/backend_modules.csv');

if (!existsSync(MODULES)) {
  console.error(`onion modules dir not found at ${MODULES}; set ONION_DIR`);
  process.exit(1);
}

function walk(dir) {
  const out = [];
  for (const e of readdirSync(dir, { withFileTypes: true })) {
    const p = join(dir, e.name);
    if (e.isDirectory()) out.push(...walk(p));
    else if (e.name.endsWith('.go') && !e.name.endsWith('_test.go')) out.push(p);
  }
  return out;
}

function csv(v) {
  const s = String(v ?? '');
  return /[",\n]/.test(s) ? `"${s.replaceAll('"', '""')}"` : s;
}

function parseCsvLine(line) {
  const out = [];
  let cur = '';
  let q = false;
  for (let i = 0; i < line.length; i++) {
    const c = line[i];
    if (q) {
      if (c === '"' && line[i + 1] === '"') {
        cur += '"';
        i++;
      } else if (c === '"') q = false;
      else cur += c;
    } else if (c === '"') q = true;
    else if (c === ',') {
      out.push(cur);
      cur = '';
    } else cur += c;
  }
  out.push(cur);
  return out;
}

const norm = (p) =>
  ('/' + p)
    .replaceAll(/\/+/g, '/')
    .replaceAll(/\$\{[^}]*\}/g, '*')
    .replaceAll(/\{[^}]*\}/g, '*')
    .replace(/\/$/, '')
    .toLowerCase();

// --- 1. frontend request paths per app, from endpoints.csv ------------------
// onekit rows carry ServiceName.method in path_or_rpc; safe/axios rows carry
// the literal path. For onekit we join through contracts.csv full paths.
const contracts = readFileSync(join(ROOT, 'docs/cross-app/inventory/contracts.csv'), 'utf8')
  .trim()
  .split('\n')
  .slice(1)
  .map(parseCsvLine);

const rpcFullPath = new Map(); // ServiceName.method -> /domain/cap/v1/path
const capOfService = new Map(); // ServiceName.method -> domain/cap/version
for (const c of contracts) {
  const [domain, capability, version, kind, name, , , , doc] = c;
  if (kind !== 'rpc') continue;
  const m = doc.match(/@(\w+)\("([^"]*)"\)/);
  if (m) rpcFullPath.set(name, norm(`/${domain}/${capability}/${version}${m[2]}`));
  capOfService.set(name, `${domain}/${capability}/${version}`);
}

const endpoints = readFileSync(join(ROOT, 'docs/cross-app/inventory/endpoints.csv'), 'utf8')
  .trim()
  .split('\n')
  .slice(1)
  .map(parseCsvLine);

// appRequests: [{app, path(normalised), capability|null}]
const appRequests = [];
for (const e of endpoints) {
  const [app, , , , pathOrRpc, , , , , , idiom] = e;
  if (idiom === 'onekit') {
    const full = rpcFullPath.get(pathOrRpc);
    if (full) appRequests.push({ app, path: full, capability: capOfService.get(pathOrRpc) });
  } else {
    appRequests.push({ app, path: norm(pathOrRpc), capability: null });
  }
}

// Also derive capability from onekit-style literal paths: /domain/cap/v1/...
const CAP_RE =
  /^\/(booking|common|console|crm|erp|finance|flow|hr|hub|insights|pos|support)\/([\w/-]+?)\/(v\d+)\//;
for (const r of appRequests) {
  if (!r.capability) {
    const m = r.path.match(CAP_RE);
    if (m) r.capability = `${m[1]}/${m[2]}/${m[3]}`;
  }
}

// --- 2. per-module evidence --------------------------------------------------
const rows = [
  [
    'module',
    'onk_domains',
    'capabilities',
    'legacy_routes',
    'consumer_apps',
    'has_frontend_consumer',
    'module_path',
  ],
];

const moduleNames = readdirSync(MODULES, { withFileTypes: true })
  .filter((e) => e.isDirectory())
  .map((e) => e.name)
  .sort();

let orphanModules = 0;
for (const mod of moduleNames) {
  const files = walk(join(MODULES, mod));
  const caps = new Set();
  const legacy = new Set();
  for (const f of files) {
    const src = readFileSync(f, 'utf8');
    for (const m of src.matchAll(/"github\.com\/voxire\/voxie\/gen\/go\/([\w/]+)"/g)) {
      caps.add(m[1]);
    }
    // Route paths are plain string literals in handler files; the previous
    // []string{...} capture broke on the '}' inside "{id}" segments, so scan
    // every path-shaped literal instead. False extras are harmless: legacy
    // routes are only ever joined against observed frontend requests.
    for (const m of src.matchAll(/"(\/[a-z0-9_\-/{}.]+)"/g)) legacy.add(norm(m[1]));
  }
  const domains = new Set([...caps].map((c) => c.split('/')[0]));
  const capsNorm = new Set([...caps].map((c) => c.replaceAll('-', '_')));

  const consumers = new Set();
  for (const r of appRequests) {
    if (r.capability && capsNorm.has(r.capability.replaceAll('-', '_'))) {
      consumers.add(r.app);
      continue;
    }
    if (legacy.size) {
      for (const lp of legacy) {
        if (lp === r.path) {
          consumers.add(r.app);
          break;
        }
        // one side has params where the other has literals: compare segments
        const a = lp.split('/');
        const b = r.path.split('/');
        if (
          a.length === b.length &&
          a.every((seg, i) => seg === '*' || b[i] === '*' || seg === b[i])
        ) {
          consumers.add(r.app);
          break;
        }
      }
    }
  }

  if (consumers.size === 0) orphanModules++;
  rows.push([
    mod,
    [...domains].sort().join(';'),
    [...caps].sort().join(';'),
    [...legacy].sort().join(';'),
    [...consumers].sort().join(';'),
    consumers.size > 0 ? 'yes' : 'no',
    `internal/modules/${mod}`,
  ]);
}

writeFileSync(OUT, rows.map((r) => r.map(csv).join(',')).join('\n') + '\n');
process.stdout.write(
  `${moduleNames.length} modules (${orphanModules} with no matched frontend consumer) -> ${relative(ROOT, OUT)}\n`
);

// --- 3. reverse check: frontend requests no module serves --------------------
// Rebuild the union of everything any module matched, then report the rest to
// stderr so the inventory can state coverage honestly.
const servedCaps = new Set();
const servedLegacy = new Set();
for (const mod of moduleNames) {
  for (const f of walk(join(MODULES, mod))) {
    const src = readFileSync(f, 'utf8');
    for (const m of src.matchAll(/"github\.com\/voxire\/voxie\/gen\/go\/([\w/]+)"/g))
      servedCaps.add(m[1]);
    for (const m of src.matchAll(/"(\/[a-z0-9_\-/{}.]+)"/g)) servedLegacy.add(norm(m[1]));
  }
}
const unserved = new Set();
for (const r of appRequests) {
  if (
    r.capability &&
    [...servedCaps].some((c) => c.replaceAll('-', '_') === r.capability.replaceAll('-', '_'))
  )
    continue;
  let hit = false;
  for (const lp of servedLegacy) {
    const a = lp.split('/');
    const b = r.path.split('/');
    if (a.length === b.length && a.every((seg, i) => seg === '*' || b[i] === '*' || seg === b[i])) {
      hit = true;
      break;
    }
  }
  if (!hit) unserved.add(`${r.app} ${r.path}`);
}
process.stdout.write(`${unserved.size} frontend request paths matched no module (stderr)\n`);
for (const u of [...unserved].sort()) process.stderr.write(`NO-MODULE ${u}\n`);
