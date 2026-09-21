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
import { OUT } from './lib.mjs';
const qcsv = rf(join(OUT, 'backend_queries.csv'), 'utf8');
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
// Go source text is scanned for SQL string literals only: prose in comments and
// log/error messages ("stock_core: update level") must not become tables. The
// scan returns every double-quoted or backtick literal with comments and rune
// literals skipped.
function sqlLiterals(text) {
  const out = [];
  let i = 0;
  while (i < text.length) {
    const c = text[i];
    if (c === '/' && text[i + 1] === '/') {
      const nl = text.indexOf('\n', i);
      i = nl < 0 ? text.length : nl + 1;
      continue;
    }
    if (c === '/' && text[i + 1] === '*') {
      const end = text.indexOf('*/', i + 2);
      i = end < 0 ? text.length : end + 2;
      continue;
    }
    if (c === "'") {
      // Rune literal: skip it so quotes and slashes inside are not read as code.
      i++;
      while (i < text.length && text[i] !== "'") i += text[i] === '\\' ? 2 : 1;
      i++;
      continue;
    }
    if (c === '"' || c === '`') {
      const quote = c;
      let j = i + 1;
      let lit = '';
      while (j < text.length && text[j] !== quote) {
        if (quote === '"' && text[j] === '\\') {
          lit += text[j + 1] ?? '';
          j += 2;
          continue;
        }
        lit += text[j];
        j++;
      }
      out.push(lit);
      i = j + 1;
      continue;
    }
    i++;
  }
  return out;
}

// rawTablesIn reads a SQL statement: FROM/JOIN read, INTO/UPDATE/DELETE FROM
// write. ON DUPLICATE KEY UPDATE names columns, not tables.
function rawTablesIn(sql) {
  let q = sql.replace(/--[^\n]*/g, ' ').replace(/\/\*[\s\S]*?\*\//g, ' ');
  q = q.replace(/\bon\s+duplicate\s+key\s+update/gi, ' ');
  // Strip DELETE FROM first so its target is not also counted as a read.
  q = q.replace(/\bdelete\s+from/gi, 'delete_from');
  const reads = [];
  const writes = [];
  for (const m of q.matchAll(/\b(?:from|join)\s+([a-zA-Z_]\w*)/gi)) reads.push(m[1]);
  for (const m of q.matchAll(/\b(?:into|update)\s+([a-zA-Z_]\w*)/gi)) writes.push(m[1]);
  for (const m of q.matchAll(/\bdelete_from\s+([a-zA-Z_]\w*)/gi)) writes.push(m[1]);
  return { reads, writes };
}

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
  // Raw SQL string literals (non-squirrel). Only literals whose first
  // statement opens with a SQL verb are scanned; error prose that merely
  // contains "update <word>" is not, and the case-insensitive match catches
  // uppercase SQL. A generated sqlc literal starts with its "-- name:" header,
  // so leading comments are skipped before the verb check.
  for (const sql of sqlLiterals(t)) {
    const head = sql.replace(/^(?:\s+|--[^\n]*|\/\*[\s\S]*?\*\/)*/, '');
    if (!/^(?:select|insert|replace|update|delete|with)\b/i.test(head)) continue;
    const { reads, writes } = rawTablesIn(sql);
    for (const name of reads) touch(name, mod, 'read');
    for (const name of writes) touch(name, mod, 'write');
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
