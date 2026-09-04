import fs from 'node:fs';
import path from 'node:path';

/**
 * Shared machinery for the "every literal t() key resolves" guard.
 *
 * Extracted from Flow's own copy of this test so Dash could be covered by the
 * same rule rather than by a second near-identical file.
 *
 * The rule is stricter than Flow's original version, which skipped any key
 * whose first segment was not the name of a locale file. That exemption was
 * precisely the shape of the Dash defect: `t('nav.account')` under
 * `useTranslation('common')` was skipped as "not a translation key" when it
 * was one, and a broken one. Every identifier-shaped literal key is checked
 * here, with no exemption for the first segment.
 *
 * Only literal keys can be checked. A computed key such as
 * `t(`priority.${p}`)` cannot be resolved statically and is out of scope.
 */

// i18next stores a `count` key as key_one / key_other, and Arabic uses all six
// categories. The empty string covers the non-plural case.
export const PLURAL_SUFFIXES = ['', '_zero', '_one', '_two', '_few', '_many', '_other'] as const;

// A `t(` that is not the tail of a longer identifier, so `signOut('dash')` and
// `split('/')` do not register as translation calls.
const LITERAL_KEY = /(?:^|[^A-Za-z0-9_$])t\(\s*(['"])([A-Za-z][A-Za-z0-9_.]*)\1/g;

function flatten(value: unknown, prefix: string, into: Set<string>): void {
  if (value !== null && typeof value === 'object' && !Array.isArray(value)) {
    for (const [key, child] of Object.entries(value as Record<string, unknown>)) {
      flatten(child, prefix ? `${prefix}.${key}` : key, into);
    }
    return;
  }
  into.add(prefix);
}

/** Every language directory under an app's `locales/`. */
export function localeLanguages(localesDir: string): string[] {
  return fs
    .readdirSync(localesDir, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => entry.name)
    .sort();
}

/** The JSON files in one language directory, sorted. */
export function localeNamespaces(localesDir: string, lang: string): string[] {
  return fs
    .readdirSync(path.join(localesDir, lang))
    .filter((file) => file.endsWith('.json'))
    .sort();
}

/**
 * Every key reachable in one language, qualified from the tree root.
 *
 * The filename is the first segment, because the apps' `loadResource` ignores
 * the namespace i18next asks for and returns the whole tree. That makes the
 * namespace cosmetic and the qualified path the only thing that resolves.
 */
export function availableKeys(localesDir: string, lang: string): Set<string> {
  const keys = new Set<string>();
  for (const file of localeNamespaces(localesDir, lang)) {
    const parsed: unknown = JSON.parse(fs.readFileSync(path.join(localesDir, lang, file), 'utf8'));
    flatten(parsed, file.replace(/\.json$/, ''), keys);
  }
  return keys;
}

function sourceFiles(dir: string, into: string[]): string[] {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      if (entry.name === 'locales' || entry.name === 'node_modules') continue;
      sourceFiles(full, into);
      continue;
    }
    if (!/\.tsx?$/.test(entry.name)) continue;
    if (/\.test\.tsx?$/.test(entry.name)) continue;
    into.push(full);
  }
  return into;
}

/** Literal t() keys used under `srcDir`, mapped to their first `file:line`. */
export function collectLiteralKeys(srcDir: string): Map<string, string> {
  const found = new Map<string, string>();
  for (const file of sourceFiles(srcDir, [])) {
    const lines = fs.readFileSync(file, 'utf8').split('\n');
    lines.forEach((line, index) => {
      for (const match of line.matchAll(LITERAL_KEY)) {
        const key = match[2];
        if (!found.has(key)) found.set(key, `${path.relative(srcDir, file)}:${index + 1}`);
      }
    });
  }
  return found;
}

/** Used keys with no resolution in `available`, as `key  (file:line)` lines. */
export function unresolvedKeys(used: Map<string, string>, available: Set<string>): string[] {
  const missing: string[] = [];
  for (const [key, where] of used) {
    if (PLURAL_SUFFIXES.some((suffix) => available.has(key + suffix))) continue;
    missing.push(`${key}  (${where})`);
  }
  return missing.sort();
}
