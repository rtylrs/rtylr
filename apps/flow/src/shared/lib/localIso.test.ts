import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import fs from 'node:fs';
import path from 'node:path';
import { toLocalIso, todayIso, addDays, startOfWeek } from './dateUtils';

/**
 * These assertions only bite east of Greenwich. On a UTC runner the broken
 * idiom and the correct one agree, so the zone is pinned here rather than
 * inherited: a CI box set to UTC would otherwise pass this file while shipping
 * the bug to every user in the product's home market.
 */
const ORIGINAL_TZ = process.env.TZ;
beforeAll(() => {
  process.env.TZ = 'Asia/Beirut';
});
afterAll(() => {
  // Assigning an undefined value would set TZ to the string "undefined", which
  // Node rejects and falls back to UTC, silently leaving every later file in a
  // different zone than it started in.
  if (ORIGINAL_TZ === undefined) delete process.env.TZ;
  else process.env.TZ = ORIGINAL_TZ;
});

describe('a calendar day is the day where the reader is standing', () => {
  it('keeps the day a locally built Date was constructed with', () => {
    // The calendar grid builds cells exactly this way. At UTC+3 this Date is
    // 2026-09-04T21:00Z, so toISOString() reports the 4th all day long.
    expect(toLocalIso(new Date(2026, 8, 5))).toBe('2026-09-05');
    expect(new Date(2026, 8, 5).toISOString().slice(0, 10)).toBe('2026-09-04');
  });

  it('holds at both ends of the year, where a slip changes the year too', () => {
    expect(toLocalIso(new Date(2026, 0, 1))).toBe('2026-01-01');
    expect(toLocalIso(new Date(2026, 11, 31))).toBe('2026-12-31');
  });

  it('pads single digit months and days', () => {
    expect(toLocalIso(new Date(2026, 2, 7))).toBe('2026-03-07');
  });

  it('reports today as the local day', () => {
    const now = new Date();
    const expected = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`;
    expect(todayIso()).toBe(expected);
  });

  it('leaves the UTC anchored string arithmetic alone', () => {
    // These take an ISO string and return one, never touching the local
    // offset. They were correct before this change and must stay that way:
    // routing them through toLocalIso would introduce the very bug above.
    expect(addDays('2026-09-05', 1)).toBe('2026-09-06');
    expect(addDays('2026-12-31', 1)).toBe('2027-01-01');
    expect(addDays('2026-03-01', -1)).toBe('2026-02-28');
    expect(startOfWeek('2026-09-05')).toBe('2026-08-31');
    expect(startOfWeek('2026-09-05', 'sunday')).toBe('2026-08-30');
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

  it('only takes the UTC day from a Date that was anchored to UTC', () => {
    const pattern = /toISOString\(\)\.(?:slice\(0, ?10\)|split\('T'\)\[0\])/;
    const offenders: string[] = [];

    for (const file of sourceFiles(SRC)) {
      const lines = fs.readFileSync(file, 'utf8').split('\n');
      lines.forEach((line, i) => {
        const trimmed = line.trim();
        // Comments describing the wrong idiom are not uses of it. Without this
        // the doc comment on toLocalIso, which names the pattern it replaces,
        // reports itself.
        if (trimmed.startsWith('*') || trimmed.startsWith('//') || trimmed.startsWith('/*')) return;
        if (!pattern.test(line)) return;
        // Legitimate use: the Date was built with Date.UTC(...) and is read
        // back through UTC getters, a few lines above.
        const anchored = lines
          .slice(Math.max(0, i - 4), i + 1)
          .some((l) => l.includes('Date.UTC('));
        if (!anchored) offenders.push(`${path.relative(SRC, file)}:${i + 1}`);
      });
    }

    expect(offenders, 'use toLocalIso(date) or todayIso() instead').toEqual([]);
  });
});
