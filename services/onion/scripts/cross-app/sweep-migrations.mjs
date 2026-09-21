// Sweep the schema/migration corpus: every SQL file under
// internal/platform/database/schemas, the tables it creates or alters, and
// whether its owning domain has a matching voxie .onk domain.
// Output: migrations.csv
import { readFileSync } from 'node:fs';
import { readdirSync, statSync, existsSync } from 'node:fs';
import { join } from 'node:path';
import { walk, paths, writeCsv, voxieApiDir } from './lib.mjs';

const apiDir = voxieApiDir();

// Contract domains on disk. An unreadable directory means the voxie checkout
// is missing (uninitialised submodule / no sibling clone): failing loudly beats
// writing every domain as has_contract_domain=no, which reads as "no domain has
// a contract" in the inventory.
let apiEntries;
try {
  apiEntries = readdirSync(apiDir);
} catch (err) {
  console.error(`voxie contracts directory is not readable: ${apiDir} (${err.code || err.message})`);
  console.error('initialise the voxie submodule or set CROSS_APP_VOXIE to the voxie/api checkout');
  process.exit(1);
}
const contractDomains = new Set();
for (const d of apiEntries) {
  try {
    if (statSync(join(apiDir, d)).isDirectory()) contractDomains.add(d);
  } catch {}
}

const rows = [];
for (const f of walk(paths.schemas, ['.sql'])) {
  const rel = f.slice(paths.schemas.length + 1);
  const domain = rel.split('/')[0];
  const t = readFileSync(f, 'utf8');
  const created = new Set();
  const altered = new Set();
  for (const m of t.matchAll(/create\s+table\s+(?:if\s+not\s+exists\s+)?`?(\w+)`?/gi))
    created.add(m[1]);
  for (const m of t.matchAll(/alter\s+table\s+`?(\w+)`?/gi)) altered.add(m[1]);
  // Views and indexes are not tracked as tables but counted for coverage.
  const views = [...t.matchAll(/create\s+(?:or\s+replace\s+)?(?:algorithm=\w+\s+)?sql\s+security\s+\w+\s+view\s+(\w+)|create\s+view\s+(\w+)/gi)].length;
  const hasContract = contractDomains.has(domain);
  rows.push({
    file: `internal/platform/database/schemas/${rel}`,
    domain,
    tables_created: [...created].sort().join(' '),
    tables_altered: [...altered].sort().join(' '),
    view_count: views,
    has_contract_domain: hasContract ? 'yes' : 'no',
  });
}
rows.sort((a, b) => a.domain.localeCompare(b.domain) || a.file.localeCompare(b.file));
writeCsv(
  'migrations.csv',
  ['file', 'domain', 'tables_created', 'tables_altered', 'view_count', 'has_contract_domain'],
  rows.map((r) => [r.file, r.domain, r.tables_created, r.tables_altered, r.view_count, r.has_contract_domain]),
);
console.log(`domains with no contract dir: ${[...new Set(rows.filter(r=>r.has_contract_domain==='no').map(r=>r.domain))].join(', ')}`);
