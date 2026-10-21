// Regression test for the round-two cross-app sweep scripts.
//
// Covers B10-R03 (tenant filters must be predicates), B22-R07 (an unreadable
// voxie contracts dir must fail instead of writing all-'no'), B42-R05 (exported
// surface must include methods and grouped declarations), B46-R05 (raw-SQL
// table sweep must not harvest Go prose and must see uppercase SQL) and
// B70-R04 (module-level plain chi routes belong in the handler inventory).
// Run locally: node scripts/cross-app/sweep_test.mjs
import { execFileSync } from 'node:child_process';
import { mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const here = dirname(fileURLToPath(import.meta.url));

let PASS = 0;
let FAIL = 0;
const ok = (name) => {
  PASS++;
  console.log(`  ok: ${name}`);
};
const fail = (name, detail = '') => {
  FAIL++;
  console.error(`  FAIL: ${name}${detail ? ` (${detail})` : ''}`);
};
const check = (cond, name, detail = '') => (cond ? ok(name) : fail(name, detail));

const work = mkdtempSync(join(tmpdir(), 'sweep-test-'));
const out = join(work, 'inventory');
const contracts = join(work, 'contracts');
const fixture = join(work, 'onion-fixture');
// A contracts checkout with two real schema domains, so the migrations sweep
// does not depend on whether the voxie submodule is initialised.
mkdirSync(join(contracts, 'crm'), { recursive: true });
mkdirSync(join(contracts, 'shared'), { recursive: true });

function run(script, env = {}) {
  try {
    // Keep the child's stderr in the error object instead of echoing it over
    // passing test output.
    const stdout = execFileSync(process.execPath, [join(here, script)], {
      env: { ...process.env, CROSS_APP_OUT: out, ...env },
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'pipe'],
    });
    return { status: 0, stdout, stderr: '' };
  } catch (err) {
    return { status: err.status ?? 1, stdout: err.stdout ?? '', stderr: err.stderr ?? '' };
  }
}

function parseCsv(file) {
  const text = readFileSync(join(out, file), 'utf8');
  const rows = [];
  let row = [];
  let cell = '';
  let quoted = false;
  for (let i = 0; i < text.length; i++) {
    const c = text[i];
    if (quoted) {
      if (c === '"' && text[i + 1] === '"') {
        cell += '"';
        i++;
      } else if (c === '"') quoted = false;
      else cell += c;
      continue;
    }
    if (c === '"') quoted = true;
    else if (c === ',') {
      row.push(cell);
      cell = '';
    } else if (c === '\n') {
      row.push(cell);
      rows.push(row);
      row = [];
      cell = '';
    } else cell += c;
  }
  if (cell !== '' || row.length) {
    row.push(cell);
    rows.push(row);
  }
  const header = rows.shift() || [];
  return rows
    .filter((r) => r.length > 1 || r[0] !== '')
    .map((r) => Object.fromEntries(header.map((h, i) => [h, r[i] ?? ''])));
}

const find = (rows, field, value) => rows.filter((r) => r[field] === value);

try {
  // ---------- queries: tenant filters are predicates ----------
  const queryModule = join(fixture, 'internal', 'modules', 'query_owner', 'sql', 'queries');
  mkdirSync(queryModule, { recursive: true });
  writeFileSync(join(queryModule, 'scoping.sql'),
    '-- name: ByID :one\nSELECT business_id FROM sample_records WHERE id = ?;\n' +
    '-- name: Ordered :many\nSELECT id FROM sample_records ORDER BY business_id;\n' +
    '-- name: Scoped :many\nSELECT id FROM sample_records WHERE business_id = ?;\n');
  let res = run('sweep-queries.mjs', { CROSS_APP_ONION: fixture });
  check(res.status === 0, 'fixture scoping query sweep runs', res.stderr.trim());
  const queries = parseCsv('backend_queries.csv');
  const query = (name) => find(queries, 'query_name', name)[0] || {};
  check(query('ByID').has_business_id_filter === 'no', 'selected tenant column is not a filter');
  check(query('Ordered').has_business_id_filter === 'no', 'ORDER BY is not a tenant filter');
  check(query('Scoped').has_business_id_filter === 'yes', 'WHERE business_id is a tenant filter');
  res = run('sweep-queries.mjs');
  check(res.status === 0, 'repository query sweep runs', res.stderr.trim());

  // ---------- tables: no prose phantoms, uppercase SQL is read ----------
  res = run('sweep-tables.mjs');
  check(res.status === 0, 'sweep-tables runs', res.stderr.trim());
  const tables = parseCsv('backend_tables.csv');
  const tableNames = new Set(tables.map((r) => r.table));
  const phantoms = ['becoming', 'hanging', 'redis', 'level', 'adding', 'modules'].filter((n) => tableNames.has(n));
  check(phantoms.length === 0, 'Go prose does not become a table', phantoms.join(','));
  const erp = find(tables, 'table', 'erp_products')[0] || {};
  check(
    (erp.modules_reading || '').split(' ').includes('erp'),
    'ERP-owned stock and product queries are attributed to the ERP context',
    erp.modules_reading,
  );
  const aicache = find(tables, 'table', 'ai_cache_events')[0] || {};
  check(
    (aicache.modules_reading || '').split(' ').includes('platform/aicache'),
    'generated sqlc literal with a leading -- name: header is read',
    aicache.modules_reading,
  );

  // The raw SQL scanner must recognize SQL keywords without depending on a
  // particular production query or its owning module layout.
  const rawModule = join(fixture, 'internal', 'modules', 'sample');
  mkdirSync(join(rawModule, 'sql', 'schema', 'sample'), { recursive: true });
  mkdirSync(join(rawModule, 'sql', 'queries', 'sample'), { recursive: true });
  mkdirSync(join(rawModule, 'internal', 'infrastructure'), { recursive: true });
  writeFileSync(
    join(rawModule, 'sql', 'schema', 'sample', 'schema.sql'),
    'CREATE TABLE raw_sql_fixture (id BIGINT PRIMARY KEY);\n' +
      'CREATE TABLE fixture_projects (id BIGINT PRIMARY KEY, business_id BIGINT NOT NULL);\n' +
      'CREATE TABLE fixture_tasks (id BIGINT PRIMARY KEY, project_id BIGINT NOT NULL);\n',
  );
  writeFileSync(
    join(rawModule, 'sql', 'migration_20261004_fixture.sql'),
    'CREATE TABLE fixture_migrated (id BIGINT PRIMARY KEY);\n' +
      'CREATE VIEW fixture_projection AS SELECT id FROM external.fixture_source;\n',
  );
  writeFileSync(
    join(rawModule, 'sql', 'queries', 'sample', 'projects.sql'),
    '-- name: ListProjectPage :many\n' +
      'WITH project_page AS (SELECT id FROM fixture_projects WHERE business_id = ?),\n' +
      'task_counts AS (SELECT project_id, COUNT(*) AS total FROM fixture_tasks GROUP BY project_id)\n' +
      'SELECT p.id, t.total FROM project_page p JOIN task_counts t ON t.project_id = p.id;\n',
  );
  writeFileSync(
    join(rawModule, 'internal', 'infrastructure', 'repository.go'),
    'package infrastructure\n' +
      'func query(db interface{ Query(string, ...any) }) { db.Query("SELECT * FROM `raw_sql_fixture`") }\n' +
      'var query = builder.From("external.qualified_source q").LeftJoin("aliased_source a ON a.id = q.id")\n' +
      'var metadata = "SELECT TABLE_NAME FROM information_schema.tables"\n' +
      'var messages = []string{"Update business is not implemented", "update active state for", "Update selected fields on a product"}\n',
  );
  res = run('sweep-queries.mjs', { CROSS_APP_ONION: fixture });
  check(res.status === 0, 'fixture query sweep runs', res.stderr.trim());
  const fixtureQuery = find(parseCsv('backend_queries.csv'), 'query_name', 'ListProjectPage')[0] || {};
  check(
    fixtureQuery.tables_read === 'fixture_projects fixture_tasks',
    'CTE aliases are excluded from query table reads',
    fixtureQuery.tables_read,
  );
  res = run('sweep-tables.mjs', { CROSS_APP_ONION: fixture });
  check(res.status === 0, 'fixture table sweep runs', res.stderr.trim());
  const fixtureTables = parseCsv('backend_tables.csv');
  const fixtureTable = find(fixtureTables, 'table', 'raw_sql_fixture')[0] || {};
  const migratedTable = find(fixtureTables, 'table', 'fixture_migrated')[0] || {};
  check(
    (fixtureTable.modules_reading || '').split(' ').includes('sample'),
    'uppercase raw SQL keywords are recognized as a table read',
    fixtureTable.modules_reading,
  );
  check(
    migratedTable.owner_schema_domain === 'sample',
    'module-owned migrations define their table owner',
    migratedTable.owner_schema_domain,
  );
  for (const name of ['fixture_source', 'qualified_source', 'aliased_source']) {
    check(
      (find(fixtureTables, 'table', name)[0]?.modules_reading || '').split(' ').includes('sample'),
      `view and aliased builder reads include ${name}`,
    );
  }
  check(!fixtureTables.some((row) => row.table === 'tables'), 'technical metadata is not an unowned business table');
  check(
    !fixtureTables.some((row) => ['project_page', 'task_counts'].includes(row.table)),
    'CTE aliases do not become unowned table rows',
    fixtureTables.map((row) => row.table).filter((name) => ['project_page', 'task_counts'].includes(name)).join(' '),
  );
  check(
    !fixtureTables.some((row) => ['business', 'active', 'selected'].includes(row.table)),
    'SQL-looking error text does not become a write target',
    fixtureTables.map((row) => row.table).filter((name) => ['business', 'active', 'selected'].includes(name)).join(' '),
  );

  // ---------- platform: methods and grouped declarations ----------
  mkdirSync(join(fixture, 'cmd', 'api'), { recursive: true });
  mkdirSync(join(fixture, 'internal', 'platform', 'cdn'), { recursive: true });
  writeFileSync(join(fixture, 'internal', 'platform', 'cdn', 'cdn.go'), 'package cdn\nfunc New() {}\n');
  writeFileSync(
    join(fixture, 'cmd', 'api', 'main.go'),
    'package main\nimport "github.com/voxire/onion/internal/platform/cdn"\nvar _ = cdn.New\n',
  );
  res = run('sweep-platform.mjs');
  check(res.status === 0, 'sweep-platform runs', res.stderr.trim());
  const platform = parseCsv('platform_packages.csv');
  const surface = (pkg) => new Set((find(platform, 'package', pkg)[0]?.exported_surface || '').split(' '));
  const logger = surface('platform/logger');
  check(logger.has('Logger') && logger.has('Info') && logger.has('Handle'), 'logger methods are in the surface', [...logger].join(' '));
  check(surface('platform/utils').has('DaysPerWeek'), 'grouped const declarations are in the surface');
  check(surface('platform/money').has('MarshalJSON'), 'value-receiver methods are in the surface');
  res = run('sweep-platform.mjs', { CROSS_APP_ONION: fixture });
  check(res.status === 0, 'fixture platform sweep runs', res.stderr.trim());
  const fixturePlatform = parseCsv('platform_packages.csv');
  const cdn = find(fixturePlatform, 'package', 'platform/cdn')[0] || {};
  check(
    cdn.verdict === 'live' && (cdn.callers || '').split(' ').includes('cmd/api'),
    'composition roots count as platform consumers',
    `${cdn.verdict || 'missing'} (${cdn.callers || 'no callers'})`,
  );

  // ---------- handlers: module-level plain chi routes ----------
  res = run('sweep-handlers.mjs');
  check(res.status === 0, 'sweep-handlers runs', res.stderr.trim());
  const handlers = parseCsv('backend_handlers.csv');
  const route = (rpc) => find(handlers, 'rpc_or_route', rpc)[0] || {};
  const anonymous = [
    'POST /subscriptions/paddle-webhook',
    'GET /flow/ws',
    'GET /dash/ws',
    'POST /pos/payments/webhook',
    'GET /pos/ws',
    'GET /oauth/authorize',
  ];
  const authenticated = [
    'POST /flow/ws-ticket',
    'POST /pos/ws-ticket',
    'GET /console/resources',
    'GET /console/resources/{table}',
    'GET /console/resources/{table}/rows',
    'GET /console/resources/{table}/rows/{id}',
  ];
  for (const rpc of [...anonymous, ...authenticated]) {
    const row = route(rpc);
    const want = anonymous.includes(rpc) ? 'no' : 'yes';
    check(!!row.module && row.auth_required === want, `${rpc} is inventoried (auth=${want})`);
  }

  // ---------- migrations: unreadable contracts dir fails loudly ----------
  const missingContracts = join(work, 'missing-voxie', 'api');
  res = run('sweep-migrations.mjs', { CROSS_APP_VOXIE: missingContracts });
  check(res.status !== 0, 'missing contracts dir exits non-zero');
  check(res.stderr.includes(missingContracts), 'missing contracts dir is named on stderr', res.stderr.trim());
  res = run('sweep-migrations.mjs', { CROSS_APP_VOXIE: contracts });
  check(res.status === 0, 'readable contracts dir succeeds', res.stderr.trim());
  const migrations = parseCsv('migrations.csv');
  check(
    find(migrations, 'domain', 'crm').every((r) => r.has_contract_domain === 'yes'),
    'crm rows see their contract domain',
  );
  check(
    find(migrations, 'domain', 'hr_time').every((r) => r.has_contract_domain === 'no'),
    'domains without a contract stay no',
  );
} finally {
  rmSync(work, { recursive: true, force: true });
}

console.log();
console.log(`  passed ${PASS}, failed ${FAIL}`);
process.exit(FAIL === 0 ? 0 : 1);
