import { describe, it, expect } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';

/**
 * Every `t()` key in the source must exist in the locale files.
 *
 * This is the guard that was missing, and the defect it catches is worse than a
 * missing translation: **an unresolved key renders as itself**. Recruit called
 * `t('emptyStates.rolesNone')` in 15 places, unqualified, so an admin with no
 * roles yet read the literal text `emptyStates.rolesNone` where a sentence
 * belonged, in **every** language including English (rtylr#1266).
 *
 * Measured, by initialising i18next exactly as `shared/services/i18next.ts`
 * does:
 *
 *   t('emptyStates.rolesNone')         -> "emptyStates.rolesNone"
 *   t('common.emptyStates.rolesNone')  -> "No roles yet. Create your first…"
 *
 * The loader keys each namespace by filename and no `defaultNS` is passed, so
 * `initI18n`'s default of `'translation'` applies and nothing lives there. Every
 * key therefore needs its file's name as a prefix.
 *
 * A missing-translation check cannot see this, because the key is absent from
 * every language equally. Only comparing source against the files does.
 */
const LOCALES = path.resolve(__dirname);
const SRC = path.resolve(__dirname, '..');

function flatten(value: unknown, prefix: string, into: Set<string>): void {
  if (value && typeof value === 'object' && !Array.isArray(value)) {
    for (const [key, nested] of Object.entries(value as Record<string, unknown>)) {
      flatten(nested, prefix ? `${prefix}.${key}` : key, into);
    }
  } else if (typeof value === 'string') {
    into.add(prefix);
  }
}

function knownKeys(lang: string): Set<string> {
  const dir = path.join(LOCALES, lang);
  const keys = new Set<string>();
  for (const file of fs.readdirSync(dir)) {
    if (!file.endsWith('.json')) continue;
    // The namespace is the filename, and the loader keys resources by it, so
    // this prefix is not decoration: it is part of the key.
    const namespace = file.replace(/\.json$/, '');
    flatten(JSON.parse(fs.readFileSync(path.join(dir, file), 'utf8')), namespace, keys);
  }
  return keys;
}

/**
 * i18next's plural forms are stored suffixed and called unsuffixed: a source
 * call of `t('team.memberCount', { count })` is satisfied by
 * `memberCount_one` / `memberCount_other` and the bare key never exists. A
 * checker that misses this reports every correctly written plural as broken; it
 * put 11 of flow's keys and 2 of dash's on my own first list, wrongly.
 */
const PLURAL_SUFFIXES = ['_zero', '_one', '_two', '_few', '_many', '_other'];

export function resolves(key: string, known: Set<string>): boolean {
  if (known.has(key)) return true;
  return PLURAL_SUFFIXES.some((suffix) => known.has(`${key}${suffix}`));
}

function sourceFiles(dir: string, into: string[] = []): string[] {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      if (entry.name === 'node_modules' || entry.name === 'locales') continue;
      sourceFiles(full, into);
      continue;
    }
    if (/\.tsx?$/.test(entry.name) && !/\.test\.tsx?$/.test(entry.name)) into.push(full);
  }
  return into;
}

describe('recruit translation keys', () => {
  it('resolves every t() key in the source', () => {
    const known = knownKeys('en');
    const unresolved: string[] = [];

    for (const file of sourceFiles(SRC)) {
      const text = fs.readFileSync(file, 'utf8');
      // `{t('…')}` and `={t('…')}`, i.e. a call in a JSX position. A bare
      // `t('…')` match would also catch `split('…')` and friends, whose names
      // happen to end in t.
      for (const match of text.matchAll(/[={(,]\s*t\('([^']+)'/g)) {
        const key = match[1];
        // Keys built at runtime cannot be checked statically.
        if (key.includes('${')) continue;
        if (!resolves(key, known)) {
          const line = text.slice(0, match.index).split('\n').length;
          unresolved.push(`${path.relative(SRC, file)}:${line} ${key}`);
        }
      }
    }

    expect(unresolved, 'an unresolved key renders as itself, in every language').toEqual([]);
  });

  // Recruit has no plural keys yet, so the real locale set would leave the
  // plural branch above unexercised and free to rot.
  it('accepts a plural key by its suffixed forms', () => {
    const known = new Set(['common.applicants_one', 'common.applicants_other']);

    expect(resolves('common.applicants', known)).toBe(true);
    expect(resolves('common.applicants_other', known)).toBe(true);
    expect(resolves('common.candidates', known)).toBe(false);
  });
});
