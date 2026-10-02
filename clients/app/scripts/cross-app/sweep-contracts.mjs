#!/usr/bin/env node
// Sweeps every .onk contract file under packages/voxie/api and emits
// docs/cross-app/inventory/contracts.csv.
//
// Row semantics:
//   kind=message, field_name=""      -> the message declaration itself
//   kind=message, field_name set     -> one field of that message
//   kind=rpc                          -> one RPC; field_name is the request
//                                        type, field_type is the response
//                                        type, and doc_comment begins with
//                                        the HTTP annotation, e.g. @get("/x")
//
// The .onk grammar swept here (verified against the corpus, 2026-08-21):
// top-level `message Name {` and `service Name {` blocks only, `///` doc
// comments, fields `name: type[?] [@annotations]`, RPCs
// `method(Request) -> Response @verb("/path")`.

import { readdirSync, readFileSync, writeFileSync, mkdirSync } from 'node:fs';
import { join, relative } from 'node:path';

const ROOT = new URL('../..', import.meta.url).pathname;
const API = join(ROOT, 'packages/voxie/api');
const OUT = join(ROOT, 'docs/cross-app/inventory/contracts.csv');

function walk(dir) {
  const out = [];
  for (const e of readdirSync(dir, { withFileTypes: true })) {
    const p = join(dir, e.name);
    if (e.isDirectory()) out.push(...walk(p));
    else if (e.name.endsWith('.onk')) out.push(p);
  }
  return out;
}

function csv(v) {
  const s = String(v ?? '');
  return /[",\n]/.test(s) ? `"${s.replaceAll('"', '""')}"` : s;
}

const rows = [
  [
    'domain',
    'capability',
    'version',
    'kind',
    'name',
    'field_name',
    'field_type',
    'nullable',
    'doc_comment',
    'file',
    'line',
  ],
];

const files = walk(API).sort();
for (const file of files) {
  const rel = relative(ROOT, file);
  // api/<domain>/<capability...>/<version>/<file>.onk
  const parts = relative(API, file).split('/');
  const domain = parts[0];
  const version = parts.length >= 3 ? parts[parts.length - 2] : '';
  const capability = parts.slice(1, -2).join('/');
  const lines = readFileSync(file, 'utf8').split('\n');

  let doc = [];
  let block = null; // { kind: 'message'|'service', name }
  for (let i = 0; i < lines.length; i++) {
    const raw = lines[i];
    const line = raw.trim();
    const ln = i + 1;
    if (line.startsWith('///')) {
      doc.push(line.slice(3).trim());
      continue;
    }
    let m;
    if ((m = raw.match(/^message (\w+) \{/))) {
      block = { kind: 'message', name: m[1] };
      rows.push([domain, capability, version, 'message', m[1], '', '', '', doc.join(' '), rel, ln]);
      doc = [];
      continue;
    }
    if ((m = raw.match(/^service (\w+) \{/))) {
      block = { kind: 'service', name: m[1] };
      doc = [];
      continue;
    }
    if (raw.startsWith('}')) {
      block = null;
      doc = [];
      continue;
    }
    if (block?.kind === 'message' && (m = raw.match(/^  (\w+): (\S+(?:, \S+\])?)(.*)$/))) {
      const [, fieldName, rawType, rest] = m;
      const nullable = rawType.endsWith('?') || rest.includes('@nullable');
      rows.push([
        domain,
        capability,
        version,
        'message',
        block.name,
        fieldName,
        rawType.replace(/\?$/, ''),
        nullable ? 'yes' : 'no',
        doc.join(' '),
        rel,
        ln,
      ]);
      doc = [];
      continue;
    }
    if (
      block?.kind === 'service' &&
      (m = raw.match(/^  (\w+)\((\w+)\) -> (\w+) (@\w+\("[^"]*"\))/))
    ) {
      const [, method, req, res, http] = m;
      rows.push([
        domain,
        capability,
        version,
        'rpc',
        `${block.name}.${method}`,
        req,
        res,
        '',
        [http, ...doc].join(' '),
        rel,
        ln,
      ]);
      doc = [];
      continue;
    }
    if (line !== '' && !line.startsWith('package ') && !line.startsWith('//')) {
      // Anything unparsed is a coverage gap; surface it rather than hide it.
      process.stderr.write(`UNPARSED ${rel}:${ln}: ${line}\n`);
    }
    doc = [];
  }
}

mkdirSync(join(ROOT, 'docs/cross-app/inventory'), { recursive: true });
writeFileSync(OUT, rows.map((r) => r.map(csv).join(',')).join('\n') + '\n');
process.stdout.write(`${files.length} files, ${rows.length - 1} rows -> ${relative(ROOT, OUT)}\n`);
