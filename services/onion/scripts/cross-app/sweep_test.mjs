// Regression test for the round-two cross-app sweep scripts.
//
// Covers B10-R03 (tenant filters must be predicates), B22-R07 (an unreadable
// voxie contracts dir must fail instead of writing all-'no'), B42-R05 (exported
// surface must include methods and grouped declarations), B46-R05 (raw-SQL
// table sweep must not harvest Go prose and must see uppercase SQL) and
// B70-R04 (module-level plain chi routes belong in the handler inventory).
// Run locally: node scripts/cross-app/sweep_test.mjs
import { execFileSync } from 'node:child_process';
import { mkdirSync, mkdtempSync, readFileSync, rmSync } from 'node:fs';
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
  let res = run('sweep-queries.mjs');
  check(res.status === 0, 'sweep-queries runs', res.stderr.trim());
  const queries = parseCsv('backend_queries.csv');
  const query = (module, name) => find(find(queries, 'module', module), 'query_name', name)[0] || {};
  check(
    query('user', 'GetUserByID').has_business_id_filter === 'no',
    'GetUserByID is not reported as tenant-filtered',
    query('user', 'GetUserByID').has_business_id_filter,
  );
  check(
    query('hr_payroll', 'ListAllPayrollConfigs').has_business_id_filter === 'no',
    'ORDER BY business_id is not a tenant filter',
    query('hr_payroll', 'ListAllPayrollConfigs').has_business_id_filter,
  );
  check(
    query('careers', 'GetCustomizationBySlug').has_business_id_filter === 'no',
    'column-list business_id is not a tenant filter',
    query('careers', 'GetCustomizationBySlug').has_business_id_filter,
  );
  check(
    query('accounts', 'ListAccounts').has_business_id_filter === 'yes',
    'WHERE business_id = ? is still a tenant filter',
    query('accounts', 'ListAccounts').has_business_id_filter,
  );

  // ---------- tables: no prose phantoms, uppercase SQL is read ----------
  res = run('sweep-tables.mjs');
  check(res.status === 0, 'sweep-tables runs', res.stderr.trim());
  const tables = parseCsv('backend_tables.csv');
  const tableNames = new Set(tables.map((r) => r.table));
  const phantoms = ['becoming', 'hanging', 'redis', 'level', 'adding', 'modules'].filter((n) => tableNames.has(n));
  check(phantoms.length === 0, 'Go prose does not become a table', phantoms.join(','));
  const erp = find(tables, 'table', 'erp_products')[0] || {};
  check(
    (erp.modules_reading || '').split(' ').includes('platform/stock_core'),
    'uppercase FROM erp_products is attributed as a read',
    erp.modules_reading,
  );
  const aicache = find(tables, 'table', 'ai_cache_events')[0] || {};
  check(
    (aicache.modules_reading || '').split(' ').includes('platform/aicache'),
    'generated sqlc literal with a leading -- name: header is read',
    aicache.modules_reading,
  );

  // ---------- platform: methods and grouped declarations ----------
  res = run('sweep-platform.mjs');
  check(res.status === 0, 'sweep-platform runs', res.stderr.trim());
  const platform = parseCsv('platform_packages.csv');
  const surface = (pkg) => new Set((find(platform, 'package', pkg)[0]?.exported_surface || '').split(' '));
  const logger = surface('platform/logger');
  check(logger.has('Logger') && logger.has('Info') && logger.has('Handle'), 'logger methods are in the surface', [...logger].join(' '));
  check(surface('platform/utils').has('DaysPerWeek'), 'grouped const declarations are in the surface');
  check(surface('platform/money').has('MarshalJSON'), 'value-receiver methods are in the surface');

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
    'GET /console/self-host-licenses',
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
