#!/usr/bin/env node
// Verifies every file:line citation and every relative doc link in
// docs/cross-app/**.md actually resolves. A citation that does not resolve is a
// fabrication and must be fixed or removed, so this exits non-zero on any
// failure.
//
//   node scripts/cross-app/verify-citations.mjs
//
// Recognised citation forms:
//   path/to/file.ext:123                  repo-relative
//   ~/Development/onion/path/file.go:123  another local repo (home-relative)
//   path/to/file.ext:123`, `:456          a bare :NNN continuation reuses the
//                                         most recent path on the same line
//
// Citations inside fenced code blocks are skipped: those are illustrative
// snippets, not claims. Line numbers must be within the file's line count.
//
// A repo-relative citation must also be TRACKED IN GIT, not merely present on
// disk. Checking the working tree alone let this script pass on a machine where
// an untracked local file happened to satisfy a citation, while the same
// citation failed for everyone else: that is how the `packages/CLAUDE.md`
// citation in 10-doc-drift.md survived (#864). Submodule contents count as
// tracked, via `git ls-files --recurse-submodules`. Paths outside this
// repository (`~/...`, `/...`) cannot be checked this way and are existence-only.

import { readdirSync, readFileSync, existsSync, statSync } from 'node:fs';
import { execFileSync } from 'node:child_process';
import { join, relative, resolve } from 'node:path';
import { homedir } from 'node:os';

const ROOT = new URL('../..', import.meta.url).pathname.replace(/\/$/, '');
const DOCS = join(ROOT, 'docs/cross-app');

// Every path git knows about, submodule contents included. A citation to a file
// that exists only in someone's working tree is a fabrication for everyone else.
const TRACKED = new Set(
  execFileSync('git', ['ls-files', '--recurse-submodules'], {
    cwd: ROOT,
    encoding: 'utf8',
    maxBuffer: 64 * 1024 * 1024,
  })
    .split('\n')
    .filter(Boolean)
);

// Locally modified files are the other half of the same problem. A line number
// checked against uncommitted content can resolve here and nowhere else, which
// is how `CLAUDE.md:901` was cited against a file that has 899 committed lines.
// This does not fail the run: citing a line you just wrote is normal while the
// change is in flight. It reports, so a green result is never mistaken for a
// reproducible one.
function locallyModified() {
  const paths = new Set();
  const add = (out, prefix) => {
    for (const f of out.split('\n').filter(Boolean)) paths.add(prefix + f);
  };
  const git = (args, cwd) => execFileSync('git', args, { cwd, encoding: 'utf8' });
  try {
    add(git(['diff', '--name-only', 'HEAD'], ROOT), '');
    const subs = git(['submodule', '--quiet', 'foreach', '--recursive', 'echo $displaypath'], ROOT)
      .split('\n')
      .filter(Boolean);
    for (const sub of subs) {
      add(git(['diff', '--name-only', 'HEAD'], join(ROOT, sub)), `${sub}/`);
    }
  } catch {
    // Not a git checkout, or git is unavailable. The tracked-set call above
    // would already have thrown, so reaching here means only the diff failed.
    return paths;
  }
  return paths;
}

const MODIFIED = locallyModified();
const citedButModified = new Set();

function walk(dir) {
  const out = [];
  for (const e of readdirSync(dir, { withFileTypes: true })) {
    const p = join(dir, e.name);
    if (e.isDirectory()) out.push(...walk(p));
    else if (e.name.endsWith('.md')) out.push(p);
  }
  return out;
}

const lineCountCache = new Map();
function lineCount(abs) {
  if (!lineCountCache.has(abs)) {
    try {
      lineCountCache.set(abs, readFileSync(abs, 'utf8').split('\n').length);
    } catch {
      lineCountCache.set(abs, null);
    }
  }
  return lineCountCache.get(abs);
}

// Round two: bare repo-relative citations are resolved against the repo they
// name, so documents can cite onion and voxie files without the ~/ prefix.
// The onion worktree is preferred (it carries this branch's tree) with the
// primary clone as fallback; voxie resolves to its own clone.
const ONION_TREE =
  process.env.CROSS_APP_ONION ||
  join(homedir(), 'Development', 'onion-worktrees', 'cross-app-round2');
const VOXIE_TREE = join(homedir(), 'Development', 'voxie');

// Tracked-file sets for the sibling repos, computed lazily on first use.
const SIBLING_TRACKED = (() => {
  const out = [];
  for (const root of [ONION_TREE, join(homedir(), 'Development', 'onion'), VOXIE_TREE]) {
    try {
      const files = new Set(
        execFileSync('git', ['ls-files', '--recurse-submodules'], {
          cwd: root,
          encoding: 'utf8',
          maxBuffer: 64 * 1024 * 1024,
        })
          .split('\n')
          .filter(Boolean)
      );
      out.push([root, files]);
    } catch {
      // Not a checkout; skip.
    }
  }
  return out;
})();

function resolveTarget(p) {
  if (p.startsWith('~/')) return join(homedir(), p.slice(2));
  if (p.startsWith('/')) return p;
  if (
    p.startsWith('internal/') ||
    p.startsWith('cmd/') ||
    p.startsWith('docs/sqlc-query-ownership.md') ||
    p.startsWith('docs/platform-patterns.md') ||
    p.startsWith('docs/tenant-index-matrix.md') ||
    p.startsWith('scripts/generate_sqlc.sh')
  ) {
    for (const root of [ONION_TREE, join(homedir(), 'Development', 'onion')]) {
      const abs = join(root, p);
      if (existsSync(abs)) return abs;
    }
    return join(ONION_TREE, p);
  }
  if (p.startsWith('api/') && p.includes('/v1')) {
    return join(VOXIE_TREE, p);
  }
  return join(ROOT, p);
}

const failures = [];
const stats = { citations: 0, links: 0, files: 0 };

// A path must look like a real path: at least one '/' and a file extension we
// expect to cite, so prose like "10:30" or "1:2" is not treated as a citation.
const CITATION =
  /((?:~\/|\/)?[\w.\-/]+\.(?:onk|go|sql|tsx|ts|jsx|js|mjs|json|md|css|yaml|yml|toml))(?::(\d+))?/g;
const BARE_LINE = /`:(\d+)`/g;
const MD_LINK = /\[[^\]]*\]\(([^)]+)\)/g;

for (const doc of walk(DOCS)) {
  stats.files++;
  const relDoc = relative(ROOT, doc);
  const lines = readFileSync(doc, 'utf8').split('\n');
  let inFence = false;
  let lastPathOnLine = null;

  lines.forEach((line, i) => {
    const ln = i + 1;
    if (/^\s*```/.test(line)) {
      inFence = !inFence;
      return;
    }
    if (inFence) return;

    lastPathOnLine = null;
    // Collect citations and bare-line continuations together and process them
    // in positional order: a `:456` continuation refers to the path that
    // precedes it on the line, not to the last path anywhere on the line.
    const events = [];
    for (const m of line.matchAll(CITATION)) events.push({ at: m.index, kind: 'cite', m });
    for (const m of line.matchAll(BARE_LINE)) events.push({ at: m.index, kind: 'cont', m });
    events.sort((a, b) => a.at - b.at);

    for (const ev of events) {
      if (ev.kind === 'cont') {
        if (!lastPathOnLine) continue;
        stats.citations++;
        const count = lineCount(lastPathOnLine);
        if (count !== null && (Number(ev.m[1]) < 1 || Number(ev.m[1]) > count)) {
          failures.push(
            `${relDoc}:${ln}  continuation line out of range: :${ev.m[1]} (file has ${count} lines)`
          );
        }
        continue;
      }
      const m = ev.m;
      {
        const [, p, lineNo] = m;
        const abs = resolveTarget(p);
        lastPathOnLine = abs;
        // Only report paths that were meant as citations: those carrying a line
        // number, or repo-relative paths under a known source dir.
        const isClaim = Boolean(lineNo) || /^(apps|packages|scripts|docs|e2e|test)\//.test(p);
        if (!existsSync(abs)) {
          if (isClaim) failures.push(`${relDoc}:${ln}  missing file: ${p}`);
          continue;
        }
        // Present on disk is not enough for a path inside this repository. An
        // untracked file satisfies the citation here and nowhere else. Paths
        // resolved into a sibling repo (onion, voxie) are tracked-checked
        // against that repo instead.
        const isRepoRelative = !p.startsWith('~/') && !p.startsWith('/');
        let trackedHere = TRACKED.has(p);
        if (isRepoRelative && !trackedHere) {
          for (const [root, set] of SIBLING_TRACKED) {
            if (abs.startsWith(root + '/')) {
              trackedHere = set.has(abs.slice(root.length + 1));
              break;
            }
          }
        }
        if (isClaim && isRepoRelative && !trackedHere) {
          failures.push(`${relDoc}:${ln}  present on disk but not tracked in git: ${p}`);
          continue;
        }
        if (isRepoRelative && MODIFIED.has(p)) citedButModified.add(p);
        if (!lineNo) continue;
        stats.citations++;
        if (statSync(abs).isDirectory()) {
          failures.push(`${relDoc}:${ln}  cited a directory with a line: ${p}:${lineNo}`);
          continue;
        }
        const count = lineCount(abs);
        if (count === null) {
          failures.push(`${relDoc}:${ln}  unreadable: ${p}`);
        } else if (Number(lineNo) < 1 || Number(lineNo) > count) {
          failures.push(
            `${relDoc}:${ln}  line out of range: ${p}:${lineNo} (file has ${count} lines)`
          );
        }
      }
    }

    // Relative markdown links between audit documents.
    for (const m of line.matchAll(MD_LINK)) {
      const target = m[1].split('#')[0];
      if (!target || /^(https?:|mailto:)/.test(target)) continue;
      stats.links++;
      const abs = target.startsWith('/') ? join(ROOT, target) : resolve(doc, '..', target);
      if (!existsSync(abs)) failures.push(`${relDoc}:${ln}  broken link: ${m[1]}`);
    }
  });
}

process.stdout.write(
  `checked ${stats.files} documents, ${stats.citations} file:line citations, ${stats.links} links\n`
);
if (citedButModified.size) {
  process.stdout.write(
    `\nWARNING: ${citedButModified.size} cited file(s) are modified locally, so their line\n` +
      `numbers were checked against uncommitted content and this result is not reproducible:\n`
  );
  for (const p of [...citedButModified].sort()) process.stdout.write(`  ${p}\n`);
}
if (failures.length) {
  process.stdout.write(`\n${failures.length} FAILURES:\n`);
  for (const f of failures) process.stdout.write(`  ${f}\n`);
  process.exit(1);
}
process.stdout.write('all citations and links resolve\n');
