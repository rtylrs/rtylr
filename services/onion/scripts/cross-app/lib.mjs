// Shared helpers for the round-two backend sweeps.
// Scripts write their CSVs into rtylr's docs/cross-app/inventory/ so one
// directory holds round one's five CSVs plus these seven.
import { readdirSync, statSync, mkdirSync, writeFileSync } from 'node:fs';
import { join, dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const here = dirname(fileURLToPath(import.meta.url));

export const ONION = process.env.CROSS_APP_ONION || resolve(here, '..', '..');
export const RTYLR = process.env.CROSS_APP_RTYLR || resolve(ONION, '..', '..');
export const OUT = process.env.CROSS_APP_OUT || join(RTYLR, 'docs', 'cross-app', 'inventory');

const CSV_PATHS = join(ONION, 'internal', 'modules');
const PLATFORM_PATHS = join(ONION, 'internal', 'platform');
const SCHEMA_ROOTS = [
  join(ONION, 'internal', 'modules'),
  join(ONION, 'internal', 'platform'),
];

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

export function modulePackages() {
  const packages = [];
  for (const root of moduleDirs()) {
    for (const file of walk(root.path, ['module.go'])) {
      const path = dirname(file);
      const relative = path.slice(root.path.length + 1);
      packages.push({ name: root.name, path, packagePath: relative });
    }
  }
  return packages;
}

export const paths = {
  modules: CSV_PATHS,
  platform: PLATFORM_PATHS,
  commands: join(ONION, 'cmd'),
  schemaRoots: SCHEMA_ROOTS,
};

export function schemaFiles() {
  return SCHEMA_ROOTS.flatMap((root) => walk(root, ['.sql'])).filter((file) => {
    const parts = file.split('/');
    const sqlIndex = parts.lastIndexOf('sql');
    const child = parts[sqlIndex + 1] || '';
    return sqlIndex >= 0 && (
      ['schema', 'schema.sql'].includes(child) ||
      (parts.length === sqlIndex + 2 && /^migration_.*\.sql$/.test(child))
    );
  });
}

export function cteNames(sql) {
  const source = sql.replace(/--[^\n]*/g, ' ').replace(/\/\*[\s\S]*?\*\//g, ' ');
  const names = new Set();
  const withStart = /^\s*with\s+(?:recursive\s+)?/i.exec(source);
  if (!withStart) return names;

  let i = withStart[0].length;
  const whitespace = () => {
    while (/\s/.test(source[i] || '')) i++;
  };
  const identifier = () => {
    whitespace();
    if (source[i] === '`') {
      const end = source.indexOf('`', i + 1);
      if (end < 0) return '';
      const value = source.slice(i + 1, end);
      i = end + 1;
      return value;
    }
    const match = /^[a-zA-Z_]\w*/.exec(source.slice(i));
    if (!match) return '';
    i += match[0].length;
    return match[0];
  };
  const parenthesized = () => {
    if (source[i] !== '(') return false;
    let depth = 0;
    let quote = '';
    while (i < source.length) {
      const c = source[i++];
      if (quote) {
        if (c === '\\') i++;
        else if (c === quote) {
          if (source[i] === quote && quote !== "'") i++;
          else quote = '';
        }
        continue;
      }
      if (c === "'" || c === '"' || c === '`') quote = c;
      else if (c === '(') depth++;
      else if (c === ')' && --depth === 0) return true;
    }
    return false;
  };

  while (i < source.length) {
    const name = identifier();
    if (!name) break;
    whitespace();
    if (source[i] === '(') {
      if (!parenthesized()) break;
      whitespace();
    }
    const as = /^as\b/i.exec(source.slice(i));
    if (!as) break;
    i += as[0].length;
    whitespace();
    if (!parenthesized()) break;
    names.add(name.toLowerCase());
    whitespace();
    if (source[i] !== ',') break;
    i++;
  }
  return names;
}

export function schemaDomain(file) {
  const parts = relRepo(file).split('/');
  if (parts[1] === 'modules') return parts[2];
  if (parts[1] === 'platform') return `platform/${parts[2]}`;
  return 'unknown';
}

export function voxieApiDir() {
 return process.env.CROSS_APP_VOXIE || join(RTYLR, 'voxie', 'api');
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
