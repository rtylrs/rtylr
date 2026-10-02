#!/usr/bin/env node
import fs from 'node:fs';
import path from 'node:path';
import process from 'node:process';
import { createRequire } from 'node:module';

const require = createRequire(import.meta.url);
const ts = require('typescript');

// Audits that every API call in the apps goes through a generated OneKit
// service client WITHOUT a type assertion defeating the contract types
// (e.g. `client.listTeam({ page } as any)`), and that every
// `@voxire/voxie/...` import resolves to a real generated file.
//
// Transport-level bypasses (axios, raw fetch, legacy safe.* wrappers) are
// enforced separately by scripts/enforce-onekit-boundary.mjs.

const root = process.cwd();
const voxieRoot = path.join(root, '../packages/contracts/src');
const failOnErrors = process.argv.includes('--fail-on-errors');
const writeReport = !process.argv.includes('--no-report');

function walk(dir, acc = []) {
  if (!fs.existsSync(dir)) return acc;
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    if (
      entry.name === 'node_modules' ||
      entry.name === 'dist' ||
      entry.name === 'build' ||
      entry.name === '.next'
    )
      continue;
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) walk(full, acc);
    else if (/\.(ts|tsx)$/.test(full) && !/\.(test|spec)\.(ts|tsx)$/.test(full)) acc.push(full);
  }
  return acc;
}

function toPosix(p) {
  return p.split(path.sep).join('/');
}

const appFiles = [...walk(path.join(root, 'src')), ...walk(path.join(root, 'entrypoints'))];

const program = ts.createProgram(appFiles, {
  allowJs: false,
  jsx: ts.JsxEmit.React,
  moduleResolution: ts.ModuleResolutionKind.NodeNext,
  module: ts.ModuleKind.NodeNext,
  target: ts.ScriptTarget.ES2022,
  skipLibCheck: true,
  noEmit: true,
});

function sourceLine(source, pos) {
  return source.getLineAndCharacterOfPosition(pos).line + 1;
}

function isClientReceiver(node) {
  // Repo convention: every generated client instance is named *Client
  // (membersClient, invoicesClient, hrCompatibilityClient, ...).
  return ts.isIdentifier(node) && /Client$/.test(node.text);
}

function containsContractDefeatingAssertion(node) {
  if (ts.isAsExpression(node) || ts.isTypeAssertionExpression(node)) {
    const targetType = node.type;
    if (ts.isTypeReferenceNode(targetType) && targetType.typeName.kind === ts.SyntaxKind.AnyKeyword)
      return true;
    if (targetType.kind === ts.SyntaxKind.AnyKeyword) return true;
    // Double assertions like `(x as unknown as Contract)` hide the real shape.
    if (
      ts.isAsExpression(node.expression) ||
      ts.isTypeAssertionExpression(node.expression) ||
      ts.isParenthesizedExpression(node.expression)
    ) {
      const inner = ts.isParenthesizedExpression(node.expression)
        ? node.expression.expression
        : node.expression;
      if (ts.isAsExpression(inner) || ts.isTypeAssertionExpression(inner)) {
        const innerTarget = inner.type;
        if (
          ts.isTypeReferenceNode(innerTarget) &&
          innerTarget.typeName.kind === ts.SyntaxKind.UnknownKeyword
        )
          return true;
        if (innerTarget.kind === ts.SyntaxKind.UnknownKeyword) return true;
      }
    }
    return false;
  }
  return ts.forEachChild(node, (child) => containsContractDefeatingAssertion(child)) === true;
}

const calls = [];
const issues = [];

for (const source of program.getSourceFiles()) {
  if (!appFiles.includes(source.fileName)) continue;

  function visit(node) {
    if (
      ts.isCallExpression(node) &&
      ts.isPropertyAccessExpression(node.expression) &&
      isClientReceiver(node.expression.expression)
    ) {
      const location = {
        file: toPosix(path.relative(root, source.fileName)),
        line: sourceLine(source, node.getStart(source)),
      };
      const call = {
        ...location,
        method: node.expression.name.text,
      };
      calls.push(call);

      for (const arg of node.arguments) {
        if (containsContractDefeatingAssertion(arg)) {
          issues.push({
            severity: 'error',
            code: 'CONTRACT_TYPE_ASSERTION',
            message:
              'Argument to a OneKit client call uses a type assertion that defeats the .onk contract typing.',
            ...location,
          });
          break;
        }
      }
    }

    if (ts.isImportDeclaration(node) && ts.isStringLiteral(node.moduleSpecifier)) {
      const specifier = node.moduleSpecifier.text;
      if (specifier.startsWith('@voxire/voxie/')) {
        const rel = specifier.slice('@voxire/voxie/'.length);
        const resolved = path.join(voxieRoot, `${rel}.ts`);
        if (!fs.existsSync(resolved)) {
          issues.push({
            severity: 'error',
            code: 'VOXIE_IMPORT_UNRESOLVED',
            message: `Import '${specifier}' does not resolve under clients/packages/contracts/src.`,
            file: toPosix(path.relative(root, source.fileName)),
            line: sourceLine(source, node.getStart(source)),
          });
        }
      }
    }

    ts.forEachChild(node, visit);
  }

  visit(source);
}

const errors = issues.filter((issue) => issue.severity === 'error');

const report = {
  generatedAt: new Date().toISOString(),
  filesAudited: appFiles.length,
  clientCalls: calls.length,
  errors: errors.length,
  issues,
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
  `request-contract audit: ${calls.length} OneKit client calls across ${appFiles.length} files, ${errors.length} errors`
);
for (const issue of issues.slice(0, 25)) {
  console.log(
    `${issue.severity.toUpperCase()} ${issue.code} ${issue.file}:${issue.line} ${issue.message}`
  );
}
if (issues.length > 25) {
  console.log(`... ${issues.length - 25} more issues in tmp/request-contract-audit.json`);
}

if (failOnErrors && errors.length > 0) process.exitCode = 1;
