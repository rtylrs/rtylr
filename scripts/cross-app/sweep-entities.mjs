#!/usr/bin/env node
// Derives the canonical entity list from contracts.csv and emits
// docs/cross-app/inventory/entities.csv.
//
// An entity is a persisted noun: any contract message that carries an `id`
// field and is not a Request/Response wrapper. The same entity name declared
// in two domains produces two rows on purpose: duplicate nouns are themselves
// a finding (see 02-entity-graph.md).

import { readFileSync, writeFileSync } from 'node:fs';
import { join, relative } from 'node:path';

const ROOT = new URL('../..', import.meta.url).pathname;
const IN = join(ROOT, 'docs/cross-app/inventory/contracts.csv');
const OUT = join(ROOT, 'docs/cross-app/inventory/entities.csv');

function parseCsvLine(line) {
  const out = [];
  let cur = '';
  let inQ = false;
  for (let i = 0; i < line.length; i++) {
    const c = line[i];
    if (inQ) {
      if (c === '"' && line[i + 1] === '"') {
        cur += '"';
        i++;
      } else if (c === '"') inQ = false;
      else cur += c;
    } else if (c === '"') inQ = true;
    else if (c === ',') {
      out.push(cur);
      cur = '';
    } else cur += c;
  }
  out.push(cur);
  return out;
}

function csv(v) {
  const s = String(v ?? '');
  return /[",\n]/.test(s) ? `"${s.replaceAll('"', '""')}"` : s;
}

const lines = readFileSync(IN, 'utf8').split('\n').filter(Boolean).slice(1);

// Group message fields by (domain, message).
const messages = new Map();
for (const line of lines) {
  const [domain, capability, version, kind, name, fieldName, fieldType, , , file, ln] =
    parseCsvLine(line);
  if (kind !== 'message') continue;
  const key = `${domain} ${name}`;
  if (!messages.has(key))
    messages.set(key, { domain, capability, version, name, fields: [], file, line: ln });
  const rec = messages.get(key);
  if (fieldName) rec.fields.push({ fieldName, fieldType });
  else {
    rec.file = file;
    rec.line = ln;
  }
}

const rows = [
  ['entity', 'owning_domain', 'capability', 'version', 'primary_key', 'fk_fields', 'file', 'line'],
];
for (const rec of messages.values()) {
  if (/(Request|Response)$/.test(rec.name)) continue;
  const hasId = rec.fields.some((f) => f.fieldName === 'id');
  if (!hasId) continue;
  const fks = rec.fields
    .filter((f) => /_ids?$/.test(f.fieldName))
    .map((f) => `${f.fieldName}:${f.fieldType}`)
    .join('; ');
  rows.push([rec.name, rec.domain, rec.capability, rec.version, 'id', fks, rec.file, rec.line]);
}

rows.splice(
  1,
  rows.length - 1,
  ...rows.slice(1).sort((a, b) => a[1].localeCompare(b[1]) || a[0].localeCompare(b[0]))
);

writeFileSync(OUT, rows.map((r) => r.map(csv).join(',')).join('\n') + '\n');
process.stdout.write(`${rows.length - 1} entities -> ${relative(ROOT, OUT)}\n`);
