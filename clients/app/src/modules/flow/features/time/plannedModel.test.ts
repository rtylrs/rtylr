import { describe, expect, it } from 'vitest';
import { buildPlannedRows, cellTone, formatHours } from './plannedModel';

describe('cellTone', () => {
  it('flags under and over the target, with an hour of slack above it', () => {
    expect(cellTone(120, 300, true)).toBe('under');
    expect(cellTone(300, 300, true)).toBe('ok');
    expect(cellTone(360, 300, true)).toBe('ok');
    expect(cellTone(361, 300, true)).toBe('over');
    expect(cellTone(0, 300, false)).toBe('off');
    expect(cellTone(30, 300, false)).toBe('ok');
  });
});

describe('buildPlannedRows', () => {
  const dates = ['2026-11-29', '2026-11-30'];
  const team = [
    { id: '7', name: 'Zed' },
    { id: '8', name: 'Amal' },
    { id: '9', name: 'Hidden' },
  ];
  const response = {
    days: [{ userId: '7', date: '2026-11-30', plannedMinutes: 330, taskIds: ['41', '42'] }],
    people: ['7', '8'],
    everyone: false,
  };

  it('shows only the people the caller may see, with zero days as gaps', () => {
    const rows = buildPlannedRows(response, team, dates, 'mon_fri', 300);
    expect(rows.map((r) => r.name)).toEqual(['Amal', 'Zed']);
    const zed = rows[1];
    expect(zed.totalMinutes).toBe(330);
    // Sunday is off on a Monday week; Monday is on target.
    expect(zed.cells.map((c) => c.tone)).toEqual(['off', 'ok']);
    expect(zed.cells[1].taskIds).toEqual(['41', '42']);
    expect(rows[0].cells.map((c) => c.tone)).toEqual(['off', 'under']);
  });

  it('treats Sunday as a working day on a Sunday week, and lists everyone when allowed', () => {
    const rows = buildPlannedRows({ ...response, everyone: true }, team, dates, 'sun_thu', 300);
    expect(rows).toHaveLength(3);
    expect(rows.find((r) => r.name === 'Amal')?.cells[0].tone).toBe('under');
  });

  it('names people missing from the team list from their tasks', () => {
    const rows = buildPlannedRows(
      {
        days: [{ userId: '3', date: '2026-11-30', plannedMinutes: 60, taskIds: ['1'] }],
        tasks: [{ assigneeId: '3', assigneeName: 'Flow QA' }],
        everyone: true,
      },
      [],
      dates,
      'mon_fri',
      300
    );
    expect(rows.map((r) => r.name)).toEqual(['Flow QA']);
  });

  it('formats hours with one decimal', () => {
    expect(formatHours(330)).toBe('5.5');
    expect(formatHours(50)).toBe('0.8');
  });
});
