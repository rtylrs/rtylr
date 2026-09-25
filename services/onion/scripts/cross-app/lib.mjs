// Shared helpers for the round-two backend sweeps.
// Scripts write their CSVs into rtylr's docs/cross-app/inventory/ so one
// directory holds round one's five CSVs plus these seven.
import { readdirSync, statSync, mkdirSync, writeFileSync } from 'node:fs';
import { join, dirname, resolve } from 'node:path';
import { homedir } from 'node:os';
import { fileURLToPath } from 'node:url';

const here = dirname(fileURLToPath(import.meta.url));

// ONION resolves wherever this checkout lives; this repo may be checked out
// in a worktree rather than at ~/Development/onion.
export const ONION = resolve(here, '..', '..');
// The audit data always lands in the canonical rtylr clone so one directory
// holds round one's CSVs plus these seven.
export const RTYLR =
  process.env.CROSS_APP_RTYLR || join(homedir(), 'Development', 'rtylr');
export const OUT =
  process.env.CROSS_APP_OUT || join(RTYLR, 'docs', 'cross-app', 'inventory');

const CSV_PATHS = join(ONION, 'internal', 'modules');
const PLATFORM_PATHS = join(ONION, 'internal', 'platform');
const SCHEMA_PATHS = join(ONION, 'internal', 'platform', 'database', 'schemas');

export function walk(dir, exts, out = []) {
  let entries;
  try {
    entries = readdirSync(dir);
  } catch {
    return out;
  }
  for (const e of entries) {
    const p = join(dir, e);
    let s;
    try {
      s = statSync(p);
    } catch {
      continue;
    }
    if (s.isDirectory()) {
      if (e === '.remember' || e === 'node_modules') continue;
      walk(p, exts, out);
    } else if (exts.some((x) => e.endsWith(x))) {
      out.push(p);
    }
  }
  return out;
}

export function moduleDirs() {
  return readdirSync(CSV_PATHS)
    .filter((d) => !d.endsWith('.md'))
    .sort()
    .map((d) => ({ name: d, path: join(CSV_PATHS, d) }));
}

export const paths = {
  modules: CSV_PATHS,
  platform: PLATFORM_PATHS,
  schemas: SCHEMA_PATHS,
};

// Voxie lives as this repo's submodule when initialised, otherwise as the
// sibling checkout. Contracts are read from whichever exists; CROSS_APP_VOXIE
// overrides both when the checkout lives somewhere else.
import { existsSync } from 'node:fs';
export function voxieApiDir() {
  if (process.env.CROSS_APP_VOXIE) return process.env.CROSS_APP_VOXIE;
  const sub = join(ONION, 'voxie', 'api');
  if (existsSync(sub)) return sub;
  return join(homedir(), 'Development', 'voxie', 'api');
}

export function csvCell(v) {
  const s = v == null ? '' : String(v);
  // Keep file:line references readable; quote only when needed.
  return /[",\n]/.test(s) ? '"' + s.replace(/"/g, '""') + '"' : s;
}

export function writeCsv(name, header, rows) {
  mkdirSync(OUT, { recursive: true });
  const body = [header.join(','), ...rows.map((r) => r.map(csvCell).join(','))].join('\n');
  const p = join(OUT, name);
  writeFileSync(p, body + '\n');
  console.log(`wrote ${p} (${rows.length} rows)`);
}

// RelPath makes citations stable across checkouts: everything is cited from
// the repo root it lives in.
export function relRepo(abs) {
  return abs.startsWith(ONION) ? abs.slice(ONION.length + 1) : abs;
}

// lineOf returns the 1-based line number of the first index of `needle`.
export function lineOf(text, needle) {
  const i = text.indexOf(needle);
  if (i < 0) return 0;
  return text.slice(0, i).split('\n').length;
}
