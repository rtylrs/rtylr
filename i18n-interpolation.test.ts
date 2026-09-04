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

describe('every plural group defines the categories its language needs', () => {
  // A group is the set of `key_<category>` entries sharing a base key. i18next
  // picks the category from Intl.PluralRules for the active language, so a
  // group that omits one has nothing to return when that category comes up.
  //
  // When the group has no un-suffixed base key, and none of these do, the
  // fallback lands nowhere and t() returns the bare key: the user sees
  // `unreadCount` where a sentence belongs. With a base key present the
  // fallback resolves and the failure is quieter, wrong number agreement
  // rather than a raw key, which is why the base key is reported alongside.
  //
  // Reachability varies sharply by language and is not a reason to skip a
  // category. French selects `many` only at 1,000,000 and above, so a French
  // gap is latent. Arabic reaches `two`, `few` and `many` at counts of 2, 3
  // and 11, so an Arabic gap is an everyday bug.
  type Group = { file: string; base: string; missing: string[]; hasBase: boolean };

  function flatten(value: unknown, prefix: string, into: Map<string, unknown>): void {
    if (value !== null && typeof value === 'object' && !Array.isArray(value)) {
      for (const [key, child] of Object.entries(value as Record<string, unknown>)) {
        flatten(child, prefix ? `${prefix}.${key}` : key, into);
      }
      return;
    }
    into.set(prefix, value);
  }

  function incompleteGroups(app: string): Group[] {
    const found: Group[] = [];
    for (const file of localeFiles(app)) {
      const lang = path.basename(path.dirname(file));
      let required: readonly string[];
      try {
        required = new Intl.PluralRules(lang).resolvedOptions().pluralCategories;
      } catch {
        // Not a language tag Intl understands, so there is no rule to check.
        continue;
      }
      const entries = new Map<string, unknown>();
      flatten(JSON.parse(fs.readFileSync(file, 'utf8')), '', entries);

      const groups = new Map<string, Set<string>>();
      for (const key of entries.keys()) {
        const match = key.match(/^(.*)_(zero|one|two|few|many|other)$/);
        if (!match) continue;
        const forms = groups.get(match[1]) ?? new Set<string>();
        forms.add(match[2]);
        groups.set(match[1], forms);
      }

      for (const [base, forms] of groups) {
        const missing = required.filter((category) => !forms.has(category));
        if (missing.length === 0) continue;
        found.push({
          file: path.relative(APPS_DIR, file),
          base,
          missing,
          hasBase: entries.has(base),
        });
      }
    }
    return found;
  }

  it('finds plural groups to check, so a passing run means something', () => {
    const total = APPS.reduce((count, app) => {
      for (const file of localeFiles(app)) {
        const entries = new Map<string, unknown>();
        flatten(JSON.parse(fs.readFileSync(file, 'utf8')), '', entries);
        const bases = new Set<string>();
        for (const key of entries.keys()) {
          const match = key.match(/^(.*)_(zero|one|two|few|many|other)$/);
          if (match) bases.add(`${file}:${match[1]}`);
        }
        count += bases.size;
      }
      return count;
    }, 0);
    expect(total).toBeGreaterThan(40);
  });

  it.each(APPS)('%s leaves no plural group short of a category', (app) => {
    const offenders = incompleteGroups(app).map(
      (group) =>
        `${group.file}  ${group.base}  missing ${group.missing.join(',')}` +
        `${group.hasBase ? '  (falls back to the base key)' : '  (no base key: renders the raw key)'}`
    );
    expect(offenders).toEqual([]);
  });

  it.each(PLURAL_CATEGORIES)('%s is a category Intl can actually return', (category) => {
    // Guards the regex above: if a category name here stopped matching what
    // Intl reports, the scan would silently find fewer groups.
    const everyCategory = new Set(
      ['en', 'ar', 'fr'].flatMap((lang) => [
        ...new Intl.PluralRules(lang).resolvedOptions().pluralCategories,
      ])
    );
    expect(everyCategory).toContain(category);
  });
});
