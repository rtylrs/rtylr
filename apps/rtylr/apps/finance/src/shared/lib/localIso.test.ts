import { describe, it, expect, afterAll } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';
import { toLocalIso, todayIso, daysBetween, formatDate, addDays } from './date';

/**
 * Finance dates decide what a customer is charged, so these run in six zones
 * spanning UTC+14 to UTC-4 rather than in whichever one the developer sits in.
 * Every assertion below failed in at least one of them before this change.
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

describe.each(ZONES)('finance dates at %s', (tz) => {
  it('keeps the day a locally built Date was constructed with', () => {
    expect(inZone(tz, () => toLocalIso(new Date(2026, 8, 5)))).toBe('2026-09-05');
    expect(inZone(tz, () => toLocalIso(new Date(2026, 0, 1)))).toBe('2026-01-01');
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
    // Parsed as UTC midnight and formatted locally, this read 14 January
    // anywhere west of Greenwich.
    expect(inZone(tz, () => formatDate('2026-01-15'))).toContain('15');
    expect(inZone(tz, () => formatDate('2026-01-15'))).toContain('2026');
    expect(inZone(tz, () => toLocalIso(addDays('2026-01-15', 1)))).toBe('2026-01-16');
  });

  it('counts calendar days across a clock change', () => {
    // Sydney's 2026-04-05 is 25 hours long and New York's 2026-03-08 is 23.
    // Dividing raw milliseconds and flooring made a ten day gap measure eleven.
    expect(inZone(tz, () => daysBetween('2026-04-01', '2026-04-11'))).toBe(10);
    expect(inZone(tz, () => daysBetween('2026-04-11', '2026-04-01'))).toBe(-10);
    expect(inZone(tz, () => daysBetween('2026-03-01', '2026-03-31'))).toBe(30);
    expect(inZone(tz, () => daysBetween('2026-10-25', '2026-11-05'))).toBe(11);
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
