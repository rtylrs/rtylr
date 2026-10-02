import fs from 'node:fs';
import path from 'node:path';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { isPast, isToday, isFuture, todayIso } from './dateUtils';
import { isDueToday, todayIso as todayIsoFromLifeRhythm } from './lifeRhythm';

/**
 * There were two functions called `todayIso`: this one and a UTC copy, three
 * hours apart every night at UTC+3. An import from the wrong module was
 * invisible at the call site, and `isPast` / `isToday` / `isFuture` all read
 * the UTC one, so in that window a task due yesterday was "today" and an
 * overdue task was not past.
 */
describe('todayIso', () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it('is one function, imported from either module', () => {
    expect(todayIsoFromLifeRhythm).toBe(todayIso);
  });

  it('has exactly one implementation in the app', () => {
    // Two identically named exports is what made this easy to miss.
    const lib = path.resolve(__dirname);
    let definitions = 0;
    for (const file of fs.readdirSync(lib)) {
      if (!file.endsWith('.ts') || file.includes('.test.')) continue;
      const source = fs.readFileSync(path.join(lib, file), 'utf8');
      definitions += [...source.matchAll(/^export function todayIso\b/gm)].length;
    }
    expect(definitions).toBe(1);
  });

  it('reports the local day, not the UTC one', () => {
    // 21:30 UTC on the 4th is 00:30 on the 5th at UTC+3. The reader's
    // calendar says the 5th, and a due date is a calendar day where the
    // reader is standing.
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-09-04T21:30:00Z'));

    const local = new Date();
    const expected = `${local.getFullYear()}-${String(local.getMonth() + 1).padStart(2, '0')}-${String(
      local.getDate()
    ).padStart(2, '0')}`;

    expect(todayIso()).toBe(expected);
    // The UTC reading of the same instant, which is what it used to return.
    expect(todayIso()).toBe(
      new Date().toLocaleDateString('en-CA', { year: 'numeric', month: '2-digit', day: '2-digit' })
    );
  });

  it('agrees with isDueToday about the same date', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-09-04T21:30:00Z'));

    const today = todayIso();
    expect(isDueToday(today)).toBe(true);
    expect(isToday(today)).toBe(true);
    expect(isPast(today)).toBe(false);
    expect(isFuture(today)).toBe(false);
  });

  it('calls yesterday past and tomorrow future across the boundary', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-09-04T21:30:00Z'));

    const today = todayIso();
    const [y, m, d] = today.split('-').map(Number);
    const shift = (days: number): string => {
      const dt = new Date(Date.UTC(y, m - 1, d + days));
      return dt.toISOString().slice(0, 10);
    };

    expect(isPast(shift(-1))).toBe(true);
    expect(isToday(shift(-1))).toBe(false);
    expect(isDueToday(shift(-1))).toBe(false);
    expect(isFuture(shift(1))).toBe(true);
  });

  it('pads a single-digit month and day', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-01-05T12:00:00Z'));
    expect(todayIso()).toMatch(/^\d{4}-\d{2}-\d{2}$/);
  });
});
