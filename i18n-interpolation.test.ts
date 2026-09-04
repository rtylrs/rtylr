import fs from 'node:fs';
import path from 'node:path';
import { describe, expect, it } from 'vitest';

/**
 * Two failure modes that ship silently to users, checked across every app.
 *
 * 1. A placeholder written with single braces. i18next interpolates `{{x}}`,
 *    so `{x}` is left in place and the reader sees the literal text.
 * 2. A `t()` key that resolves to an object rather than a string. i18next
 *    returns the key itself, so the reader sees something like
 *    `erp.supplierBills`.
 *
 * Neither throws, neither fails a typecheck, and both were live in three apps
 * at once when this was written.
 */
const APPS_DIR = path.resolve(__dirname, 'apps');

// A single-brace placeholder that is not one half of a {{...}} pair.
const SINGLE_BRACE = /(?<!\{)\{([A-Za-z_][A-Za-z0-9_]*)\}(?!\})/g;
const USED_KEY = /\bt\(\s*(['"])([A-Za-z0-9_.\-]+)\1/g;
const PLURAL_CATEGORIES = ['zero', 'one', 'two', 'few', 'many', 'other'];

function localeApps(): string[] {
  return fs
    .readdirSync(APPS_DIR)
    .filter((app) => fs.existsSync(path.join(APPS_DIR, app, 'src/locales')))
    .sort();
}

function localeFiles(app: string): string[] {
  const dir = path.join(APPS_DIR, app, 'src/locales');
  const found: string[] = [];
  for (const lang of fs.readdirSync(dir)) {
    const langDir = path.join(dir, lang);
    if (!fs.statSync(langDir).isDirectory()) continue;
    for (const file of fs.readdirSync(langDir)) {
      if (file.endsWith('.json')) found.push(path.join(langDir, file));
    }
  }
  return found;
}

/** The i18next resource tree for one language: filename becomes the namespace. */
function resources(app: string, lang: string): Record<string, unknown> {
  const dir = path.join(APPS_DIR, app, 'src/locales', lang);
  const tree: Record<string, unknown> = {};
  if (!fs.existsSync(dir)) return tree;
  for (const file of fs.readdirSync(dir).filter((f) => f.endsWith('.json'))) {
    tree[file.replace(/\.json$/, '')] = JSON.parse(fs.readFileSync(path.join(dir, file), 'utf8'));
  }
  return tree;
}

function resolve(tree: Record<string, unknown>, key: string): unknown {
  let node: unknown = tree;
  for (const part of key.split('.')) {
    if (node === null || typeof node !== 'object') return undefined;
    node = (node as Record<string, unknown>)[part];
  }
  return node;
}

function sourceFiles(dir: string, into: string[] = []): string[] {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      if (entry.name === 'locales' || entry.name === 'node_modules') continue;
      sourceFiles(full, into);
      continue;
    }
    if (/\.tsx?$/.test(entry.name) && !/\.test\.tsx?$/.test(entry.name)) into.push(full);
  }
  return into;
}

const APPS = localeApps();

describe('i18n placeholders use the braces i18next reads', () => {
  it('finds locale files to check, so a pass means something', () => {
    expect(APPS.length).toBeGreaterThan(5);
    expect(APPS.flatMap(localeFiles).length).toBeGreaterThan(20);
  });

  it.each(APPS)('%s has no single-brace placeholder in its copy', (app) => {
    const offenders: string[] = [];
    for (const file of localeFiles(app)) {
      const walk = (value: unknown, keyPath: string): void => {
        if (value !== null && typeof value === 'object') {
          for (const [key, child] of Object.entries(value as Record<string, unknown>)) {
            walk(child, keyPath ? `${keyPath}.${key}` : key);
          }
          return;
        }
        if (typeof value !== 'string') return;
        for (const match of value.matchAll(SINGLE_BRACE)) {
          offenders.push(`${path.relative(APPS_DIR, file)}  ${keyPath}  ${match[0]}`);
        }
      };
      walk(JSON.parse(fs.readFileSync(file, 'utf8')), '');
    }
    expect(offenders).toEqual([]);
  });
});

describe('plural selection is left to i18next', () => {
  // Picking the suffix by hand, as in
  // `t(n === 1 ? 'x' : 'x_other', { count: n })`, reimplements a job i18next
  // already does from Intl.PluralRules, and it is wrong outside English: it
  // cannot express Arabic's dual or French's `many`, so those languages get
  // the plural form for counts that need a different one.
  const EXPLICIT_SUFFIX = /['"][A-Za-z0-9_.]+_(?:zero|one|two|few|many|other)['"]/;

  it.each(APPS)('%s names no plural suffix in its source', (app) => {
    const offenders: string[] = [];
    for (const file of sourceFiles(path.join(APPS_DIR, app, 'src'))) {
      const lines = fs.readFileSync(file, 'utf8').split('\n');
      lines.forEach((line, index) => {
        if (EXPLICIT_SUFFIX.test(line)) {
          offenders.push(`${path.relative(APPS_DIR, file)}:${index + 1}  ${line.trim()}`);
        }
      });
    }
    expect(offenders).toEqual([]);
  });
});

describe('every literal t() key resolves to a string', () => {
  it.each(APPS)('%s has no key that resolves to an object', (app) => {
    const tree = resources(app, 'en');
    const namespaces = new Set(Object.keys(tree));
    const offenders = new Set<string>();

    for (const file of sourceFiles(path.join(APPS_DIR, app, 'src'))) {
      const lines = fs.readFileSync(file, 'utf8').split('\n');
      lines.forEach((line, index) => {
        for (const match of line.matchAll(USED_KEY)) {
          const key = match[2];
          if (!namespaces.has(key.split('.')[0])) continue;
          const value = resolve(tree, key);
          if (value === null || typeof value !== 'object') continue;
          // A pluralised key is legitimately an object at the parent level only
          // when the suffixed children are the strings, which is not this case:
          // here the key itself is what resolves, so check its own children.
          const asRecord = value as Record<string, unknown>;
          const isPluralGroup = PLURAL_CATEGORIES.some(
            (category) => typeof asRecord[category] === 'string'
          );
          if (isPluralGroup) continue;
          offenders.add(`${key}  (${path.relative(APPS_DIR, file)}:${index + 1})`);
        }
      });
    }

    expect([...offenders]).toEqual([]);
  });
});
