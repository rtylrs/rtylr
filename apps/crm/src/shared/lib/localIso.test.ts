import { describe, it, expect, afterAll } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';
import { toLocalIso, todayIso } from './dateUtils';
import { isToday, isThisMonth, relativeDate, daysAgo, daysFromNow } from './dateUtils';

/** Run in six zones spanning UTC+14 to UTC-4, not just the one we sit in. */
const ZONES = [
  'Asia/Beirut',
  'UTC',
  'America/New_York',
  'Australia/Sydney',
  'America/Santiago',
  'Pacific/Kiritimati',
];
const ORIGINAL_TZ = process.env.TZ;

afterAll(() => {
  // Assigning undefined sets TZ to the string "undefined", which Node rejects
  // and falls back to UTC, leaving later files in another zone.
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

describe.each(ZONES)('crm dates at %s', (tz) => {
  it('keeps the day a locally built Date was constructed with', () => {
    expect(inZone(tz, () => toLocalIso(new Date(2026, 8, 5)))).toBe('2026-09-05');
    expect(inZone(tz, () => toLocalIso(new Date(2026, 0, 1)))).toBe('2026-01-01');
    expect(inZone(tz, () => toLocalIso(new Date(2026, 11, 31)))).toBe('2026-12-31');
  });

  it('labels a record by the calendar day it fell on', () => {
    const today = inZone(tz, () => todayIso());
    expect(inZone(tz, () => isToday(today))).toBe(true);
    expect(inZone(tz, () => relativeDate(today))).toBe('Today');
    // Parsed as UTC midnight, this read as the previous month west of
    // Greenwich, so a 1 January record was not "this month" in January.
    expect(inZone(tz, () => isThisMonth(today))).toBe(true);
  });

  it('steps whole days without drifting', () => {
    // The reference Date is built inside the zone. Built outside it, it would
    // be this machine's local midnight reinterpreted elsewhere, which is the
    // very mistake these helpers are being tested for.
    expect(inZone(tz, () => daysAgo(0, new Date(2026, 5, 15)))).toBe('2026-06-15');
    expect(inZone(tz, () => daysAgo(10, new Date(2026, 5, 15)))).toBe('2026-06-05');
    expect(inZone(tz, () => daysFromNow(10, new Date(2026, 5, 15)))).toBe('2026-06-25');
    // Across the local spring forward, where a day is 23 hours long.
    expect(inZone(tz, () => daysFromNow(1, new Date(2026, 2, 28)))).toBe('2026-03-29');
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

  it('has no toISOString day extraction left in the app', () => {
    const pattern = /toISOString\(\)\.(?:slice\(0, ?10\)|split\('T'\)\[0\])/;
    const offenders: string[] = [];

    for (const file of sourceFiles(SRC)) {
      const lines = fs.readFileSync(file, 'utf8').split('\n');
      lines.forEach((line, i) => {
        const trimmed = line.trim();
        // A comment naming the wrong idiom is not a use of it.
        if (trimmed.startsWith('*') || trimmed.startsWith('//') || trimmed.startsWith('/*')) return;
        if (pattern.test(line)) offenders.push(`${path.relative(SRC, file)}:${i + 1}`);
      });
    }

    expect(offenders, 'use toLocalIso(date) or todayIso() instead').toEqual([]);
  });
});
