#!/usr/bin/env node

import { execFileSync } from 'node:child_process';
import { existsSync, readdirSync, readFileSync } from 'node:fs';
import { join, relative, sep } from 'node:path';

const root = process.cwd();
const changedOnly = process.argv.includes('--changed-only');
const scanAll = process.argv.includes('--all') || !changedOnly;

const sourceExtensions = new Set(['.js', '.jsx', '.mjs', '.ts', '.tsx']);
const rawFetchAllowlist = new Set([
  'apps/flow/src/shared/lib/realtime.ts',
  // Recruit uploads directly to a presigned R2 URL; it must not receive the
  // authenticated API transport's headers or base URL.
  'apps/recruit/src/shared/hooks/queries/usePublic.ts',
]);

const rules = [
  {
    code: 'AXIOS_IMPORT',
    pattern: /(?:from\s+['"]axios['"]|require\(\s*['"]axios['"]\s*\))/,
    message: 'Importing Axios in an app bypasses the generated OneKit client.',
  },
  {
    code: 'LEGACY_API_IMPORT',
    pattern: /['"][^'"]*\/lib\/api['"]/,
    message: 'Importing the legacy app API wrapper bypasses the generated OneKit client.',
  },
  {
    code: 'LEGACY_API_WRAPPER',
    pattern: /\b(?:safeFetch|SafeApi|createApiClient)\b/,
    message: 'Defining or using the legacy app API wrapper bypasses the generated OneKit client.',
  },
  {
    code: 'DIRECT_HTTP_CALL',
    pattern: /\b(?:axios|api)\.(?:get|post|put|patch|delete|request)(?:<[^>\n]+>)?\s*\(/,
    message: 'Direct HTTP calls must go through a generated OneKit service client.',
  },
  {
    code: 'SAFE_HTTP_CALL',
    pattern: /\bsafe\.(?:get|post|put|patch|delete)(?:<[^>\n]+>)?\s*\(/,
    message: 'Legacy safe.* calls must be migrated to a generated OneKit service client.',
  },
  {
    code: 'RAW_FETCH',
    pattern: /\bfetch\s*\(/,
    message: 'Raw fetch is reserved for the shared OneKit transport adapter.',
  },
];

function relativePath(file) {
  return relative(root, file).split(sep).join('/');
}

function walk(directory) {
  if (!existsSync(directory)) return [];
  const files = [];
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    if (entry.name === 'node_modules' || entry.name === 'dist' || entry.name === 'build') continue;
    const file = join(directory, entry.name);
    if (entry.isDirectory()) {
      files.push(...walk(file));
      continue;
    }
    if (sourceExtensions.has(file.slice(file.lastIndexOf('.')))) files.push(file);
  }
  return files;
}

function gitOutput(args) {
  try {
    return execFileSync('git', args, { cwd: root, encoding: 'utf8' });
  } catch (error) {
    const detail = error instanceof Error ? error.message : String(error);
    throw new Error(`Unable to inspect git changes: ${detail}`);
  }
}

function diffBase() {
  if (process.env.GITHUB_BASE_REF) return `origin/${process.env.GITHUB_BASE_REF}`;
  if (process.env.GITHUB_EVENT_BEFORE && !/^0+$/.test(process.env.GITHUB_EVENT_BEFORE)) {
    return process.env.GITHUB_EVENT_BEFORE;
  }
  if (process.env.GITHUB_SHA) return `${process.env.GITHUB_SHA}^`;
  return 'HEAD';
}

function changedLines() {
  const diff = gitOutput(['diff', '--unified=0', diffBase(), 'HEAD']);
  const findings = [];
  let file = null;
  let lineNumber = 0;

  for (const line of diff.split('\n')) {
    if (line.startsWith('+++ b/')) {
      file = line.slice('+++ b/'.length);
      continue;
    }
    const hunk = line.match(/^@@ -\d+(?:,\d+)? \+(\d+)(?:,\d+)? @@/);
    if (hunk) {
      lineNumber = Number(hunk[1]);
      continue;
    }
    if (
      !file ||
      !file.startsWith('apps/') ||
      !sourceExtensions.has(file.slice(file.lastIndexOf('.')))
    ) {
      continue;
    }
    if (line.startsWith('+') && !line.startsWith('+++')) {
      findings.push({ file, line: lineNumber, text: line.slice(1) });
      lineNumber += 1;
    } else if (!line.startsWith('-')) {
      lineNumber += 1;
    }
  }

  // git diff does not include untracked files. Treat new app source files as
  // entirely new so the guard is useful before the first commit as well.
  const untracked = gitOutput(['ls-files', '--others', '--exclude-standard', '--', 'apps'])
    .split('\n')
    .filter(Boolean)
    .filter((fileName) => sourceExtensions.has(fileName.slice(fileName.lastIndexOf('.'))));
  for (const fileName of untracked) {
    const contents = readFileSync(join(root, fileName), 'utf8').split('\n');
    contents.forEach((text, index) => findings.push({ file: fileName, line: index + 1, text }));
  }

  return findings;
}

function findingsForLine(file, line, text) {
  return rules.flatMap((rule) => {
    if (rule.code === 'RAW_FETCH' && rawFetchAllowlist.has(file)) return [];
    return rule.pattern.test(text) ? [{ file, line, code: rule.code, message: rule.message }] : [];
  });
}

function appDependencyFindings() {
  const findings = [];
  const appsDirectory = join(root, 'apps');
  if (!existsSync(appsDirectory)) return findings;

  for (const entry of readdirSync(appsDirectory, { withFileTypes: true })) {
    if (!entry.isDirectory()) continue;
    const fileName = join(appsDirectory, entry.name, 'package.json');
    if (!existsSync(fileName)) continue;

    const manifest = JSON.parse(readFileSync(fileName, 'utf8'));
    const dependencies = manifest.dependencies ?? {};
    if (!dependencies['@voxire/voxie']) {
      findings.push({
        file: relativePath(fileName),
        line: 1,
        code: 'MISSING_ONEKIT_DEPENDENCY',
        message: 'Every rtylr app must declare @voxire/voxie as a workspace dependency.',
      });
    }
  }
  return findings;
}

const findings = [];
findings.push(...appDependencyFindings());
if (changedOnly) {
  for (const entry of changedLines())
    findings.push(...findingsForLine(entry.file, entry.line, entry.text));
} else if (scanAll) {
  for (const file of walk(join(root, 'apps'))) {
    const fileName = relativePath(file);
    const contents = readFileSync(file, 'utf8').split('\n');
    contents.forEach((text, index) => findings.push(...findingsForLine(fileName, index + 1, text)));
  }
}

if (findings.length > 0) {
  console.error(
    `OneKit boundary: ${findings.length} violation(s)${changedOnly ? ' in added lines' : ''}.`
  );
  for (const finding of findings.slice(0, 100)) {
    console.error(`- ${finding.file}:${finding.line} [${finding.code}] ${finding.message}`);
  }
  if (findings.length > 100) console.error(`- ... ${findings.length - 100} more`);
  process.exitCode = 1;
} else {
  console.log(`OneKit boundary: no ${changedOnly ? 'new ' : ''}app transport violations found.`);
}
