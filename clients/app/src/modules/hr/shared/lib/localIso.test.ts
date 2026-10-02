import { describe, it, expect, afterAll } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';
import { toLocalIso, todayIso } from './dateUtils';
import { addDays, isWeekend, isToday } from './hrDateUtils';

/**
 * HR's date bugs pointed in both directions at once, so one pinned zone is not
 * enough to catch them: `isWeekend` was wrong only west of Greenwich, the
 * calendar was wrong only east of it, and `addDays` was wrong only on the day
 * the clocks go forward. Every assertion here therefore runs in each zone.
 *
 * Before this, the suite was green purely because everyone ran it at UTC+3.
 * The same assertions failed in New York against the old implementation.
 */
const ZONES = ['Asia/Beirut', 'UTC', 'America/New_York', 'Australia/Sydney'];
const ORIGINAL_TZ = process.env.TZ;

afterAll(() => {
  // Assigning undefined would set TZ to the string "undefined", which Node
  // rejects and falls back to UTC, leaving later files in another zone.
  if (ORIGINAL_TZ === undefined) delete process.env.TZ;
  else process.env.TZ = ORIGINAL_TZ;
});

function inZone<T>(tz: string, run: () => T): T {
  process.env.TZ = tz;
  try {
    return run();
  } finally {
    if (ORIGINAL_TZ === undefined) delete process.env.TZ;
    else process.env.TZ = ORIGINAL_TZ;
  }
}

describe.each(ZONES)('date handling at %s', (tz) => {
  it('keeps the day a locally built Date was constructed with', () => {
    // How the leave calendar builds its cells.
    expect(inZone(tz, () => toLocalIso(new Date(2026, 8, 5)))).toBe('2026-09-05');
    expect(inZone(tz, () => toLocalIso(new Date(2026, 0, 1)))).toBe('2026-01-01');
    expect(inZone(tz, () => toLocalIso(new Date(2026, 11, 31)))).toBe('2026-12-31');
  });

  it('agrees with the local clock about today', () => {
    const [today, expected] = inZone(tz, () => {
      const now = new Date();
      return [
        todayIso(),
        `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`,
      ];
    });
    expect(today).toBe(expected);
    expect(inZone(tz, () => isToday(today))).toBe(true);
  });

  it('knows which days are the weekend', () => {
    // 5 September 2026 is a Saturday, 6th a Sunday, 7th a Monday. The local
    // getter reported the 5th as a Friday west of Greenwich.
    expect(inZone(tz, () => isWeekend('2026-09-05'))).toBe(true);
    expect(inZone(tz, () => isWeekend('2026-09-06'))).toBe(true);
    expect(inZone(tz, () => isWeekend('2026-09-07'))).toBe(false);
  });

  it('advances a day across a clock change', () => {
    // Spring forward: Beirut and Paris on 2026-03-29, New York on 2026-03-08,
    // Sydney falls back on 2026-04-05. With local setters, addDays returned
    // the same date it was given on the forward day and lost a day outright.
    for (const start of [
      '2026-03-07',
      '2026-03-08',
      '2026-03-28',
      '2026-03-29',
      '2026-04-04',
      '2026-10-31',
    ]) {
      const [y, m, d] = start.split('-').map(Number);
      const expected = new Date(Date.UTC(y, m - 1, d + 1)).toISOString().slice(0, 10);
      expect(
        inZone(tz, () => addDays(start, 1)),
        `addDays('${start}', 1) at ${tz}`
      ).toBe(expected);
    }
  });
});

describe('no local Date is serialised through UTC', () => {
  const SRC = path.resolve(__dirname, '../..');

  function sourceFiles(dir: string, into: string[] = []): string[] {
    for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
      const full = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        if (entry.name === 'node_modules' || entry.name === 'locales') continue;
        sourceFiles(full, into);
        continue;
      }
      if (!/\.tsx?$/.test(entry.name)) continue;
      if (/\.test\.tsx?$/.test(entry.name)) continue;
      into.push(full);
    }
    return into;
  }

  it('only takes the UTC day from a Date read through UTC getters', () => {
    const pattern = /toISOString\(\)\.(?:slice\(0, ?10\)|split\('T'\)\[0\])/;
    const offenders: string[] = [];

    for (const file of sourceFiles(SRC)) {
      const lines = fs.readFileSync(file, 'utf8').split('\n');
      lines.forEach((line, i) => {
        const trimmed = line.trim();
        // A comment naming the wrong idiom is not a use of it.
        if (trimmed.startsWith('*') || trimmed.startsWith('/hr/') || trimmed.startsWith('/*'))
          return;
        if (!pattern.test(line)) return;
        const window = lines.slice(Math.max(0, i - 6), i + 1).join('\n');
        const utcOnly = /Date\.UTC\(|setUTC|getUTC/.test(window);
        if (!utcOnly) offenders.push(`${path.relative(SRC, file)}:${i + 1}`);
      });
    }

    expect(offenders, 'use toLocalIso(date) or todayIso() instead').toEqual([]);
  });
});
