import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { monthCells, tasksByLocalDay, agendaDays } from './monthGrid';

/**
 * The two calendars each built their own month grid and the two disagreed on
 * which day a task belongs to.
 *
 * The global calendar keyed cells with `toLocalIso` and bucketed on
 * `dueDate.split('T')[0]`, both the reader's own calendar day. The project
 * calendar keyed cells from a local Date but bucketed on
 * `new Date(dueDate).toDateString()`, which parses a bare `YYYY-MM-DD` as UTC
 * midnight and reads it back locally. West of Greenwich that lands on the day
 * before, so every task on that screen sat on the wrong day.
 *
 * It read as correct from Beirut, which is why the suite never said otherwise
 * (rtylr#1192).
 */
describe('monthCells', () => {
  it('pads to whole Sunday-first weeks', () => {
    const cells = monthCells(new Date(2026, 8, 15)); // September 2026
    expect(cells.length % 7).toBe(0);
    expect(cells[0].date.getDay()).toBe(0);
    expect(cells[cells.length - 1].date.getDay()).toBe(6);
  });

  it('marks the padding days as outside the month', () => {
    // 1 September 2026 is a Tuesday, so two leading days belong to August.
    const cells = monthCells(new Date(2026, 8, 1));
    expect(cells.slice(0, 2).every((c) => !c.inMonth)).toBe(true);
    expect(cells[2].inMonth).toBe(true);
    expect(cells[2].iso).toBe('2026-09-01');
    expect(cells.filter((c) => c.inMonth)).toHaveLength(30);
  });

  it('keys every cell by the local calendar day, never the UTC one', () => {
    // The specific defect. A cell built as local midnight has a UTC date one day
    // earlier for any positive offset, so `toISOString().slice(0, 10)` names the
    // wrong day for the whole day rather than only near midnight.
    for (const cell of monthCells(new Date(2026, 8, 15))) {
      const localParts = [
        cell.date.getFullYear(),
        String(cell.date.getMonth() + 1).padStart(2, '0'),
        String(cell.date.getDate()).padStart(2, '0'),
      ].join('-');
      expect(cell.iso).toBe(localParts);
    }
  });
});

describe('tasksByLocalDay', () => {
  it('buckets on the stored date part rather than parsing it', () => {
    const byDay = tasksByLocalDay([
      { id: 'a', dueDate: '2026-09-08' },
      { id: 'b', dueDate: '2026-09-08T14:30:00Z' },
      { id: 'c', dueDate: '2026-09-09' },
    ]);
    expect(byDay.get('2026-09-08')?.map((t) => t.id)).toEqual(['a', 'b']);
    expect(byDay.get('2026-09-09')?.map((t) => t.id)).toEqual(['c']);
  });

  it('agrees with the cell key for a date-only due date', () => {
    const cells = monthCells(new Date(2026, 8, 15));
    const byDay = tasksByLocalDay([{ id: 'a', dueDate: '2026-09-08' }]);
    const cell = cells.find((c) => c.inMonth && c.date.getDate() === 8);

    expect(cell).toBeDefined();
    expect(byDay.get(cell!.iso)).toHaveLength(1);
  });

  it('ignores a task with no due date rather than bucketing it under undefined', () => {
    expect(tasksByLocalDay([{ id: 'a' }, { id: 'b', dueDate: null }]).size).toBe(0);
  });
});

describe('agendaDays', () => {
  it('lists only days that have tasks, in order', () => {
    const cells = monthCells(new Date(2026, 8, 15));
    const byDay = tasksByLocalDay([
      { id: 'late', dueDate: '2026-09-20' },
      { id: 'early', dueDate: '2026-09-03' },
    ]);
    const days = agendaDays(cells, byDay);

    expect(days.map((d) => d.cell.iso)).toEqual(['2026-09-03', '2026-09-20']);
  });

  it('leaves out the padding days, so a neighbouring month cannot appear', () => {
    const cells = monthCells(new Date(2026, 8, 1));
    // 30 August is a padding cell in this grid.
    const byDay = tasksByLocalDay([{ id: 'x', dueDate: '2026-08-30' }]);
    expect(agendaDays(cells, byDay)).toEqual([]);
  });
});

describe('west of Greenwich, where the old idiom lost a day', () => {
  /**
   * Pinned to a western zone rather than inherited. From Beirut, which is where
   * this was written, the broken idiom and the correct one agree, so an
   * unpinned version of this file would pass while every task on the project
   * calendar sat on the wrong day for users in the Americas.
   *
   * The same blind spot as rtylr#1192, and the reason that issue existed.
   */
  const ORIGINAL_TZ = process.env.TZ;
  beforeAll(() => {
    process.env.TZ = 'America/New_York';
  });
  afterAll(() => {
    // Assigning undefined sets TZ to the string "undefined", which Node rejects
    // and falls back to UTC, silently leaving later files in another zone.
    if (ORIGINAL_TZ === undefined) delete process.env.TZ;
    else process.env.TZ = ORIGINAL_TZ;
  });

  it('shows the idiom the project calendar used naming the day before', () => {
    // `new Date('2026-09-08')` is UTC midnight, which is 20:00 on the 7th in
    // New York. This is the assertion that would have caught it.
    expect(new Date('2026-09-08').toDateString()).toBe('Mon Sep 07 2026');
    expect(new Date(2026, 8, 8).toDateString()).toBe('Tue Sep 08 2026');
  });

  it('puts the task on the day it is due', () => {
    const cells = monthCells(new Date(2026, 8, 15));
    const byDay = tasksByLocalDay([{ id: 'a', dueDate: '2026-09-08' }]);
    const eighth = cells.find((c) => c.inMonth && c.date.getDate() === 8);

    expect(eighth?.iso).toBe('2026-09-08');
    expect(byDay.get(eighth!.iso)).toHaveLength(1);
    // And nothing landed on the 7th, which is where it used to go.
    const seventh = cells.find((c) => c.inMonth && c.date.getDate() === 7);
    expect(byDay.get(seventh!.iso)).toBeUndefined();
  });
});
