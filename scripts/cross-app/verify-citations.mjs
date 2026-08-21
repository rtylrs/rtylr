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

import { readdirSync, readFileSync, existsSync, statSync } from 'node:fs';
import { join, relative, resolve } from 'node:path';
import { homedir } from 'node:os';

const ROOT = new URL('../..', import.meta.url).pathname.replace(/\/$/, '');
const DOCS = join(ROOT, 'docs/cross-app');

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

function resolveTarget(p) {
  if (p.startsWith('~/')) return join(homedir(), p.slice(2));
  if (p.startsWith('/')) return p;
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
    for (const m of line.matchAll(CITATION))
      events.push({ at: m.index, kind: 'cite', m });
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
      if (!existsSync(abs)) {
        // Only report paths that were meant as citations: those carrying a
        // line number, or repo-relative paths under a known source dir.
        if (lineNo || /^(apps|packages|scripts|docs|e2e|test)\//.test(p)) {
          failures.push(`${relDoc}:${ln}  missing file: ${p}`);
        }
        continue;
      }
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
if (failures.length) {
  process.stdout.write(`\n${failures.length} FAILURES:\n`);
  for (const f of failures) process.stdout.write(`  ${f}\n`);
  process.exit(1);
}
process.stdout.write('all citations and links resolve\n');
