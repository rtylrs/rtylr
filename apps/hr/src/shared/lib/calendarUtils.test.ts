import { describe, expect, it } from 'vitest';
import {
  eventDurationMinutes,
  eventTypeColor,
  eventTypeLabel,
  eventsOnDate,
  eventsOverlap,
  isMultiDayEvent,
  isPublicHoliday,
  recurrenceLabel,
  workingDaysInMonth,
} from './calendarUtils';

describe('eventTypeLabel', () => {
  it('returns correct label for each type', () => {
    expect(eventTypeLabel('Meeting')).toBe('Meeting');
    expect(eventTypeLabel('Holiday')).toBe('Holiday');
    expect(eventTypeLabel('Training')).toBe('Training');
    expect(eventTypeLabel('Review')).toBe('Review');
    expect(eventTypeLabel('Other')).toBe('Other');
  });
});

describe('eventTypeColor', () => {
  it('returns correct color for each type', () => {
    expect(eventTypeColor('Meeting')).toBe('blue');
    expect(eventTypeColor('Holiday')).toBe('green');
    expect(eventTypeColor('Training')).toBe('purple');
    expect(eventTypeColor('Review')).toBe('orange');
    expect(eventTypeColor('Other')).toBe('gray');
  });
});

describe('recurrenceLabel', () => {
  it('returns correct label for each rule', () => {
    expect(recurrenceLabel('once')).toBe('One Time');
    expect(recurrenceLabel('daily')).toBe('Daily');
    expect(recurrenceLabel('weekly')).toBe('Weekly');
    expect(recurrenceLabel('monthly')).toBe('Monthly');
    expect(recurrenceLabel('yearly')).toBe('Yearly');
  });
});

describe('eventsOverlap', () => {
  it('returns true when events overlap', () => {
    expect(
      eventsOverlap('2024-01-10T09:00', '2024-01-10T11:00', '2024-01-10T10:00', '2024-01-10T12:00')
    ).toBe(true);
  });

  it('returns false when events are back-to-back', () => {
    expect(
      eventsOverlap('2024-01-10T09:00', '2024-01-10T10:00', '2024-01-10T10:00', '2024-01-10T11:00')
    ).toBe(false);
  });

  it('returns false when events do not overlap', () => {
    expect(
      eventsOverlap('2024-01-10T09:00', '2024-01-10T10:00', '2024-01-10T11:00', '2024-01-10T12:00')
    ).toBe(false);
  });
});

describe('isMultiDayEvent', () => {
  it('returns true when end date is different from start date', () => {
    expect(isMultiDayEvent('2024-01-10', '2024-01-12')).toBe(true);
  });

  it('returns false when start and end are the same date', () => {
    expect(isMultiDayEvent('2024-01-10', '2024-01-10')).toBe(false);
  });
});

describe('eventDurationMinutes', () => {
  it('calculates duration in minutes', () => {
    expect(eventDurationMinutes('09:00', '10:30')).toBe(90);
    expect(eventDurationMinutes('08:00', '17:00')).toBe(540);
    expect(eventDurationMinutes('13:00', '13:45')).toBe(45);
  });
});

describe('eventsOnDate', () => {
  const events = [
    { date: '2024-01-10', title: 'Meeting' },
    { date: '2024-01-11', title: 'Training' },
    { date: '2024-01-10', title: 'Review' },
  ];

  it('returns events matching the given date', () => {
    const result = eventsOnDate(events, '2024-01-10');
    expect(result).toHaveLength(2);
    expect(result.every((e) => e.date.startsWith('2024-01-10'))).toBe(true);
  });

  it('returns empty array when no events on date', () => {
    expect(eventsOnDate(events, '2024-01-15')).toHaveLength(0);
  });
});

describe('isPublicHoliday', () => {
  const holidays = ['2024-01-01', '2024-12-25'];

  it('returns true when date is a public holiday', () => {
    expect(isPublicHoliday('2024-01-01', holidays)).toBe(true);
    expect(isPublicHoliday('2024-12-25', holidays)).toBe(true);
  });

  it('returns false when date is not a holiday', () => {
    expect(isPublicHoliday('2024-06-15', holidays)).toBe(false);
  });
});

describe('workingDaysInMonth', () => {
  it('calculates working days for January 2024 (31 days)', () => {
    const result = workingDaysInMonth(2024, 1);
    expect(result).toBeGreaterThan(20);
    expect(result).toBeLessThanOrEqual(23);
  });

  it('calculates working days for February 2024 (29 days - leap year)', () => {
    const result = workingDaysInMonth(2024, 2);
    expect(result).toBeGreaterThan(18);
    expect(result).toBeLessThanOrEqual(21);
  });
});
