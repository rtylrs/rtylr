// Sweep every table: its owning schema domain, and every module that reads or
// writes it via any idiom: sqlc query files AND raw squirrel/SQL strings in Go
// (e.g. pos_customers reading crm_customers straight from repository.go).
// Output: backend_tables.csv
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { walk, paths, writeCsv } from './lib.mjs';

// table -> { owner, readers:Set(module), writers:Set(module) }
const tables = new Map();
function touch(name, module, mode) {
  if (!tables.has(name)) tables.set(name, { owner: '', readers: new Set(), writers: new Set() });
  const t = tables.get(name);
  if (mode === 'read') t.readers.add(module);
  else t.writers.add(module);
}

const IGNORE = new Set(['select', 'dual', 'information_schema']);

// 1. Ownership from the schemas submodule: the directory that CREATE TABLEs it.
for (const f of walk(paths.schemas, ['.sql'])) {
  const domain = f.slice(paths.schemas.length + 1).split('/')[0];
  const t = readFileSync(f, 'utf8');
  for (const m of t.matchAll(/create\s+table\s+(?:if\s+not\s+exists\s+)?[`]?(\w+)[`]?/gi)) {
    const name = m[1];
    if (!tables.has(name)) tables.set(name, { owner: '', readers: new Set(), writers: new Set() });
    // First declarer wins; later duplicates recorded as re-declared below.
    if (!tables.get(name).owner) tables.get(name).owner = domain;
    else if (tables.get(name).owner !== domain)
      tables.get(name).owner += '|' + domain;
  }
}

// 2. sqlc query access (module attribution from the queries sweep output).
import { readFileSync as rf } from 'node:fs';
const qcsv = rf(join(paths.schemas, '..', '..', '..', '..', '..', 'rtylr', 'docs', 'cross-app', 'inventory', 'backend_queries.csv'), 'utf8');
for (const line of qcsv.split('\n').slice(1)) {
  if (!line.trim()) continue;
  // Parse conservatively: module is field 1, reads field 4, writes field 5.
  const cols = line.split(',');
  const mod = cols[0];
  for (const tbl of (cols[3] || '').split(' ')) if (tbl && !IGNORE.has(tbl)) touch(tbl, mod, 'read');
  for (const tbl of (cols[4] || '').split(' ')) if (tbl && !IGNORE.has(tbl)) touch(tbl, mod, 'write');
}

// 3. Raw squirrel / SQL strings inside Go code.
const fromRe = /\.(?:From|Join)\("([a-zA-Z_]\w*)"/g;
const insertRe = /\.Insert\("([a-zA-Z_]\w*)"/g;
const updateRe = /\.Update\("([a-zA-Z_]\w*)"/g;
const deleteRe = /\.Delete\("([a-zA-Z_]\w*)"/g;
const rawRe = /\b(?:FROM|INTO|UPDATE|JOIN)\s+([a-z][a-z0-9_]*)\b/g;

for (const f of walk(paths.modules, ['.go'])) {
  const mod = f.slice(paths.modules.length + 1).split('/')[0];
  if (mod === 'platform') continue;
  const t = readFileSync(f, 'utf8');
  let m;
  while ((m = fromRe.exec(t))) touch(m[1], mod, 'read');
  while ((m = insertRe.exec(t))) touch(m[1], mod, 'write');
  while ((m = updateRe.exec(t))) touch(m[1], mod, 'write');
  while ((m = deleteRe.exec(t))) touch(m[1], mod, 'write');
}
for (const f of walk(paths.platform, ['.go'])) {
  if (f.includes(`${join('database', 'schemas')}`)) continue;
  const rel = f.slice(paths.platform.length + 1);
  const pkg = rel.split('/')[0];
  const mod = `platform/${pkg}`;
  const t = readFileSync(f, 'utf8');
  let m;
  while ((m = fromRe.exec(t))) touch(m[1], mod, 'read');
  while ((m = insertRe.exec(t))) touch(m[1], mod, 'write');
  while ((m = updateRe.exec(t))) touch(m[1], mod, 'write');
  while ((m = deleteRe.exec(t))) touch(m[1], mod, 'write');
  // Raw SQL string literals (non-squirrel), lowercased word match.
  const rawLower = /\b(?:from|into|update|join)\s+([a-z][a-z0-9_]{3,})\b/g;
  while ((m = rawLower.exec(t))) {
    const name = m[1];
    if (/^(the|this|that|which|where|table|their|them|these|those|there)$/.test(name)) continue;
    touch(name, mod, /from|join/.test(m[0]) ? 'read' : 'write');
  }
}

const rows = [];
for (const [name, t] of [...tables.entries()].sort()) {
  const modules = new Set([...t.readers, ...t.writers]);
  if (!name || name.length < 3 || IGNORE.has(name)) continue;
  rows.push({
    table: name,
    owner_schema_domain: t.owner,
    modules_reading: [...t.readers].sort().join(' '),
    modules_writing: [...t.writers].sort().join(' '),
    module_count: modules.size,
    multi_module: modules.size > 1 ? 'yes' : 'no',
    owner_declared_in_schemas: t.owner ? 'yes' : 'NO-TABLE-CREATION-FOUND',
  });
}
rows.sort((a, b) => Number(b.multi_module === 'yes') - Number(a.multi_module === 'yes') || b.module_count - a.module_count || a.table.localeCompare(b.table));
writeCsv(
  'backend_tables.csv',
  ['table', 'owner_schema_domain', 'modules_reading', 'modules_writing', 'module_count', 'multi_module', 'owner_declared_in_schemas'],
  rows.map((r) => [r.table, r.owner_schema_domain, r.modules_reading, r.modules_writing, r.module_count, r.multi_module, r.owner_declared_in_schemas]),
);
