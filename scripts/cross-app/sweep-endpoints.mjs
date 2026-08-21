#!/usr/bin/env node
// Sweeps every browser-leaving request across the three client idioms and
// emits docs/cross-app/inventory/endpoints.csv.
//
// Idioms swept (see docs/cross-app/01-inventory.md):
//   safeFetch  safe.get/post/put/patch/delete(...) on the SafeApi instance
//   onekit     createOneKitClient(SomeServiceClient, app) generated methods
//   axios      raw axios instances (apps/menu, and any api.get style caller
//              whose client file creates an axios instance directly)
//
// Test files (*.test.ts, *.test.tsx) are excluded: mocks are not requests.

import { readdirSync, readFileSync, writeFileSync } from 'node:fs';
import { join, relative } from 'node:path';

const ROOT = new URL('../..', import.meta.url).pathname;
const OUT = join(ROOT, 'docs/cross-app/inventory/endpoints.csv');

function walk(dir) {
  const out = [];
  for (const e of readdirSync(dir, { withFileTypes: true })) {
    if (e.name === 'node_modules' || e.name === 'dist' || e.name === 'gen') continue;
    const p = join(dir, e.name);
    if (e.isDirectory()) out.push(...walk(p));
    else if (/\.(ts|tsx|js|jsx)$/.test(e.name) && !/\.test\.(ts|tsx|js|jsx)$/.test(e.name))
      out.push(p);
  }
  return out;
}

function csv(v) {
  const s = String(v ?? '');
  return /[",\n]/.test(s) ? `"${s.replaceAll('"', '""')}"` : s;
}

function lineOf(src, idx) {
  let n = 1;
  for (let i = 0; i < idx; i++) if (src[i] === '\n') n++;
  return n;
}

// Nearest enclosing named function or exported const above the call site.
function enclosingName(src, idx) {
  const head = src.slice(0, idx);
  const re = /(?:export\s+)?(?:async\s+)?(?:function|const)\s+([A-Za-z_$][\w$]*)/g;
  let m;
  let name = '';
  while ((m = re.exec(head))) name = m[1];
  return name;
}

function nearbyQueryKey(src, idx) {
  const windowText = src.slice(Math.max(0, idx - 800), idx + 200);
  const m = [...windowText.matchAll(/queryKey:\s*(\[[^\]\n]*\])/g)];
  return m.length ? m[m.length - 1][1].replace(/\s+/g, ' ') : '';
}

function featureOf(rel) {
  const m =
    rel.match(/src\/features\/([^/]+)/) ||
    rel.match(/src\/shared\/(hooks\/queries|services|stores|lib)/) ||
    rel.match(/src\/(shared|stores|layouts|features)/);
  return m ? m[1] : '';
}

// Load RPC metadata from contracts.csv: ServiceName.method -> details.
const rpcMeta = new Map();
{
  const lines = readFileSync(join(ROOT, 'docs/cross-app/inventory/contracts.csv'), 'utf8')
    .split('\n')
    .slice(1);
  for (const line of lines) {
    if (!line) continue;
    // contracts.csv quotes only doc_comment; split defensively on the first 8 commas.
    const cols = [];
    let rest = line;
    for (let i = 0; i < 8; i++) {
      const j = rest.indexOf(',');
      cols.push(rest.slice(0, j));
      rest = rest.slice(j + 1);
    }
    const [domain, capability, version, kind, name, fieldName, fieldType] = cols;
    if (kind !== 'rpc') continue;
    const http = rest.match(/@(\w+)\("([^"]*)"\)/);
    rpcMeta.set(name, {
      domain,
      capability,
      version,
      request: fieldName,
      response: fieldType,
      method: http ? http[1] : '',
      path: http ? `/${domain}/${capability}/${version}${http[2]}` : '',
    });
  }
}

function domainOfPath(p) {
  const m = p.match(
    /^\/?(booking|common|console|crm|erp|finance|flow|hr|hub|insights|pos|support)\//
  );
  return m ? m[1] : '';
}

const rows = [
  [
    'app',
    'feature',
    'hook_or_service',
    'method',
    'path_or_rpc',
    'request_type',
    'response_type',
    'query_key',
    'file',
    'line',
    'idiom',
    'backend_domain',
  ],
];

const scopes = [];
for (const app of readdirSync(join(ROOT, 'apps'))) scopes.push(['apps/' + app, app]);
for (const pkg of readdirSync(join(ROOT, 'packages'))) {
  const p = join(ROOT, 'packages', pkg);
  try {
    if (readdirSync(p).includes('src'))
      scopes.push(['packages/' + pkg + '/src', 'packages/' + pkg]);
  } catch {
    /* not a directory */
  }
}

let unresolvedOnekit = 0;
for (const [scope, app] of scopes) {
  let files;
  try {
    files = walk(join(ROOT, scope));
  } catch {
    continue;
  }
  for (const file of files) {
    const rel = relative(ROOT, file);
    // Blank out full-line comments so documentation examples (e.g. the
    // api.get('/user/me') sample in AppGuard's JSDoc) are not counted as
    // requests. Offsets are preserved so file:line citations stay exact.
    const src = readFileSync(file, 'utf8')
      .split('\n')
      .map((l) => (/^\s*(\/\/|\*|\/\*)/.test(l) ? ' '.repeat(l.length) : l))
      .join('\n');

    // --- Idiom: safeFetch and axios-style verb calls ------------------------
    // A regex cannot span generics that contain ';', '(' or nested braces
    // (e.g. safe.get<{ a?: Record<string, unknown> }>(...)), so parse forward
    // manually from each <recv>.<verb> hit: consume a balanced <...> generic if
    // present, expect '(', then read the first argument.
    const headRe = /\b([A-Za-z_$][\w$]*)\.(get|post|put|patch|delete)\b/g;
    let m;
    while ((m = headRe.exec(src))) {
      const [, recv, verb] = m;
      const isSafe = recv === 'safe';
      const isAxiosish = ['api', 'publicApi', 'client', 'http'].includes(recv);
      if (!isSafe && !isAxiosish) continue;
      let i = m.index + m[0].length;
      let generic = '';
      if (src[i] === '<') {
        let depth = 0;
        const start = i;
        while (i < src.length) {
          if (src[i] === '<') depth++;
          else if (src[i] === '>') {
            depth--;
            if (depth === 0) {
              i++;
              break;
            }
          }
          i++;
        }
        generic = src.slice(start + 1, i - 1);
      }
      if (src[i] !== '(') continue; // property access, not a call
      i++;
      while (/\s/.test(src[i])) i++;
      let path;
      const q = src[i];
      if (q === "'" || q === '"' || q === '`') {
        let j = i + 1;
        while (j < src.length && src[j] !== q) {
          if (src[j] === '\\') j++;
          j++;
        }
        path = src.slice(i + 1, j);
        if (!path.startsWith('/')) continue;
      } else {
        // Path built elsewhere (a const or an expression). Record the
        // identifier so the row is countable rather than silently dropped.
        const id = src.slice(i).match(/^[\w$.]+/);
        path = `<dynamic:${id ? id[0] : 'expr'}>`;
      }
      rows.push([
        app,
        featureOf(rel),
        enclosingName(src, m.index),
        verb.toUpperCase(),
        path,
        '',
        generic.replace(/\s+/g, ' '),
        nearbyQueryKey(src, m.index),
        rel,
        lineOf(src, m.index),
        isSafe ? 'safeFetch' : 'axios',
        domainOfPath(path),
      ]);
    }

    // --- Idiom: raw fetch (caret streaming, presigned uploads) --------------
    const fetchRe = /\bfetch\(\s*(['"`])((?:\\.|(?!\1).)*)\1|\bfetch\(\s*([\w$.]+)/g;
    while ((m = fetchRe.exec(src))) {
      const path = m[2] ?? `<dynamic:${m[3]}>`;
      if (m[3] === 'input') continue; // pass-through wrapper, not a request site
      rows.push([
        app,
        featureOf(rel),
        enclosingName(src, m.index),
        '',
        path,
        '',
        '',
        nearbyQueryKey(src, m.index),
        rel,
        lineOf(src, m.index),
        'fetch',
        domainOfPath(path.replace(/^\$\{[^}]*\}/, '')),
      ]);
    }

    // --- Idiom: OneKit generated clients ------------------------------------
    const clients = new Map(); // varName -> ServiceName
    const declRe = /(?:const|let)\s+(\w+)\s*=\s*createOneKitClient\(\s*(\w+)Client\b/g;
    while ((m = declRe.exec(src))) clients.set(m[1], m[2]);
    if (clients.size) {
      const callRe = new RegExp(`\\b(${[...clients.keys()].join('|')})\\.(\\w+)\\(`, 'g');
      while ((m = callRe.exec(src))) {
        const service = clients.get(m[1]);
        const method = m[2];
        if (['toString', 'then', 'catch'].includes(method)) continue;
        const meta = rpcMeta.get(`${service}.${method}`);
        if (!meta) {
          unresolvedOnekit++;
          process.stderr.write(`NO-RPC ${rel}:${lineOf(src, m.index)}: ${service}.${method}\n`);
        }
        rows.push([
          app,
          featureOf(rel),
          enclosingName(src, m.index),
          meta ? meta.method.toUpperCase() : '',
          `${service}.${method}`,
          meta ? meta.request : '',
          meta ? meta.response : '',
          nearbyQueryKey(src, m.index),
          rel,
          lineOf(src, m.index),
          'onekit',
          meta ? meta.domain : '',
        ]);
      }
    }
  }
}

writeFileSync(OUT, rows.map((r) => r.map(csv).join(',')).join('\n') + '\n');
process.stdout.write(
  `${rows.length - 1} requests -> ${relative(ROOT, OUT)} (${unresolvedOnekit} onekit calls had no contract RPC)\n`
);
