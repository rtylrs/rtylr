import fs from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';

import { SEGMENT_RULE_FIELDS, SEGMENT_RULE_OPERATORS } from '../shared/lib/segmentRules';

/**
 * Keys that are built rather than written, and so are invisible to
 * `i18n-key-resolution.test.ts`, which collects literal `t('a.b')` keys only.
 *
 * crm grew three such families when its module-level label maps became id
 * lists: lead stages, segment rule fields and segment rule operators. Each is
 * read as `` t(`crm.<family>.${id}`) ``, so nothing else checks that the ids
 * and the locale files still agree.
 *
 * dash has the same guard for the same reason (rtylr#1266). The precedent there
 * was `nav.createItem`, which existed in no locale while the screen rendered
 * English, because the call site built its key.
 */
const LOCALES = path.resolve(__dirname);
const LANGS = ['en', 'ar', 'fr'] as const;

function load(lang: string): Record<string, unknown> {
  const dir = path.join(LOCALES, lang);
  const tree: Record<string, unknown> = {};
  for (const file of fs.readdirSync(dir)) {
    if (!file.endsWith('.json')) continue;
    tree[file.replace('.json', '')] = JSON.parse(fs.readFileSync(path.join(dir, file), 'utf8'));
  }
  return tree;
}

function resolves(tree: Record<string, unknown>, dotted: string): boolean {
  let node: unknown = tree;
  for (const part of dotted.split('.')) {
    if (typeof node !== 'object' || node === null) return false;
    node = (node as Record<string, unknown>)[part];
  }
  return typeof node === 'string';
}

/** Operators are symbols in code; these are the key suffixes they map to. */
const OPERATOR_SUFFIX: Record<string, string> = { '>': 'gt', '<': 'lt', '=': 'eq', '!=': 'ne' };

const LEAD_STAGES = [
  'cold',
  'contacted',
  'demo_booked',
  'demo_done',
  'proposal_sent',
  'contracted',
  'lost',
];

const FAMILIES: Record<string, string[]> = {
  leadStages: LEAD_STAGES.map((s) => `crm.leadStages.${s}`),
  segmentFields: SEGMENT_RULE_FIELDS.map((f) => `crm.segmentFields.${f}`),
  segmentOperators: SEGMENT_RULE_OPERATORS.map((o) => `crm.segmentOperators.${OPERATOR_SUFFIX[o]}`),
};

describe.each(LANGS)('crm dynamic i18n keys resolve in %s', (lang) => {
  const tree = load(lang);

  it.each(Object.entries(FAMILIES))('%s', (_family, keys) => {
    expect(keys.length).toBeGreaterThan(0);
    expect(keys.filter((k) => !resolves(tree, k))).toEqual([]);
  });
});
