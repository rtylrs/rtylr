import { describe, it, expect, beforeAll } from 'vitest';
import i18next from 'i18next';
import toasts from './ar/toasts.json';
import archive from './ar/archive.json';
import workspaceManager from './ar/workspaceManager.json';

// Arabic has six Intl.PluralRules categories (zero/one/two/few/many/other).
// "other" is only reachable for exact multiples of 100 or non-integer counts,
// so a resource that defines only _one/_other silently fails to render the
// count for the whole 2-99 range: i18next.exists() returns false and t()
// falls back to the bare key instead of interpolating text. These counts
// exercise every category boundary.
const COUNTS_BY_CATEGORY: Array<[string, number]> = [
  ['zero', 0],
  ['one', 1],
  ['two', 2],
  ['few', 3],
  ['many', 15],
  ['other', 100],
];

const PLURALIZED_KEYS = [
  'toasts:task.bulkUpdated',
  'toasts:task.bulkDeleted',
  'toasts:weeklyReview.streak',
  'toasts:weeklyReview.alreadyLoggedStreak',
  'archive:count',
  'workspaceManager:memberCount',
];

const CASES: Array<[string, string, number]> = PLURALIZED_KEYS.flatMap((key) =>
  COUNTS_BY_CATEGORY.map(([category, count]): [string, string, number] => [key, category, count])
);

describe('ar locale pluralization', () => {
  beforeAll(async () => {
    await i18next.init({
      lng: 'ar',
      resources: { ar: { toasts, archive, workspaceManager } },
      interpolation: { escapeValue: false },
    });
  });

  it.each(CASES)('resolves %s for the %s category (count=%i)', (key, _category, count) => {
    expect(i18next.exists(key, { count })).toBe(true);
    expect(i18next.t(key, { count })).not.toContain('{{count}}');
  });
});
