import { describe, it, expect } from 'vitest';
import {
  parseTimeToMinutes,
  netWorkingMinutes,
  minutesToDecimalHours,
  totalHoursForEntries,
  isEditable,
  isPendingReview,
  timesheetStatusLabel,
  groupEntriesByDate,
} from './timesheetUtils';
import type { TimeEntry } from './timesheetUtils';

describe('parseTimeToMinutes', () => {
  it('converts 08:00 to 480 minutes', () => {
    expect(parseTimeToMinutes('08:00')).toBe(480);
  });

  it('converts 00:00 to 0', () => {
    expect(parseTimeToMinutes('00:00')).toBe(0);
  });

  it('converts 17:30 to 1050 minutes', () => {
    expect(parseTimeToMinutes('17:30')).toBe(1050);
  });
});

describe('netWorkingMinutes', () => {
  it('calculates net minutes after break', () => {
    const entry: TimeEntry = {
      date: '2025-01-01',
      startTime: '09:00',
      endTime: '17:00',
      breakMinutes: 30,
    };
    expect(netWorkingMinutes(entry)).toBe(450);
  });

  it('returns 0 when end is before start', () => {
    const entry: TimeEntry = {
      date: '2025-01-01',
      startTime: '17:00',
      endTime: '09:00',
      breakMinutes: 0,
    };
    expect(netWorkingMinutes(entry)).toBe(0);
  });
});

describe('minutesToDecimalHours', () => {
  it('converts 90 minutes to 1.5 hours', () => {
    expect(minutesToDecimalHours(90)).toBe(1.5);
  });

  it('converts 480 minutes to 8 hours', () => {
    expect(minutesToDecimalHours(480)).toBe(8);
  });
});

describe('totalHoursForEntries', () => {
  it('sums hours across multiple entries', () => {
    const entries: TimeEntry[] = [
      { date: '2025-01-01', startTime: '09:00', endTime: '17:00', breakMinutes: 0 },
      { date: '2025-01-02', startTime: '09:00', endTime: '13:00', breakMinutes: 0 },
    ];
    expect(totalHoursForEntries(entries)).toBe(12);
  });

  it('returns 0 for empty entries', () => {
    expect(totalHoursForEntries([])).toBe(0);
  });
});

describe('isEditable', () => {
  it('returns true for draft', () => {
    expect(isEditable('draft')).toBe(true);
  });

  it('returns true for rejected', () => {
    expect(isEditable('rejected')).toBe(true);
  });

  it('returns false for submitted', () => {
    expect(isEditable('submitted')).toBe(false);
  });

  it('returns false for approved', () => {
    expect(isEditable('approved')).toBe(false);
  });
});

describe('isPendingReview', () => {
  it('returns true for submitted', () => {
    expect(isPendingReview('submitted')).toBe(true);
  });

  it('returns false for draft', () => {
    expect(isPendingReview('draft')).toBe(false);
  });
});

describe('timesheetStatusLabel', () => {
  it('returns Draft for draft', () => {
    expect(timesheetStatusLabel('draft')).toBe('Draft');
  });

  it('returns Pending Review for submitted', () => {
    expect(timesheetStatusLabel('submitted')).toBe('Pending Review');
  });
});

describe('groupEntriesByDate', () => {
  it('groups entries by their date field', () => {
    const entries: TimeEntry[] = [
      { date: '2025-01-01', startTime: '09:00', endTime: '17:00', breakMinutes: 30 },
      { date: '2025-01-02', startTime: '09:00', endTime: '17:00', breakMinutes: 30 },
      { date: '2025-01-01', startTime: '18:00', endTime: '20:00', breakMinutes: 0 },
    ];
    const result = groupEntriesByDate(entries);
    expect(Object.keys(result)).toHaveLength(2);
    expect(result['2025-01-01']).toHaveLength(2);
  });

  it('returns empty object for empty entries', () => {
    expect(groupEntriesByDate([])).toEqual({});
  });
});
