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

import { existsSync, readdirSync, readFileSync, writeFileSync } from 'node:fs';
import { join, relative } from 'node:path';

const ROOT = new URL('../..', import.meta.url).pathname;
const OUT = join(ROOT, 'docs/cross-app/inventory/endpoints.csv');
// The shared packages stayed at the monorepo root when this workspace moved
// under apps/rtylr; only the workspace's own apps live under ROOT.
const PACKAGES = existsSync(join(ROOT, 'packages'))
  ? join(ROOT, 'packages')
  : join(ROOT, '..', '..', 'packages');
if (!existsSync(PACKAGES)) {
  console.error(`shared packages not found at ${PACKAGES}; expected a full checkout`);
  process.exit(1);
}

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
    const meta = {
      domain,
      capability,
      version,
      request: fieldName,
      response: fieldType,
      method: http ? http[1] : '',
      path: http ? `/${domain}/${capability}/${version}${http[2]}` : '',
    };
    rpcMeta.set(name, meta);
    // Frontend vars are often prefixed (InsightsProductServiceClient) while
    // the contract name is ProductService.getBestSellers, so index the bare
    // method too for a second-chance lookup.
    const bare = name.includes('.') ? name.split('.')[1] : name;
    if (!rpcMeta.has('bare:' + bare)) rpcMeta.set('bare:' + bare, []);
    rpcMeta.get('bare:' + bare).push(meta);
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
for (const app of readdirSync(join(ROOT, 'apps')))
  scopes.push({ app, dir: join(ROOT, 'apps', app), rel: 'apps/' + app });
for (const pkg of readdirSync(PACKAGES)) {
  const p = join(PACKAGES, pkg);
  try {
    if (readdirSync(p).includes('src'))
      scopes.push({ app: 'packages/' + pkg, dir: join(p, 'src'), rel: 'packages/' + pkg + '/src' });
  } catch {
    /* not a directory */
  }
}
// Round two: the pos terminal repo is swept as a pseudo-app when cloned.
import { homedir as _homedir } from 'node:os';
const POS_ROOT = process.env.CROSS_APP_POS || join(_homedir(), 'Development', 'pos');
if (existsSync(join(POS_ROOT, 'src'))) scopes.push({ app: 'pos', dir: POS_ROOT, rel: 'pos' });

// ---- Constant folding for <dynamic:...> paths (round two) -----------------
// Two passes: first collect every module-scope const whose initializer is a
// string literal or an object literal of strings, per scope; then resolve
// dotted identifier expressions like API_ROUTES.GET.GET_ALL at call sites.
function parseStringInit(src, eqIdx) {
  // Returns the raw literal contents starting after '=', or null.
  let i = eqIdx + 1;
  while (/\s/.test(src[i])) i++;
  const q = src[i];
  if (q === "'" || q === '"') {
    let j = i + 1;
    let out = '';
    while (j < src.length && src[j] !== q) {
      if (src[j] === '\\') {
        out += src[j + 1];
        j += 2;
      } else out += src[j++];
    }
    return { value: out, template: false };
  }
  if (q === '`') return { value: src.slice(i), template: true };
  if (q === '{') {
    // Object literal of nested string values.
    let depth = 0;
    const start = i;
    while (i < src.length) {
      if (src[i] === '{') depth++;
      else if (src[i] === '}') {
        depth--;
        if (depth === 0) break;
      }
      i++;
    }
    const body = src.slice(start + 1, i);
    const map = {};
    for (const mm of body.matchAll(/([A-Za-z_$][\w$]*)\s*:\s*(?:'([^']*)'|"([^"]*)")/g))
      map[mm[1]] = mm[2] ?? mm[3];
    for (const mm of body.matchAll(/([A-Za-z_$][\w$]*)\s*:\s*\{/g)) {
      // Nested objects handled by recursion on the substring.
      const key = mm[1];
      const subStart = body.indexOf('{', mm.index);
      let depth2 = 0;
      let k = subStart;
      while (k < body.length) {
        if (body[k] === '{') depth2++;
        else if (body[k] === '}') {
          depth2--;
          if (depth2 === 0) break;
        }
        k++;
      }
      map[key] = foldObject(body.slice(subStart + 1, k));
    }
    return { value: map };
  }
  return null;
}
function foldObject(body) {
  const map = {};
  for (const mm of body.matchAll(/([A-Za-z_$][\w$]*)\s*:\s*(?:'([^']*)'|"([^"]*)")/g))
    map[mm[1]] = mm[2] ?? mm[3];
  return map;
}
function buildConstMap(scopeAbs) {
  const map = new Map();
  let files;
  try {
    files = walk(scopeAbs);
  } catch {
    return { consts: map, onekitClients: new Map() };
  }
  // Round two: apps now centralise OneKit client instantiation in shared/lib
  // and import the variables elsewhere, so client declarations must be
  // collected for the whole scope, not per file.
  const onekitClients = new Map();
  const declRe = /(?:const|let)\s+(\w+)\s*=\s*createOneKitClient\(\s*(\w+)Client\b/g;
  const constRe = /(?:export\s+)?(?:const|let|var)\s+([A-Za-z_$][\w$]*)\s*=\s*/g;
  for (const f of files) {
    const src = readFileSync(f, 'utf8');
    let mm;
    while ((mm = constRe.exec(src))) {
      const parsed = parseStringInit(src, mm.index + mm[0].length - 1);
      if (parsed && !parsed.template && typeof parsed.value !== 'undefined')
        map.set(mm[1], parsed.value);
    }
    while ((mm = declRe.exec(src))) onekitClients.set(mm[1], mm[2]);
  }
  return { consts: map, onekitClients };
}
// Fold a template-literal body into a route shape, resolving known
// identifiers inside interpolations where possible.
function resolveTemplate(body, constMap) {
  return body.replace(/\$\{([^}]*)\}/g, (_, expr) => {
    const trimmed = expr.trim();
    const folded = resolveIdent(trimmed.replace(/^\w+\./, ''), constMap);
    if (folded) return folded;
    return '{' + trimmed.split(/[^\w$]/)[0] + '}';
  });
}
function resolveIdent(expr, constMap) {
  const parts = expr.split('.');
  let v = constMap.get(parts[0]);
  if (v === undefined) return null;
  for (let i = 1; i < parts.length; i++) {
    if (v && typeof v === 'object' && parts[i] in v) v = v[parts[i]];
    else return null;
  }
  return typeof v === 'string' ? v : null;
}

let unresolvedOnekit = 0;
for (const { app, dir, rel: relRoot } of scopes) {
  let files;
  try {
    files = walk(dir);
  } catch {
    continue;
  }
  // Per-scope constant table for folding dynamic paths.
  const { consts: constMap, onekitClients } = buildConstMap(dir);
  for (const file of files) {
    const rel = join(relRoot, relative(dir, file));
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
      // Round two: apps migrated to named transports (e.g. hrTransport) that
      // delegate straight to safe; treat them as the same idiom.
      const isTransport = /Transport$/.test(recv);
      const isAxiosish = ['api', 'publicApi', 'client', 'http'].includes(recv);
      if (!isSafe && !isAxiosish && !isTransport) continue;
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
        // Template literals become route shapes: /a/${id} -> /a/{id}.
        if (q === '`') {
          const foldedTpl = resolveTemplate(path, constMap);
          if (!foldedTpl.startsWith('/')) continue;
          path = foldedTpl;
        }
      } else {
        // Path built elsewhere: try to fold the identifier through this
        // scope's constant table before recording it as unresolved. Bare
        // parameter names (path, url) are pass-through wrapper definitions,
        // not request sites.
        const id = src.slice(i).match(/^[\w$.]+/);
        const expr = id ? id[0] : '';
        if (['path', 'url', 'input'].includes(expr)) continue;
        const folded = expr ? resolveIdent(expr, constMap) : null;
        path = folded ?? `<dynamic:${expr || 'expr'}>`;
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
        isSafe || isTransport ? 'safeFetch' : 'axios',
        domainOfPath(path),
      ]);
    }

    // --- Idiom: raw fetch (caret streaming, presigned uploads) --------------
    const fetchRe = /\bfetch\(\s*(['"`])((?:\\.|(?!\1).)*)\1|\bfetch\(\s*([\w$.]+)/g;
    while ((m = fetchRe.exec(src))) {
      const foldedFetch = m[3] ? resolveIdent(m[3], constMap) : null;
      const path = m[2] ?? foldedFetch ?? `<dynamic:${m[3]}>`;
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
    const clients = new Map(onekitClients); // varName -> ServiceName
    const declRe = /(?:const|let)\s+(\w+)\s*=\s*createOneKitClient\(\s*(\w+)Client\b/g;
    while ((m = declRe.exec(src))) clients.set(m[1], m[2]);
    if (clients.size) {
      const callRe = new RegExp(`\\b(${[...clients.keys()].join('|')})\\.(\\w+)\\(`, 'g');
      while ((m = callRe.exec(src))) {
        const service = clients.get(m[1]);
        const method = m[2];
        if (['toString', 'then', 'catch'].includes(method)) continue;
        let meta = rpcMeta.get(`${service}.${method}`);
        if (!meta) {
          // Second chance: strip the app prefix from the client-side service
          // name and match on the bare RPC name when it is unambiguous.
          const candidates = rpcMeta.get('bare:' + method);
          if (candidates && candidates.length === 1) meta = candidates[0];
        }
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
