#!/usr/bin/env node

import { execFileSync } from 'node:child_process';
import { existsSync, readdirSync, readFileSync } from 'node:fs';
import { join, relative, resolve, sep } from 'node:path';

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

function toPosix(value) {
  return value.split(sep).join('/');
}

function fatal(message) {
  console.error(`OneKit boundary: ${message}`);
  process.exit(2);
}

function git(args) {
  try {
    return execFileSync('git', args, {
      cwd: root,
      encoding: 'utf8',
      maxBuffer: 64 * 1024 * 1024,
    });
  } catch (error) {
    const detail = String(error.stderr || error.message).trim();
    fatal(`git ${args.join(' ')} failed: ${detail}`);
  }
}

// A ref that does not exist yet is a normal answer to "what is the base?", so
// resolution failures are collected rather than fatal.
function gitOrNull(args) {
  try {
    return execFileSync('git', args, {
      cwd: root,
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'ignore'],
      maxBuffer: 64 * 1024 * 1024,
    });
  } catch {
    return null;
  }
}

const repoRoot = git(['rev-parse', '--show-toplevel']).trim();
const appsDirectory = join(root, 'apps');
if (!appsDirectory.startsWith(`${repoRoot}${sep}`)) {
  fatal(
    `${root} is not inside ${repoRoot}. Run this from the client workspace root (clients/web/rtylr).`
  );
}

// git reports every path relative to the repository root, never to this
// package. Filtering on a bare `apps/` prefix therefore matched nothing once
// the workspace moved under clients/.
const appsPrefix = toPosix(relative(repoRoot, appsDirectory));
// `:(top)` keeps the pathspec anchored at the repository root instead of the
// process cwd, which is this package rather than the repository.
const appsPathspec = `:(top)${appsPrefix}/`;

function toPackagePath(gitPath) {
  return toPosix(relative(root, resolve(repoRoot, gitPath)));
}

function isAppSource(gitPath) {
  return (
    gitPath.startsWith(`${appsPrefix}/`) &&
    sourceExtensions.has(gitPath.slice(gitPath.lastIndexOf('.')))
  );
}

function relativePath(file) {
  return toPosix(relative(root, file));
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

function explicitBaseRef() {
  if (process.env.GITHUB_BASE_REF) return `origin/${process.env.GITHUB_BASE_REF}`;
  if (process.env.GITHUB_EVENT_BEFORE && !/^0+$/.test(process.env.GITHUB_EVENT_BEFORE)) {
    return process.env.GITHUB_EVENT_BEFORE;
  }
  if (process.env.GITHUB_SHA) return `${process.env.GITHUB_SHA}^`;
  return null;
}

// Ordered best-first. Only refs that actually resolve are used, and the result
// is always a merge base, so a wrong guess here widens the diff at worst and
// never narrows it.
function candidateBaseRefs() {
  const refs = [];
  const upstream = gitOrNull(['rev-parse', '--abbrev-ref', '--symbolic-full-name', '@{upstream}']);
  if (upstream) refs.push(upstream.trim());

  for (const remote of (gitOrNull(['remote']) ?? '').split('\n').filter(Boolean)) {
    const head = gitOrNull(['symbolic-ref', '--quiet', `refs/remotes/${remote}/HEAD`]);
    if (head) refs.push(head.trim());
    for (const branch of ['main', 'master']) {
      if (refs.includes(`${remote}/${branch}`)) continue;
      const ref = gitOrNull([
        'rev-parse',
        '--verify',
        '--quiet',
        `refs/remotes/${remote}/${branch}^{commit}`,
      ]);
      if (ref) refs.push(`${remote}/${branch}`);
    }
  }
  return refs;
}

function isShallow() {
  return (gitOrNull(['rev-parse', '--is-shallow-repository']) ?? '').trim() === 'true';
}

// The commit the branch is compared against, or null plus the reason it could
// not be found. `--changed-only` promises that nothing new crossed the
// boundary, and reporting success after inspecting an empty `HEAD..HEAD` diff
// is the defect this replaced, so an unresolved base is never a pass.
function resolveBase() {
  const explicit = explicitBaseRef();
  for (const ref of explicit ? [explicit] : candidateBaseRefs()) {
    const commit = gitOrNull(['rev-parse', '--verify', '--quiet', `${ref}^{commit}`]);
    if (!commit) continue;
    const mergeBase = gitOrNull(['merge-base', 'HEAD', commit.trim()]);
    return { ref, sha: (mergeBase ?? commit).trim() };
  }

  if (isShallow()) {
    return {
      error:
        'this is a shallow clone and no diff base could be resolved, so the branch cannot be checked. Run `git fetch --unshallow` and commit again.',
    };
  }
  if (explicit) {
    return {
      error: `the requested diff base "${explicit}" does not exist locally. Fetch it (git fetch origin "${explicit}") or unset the CI base variables, and commit again.`,
    };
  }
  return {
    error:
      'no diff base could be resolved (no upstream branch and no remote HEAD), so the branch cannot be checked. Run `git fetch origin` and `git branch --set-upstream-to=origin/<branch>`, then commit again.',
  };
}

function addedLines(args) {
  const diff = git([
    'diff',
    '--unified=0',
    '--no-color',
    '--no-ext-diff',
    ...args,
    '--',
    appsPathspec,
  ]);
  const added = [];
  let file = null;
  let lineNumber = 0;

  for (const line of diff.split('\n')) {
    if (line.startsWith('+++ ')) {
      const target = line.slice(4).replace(/\t.*$/, '').trim();
      file = target === '/dev/null' ? null : target.replace(/^b\//, '');
      continue;
    }
    const hunk = line.match(/^@@ -\d+(?:,\d+)? \+(\d+)(?:,\d+)? @@/);
    if (hunk) {
      lineNumber = Number(hunk[1]);
      continue;
    }
    if (file === null) continue;
    if (line.startsWith('+')) {
      added.push({ file, line: lineNumber, text: line.slice(1) });
      lineNumber += 1;
    } else if (!line.startsWith('-')) {
      lineNumber += 1;
    }
  }
  return added;
}

function wholeFile(fileName) {
  return readFileSync(join(repoRoot, fileName), 'utf8')
    .split('\n')
    .map((text, index) => ({ file: fileName, line: index + 1, text }));
}

// The pre-commit hook runs after `git add`, so the staged index is the one that
// decides the commit. The branch diff covers what is already committed and the
// worktree diff covers edits made after staging.
function reviewedLines(base) {
  const hasHead = gitOrNull(['rev-parse', '--verify', '--quiet', 'HEAD^{commit}']) !== null;
  const head = hasHead ? 'HEAD' : '4b825dc642cb6eb9a060e54bf8d69288fbee4904';
  const lines = [];

  // Skipped when there is no base to diff against; the caller fails loudly
  // rather than treating the missing branch diff as a clean result.
  if (base && base.sha !== head) lines.push(...addedLines([base.sha, head]));
  lines.push(...addedLines(['--cached', head]));
  lines.push(...addedLines([]));

  // git diff does not include untracked files. Treat new app source files as
  // entirely new so the guard is useful before the first commit as well.
  const untracked = git([
    'ls-files',
    '--others',
    '--exclude-standard',
    '--full-name',
    '--',
    appsPathspec,
  ])
    .split('\n')
    .filter(Boolean)
    .filter((fileName) => sourceExtensions.has(fileName.slice(fileName.lastIndexOf('.'))));
  for (const fileName of untracked) lines.push(...wholeFile(fileName));

  return lines
    .filter((entry) => isAppSource(entry.file))
    .map((entry) => ({
      file: toPackagePath(entry.file),
      line: entry.line,
      text: entry.text,
    }));
}

function findingsForLine(file, line, text) {
  return rules.flatMap((rule) => {
    if (rule.code === 'RAW_FETCH' && rawFetchAllowlist.has(file)) return [];
    return rule.pattern.test(text) ? [{ file, line, code: rule.code, message: rule.message }] : [];
  });
}

function appDependencyFindings() {
  const findings = [];
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
const seen = new Set();
const record = (finding) => {
  const key = `${finding.file}:${finding.line}:${finding.code}`;
  if (seen.has(key)) return;
  seen.add(key);
  findings.push(finding);
};

findings.push(...appDependencyFindings());
let baseError = null;
if (changedOnly) {
  const base = resolveBase();
  if (base.error) {
    baseError = base.error;
  } else if (isShallow()) {
    console.error(
      `OneKit boundary: warning: shallow clone; the diff against ${base.ref} may be incomplete.`
    );
  }
  // Staged, unstaged and untracked lines need no base, so they are still judged
  // when the base is unresolvable: the author sees every real violation in one
  // run instead of fixing them one commit attempt at a time.
  for (const entry of reviewedLines(base.error ? null : base)) {
    for (const finding of findingsForLine(entry.file, entry.line, entry.text)) record(finding);
  }
} else if (scanAll) {
  for (const file of walk(appsDirectory)) {
    const fileName = relativePath(file);
    const contents = readFileSync(file, 'utf8').split('\n');
    contents.forEach((text, index) => {
      for (const finding of findingsForLine(fileName, index + 1, text)) record(finding);
    });
  }
}

if (findings.length > 0) {
  console.error(
    `OneKit boundary: ${findings.length} violation(s)${changedOnly ? ' in added lines' : ''} under ${appsPrefix}/.`
  );
  for (const finding of findings.slice(0, 100)) {
    console.error(`- ${finding.file}:${finding.line} [${finding.code}] ${finding.message}`);
  }
  if (findings.length > 100) console.error(`- ... ${findings.length - 100} more`);
  console.error(
    'Apps must reach the backend through a generated OneKit service client. Revert these lines, or allowlist a deliberate exception in scripts/enforce-onekit-boundary.mjs.'
  );
  process.exitCode = 1;
} else if (!baseError) {
  console.log(`OneKit boundary: no ${changedOnly ? 'new ' : ''}app transport violations found.`);
}

if (baseError) {
  // Exit 2 keeps this distinguishable from an actual boundary violation (1).
  console.error(`OneKit boundary: ${baseError}`);
  process.exit(2);
}
