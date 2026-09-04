import fs from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';

/**
 * Every `t('literal.key')` in Flow must resolve in every locale.
 *
 * This exists because #1055 was filed for one raw key on screen
 * (`toasts.time.deleted`) and a scan found eight. The failure mode is quiet:
 * i18next renders the key itself, so a missed key ships as a user-visible
 * string like "toasts.time.deleted" and nothing fails.
 *
 * Only literal keys are checked. `t(`priority.${p}`)` and keys held in a
 * variable cannot be resolved statically, so they are out of scope here.
 */
const LOCALES_DIR = path.resolve(__dirname, '../../locales');
const SRC_DIR = path.resolve(__dirname, '../..');

// i18next plural suffixes: a `count` key is stored as key_one / key_other.
const PLURAL_SUFFIXES = ['', '_zero', '_one', '_two', '_few', '_many', '_other'];

function flatten(value: unknown, prefix: string, into: Set<string>): void {
  if (value !== null && typeof value === 'object' && !Array.isArray(value)) {
    for (const [key, child] of Object.entries(value as Record<string, unknown>)) {
      flatten(child, prefix ? `${prefix}.${key}` : key, into);
    }
    return;
  }
  into.add(prefix);
}

function keysFor(lang: string): Set<string> {
  const dir = path.join(LOCALES_DIR, lang);
  const keys = new Set<string>();
  for (const file of fs.readdirSync(dir).filter((f) => f.endsWith('.json'))) {
    // The filename is the namespace: t('common.save') reads common.json.
    const namespace = file.replace(/\.json$/, '');
    const parsed: unknown = JSON.parse(fs.readFileSync(path.join(dir, file), 'utf8'));
    flatten(parsed, namespace, keys);
  }
  return keys;
}

function collectSourceFiles(dir: string, into: string[]): string[] {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      if (entry.name === 'locales' || entry.name === 'node_modules') continue;
      collectSourceFiles(full, into);
      continue;
    }
    if (!/\.tsx?$/.test(entry.name)) continue;
    if (/\.test\.tsx?$/.test(entry.name)) continue;
    into.push(full);
  }
  return into;
}

const USED_KEY = /\bt\(\s*(['"])([A-Za-z0-9_.\-]+)\1/g;

function usedKeys(): Map<string, string> {
  const found = new Map<string, string>();
  for (const file of collectSourceFiles(SRC_DIR, [])) {
    const lines = fs.readFileSync(file, 'utf8').split('\n');
    lines.forEach((line, index) => {
      for (const match of line.matchAll(USED_KEY)) {
        const key = match[2];
        if (!found.has(key)) found.set(key, `${path.relative(SRC_DIR, file)}:${index + 1}`);
      }
    });
  }
  return found;
}

const LANGS = ['en', 'ar', 'fr'] as const;

describe('Flow i18n keys resolve', () => {
  const used = usedKeys();

  it('finds translation calls to check, so a passing run means something', () => {
    expect(used.size).toBeGreaterThan(500);
  });

  it.each(LANGS)('every literal key used in the app exists in %s', (lang) => {
    const available = keysFor(lang);
    const namespaces = new Set([...available].map((key) => key.split('.')[0]));

    const missing: string[] = [];
    for (const [key, where] of used) {
      // A key whose first segment is not a namespace file is not a translation
      // key at all, so it is not this test's business.
      if (!namespaces.has(key.split('.')[0])) continue;
      if (PLURAL_SUFFIXES.some((suffix) => available.has(key + suffix))) continue;
      missing.push(`${key}  (${where})`);
    }

    expect(missing).toEqual([]);
  });

  it.each(LANGS)('%s declares no namespace the other locales lack', (lang) => {
    const namespacesOf = (l: string): string[] =>
      fs
        .readdirSync(path.join(LOCALES_DIR, l))
        .filter((f) => f.endsWith('.json'))
        .sort();
    expect(namespacesOf(lang)).toEqual(namespacesOf('en'));
  });
});
