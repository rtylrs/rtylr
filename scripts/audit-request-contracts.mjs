#!/usr/bin/env node
import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const ts = require('typescript');

const root = process.cwd();
const onionRoot = path.resolve(root, '..', 'onion');
const metadataPath = path.join(root, 'packages/contracts/src/validation/metadata.json');
const failOnErrors = process.argv.includes('--fail-on-errors');
const writeReport = !process.argv.includes('--no-report');

const metadata = JSON.parse(fs.readFileSync(metadataPath, 'utf8')).messages ?? {};

function walk(dir, predicate, acc = []) {
  if (!fs.existsSync(dir)) return acc;
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    if (entry.name === 'node_modules' || entry.name === 'dist' || entry.name === '.next') continue;
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) walk(full, predicate, acc);
    else if (predicate(full)) acc.push(full);
  }
  return acc;
}

function toPosix(p) {
  return p.split(path.sep).join('/');
}

function camelToSnake(value) {
  return value.replace(/[A-Z]/g, (char) => `_${char.toLowerCase()}`);
}

function upperFirst(value) {
  return value ? value[0].toUpperCase() + value.slice(1) : value;
}

function normalizeRoute(value) {
  if (!value) return '';
  return (
    value
      .replace(/^`|`$/g, '')
      .replace(/[?#].*$/, '')
      .replace(/\$\{[^}]+\}/g, '{param}')
      .replace(/:[^/]+/g, '{param}')
      .replace(/\{[^}]+\}/g, '{param}')
      .replace(/\/+/g, '/')
      .replace(/\/$/, '') || '/'
  );
}

function sourceLine(source, pos) {
  return source.getLineAndCharacterOfPosition(pos).line + 1;
}

function readText(file) {
  return fs.readFileSync(file, 'utf8');
}

function parseRoutesFunctionPaths(handlerFiles) {
  const routeFns = new Map();
  const routeFnRegex = /func\s+([A-Za-z0-9_]+Routes)\s*\(\)\s*\[\]string\s*\{/g;

  for (const file of handlerFiles) {
    const text = readText(file);
    const moduleRoot = moduleRootForFile(file);
    let match;
    while ((match = routeFnRegex.exec(text))) {
      const openBrace = text.indexOf('{', match.index);
      const closeBrace = findMatchingBrace(text, openBrace);
      const body = text.slice(match.index, closeBrace > openBrace ? closeBrace + 1 : match.index);
      const paths = [...body.matchAll(/"([^"]+)"/g)].map((m) => m[1]);
      routeFns.set(`${moduleRoot}:${match[1]}`, paths);
    }
  }
  return routeFns;
}

function findMatchingBrace(text, openIndex) {
  let depth = 0;
  let inString = false;
  let escaped = false;
  for (let index = openIndex; index < text.length; index++) {
    const char = text[index];
    if (inString) {
      if (escaped) {
        escaped = false;
      } else if (char === '\\') {
        escaped = true;
      } else if (char === '"') {
        inString = false;
      }
      continue;
    }
    if (char === '"') {
      inString = true;
    } else if (char === '{') {
      depth++;
    } else if (char === '}') {
      depth--;
      if (depth === 0) return index;
    }
  }
  return -1;
}

function moduleRootForFile(file) {
  const marker = '/internal/routes/';
  const index = toPosix(file).indexOf(marker);
  return index === -1 ? null : toPosix(file).slice(0, index);
}

function parseRequestStructs(handlerFiles) {
  const structs = new Map();
  const validateRegex =
    /func\s+\(h \*\s*([A-Za-z0-9_]+Handler)\)\s+Validate\s*\([^)]*req\s+\*([A-Za-z0-9_]+)\)[^{]*\{/g;

  for (const file of handlerFiles) {
    const text = readText(file);
    const lines = text.split('\n');
    for (let index = 0; index < lines.length; index++) {
      const declaration = lines[index].match(
        /^\s*(?:type\s+)?([A-Z][A-Za-z0-9_]*)\s+struct\s*\{\s*$/
      );
      if (!declaration) continue;

      const body = [];
      index++;
      while (index < lines.length && !/^\s*}\s*$/.test(lines[index])) {
        body.push(lines[index]);
        index++;
      }

      const fields = {};
      for (const line of body) {
        const json = line.match(/`[^`]*json:"([^",`]+)[^"`]*"[^`]*`/);
        const field = line.trim().match(/^([A-Z][A-Za-z0-9_]*)\s+/);
        if (!json || !field || json[1] === '-') continue;
        fields[json[1]] = { goName: field[1], jsonName: json[1], required: false };
      }
      const item = {
        absFile: file,
        file: toPosix(path.relative(root, file)),
        name: declaration[1],
        fields,
      };
      if (!structs.has(declaration[1])) structs.set(declaration[1], []);
      structs.get(declaration[1]).push(item);
    }

    let match;
    while ((match = validateRegex.exec(text))) {
      const reqName = match[2];
      const current = structs.get(reqName)?.find((item) => item.absFile === file);
      if (!current) continue;

      const openBrace = text.indexOf('{', match.index);
      const closeBrace = findMatchingBrace(text, openBrace);
      if (closeBrace <= openBrace) continue;
      const body = text.slice(openBrace + 1, closeBrace);
      for (const requiredCheck of body.matchAll(/if\s+([^{]+req\.[^{]+)\{/g)) {
        const condition = requiredCheck[1];
        const ifOpenBrace = body.indexOf('{', requiredCheck.index);
        const ifCloseBrace = findMatchingBrace(body, ifOpenBrace);
        if (ifCloseBrace <= ifOpenBrace) continue;
        const block = body.slice(ifOpenBrace + 1, ifCloseBrace);
        if (!/\brequired\b/i.test(block)) continue;
        for (const reqField of condition.matchAll(/req\.([A-Z][A-Za-z0-9_]*)/g)) {
          for (const field of Object.values(current.fields)) {
            if (field.goName === reqField[1]) field.required = true;
          }
        }
      }
    }
  }

  return structs;
}

function chooseRequestStruct(structs, request, routeFile) {
  const candidates = structs.get(request) ?? [];
  if (candidates.length <= 1) return candidates[0] ?? null;
  const moduleRoot = moduleRootForFile(routeFile);
  if (!moduleRoot) return candidates[0] ?? null;
  return (
    candidates.find((candidate) => candidate.absFile.startsWith(`${moduleRoot}/`)) ??
    candidates[0] ??
    null
  );
}

function parseBackendRoutes() {
  const routeFiles = walk(path.join(onionRoot, 'internal/modules'), (file) =>
    file.endsWith('/internal/routes/routes.go')
  );
  const handlerFiles = walk(
    path.join(onionRoot, 'internal/modules'),
    (file) => file.includes('/internal/routes/handlers/') && file.endsWith('.go')
  );

  const routeFns = parseRoutesFunctionPaths(handlerFiles);
  const structs = parseRequestStructs(handlerFiles);
  const routes = [];
  const regRegex =
    /(?:http|serverHttp)\.Register(GET|POST|PUT|PATCH|DELETE)(?:WithValidation(?:AndMiddleware)?|AndMiddleware)?\s*\[\s*handlers\.([A-Za-z0-9_]+)\s*,\s*handlers\.([A-Za-z0-9_]+)\s*\]\s*\(\s*ctx\s*,\s*router\s*,\s*handlers\.([A-Za-z0-9_]+Routes)\(\)/g;
  const rawRouteRegex = /router\.(Get|Post|Put|Patch|Delete)\s*\(\s*"([^"]+)"/g;

  for (const file of routeFiles) {
    const text = readText(file);
    let match;
    while ((match = regRegex.exec(text))) {
      const [, method, request, response, routeFn] = match;
      const moduleRoot = moduleRootForFile(file);
      for (const routePath of routeFns.get(`${moduleRoot}:${routeFn}`) ?? []) {
        routes.push({
          method,
          path: routePath,
          normalizedPath: normalizeRoute(routePath),
          request,
          response,
          routeFile: toPosix(path.relative(root, file)),
          requestStruct: chooseRequestStruct(structs, request, file),
          validated: match[0].includes('WithValidation'),
        });
      }
    }
    while ((match = rawRouteRegex.exec(text))) {
      const [, methodName, routePath] = match;
      const method = methodName.toUpperCase();
      routes.push({
        method,
        path: routePath,
        normalizedPath: normalizeRoute(routePath),
        request: null,
        response: null,
        routeFile: toPosix(path.relative(root, file)),
        requestStruct: null,
        validated: false,
        raw: true,
      });
    }
  }

  return routes;
}

function parseContractInterfaces() {
  const files = walk(path.join(root, 'packages/contracts/src'), (file) => file.endsWith('.ts'));
  const byDeclFileAndName = new Map();
  const byName = new Map();

  for (const file of files) {
    const rel = toPosix(path.relative(path.join(root, 'packages/contracts/src'), file));
    const sourceRoot = rel.replace(/\.ts$/, '');
    if (sourceRoot.startsWith('validation/') || sourceRoot.startsWith('google/')) continue;
    const namespace = sourceRoot.split('/').slice(0, -1).join('.');
    const text = readText(file);
    for (const match of text.matchAll(/export interface\s+([A-Za-z0-9_]+Request)\b/g)) {
      const name = match[1];
      const fqn = `${namespace}.${name}`;
      const item = { name, fqn, file };
      byDeclFileAndName.set(`${file}:${name}`, item);
      if (!byName.has(name)) byName.set(name, []);
      byName.get(name).push(item);
    }
  }

  return { byDeclFileAndName, byName };
}

function getStringLike(node) {
  if (!node) return null;
  if (ts.isStringLiteral(node) || ts.isNoSubstitutionTemplateLiteral(node)) return node.text;
  if (ts.isTemplateExpression(node)) {
    let out = node.head.text;
    for (const span of node.templateSpans) out += '{param}' + span.literal.text;
    return out;
  }
  return null;
}

function objectKeys(node) {
  if (!node || !ts.isObjectLiteralExpression(node)) return { keys: null, hasSpread: false };
  const keys = [];
  let hasSpread = false;
  for (const prop of node.properties) {
    if (ts.isSpreadAssignment(prop)) {
      hasSpread = true;
      continue;
    }
    if (!ts.isPropertyAssignment(prop) && !ts.isShorthandPropertyAssignment(prop)) continue;
    const name = prop.name;
    if (ts.isIdentifier(name) || ts.isStringLiteral(name) || ts.isNumericLiteral(name)) {
      keys.push(name.text);
    } else {
      hasSpread = true;
    }
  }
  return { keys, hasSpread };
}

function propertyFromConfig(node, propName) {
  if (!node || !ts.isObjectLiteralExpression(node)) return null;
  for (const prop of node.properties) {
    if (!ts.isPropertyAssignment(prop)) continue;
    const name = prop.name;
    if ((ts.isIdentifier(name) || ts.isStringLiteral(name)) && name.text === propName) {
      return prop.initializer;
    }
  }
  return null;
}

function requestContractFromType(type, contracts) {
  const symbol = type.aliasSymbol ?? type.symbol;
  const declarations = symbol?.declarations ?? [];
  for (const declaration of declarations) {
    const file = declaration.getSourceFile().fileName;
    const item = contracts.byDeclFileAndName.get(`${file}:${symbol.name}`);
    if (item) return item;
  }
  return null;
}

function fieldInfoFromExpression(checker, expression, contracts) {
  if (!expression) return { keys: null, snakeKeys: null, hasSpread: false, contract: null };

  const type = checker.getTypeAtLocation(expression);
  if (type.symbol?.name === 'FormData' || type.aliasSymbol?.name === 'FormData') {
    return { keys: null, snakeKeys: null, hasSpread: true, contract: null, multipart: true };
  }

  const literal = objectKeys(expression);
  if (literal.keys) {
    return {
      keys: literal.keys,
      snakeKeys: literal.keys.map(camelToSnake),
      hasSpread: literal.hasSpread,
      contract: requestContractFromType(checker.getTypeAtLocation(expression), contracts),
    };
  }

  const contract = requestContractFromType(type, contracts);
  const keys = type
    .getProperties()
    .map((prop) => prop.name)
    .filter((name) => !name.startsWith('__'));

  return {
    keys: keys.length ? keys : null,
    snakeKeys: keys.length ? keys.map(camelToSnake) : null,
    hasSpread: Boolean(type.getStringIndexType() || type.getNumberIndexType()),
    contract,
  };
}

function analyzeFrontendCalls(backendRoutes, contracts) {
  const files = walk(path.join(root, 'apps'), (file) => {
    if (!/\.(ts|tsx)$/.test(file)) return false;
    if (/\.(test|spec)\.(ts|tsx)$/.test(file)) return false;
    return !file.includes('/node_modules/');
  });
  const program = ts.createProgram(files, {
    allowJs: false,
    jsx: ts.JsxEmit.React,
    moduleResolution: ts.ModuleResolutionKind.NodeNext,
    module: ts.ModuleKind.NodeNext,
    target: ts.ScriptTarget.ES2022,
    skipLibCheck: true,
    noEmit: true,
  });
  const checker = program.getTypeChecker();
  const routesByKey = new Map();
  for (const route of backendRoutes)
    routesByKey.set(`${route.method} ${route.normalizedPath}`, route);

  const calls = [];
  const issues = [];

  for (const source of program.getSourceFiles()) {
    if (!source.fileName.startsWith(path.join(root, 'apps'))) continue;

    function visit(node) {
      if (
        ts.isCallExpression(node) &&
        ts.isPropertyAccessExpression(node.expression) &&
        ts.isIdentifier(node.expression.expression) &&
        node.expression.expression.text === 'safe'
      ) {
        const methodName = node.expression.name.text;
        const method = methodName.toUpperCase();
        if (!['GET', 'POST', 'PUT', 'PATCH', 'DELETE'].includes(method)) return;

        const rawUrl = getStringLike(node.arguments[0]);
        if (!rawUrl) return;

        let payloadExpression = null;
        let payloadKind = 'body';
        if (['POST', 'PUT', 'PATCH'].includes(method)) {
          payloadExpression = node.arguments[1] ?? null;
        } else {
          payloadKind = 'params';
          payloadExpression = propertyFromConfig(node.arguments[1], 'params');
          if (method === 'DELETE') {
            const data = propertyFromConfig(node.arguments[1], 'data');
            if (data) {
              payloadExpression = data;
              payloadKind = 'body';
            }
          }
        }

        const payload = fieldInfoFromExpression(checker, payloadExpression, contracts);
        const normalizedPath = normalizeRoute(rawUrl);
        const backend = routesByKey.get(`${method} ${normalizedPath}`) ?? null;
        const location = {
          file: toPosix(path.relative(root, source.fileName)),
          line: sourceLine(source, node.getStart(source)),
        };
        const call = {
          ...location,
          method,
          path: rawUrl,
          normalizedPath,
          payloadKind,
          payloadKeys: payload.keys,
          payloadSnakeKeys: payload.snakeKeys,
          hasSpread: payload.hasSpread,
          contract: payload.contract?.fqn ?? null,
          backendRequest: backend?.request ?? null,
        };
        calls.push(call);

        const addIssue = (severity, code, message, extra = {}) => {
          issues.push({ severity, code, message, ...location, method, path: rawUrl, ...extra });
        };

        if (!backend && rawUrl.startsWith('/')) {
          addIssue('warning', 'NO_BACKEND_ROUTE', 'No Onion route matched this frontend request.');
        }

        if (payload.contract && !metadata[payload.contract.fqn]) {
          addIssue(
            'warning',
            'CONTRACT_WITHOUT_VALIDATION_METADATA',
            'Request type has no Voxie validation metadata.',
            {
              contract: payload.contract.fqn,
            }
          );
        }

        const contractFields = payload.contract ? metadata[payload.contract.fqn] : null;
        if (contractFields && payload.keys && !payload.hasSpread) {
          const provided = new Set(payload.keys);
          const missing = Object.values(contractFields)
            .filter((field) => field.required)
            .map((field) => field.name)
            .filter((fieldName) => !provided.has(fieldName));
          if (missing.length > 0) {
            addIssue(
              'error',
              'MISSING_REQUIRED_CONTRACT_FIELDS',
              'Frontend payload may omit required Voxie fields.',
              {
                contract: payload.contract.fqn,
                missing,
              }
            );
          }
        }

        if (backend?.requestStruct && payload.snakeKeys && !payload.hasSpread) {
          const backendFields = new Set(Object.keys(backend.requestStruct.fields));
          const frontendFields = new Set(payload.snakeKeys);
          const missingBackendRequired = Object.values(backend.requestStruct.fields)
            .filter((field) => field.required)
            .map((field) => field.jsonName)
            .filter((fieldName) => !frontendFields.has(fieldName));
          const extraFrontend = [...frontendFields].filter(
            (fieldName) => !backendFields.has(fieldName)
          );

          if (missingBackendRequired.length > 0) {
            addIssue(
              'error',
              'MISSING_REQUIRED_BACKEND_FIELDS',
              'Frontend payload may omit required Onion fields.',
              {
                backendRequest: backend.request,
                missing: missingBackendRequired,
              }
            );
          }
          if (extraFrontend.length > 0 && payloadKind === 'body') {
            addIssue(
              'warning',
              'FRONTEND_FIELDS_NOT_IN_BACKEND_REQUEST',
              'Frontend sends fields not declared by Onion request struct.',
              {
                backendRequest: backend.request,
                extra: extraFrontend,
              }
            );
          }
        }

        if (['POST', 'PUT', 'PATCH'].includes(method) && payloadExpression && !payload.contract) {
          const inferred = backend
            ? (contracts.byName.get(backend.request)?.map((item) => item.fqn) ?? [])
            : [];
          addIssue(
            'warning',
            'UNTYPED_MUTATION_PAYLOAD',
            'Mutation payload is not directly typed as a generated Voxie request.',
            {
              backendRequest: backend?.request ?? null,
              possibleContracts: inferred,
            }
          );
        }
      }
      ts.forEachChild(node, visit);
    }

    visit(source);
  }

  return { calls, issues };
}

const backendRoutes = parseBackendRoutes();
const contracts = parseContractInterfaces();
const frontend = analyzeFrontendCalls(backendRoutes, contracts);
const errors = frontend.issues.filter((issue) => issue.severity === 'error');
const warnings = frontend.issues.filter((issue) => issue.severity === 'warning');

const report = {
  generatedAt: new Date().toISOString(),
  backendRoutes: backendRoutes.length,
  backendRouteDetails: backendRoutes.map((route) => ({
    method: route.method,
    path: route.path,
    request: route.request,
    routeFile: route.routeFile,
    requestStruct: route.requestStruct
      ? {
          name: route.requestStruct.name,
          file: route.requestStruct.file,
          fields: route.requestStruct.fields,
        }
      : null,
    validated: route.validated,
    raw: Boolean(route.raw),
  })),
  frontendCalls: frontend.calls.length,
  errors: errors.length,
  warnings: warnings.length,
  issues: frontend.issues,
  calls: frontend.calls,
};

if (writeReport) {
  const outDir = path.join(root, 'tmp');
  fs.mkdirSync(outDir, { recursive: true });
  fs.writeFileSync(
    path.join(outDir, 'request-contract-audit.json'),
    `${JSON.stringify(report, null, 2)}\n`
  );
}

console.log(
  `request-contract audit: ${frontend.calls.length} frontend calls, ${backendRoutes.length} backend routes, ${errors.length} errors, ${warnings.length} warnings`
);
for (const issue of frontend.issues.slice(0, 25)) {
  const details = issue.missing
    ? ` missing=${issue.missing.join(',')}`
    : issue.extra
      ? ` extra=${issue.extra.join(',')}`
      : '';
  console.log(
    `${issue.severity.toUpperCase()} ${issue.code} ${issue.file}:${issue.line} ${issue.method} ${issue.path}${details}`
  );
}
if (frontend.issues.length > 25) {
  console.log(`... ${frontend.issues.length - 25} more issues in tmp/request-contract-audit.json`);
}

if (failOnErrors && errors.length > 0) {
  process.exitCode = 1;
}
