import { describe, it, expect } from 'vitest';
import en from './en/flow.json';
import ar from './ar/flow.json';
import fr from './fr/flow.json';

/**
 * Guards the one i18n failure the app cannot report: a key that exists in en and
 * not in ar or fr renders the raw key path to an Arabic or French user, and
 * nothing in the build catches it.
 */

const PLURAL_SUFFIX = /_(zero|one|two|few|many|other)$/;

function keyPaths(value: unknown, prefix = ''): string[] {
  if (typeof value !== 'object' || value === null) return [prefix];
  return Object.entries(value as Record<string, unknown>).flatMap(([k, v]) =>
    keyPaths(v, prefix ? `${prefix}.${k}` : k)
  );
}

/** Plural variants of one string count as that one string. */
function logicalKeys(bundle: unknown): Set<string> {
  return new Set(keyPaths(bundle).map((k) => k.replace(PLURAL_SUFFIX, '')));
}

const EN = logicalKeys(en);

describe('flow locale parity', () => {
  it('carries a real English string for every key', () => {
    for (const path of keyPaths(en)) {
      const value = path
        .split('.')
        .reduce<unknown>((acc, k) => (acc as Record<string, unknown>)[k], en);
      expect(typeof value, path).toBe('string');
      expect((value as string).trim().length, path).toBeGreaterThan(0);
    }
  });

  it.each([
    ['ar', ar],
    ['fr', fr],
  ])('translates every English key in %s', (_lang, bundle) => {
    const missing = [...EN].filter((k) => !logicalKeys(bundle).has(k));

    expect(missing).toEqual([]);
  });

  it.each([
    ['ar', ar],
    ['fr', fr],
  ])('has no key in %s that English does not define', (_lang, bundle) => {
    const orphaned = [...logicalKeys(bundle)].filter((k) => !EN.has(k));

    expect(orphaned).toEqual([]);
  });

  it('leaves no string untranslated by copying the English through', () => {
    // Deliberate matches: the duration notations the parser accepts are literal
    // input, not prose, and "Date" is the same word in French.
    const allowed = new Set([
      'time.durationPlaceholder',
      'time.durationInvalid',
      'board.confirmRemoveColumn',
      'time.date',
      // "Type" is the ordinary French word too, same as "Date" above.
      'leave.leaveType',
      // "Section" is the ordinary French word too; taskDetail.noSection
      // ("Aucune section") already establishes it as the correct term here.
      'taskDetail.section',
      // Pure placeholder with no prose to translate, same as the duration
      // placeholders above.
      'taskDetail.genericActivityNoValue',
      // "Description" is the same word in French too.
      'settings.workTypes.description',
      // "Mentions" (as in @mentions) is the same word in French too.
      'settings.notifications.items.mentions.label',
    ]);

    const copied = keyPaths(en).filter((path) => {
      if (allowed.has(path.replace(PLURAL_SUFFIX, ''))) return false;
      const read = (b: unknown): string =>
        path.split('.').reduce<unknown>((acc, k) => {
          const next = acc as Record<string, unknown> | undefined;
          return next?.[k];
        }, b) as string;
      const enValue = read(en);
      return read(ar) === enValue || read(fr) === enValue;
    });

    expect(copied).toEqual([]);
  });
});
