import { describe, it, expect, afterAll } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';
import { toIsoDate, todayIso } from './dateUtils';
import { formatDate } from './format';
import { aggregateByPeriod } from './reportUtils';

/**
 * Run in six zones spanning UTC+14 to UTC-4 rather than in whichever one the
 * developer sits in. Before this change the app's own suite passed in two of
 * them.
 */
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

describe.each(ZONES)('erp dates at %s', (tz) => {
  it('keeps the day a locally built Date was constructed with', () => {
    expect(inZone(tz, () => toIsoDate(new Date(2026, 8, 5)))).toBe('2026-09-05');
    expect(inZone(tz, () => toIsoDate(new Date(2026, 0, 1)))).toBe('2026-01-01');
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

  it('renders a calendar day as the day it names', () => {
    expect(inZone(tz, () => formatDate('2026-01-15'))).toContain('15');
    expect(inZone(tz, () => formatDate('2026-01-15'))).toContain('2026');
  });

  it('groups a new year row under the new year', () => {
    // getPeriodKey used local getters on a UTC parsed date, so a 1 January row
    // was filed under the previous year west of Greenwich.
    const points = [{ label: 'new year', value: 10, date: '2026-01-01' }];
    expect(inZone(tz, () => aggregateByPeriod(points, 'yearly')[0].label)).toBe('2026');
    expect(inZone(tz, () => aggregateByPeriod(points, 'quarterly')[0].label)).toBe('2026-Q1');
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
        if (trimmed.startsWith('*') || trimmed.startsWith('//') || trimmed.startsWith('/*')) return;
        if (!pattern.test(line)) return;
        const window = lines.slice(Math.max(0, i - 6), i + 1).join('\n');
        if (!/Date\.UTC\(|setUTC|getUTC/.test(window))
          offenders.push(`${path.relative(SRC, file)}:${i + 1}`);
      });
    }

    expect(offenders, 'use toIsoDate(date) or todayIso() instead').toEqual([]);
  });
});
