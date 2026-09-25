// Sweep every backend request surface: legacy chi route registrations and
// oneKit generated-service RPCs, plus platform-level routes.
// Output: backend_handlers.csv
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { walk, moduleDirs, paths, writeCsv, relRepo, lineOf, voxieApiDir } from './lib.mjs';

const rows = [];

// ---------- legacy chi handlers ----------
const goFiles = [];
for (const m of moduleDirs()) goFiles.push(...walk(m.path, ['.go']));

// Resolve XRoutes() helpers: name -> [{path, file}]. Uses a balanced-brace
// scan because many of these are one-liners and a lazy newline-anchored
// regex over-captures into following structs.
function funcBody(text, nameIdx) {
  const open = text.indexOf('{', nameIdx);
  if (open < 0) return null;
  let depth = 0;
  for (let i = open; i < text.length; i++) {
    const c = text[i];
    if (c === '"' || c === '`') {
      const q = c;
      i++;
      while (i < text.length && text[i] !== q) {
        if (text[i] === '\\') i++;
        i++;
      }
      continue;
    }
    if (c === '{') depth++;
    else if (c === '}') {
      depth--;
      if (depth === 0) return text.slice(open + 1, i);
    }
  }
  return null;
}

const routesFn = new Map();
for (const f of goFiles) {
  const t = readFileSync(f, 'utf8');
  const re = /func\s+(\w+Routes)\(\)\s+\[\]string/g;
  let m2;
  while ((m2 = re.exec(t))) {
    const body = funcBody(t, m2.index);
    if (!body) continue;
    const paths = [...body.matchAll(/"([^"]+)"/g)].map((x) => x[1]);
    if (paths.length)
      routesFn.set(m2[1], {
        paths,
        file: relRepo(f),
        line: t.slice(0, m2.index).split('\n').length,
      });
  }
}

// Anonymous (unauthenticated) route literals per module.
const anonByModule = new Map();
for (const m of moduleDirs()) {
  const modFile = join(m.path, 'module.go');
  try {
    const t = readFileSync(modFile, 'utf8');
    const re = /GetAnonymousRoutes\(\)\s+\[\]string\s*\{([\s\S]*?)\n\}/;
    const mm = re.exec(t);
    if (mm) {
      const p = [...mm[1].matchAll(/"([^"]+)"/g)].map((x) => x[1]);
      if (p.length) anonByModule.set(m.name, p);
    }
  } catch {}
}

function tenantPredicate(fileText) {
  const tokens = new Set();
  for (const tok of [
    'ResolveBusinessID',
    'device.User.BusinessID',
    'device.User.UserID',
    'businessID',
    'ownerID',
    'business_id',
  ]) {
    if (fileText.includes(tok)) tokens.add(tok);
  }
  return tokens.size ? [...tokens].sort().join('|') : 'none-found-in-file';
}

// Tenancy may live one hop below the handler, in the module's application
// service. Check the handler text first, then the service layer of the same
// module, and label which place the evidence came from.
function tenantEvidence(fileText, moduleName) {
  const direct = tenantPredicate(fileText);
  if (direct !== 'none-found-in-file') return direct;
  let serviceText = '';
  for (const f of walk(join(paths.modules, moduleName, 'internal', 'application'), ['.go'])) {
    serviceText += readFileSync(f, 'utf8');
  }
  if (serviceText) {
    const svc = tenantPredicate(serviceText);
    if (svc !== 'none-found-in-file') return `service-layer:${svc}`;
  }
  return 'none-found-in-module';
}

// Balanced scanner: given the index of the opening paren of a call, returns
// its full argument list split on top-level commas. Handles (), [], {} and
// string literals so inline route arrays survive.
function callArgs(text, openParenIdx) {
  let depth = 0;
  let argStart = openParenIdx + 1;
  const args = [];
  for (let i = openParenIdx; i < text.length; i++) {
    const c = text[i];
    if (c === '"' || c === '`') {
      const q = c;
      i++;
      while (i < text.length && text[i] !== q) {
        if (text[i] === '\\') i++;
        i++;
      }
      continue;
    }
    if (c === '(' || c === '[' || c === '{') depth++;
    else if (c === ')' || c === ']' || c === '}') {
      depth--;
      if (depth === 0 && c === ')') {
        args.push(text.slice(argStart, i));
        return args;
      }
    } else if (c === ',' && depth === 1) {
      args.push(text.slice(argStart, i));
      argStart = i + 1;
    }
  }
  return args;
}


for (const m of moduleDirs()) {
  for (const f of walk(m.path, ['.go'])) {
    const t = readFileSync(f, 'utf8');
    // Register<VERB>[Req, Res](ctx, router, <routes>, &handlers.<Name>{...})
    const re = /serverHttp\.Register(DELETE|GET|POST|PUT|PATCH)(WithValidation)?\[([^\]]+)\]\(/g;
    let mm;
    while ((mm = re.exec(t))) {
      const verb = mm[1];
      const reqRes = mm[3].split(',').map((s) => s.trim());
      const reqT = reqRes[0] || '';
      const resT = reqRes[1] || '';
      const args = callArgs(t, re.lastIndex - 1);
      if (args.length < 4) continue;
      const routesExpr = args[2].trim();
      let paths = [];
      let routeSrc = '';
      const regLine = t.slice(0, mm.index).split('\n').length;
      const fnMatch = /^handlers\.(\w+)\(\)$/.exec(routesExpr);
      if (fnMatch && routesFn.has(fnMatch[1])) {
        paths = routesFn.get(fnMatch[1]).paths;
        routeSrc = `${routesFn.get(fnMatch[1]).file}:${routesFn.get(fnMatch[1]).line}`;
      } else {
        paths = [...routesExpr.matchAll(/"([^"]+)"/g)].map((x) => x[1]);
        routeSrc = `inline:${relRepo(f)}:${regLine}`;
      }
      // Locate the Serve method of this handler in the same file set.
      const handlerName = /&(\w+)\{/.exec(args[3] || '');
      const hName = handlerName ? handlerName[1] : '';
      let serveText = t;
      let serveFile = relRepo(f);
      if (hName) {
        const serveRe = new RegExp(`func \\([\\w ]*\\*?${hName}\\) Serve`);
        const hit = goFiles.find((gf) => serveRe.test(readFileSync(gf, 'utf8')));
        if (hit) {
          serveFile = relRepo(hit);
          serveText = readFileSync(hit, 'utf8');
        }
      }
      const anon = anonByModule.get(m.name) || [];
      for (const p of paths) {
        rows.push({
          module: m.name,
          service: 'legacy-chi',
          rpc_or_route: `${verb} ${p}`,
          request_type: reqT,
          response_type: resT,
          auth_required: anon.includes(p) ? 'no' : 'yes',
          tenant_predicate: tenantEvidence(serveText, m.name),
          file: `${routeSrc} serve=${serveFile}`,
          line: regLine,
        });
      }
    }
  }
}

// Module routes registered directly on the chi router (websocket upgrades,
// signed webhooks, raw HTTP) bypass serverHttp.Register, so sweep those
// literals too. Auth follows the module's GetAnonymousRoutes list.
for (const m of moduleDirs()) {
  for (const f of walk(m.path, ['.go'])) {
    const t = readFileSync(f, 'utf8');
    const re = /\.(Get|Post|Put|Patch|Delete)\("(\/[^"]+)"/g;
    let mm;
    while ((mm = re.exec(t))) {
      const anon = anonByModule.get(m.name) || [];
      rows.push({
        module: m.name,
        service: 'legacy-chi',
        rpc_or_route: `${mm[1].toUpperCase()} ${mm[2]}`,
        request_type: '',
        response_type: '',
        auth_required: anon.includes(mm[2]) ? 'no' : 'yes',
        tenant_predicate: tenantEvidence(t, m.name),
        file: relRepo(f),
        line: t.slice(0, mm.index).split('\n').length,
      });
    }
  }
}

// ---------- oneKit RPC handlers ----------
// .onk services -> generated RegisterXServiceServer calls in module.go.
const onkRoot = voxieApiDir();
import { readdirSync, statSync as st } from 'node:fs';
function listOnk(dir) {
  const out = [];
  let es;
  try {
    es = readdirSync(dir);
  } catch {
    return out;
  }
  for (const e of es) {
    const p = join(dir, e);
    let s;
    try {
      s = st(p);
    } catch {
      continue;
    }
    if (s.isDirectory()) out.push(...listOnk(p));
    else if (e.endsWith('.onk') && !e.includes('gen')) out.push(p);
  }
  return out;
}
const services = []; // {pkgPath, service, rpcs:[{name,req,res,verb,path,line,file}]}
for (const f of listOnk(onkRoot)) {
  const t = readFileSync(f, 'utf8');
  const pkg = /^\s*package\s+([\w.]+)/m.exec(t);
  const svcRe = /service\s+(\w+)\s*\{([\s\S]*?)\n\}/g;
  let sm;
  while ((sm = svcRe.exec(t))) {
    const svc = sm[1];
    const body = sm[2];
    const rpcRe =
      /(\w+)\(([\w.,\s*]+)\)\s*->\s*(\w+)\s*@(\w+)\("([^"]+)"\)/g;
    let rm;
    const rpcs = [];
    while ((rm = rpcRe.exec(body))) {
      rpcs.push({
        name: rm[1],
        req: rm[2].trim(),
        res: rm[3],
        verb: rm[4].toUpperCase(),
        path: rm[5],
        line: lineOf(t, rm[1]),
        file: relRepo(f).replace(/^.*voxie\//, 'voxie/'),
      });
    }
    if (rpcs.length) services.push({ pkg: pkg ? pkg[1] : '', pkgPath: f, service: svc, rpcs });
  }
}

// Which module registers which generated service.
const genImportRe = /(\w+)\s+"github\.com\/voxire\/onion\/internal\/modules\/([\w/]+)"/g;
for (const m of moduleDirs()) {
  const modFile = join(m.path, 'module.go');
  let t;
  try {
    t = readFileSync(modFile, 'utf8');
  } catch {
    continue;
  }
  for (const s of services) {
    // gen package alias e.g. flowprojectv1 for voxie/gen/go/flow/project/v1
    const seg = s.pkg.split('.');
    const genPkg = `github.com/voxire/voxie/gen/go/${seg.join('/')}`;
    const aliasRe = new RegExp(`(\\w+)\\s+"${genPkg.replace(/\//g, '\\/')}"`);
    const am = aliasRe.exec(t);
    if (!am) continue;
    const alias = am[1];
    if (!t.includes(`${alias}.Register${s.service}Server`)) continue;
    for (const r of s.rpcs) {
      // Generated Go methods are exported: listTimeEntries -> ListTimeEntries.
      const goName = r.name.charAt(0).toUpperCase() + r.name.slice(1);
      // Find the implementing API file for the method to read tenancy tokens.
      let implText = '';
      let implFile = '';
      for (const f of walk(join(m.path, 'internal'), ['.go'])) {
        const ft = readFileSync(f, 'utf8');
        if (new RegExp(`func \\([\\w ]*\\*?[\\w]*\\) ${goName}\\(`).test(ft)) {
          implText = ft;
          implFile = relRepo(f);
          break;
        }
      }
      const prefix = '/' + seg.join('/');
      const full = prefix + (r.path.startsWith('/') ? r.path : '/' + r.path);
      rows.push({
        module: m.name,
        service: `${s.pkg} ${s.service}`,
        rpc_or_route: `${r.verb} ${full}`,
        request_type: r.req,
        response_type: r.res,
        auth_required: 'yes(onekit)',
        tenant_predicate: implText ? tenantEvidence(implText, m.name) : 'impl-not-located',
        file: `${r.file}:${r.line}${implFile ? ' impl=' + implFile : ''}`,
        line: r.line,
      });
    }
  }
}

// ---------- platform-level route literals (outside modules) ----------
for (const f of walk(paths.platform, ['.go'])) {
  const t = readFileSync(f, 'utf8');
  const re = /(RegisterGET|RegisterPOST|RegisterPUT|RegisterDELETE|r\.HandleFunc|r\.Handle|mux\.Handle)\b/g;
  let mm;
  while ((mm = re.exec(t))) {
    const rest = t.slice(mm.index, mm.index + 600);
    const pm = /"\/([a-z0-9\-/{}_:]*)"/.exec(rest);
    if (pm && !pm[1].startsWith('/')) {
      rows.push({
        module: 'platform',
        service: 'platform-route',
        rpc_or_route: `${mm[1].replace('Register', '').toUpperCase()} /${pm[1]}`,
        request_type: '',
        response_type: '',
        auth_required: t.includes('license_routes') ? 'mixed' : 'yes',
        tenant_predicate: tenantPredicate(t),
        file: relRepo(f),
        line: t.slice(0, mm.index).split('\n').length,
      });
    }
  }
  // Plain chi route literals registered via router/r.Get etc.
  const re2 = /\.(Get|Post|Put|Delete)\("\/[^"]+"/g;
  while ((mm = re2.exec(t))) {
    rows.push({
      module: 'platform',
      service: 'platform-route',
      rpc_or_route: `${mm[1].toUpperCase()} ${/"([^"]+)"/.exec(mm[0])[1]}`,
      request_type: '',
      response_type: '',
      auth_required: 'yes',
      tenant_predicate: tenantPredicate(t),
      file: relRepo(f),
      line: t.slice(0, mm.index).split('\n').length,
    });
  }
}

rows.sort((a, b) => a.module.localeCompare(b.module) || a.rpc_or_route.localeCompare(b.rpc_or_route));
writeCsv(
  'backend_handlers.csv',
  ['module', 'service', 'rpc_or_route', 'request_type', 'response_type', 'auth_required', 'tenant_predicate', 'file', 'line'],
  rows.map((r) => [r.module, r.service, r.rpc_or_route, r.request_type, r.response_type, r.auth_required, r.tenant_predicate, r.file, r.line]),
);
