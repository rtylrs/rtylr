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
      // "Actions" and "Navigation" are the same words in French too.
      'commandBar.groupActions',
      'commandBar.groupNavigation',
      // "Minutes" is the same word in French too, same reasoning as "Date" above.
      'time.add.minutes',
      // Pure placeholder with no prose to translate, same as the duration
      // placeholders above.
      'time.add.noteCounter',
      // "Mode", "Palette" and "Compact" are all ordinary French words, same
      // reasoning as "Date" and "Type" above. Worth noting for whoever adds the
      // next appearance string: this list is the mechanism, not a workaround,
      // but it only covers a genuine cognate. An entry here for a word that
      // does have a French translation would hide a real gap.
      'settings.appearance.groups.mode',
      'settings.appearance.groups.palette',
      'settings.appearance.density.compact',
      // A URL placeholder is intentionally identical in every locale.
      'sidebar.externalUrlPlaceholder',
      // "Public" is the same word in French too. Its counterpart
      // visibility.private ("Privé") does differ, which is the check that this
      // is a cognate rather than an untranslated string.
      'visibility.public',
      // "Backlog" is the standard French agile/kanban term too - it is not
      // translated in French project-management tooling generally. Its
      // sibling bucket.active/done/cancelled all differ, which is the check
      // that this is a cognate rather than a missed translation.
      'status.bucket.backlog',
      // "Total" is the same word in French. Its siblings in the same block,
      // projectOverview.health.onTrack and .recentComments, both differ, which
      // is the check that this is a cognate rather than a missed translation.
      'projectOverview.total',
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
