import path from 'node:path';
import { describe, expect, it } from 'vitest';

import { collectTranslatedKeys, formatTranslatedKeys } from './test/translatedKeys';

/**
 * No list may use a translated string as its React key.
 *
 * Switching language then changes every key, so React remounts the whole list:
 * cards replay their entrance animation and any state inside a row is lost. It
 * is invisible to every other gate, because happy-dom runs no animations and no
 * test switches language mid-render (rtylr#1324).
 *
 * The same family as `data-tour={label.toLowerCase()}` in rtylr#1316: an
 * identifier taken from user-facing text, correct until that text is translated.
 */
const APPS = [
  'auth',
  'crm',
  'dash',
  'erp',
  'finance',
  'flow',
  'hr',
  'insights',
  'menu',
  'recruit',
] as const;

/**
 * Known and not yet fixed.
 *
 * All seven are in `apps/hr`, which is owned by another developer, so they are
 * recorded rather than changed. Every other app is at zero. Anything new fails
 * this suite; these shrink as hr's owner takes them.
 */
const KNOWN: readonly string[] = [
  'apps/hr/src/features/dashboard/Dashboard.tsx:427',
  'apps/hr/src/features/dashboard/Dashboard.tsx:793',
  'apps/hr/src/features/dashboard/widgets/HeadcountWidget.tsx:55',
  'apps/hr/src/features/dashboard/widgets/TeamAttendanceTodayWidget.tsx:54',
  'apps/hr/src/features/payroll/Payroll.tsx:130',
  'apps/hr/src/features/payroll/PayrollRunDetail.tsx:204',
  'apps/hr/src/features/shifts/Shifts.tsx:149',
];

describe('no list keys off a translated label', () => {
  const root = path.resolve(__dirname);

  it.each(APPS)('%s', (app) => {
    const found = formatTranslatedKeys(
      collectTranslatedKeys(path.resolve(root, 'apps', app, 'src')),
      root
    );
    expect(found.filter((f) => !KNOWN.includes(f))).toEqual([]);
  });

  it('every known entry still exists, so the list cannot rot silently', () => {
    const all = APPS.flatMap((app) =>
      formatTranslatedKeys(collectTranslatedKeys(path.resolve(root, 'apps', app, 'src')), root)
    );
    expect(KNOWN.filter((k) => !all.includes(k))).toEqual([]);
  });
});
