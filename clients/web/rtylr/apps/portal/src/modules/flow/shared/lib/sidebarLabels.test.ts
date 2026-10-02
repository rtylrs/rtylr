import { describe, it, expect } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';
import { createDefaultSidebarPreferences } from './sidebarPreferences';

/**
 * Two sidebar entries were called "Time" and two were called "Team": a group
 * and a leaf inside it, in both cases (#1097, D-057). A duplicate name in a
 * navigation list is not a cosmetic problem. It makes the two rows
 * indistinguishable by their accessible name, so a screen reader announces the
 * same thing twice and "click Team" is ambiguous instruction.
 *
 * Checked in every shipped language, because a collision can exist in one
 * translation and not in English. That is how it would go unnoticed: the
 * reviewer reads English.
 */
const LOCALES = path.resolve(__dirname, '../../locales');
const LANGUAGES = fs.readdirSync(LOCALES).filter((entry) => /^[a-z]{2}$/.test(entry));

function loadLanguage(lang: string): Record<string, unknown> {
  const dir = path.join(LOCALES, lang);
  const tree: Record<string, unknown> = {};
  for (const file of fs.readdirSync(dir)) {
    if (!file.endsWith('.json')) continue;
    tree[path.basename(file, '.json')] = JSON.parse(fs.readFileSync(path.join(dir, file), 'utf8'));
  }
  return tree;
}

function resolve(tree: Record<string, unknown>, key: string): string | undefined {
  let node: unknown = tree;
  for (const segment of key.split('.')) {
    if (typeof node !== 'object' || node === null) return undefined;
    node = (node as Record<string, unknown>)[segment];
  }
  return typeof node === 'string' ? node : undefined;
}

describe('no two sidebar entries share a name', () => {
  it('finds the shipped languages, so this cannot pass by checking nothing', () => {
    expect(LANGUAGES).toContain('en');
    expect(LANGUAGES.length).toBeGreaterThanOrEqual(2);
  });

  it.each(LANGUAGES)('in %s', (lang) => {
    const tree = loadLanguage(lang);
    const defaults = createDefaultSidebarPreferences();

    const keys = [
      ...defaults.groups.map((group) => group.label),
      ...defaults.items.map((item) => item.label),
    ];

    const byLabel = new Map<string, string[]>();
    for (const key of keys) {
      const label = resolve(tree, key);
      // A key with no translation is a different defect, covered elsewhere.
      if (!label) continue;
      byLabel.set(label, [...(byLabel.get(label) ?? []), key]);
    }

    const collisions = [...byLabel.entries()]
      .filter(([, usedBy]) => usedBy.length > 1)
      .map(([label, usedBy]) => `${label}: ${usedBy.join(', ')}`);

    expect(collisions, `duplicate sidebar names in ${lang}`).toEqual([]);
  });
});
