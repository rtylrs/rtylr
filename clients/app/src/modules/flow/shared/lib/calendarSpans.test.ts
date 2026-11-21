import { describe, expect, it } from 'vitest';
import {
  daysInRange,
  hiddenPerDay,
  isUnscheduled,
  isoDay,
  layoutWeek,
  moveRange,
  reschedule,
  resizeRange,
  taskRange,
  weekCells,
} from './calendarSpans';

describe('isoDay', () => {
  it('takes the day part of a stored date', () => {
    expect(isoDay('2026-11-19')).toBe('2026-11-19');
    expect(isoDay('2026-11-19T00:00:00Z')).toBe('2026-11-19');
  });

  it('answers null for nothing or for something that is not a date', () => {
    expect(isoDay(undefined)).toBeNull();
    expect(isoDay('')).toBeNull();
    expect(isoDay('soon')).toBeNull();
  });
});

describe('taskRange', () => {
  it('runs from the start date to the due date', () => {
    expect(taskRange({ startDate: '2026-11-02', dueDate: '2026-11-05' })).toEqual({
      start: '2026-11-02',
      end: '2026-11-05',
    });
  });

  it('is one day for a due date alone or a start date alone', () => {
    expect(taskRange({ dueDate: '2026-11-05' })).toEqual({
      start: '2026-11-05',
      end: '2026-11-05',
    });
    expect(taskRange({ startDate: '2026-11-02' })).toEqual({
      start: '2026-11-02',
      end: '2026-11-02',
    });
  });

  it('keeps the due date when the start is after it', () => {
    expect(taskRange({ startDate: '2026-11-09', dueDate: '2026-11-05' })).toEqual({
      start: '2026-11-05',
      end: '2026-11-05',
    });
  });

  it('is null with no dates, which is what the Unscheduled panel lists', () => {
    expect(taskRange({})).toBeNull();
    expect(isUnscheduled({ dueDate: null, startDate: '' })).toBe(true);
    expect(isUnscheduled({ dueDate: '2026-11-05' })).toBe(false);
  });
});

describe('weekCells', () => {
  it('is the Sunday-first week holding the day, every cell in range', () => {
    const cells = weekCells(new Date(2026, 10, 19));
    expect(cells.map((cell) => cell.iso)).toEqual([
      '2026-11-15',
      '2026-11-16',
      '2026-11-17',
      '2026-11-18',
      '2026-11-19',
      '2026-11-20',
      '2026-11-21',
    ]);
    expect(cells.every((cell) => cell.inMonth)).toBe(true);
  });
});

describe('layoutWeek', () => {
  const week = '2026-11-15';

  it('clips a bar to the week and records which edges are real', () => {
    const [segment] = layoutWeek(week, [
      { item: 'a', range: { start: '2026-11-10', end: '2026-11-17' } },
    ]);
    expect(segment).toMatchObject({ col: 0, span: 3, lane: 0, startsHere: false, endsHere: true });
  });

  it('stacks overlapping bars and reuses a lane once it is free', () => {
    const segments = layoutWeek(week, [
      { item: 'long', range: { start: '2026-11-15', end: '2026-11-18' } },
      { item: 'overlap', range: { start: '2026-11-16', end: '2026-11-16' } },
      { item: 'after', range: { start: '2026-11-19', end: '2026-11-20' } },
    ]);
    const lane = (item: string) => segments.find((s) => s.item === item)?.lane;
    expect(lane('long')).toBe(0);
    expect(lane('overlap')).toBe(1);
    expect(lane('after')).toBe(0);
  });

  it('leaves out bars outside the week', () => {
    expect(
      layoutWeek(week, [{ item: 'x', range: { start: '2026-11-22', end: '2026-11-23' } }])
    ).toHaveLength(0);
  });

  it('counts the bars past the lane limit per day', () => {
    const segments = layoutWeek(
      week,
      Array.from({ length: 4 }, (_, index) => ({
        item: index,
        range: { start: '2026-11-16', end: '2026-11-16' },
      }))
    );
    expect(hiddenPerDay(segments, 3)).toEqual([0, 1, 0, 0, 0, 0, 0]);
  });
});

describe('ranges', () => {
  const range = { start: '2026-11-10', end: '2026-11-12' };

  it('stops an edge at the other edge rather than turning the bar inside out', () => {
    expect(resizeRange(range, 'start', '2026-11-14')).toEqual({
      start: '2026-11-12',
      end: '2026-11-12',
    });
    expect(resizeRange(range, 'end', '2026-11-01')).toEqual({
      start: '2026-11-10',
      end: '2026-11-10',
    });
  });

  it('moves a range whole, across a month end', () => {
    expect(moveRange(range, '2026-11-29')).toEqual({ start: '2026-11-29', end: '2026-12-01' });
  });

  it('lists the days of a range inside a window', () => {
    expect(daysInRange(range, '2026-11-11', '2026-11-30')).toEqual(['2026-11-11', '2026-11-12']);
  });
});

describe('reschedule', () => {
  it('gives an unscheduled task the day as its due date', () => {
    expect(reschedule({}, 'schedule', '2026-11-20')).toEqual({ dueDate: '2026-11-20' });
  });

  it('moves a due-only task without inventing a start date', () => {
    expect(reschedule({ dueDate: '2026-11-10' }, 'move', '2026-11-13')).toEqual({
      dueDate: '2026-11-13',
    });
  });

  it('moves a range by how far the grabbed day travelled, keeping its length', () => {
    expect(
      reschedule(
        { startDate: '2026-11-10', dueDate: '2026-11-12' },
        'move',
        '2026-11-14',
        '2026-11-11'
      )
    ).toEqual({ startDate: '2026-11-13', dueDate: '2026-11-15' });
  });

  it('moves a start-only task without inventing a due date', () => {
    expect(reschedule({ startDate: '2026-11-10' }, 'move', '2026-11-11')).toEqual({
      startDate: '2026-11-11',
    });
  });

  it('turns a one-day task into a range when its end edge is dragged out', () => {
    expect(reschedule({ dueDate: '2026-11-10' }, 'end', '2026-11-12')).toEqual({
      startDate: '2026-11-10',
      dueDate: '2026-11-12',
    });
  });

  it('writes only the start when the start edge of a range moves', () => {
    expect(
      reschedule({ startDate: '2026-11-10', dueDate: '2026-11-12' }, 'start', '2026-11-08')
    ).toEqual({ startDate: '2026-11-08' });
  });

  it('fills the due date when the start edge of a start-only task moves', () => {
    expect(reschedule({ startDate: '2026-11-10' }, 'start', '2026-11-08')).toEqual({
      startDate: '2026-11-08',
      dueDate: '2026-11-10',
    });
  });

  it('answers null for a drop that changes nothing', () => {
    expect(reschedule({ dueDate: '2026-11-10' }, 'move', '2026-11-10')).toBeNull();
    expect(reschedule({ dueDate: '2026-11-10' }, 'schedule', '2026-11-10')).toBeNull();
    expect(
      reschedule({ startDate: '2026-11-10', dueDate: '2026-11-12' }, 'end', '2026-11-12')
    ).toBeNull();
  });

  it('repairs a start after the due date when an edge is dragged', () => {
    expect(
      reschedule({ startDate: '2026-11-15', dueDate: '2026-11-10' }, 'end', '2026-11-12')
    ).toEqual({ startDate: '2026-11-10', dueDate: '2026-11-12' });
  });
});
