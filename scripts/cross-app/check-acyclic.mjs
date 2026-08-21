#!/usr/bin/env node
// Asserts that the proposed cross-app dependency graph is acyclic, and prints
// the graph. Reads docs/cross-app/connections.json.
//
//   node scripts/cross-app/check-acyclic.mjs
//
// Two granularities are checked, because they answer different questions and
// only one of them is the real constraint:
//
//   1. DOMAIN graph: one node per contract domain, one edge per
//      producer->consumer pair. This is the coarse view a reader expects, and
//      it DOES contain cycles: erp<->finance and crm<->finance are each
//      proposed in both directions. Reported, not hidden.
//
//   2. FACT graph: one node per (domain, fact) where a fact is the specific
//      thing a connection moves. Rule 2 is really "no connection may depend on
//      the output of a connection that depends on it", and that is a statement
//      about facts, not about domains. A cycle here would be a genuine design
//      error. This is the assertion that must pass.
//
// The fact graph is derived from the `producerFact` and `consumerFact` fields
// on each connection in connections.json; a connection without them fails the
// check rather than being skipped, so the graph cannot be silently incomplete.

import { readFileSync } from 'node:fs';
import { join } from 'node:path';

const ROOT = new URL('../..', import.meta.url).pathname.replace(/\/$/, '');
const data = JSON.parse(readFileSync(join(ROOT, 'docs/cross-app/connections.json'), 'utf8'));

const proposed = data.connections.filter((c) => c.status === 'proposed');

// --- 1. domain graph --------------------------------------------------------
const domainEdges = new Map(); // "a>b" -> [ids]
for (const c of proposed) {
  if (c.consumer === 'all' || c.producer === c.consumer) continue;
  const k = `${c.producer}>${c.consumer}`;
  if (!domainEdges.has(k)) domainEdges.set(k, []);
  domainEdges.get(k).push(c.id);
}

function findCycles(nodes, edgesOf) {
  const WHITE = 0;
  const GREY = 1;
  const BLACK = 2;
  const colour = new Map(nodes.map((n) => [n, WHITE]));
  const stack = [];
  const cycles = [];
  function visit(n) {
    colour.set(n, GREY);
    stack.push(n);
    for (const m of edgesOf(n)) {
      const c = colour.get(m) ?? WHITE;
      if (c === GREY) {
        cycles.push([...stack.slice(stack.indexOf(m)), m]);
      } else if (c === WHITE) {
        visit(m);
      }
    }
    stack.pop();
    colour.set(n, BLACK);
  }
  for (const n of nodes) if ((colour.get(n) ?? WHITE) === WHITE) visit(n);
  return cycles;
}

const domainNodes = [...new Set([...domainEdges.keys()].flatMap((k) => k.split('>')))].sort();
const domainCycles = findCycles(domainNodes, (n) =>
  [...domainEdges.keys()].filter((k) => k.startsWith(n + '>')).map((k) => k.split('>')[1])
);

process.stdout.write('=== Domain graph ===\n');
for (const [k, ids] of [...domainEdges].sort())
  process.stdout.write(`  ${k.replace('>', ' -> ')}  (${ids.join(', ')})\n`);
process.stdout.write(`  ${domainEdges.size} directed edges over ${domainNodes.length} domains\n`);
if (domainCycles.length) {
  process.stdout.write(`  ${domainCycles.length} cycle(s) at domain granularity:\n`);
  for (const cy of domainCycles) process.stdout.write(`    ${cy.join(' -> ')}\n`);
  process.stdout.write(
    '  These are expected and documented in 12-coupling.md. The binding\n' +
      '  assertion is the fact graph below.\n'
  );
} else {
  process.stdout.write('  no cycles at domain granularity\n');
}

// --- 2. fact graph ----------------------------------------------------------
const missing = proposed.filter(
  (c) => c.consumer !== 'all' && c.producer !== c.consumer && (!c.producerFact || !c.consumerFact)
);

const factEdges = new Map(); // "producerFact" -> Set(consumerFact)
const factOwner = new Map(); // fact -> domain
for (const c of proposed) {
  if (c.consumer === 'all' || c.producer === c.consumer) continue;
  if (!c.producerFact || !c.consumerFact) continue;
  const from = `${c.producer}.${c.producerFact}`;
  const to = `${c.consumer}.${c.consumerFact}`;
  factOwner.set(from, c.producer);
  factOwner.set(to, c.consumer);
  if (!factEdges.has(from)) factEdges.set(from, new Set());
  factEdges.get(from).add(to);
}

const factNodes = [...new Set([...factOwner.keys()])].sort();
const factCycles = findCycles(factNodes, (n) => [...(factEdges.get(n) ?? [])]);

process.stdout.write('\n=== Fact graph ===\n');
for (const [from, tos] of [...factEdges].sort())
  for (const to of [...tos].sort()) process.stdout.write(`  ${from} -> ${to}\n`);
process.stdout.write(
  `  ${[...factEdges.values()].reduce((a, s) => a + s.size, 0)} edges over ${factNodes.length} facts\n`
);

let failed = false;
if (missing.length) {
  process.stdout.write(
    `\nFAIL: ${missing.length} connection(s) lack producerFact/consumerFact and cannot be placed in the fact graph:\n`
  );
  for (const c of missing) process.stdout.write(`  ${c.id}\n`);
  failed = true;
}
if (factCycles.length) {
  process.stdout.write(`\nFAIL: ${factCycles.length} cycle(s) in the fact graph:\n`);
  for (const cy of factCycles) process.stdout.write(`  ${cy.join(' -> ')}\n`);
  failed = true;
}

if (failed) process.exit(1);
process.stdout.write('\nPASS: the fact graph is acyclic and every connection is placed in it.\n');
