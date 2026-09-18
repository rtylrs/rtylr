#!/usr/bin/env node
// Sweeps user-initiated actions (JSX handlers and mutation declarations)
// across all apps and emits docs/cross-app/inventory/actions.csv.
//
// Mechanical extraction only. The columns:
//   app, screen            file-derived feature/screen name
//   label_or_i18n_key      nearest t('...') key or aria-label near the handler
//   handler                the handler expression (trimmed)
//   mutation_called        mutate/mutateAsync target if the handler names one
//   entities_touched       the backend DOMAINS the same file makes requests to
//                          (a file-level approximation; per-action entity
//                          attribution needs a human pass and is done in the
//                          analysis documents, not here)
//   file, line
//   cross_domain_candidate yes when the file requests 2+ backend domains
//
// Test files are excluded. "Close modal" style handlers are included here and
// filtered during analysis; a superset CSV keeps the sweep honest.

import { readdirSync, readFileSync, writeFileSync, statSync } from 'node:fs';
import { join, relative } from 'node:path';

const ROOT = new URL('../..', import.meta.url).pathname;
const OUT = join(ROOT, 'docs/cross-app/inventory/actions.csv');

const DOMAINS = [
  'booking',
  'common',
  'console',
  'crm',
  'erp',
  'finance',
  'flow',
  'hr',
  'hub',
  'insights',
  'pos',
  'support',
];

function walk(dir) {
  const out = [];
  for (const e of readdirSync(dir, { withFileTypes: true })) {
    if (e.name === 'node_modules' || e.name === 'dist') continue;
    const p = join(dir, e.name);
    if (e.isDirectory()) out.push(...walk(p));
    else if (/\.(tsx|jsx)$/.test(e.name) && !/\.test\.(tsx|jsx)$/.test(e.name)) out.push(p);
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

function domainsOfFile(src) {
  const found = new Set();
  for (const d of DOMAINS) {
    if (new RegExp(`['"\`]/${d}/`).test(src)) found.add(d);
    if (new RegExp(`@voxire/voxie/${d}/`).test(src)) found.add(d);
  }
  return [...found].sort();
}

function nearestLabel(src, idx) {
  // Look around the handler for a t('...') key or an aria-label.
  const windowText = src.slice(Math.max(0, idx - 300), idx + 500);
  const t = windowText.match(/\bt\(\s*['"`]([^'"`]+)['"`]/);
  if (t) return t[1];
  const aria = windowText.match(/aria-label=\{?['"`]?([^'"}`>\n]+)/);
  if (aria) return 'aria:' + aria[1].trim();
  return '';
}

const rows = [
  [
    'app',
    'screen',
    'label_or_i18n_key',
    'handler',
    'mutation_called',
    'entities_touched',
    'file',
    'line',
    'cross_domain_candidate',
  ],
];

for (const app of readdirSync(join(ROOT, 'apps'))) {
  const dir = join(ROOT, 'apps', app, 'src');
  try {
    if (!statSync(dir).isDirectory()) continue;
  } catch {
    continue;
  }
  for (const file of walk(dir)) {
    const rel = relative(ROOT, file);
    const src = readFileSync(file, 'utf8');
    const fileDomains = domainsOfFile(src);
    const cross = fileDomains.length >= 2 ? 'yes' : '';
    const screen = (rel.match(/src\/(?:features\/)?([^/]+\/[^/]+)\.tsx$/) || [, rel])[1];

    const handlerRe = /\bon(?:Click|Submit|Confirm|Save|Delete|Create|Export|Import)=\{/g;
    let m;
    while ((m = handlerRe.exec(src))) {
      // capture the balanced {...} expression
      let depth = 0;
      let i = m.index + m[0].length - 1;
      const start = i + 1;
      while (i < src.length) {
        if (src[i] === '{') depth++;
        else if (src[i] === '}') {
          depth--;
          if (depth === 0) break;
        }
        i++;
      }
      const expr = src.slice(start, i).replace(/\s+/g, ' ').trim().slice(0, 80);
      const mut = expr.match(/(\w+)\.(?:mutate|mutateAsync)\b/) || expr.match(/\b(mutate\w*)\(/);
      rows.push([
        app,
        screen,
        nearestLabel(src, m.index),
        expr,
        mut ? mut[1] : '',
        fileDomains.join(';'),
        rel,
        lineOf(src, m.index),
        cross,
      ]);
    }

    // Cross-app navigation is a boundary crossing even with no mutation:
    // APP_URLS.<app> links, window.open, and location.href assignments.
    const navRe = /\b(APP_URLS\.\w+|window\.open\(|(?:window\.)?location\.href\s*=)/g;
    while ((m = navRe.exec(src))) {
      rows.push([
        app,
        screen,
        nearestLabel(src, m.index),
        `navigate:${m[1].replace(/\s*=$/, '').replace(/\($/, '')}`,
        '',
        fileDomains.join(';'),
        rel,
        lineOf(src, m.index),
        cross,
      ]);
    }
  }
}

writeFileSync(OUT, rows.map((r) => r.map(csv).join(',')).join('\n') + '\n');
process.stdout.write(`${rows.length - 1} actions -> ${relative(ROOT, OUT)}\n`);
