// Inventory named queries across business modules and platform packages.
// Output: backend_queries.csv
// Columns: module, query_name, tables_read, tables_written,
//          has_business_id_filter, has_owner_id_filter, file, line
import { readFileSync } from 'node:fs';
import { walk, paths, writeCsv, relRepo, cteNames } from './lib.mjs';

const rows = [];

function tablesIn(sql) {
  // Strip comments so commented examples do not count.
  let t = sql.replace(/--[^\n]*/g, ' ').replace(/\/\*[\s\S]*?\*\//g, ' ').replace(/`/g, '');
  // ON DUPLICATE KEY UPDATE names columns, not tables.
  t = t.replace(/on\s+duplicate\s+key\s+update/gi, ' ');
  const reads = new Set();
  const writes = new Set();
  const ctes = cteNames(t);
  const push = (set, name) => {
    name = name.trim().replace(/[;,)]/g, '').split('.').at(-1).toLowerCase();
    if (!name || /\s/.test(name)) return;
    if (/^(select|from|where|group|order|left|inner|outer|join|on|and|or|set|values|as)$/.test(name)) return;
    if (set === reads && ctes.has(name.split('.').at(-1).replace(/`/g, '').toLowerCase())) return;
    set.add(name);
  };
  // FROM / JOIN clauses (reads), including subqueries. Strip write keywords
  // first so DELETE FROM does not double-count its target as a read.
  const tReads = t.replace(/\bdelete\s+from/gi, 'delete_from');
  for (const m of tReads.matchAll(/\bfrom\s+([a-zA-Z_][\w.]*)/gi)) push(reads, m[1]);
  for (const m of t.matchAll(/\b(?:left\s+|right\s+|inner\s+|outer\s+)?join\s+([a-zA-Z_][\w.]*)/gi)) push(reads, m[1]);
  // Writes.
  for (const m of t.matchAll(/\binsert\s+into\s+([a-zA-Z_][\w.]*)/gi)) push(writes, m[1]);
  for (const m of t.matchAll(/\bupdate\s+([a-zA-Z_][\w.]*)/gi)) push(writes, m[1]);
  for (const m of t.matchAll(/\bdelete\s+from\s+([a-zA-Z_][\w.]*)/gi)) push(writes, m[1]);
  // UPDATE ... FROM or DELETE ... USING also read their target? Keep it simple:
  // a CTE-referencing write still shows its read targets via from/join above.
  return { reads: [...reads], writes: [...writes] };
}

// Module query files.
const seen = new Set();
function addQueryFile(file, module) {
  if (seen.has(file)) return;
  seen.add(file);
  const t = readFileSync(file, 'utf8');
  const re = /--\s*name:\s*(\w+)\s*:([a-z]+)\s*\n([\s\S]*?)(?=;\s*\n|--\s*name:|$)/g;
  let m;
  while ((m = re.exec(t))) {
    const name = m[1];
    const kind = m[2];
    const body = m[3];
    const { reads, writes } = tablesIn(body);
    // Only occurrences inside the WHERE clause are tenant *filters*: a SELECT
    // column list, a DISTINCT, an ORDER BY or an INSERT column list mentions
    // business_id without scoping anything. Trailing GROUP BY / ORDER BY /
    // HAVING / LIMIT clauses are not predicates either.
    const seg = body.match(/\bwhere\b[\s\S]*?(?:$|\n\n)/i);
    const predText = seg
      ? seg[0].replace(/\b(?:group\s+by|order\s+by|having|limit)\b[\s\S]*$/i, '')
      : '';
    rows.push({
      module,
      query_name: name,
      kind,
      tables_read: reads.join(' '),
      tables_written: writes.join(' '),
      has_business_id_filter: /\bbusiness_id\b/.test(predText) ? 'yes' : 'no',
      has_owner_id_filter: /\bowner_id\b/.test(predText) ? 'yes' : 'no',
      file: relRepo(file),
      line: t.slice(0, m.index).split('\n').length,
    });
  }
}

for (const dir of walk(paths.modules, ['.sql'])) {
  if (!dir.includes('/sql/queries/') && !dir.includes('sqlc')) continue;
  const module = dir.slice(paths.modules.length + 1).split('/')[0];
  addQueryFile(dir, module);
}
// Some modules keep queries elsewhere (e.g. internal/**/queries). Catch any
// .sql under a module not yet seen that declares sqlc-style names.
for (const f of walk(paths.modules, ['.sql'])) {
  const module = f.slice(paths.modules.length + 1).split('/')[0];
  const t = readFileSync(f, 'utf8');
  if (/--\s*name:\s*\w+\s*:/.test(t)) addQueryFile(f, module);
}

// Platform packages with their own SQL.
for (const f of walk(paths.platform, ['.sql'])) {
  const owner = f.slice(paths.platform.length + 1).split('/')[0];
  addQueryFile(f, `platform/${owner}`);
}

rows.sort((a, b) => a.module.localeCompare(b.module) || a.query_name.localeCompare(b.query_name));
writeCsv(
  'backend_queries.csv',
  ['module', 'query_name', 'kind', 'tables_read', 'tables_written', 'has_business_id_filter', 'has_owner_id_filter', 'file', 'line'],
  rows.map((r) => [r.module, r.query_name, r.kind, r.tables_read, r.tables_written, r.has_business_id_filter, r.has_owner_id_filter, r.file, r.line]),
);
