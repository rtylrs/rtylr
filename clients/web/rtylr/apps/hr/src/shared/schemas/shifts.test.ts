import { describe, it, expect } from 'vitest';
import {
  shiftSchema,
  scheduleShiftSchema,
  availabilitySchema,
  swapRequestSchema,
  breakLogSchema,
  overtimeAlertSchema,
} from './shifts';

// ─── Zod schema unit tests - no mocks, no network ────────────────────────────
//
// Each schema has a valid "happy path" case plus edge-case coercions and the
// `.catch()` fallback that exists on every enum field.

describe('shiftSchema', () => {
  it('parses a valid shift with numeric id', () => {
    const result = shiftSchema.parse({
      id: 1,
      name: 'Morning Shift',
      role: 'Cashier',
      start: '08:00',
      end: '16:00',
      status: 'Scheduled',
      day: 'Monday',
    });
    expect(result.id).toBe(1);
    expect(result.status).toBe('Scheduled');
  });

  it('coerces string id to number', () => {
    const result = shiftSchema.parse({
      id: '42',
      name: 'Evening',
      role: 'Barista',
      start: '16:00',
      end: '22:00',
      status: 'In Progress',
      day: 'Tuesday',
    });
    expect(result.id).toBe(42);
  });

  it('falls back to "Scheduled" for unrecognised status', () => {
    const result = shiftSchema.parse({
      id: 5,
      name: 'Night Shift',
      role: 'Security',
      start: '22:00',
      end: '06:00',
      status: 'Unknown Status',
      day: 'Wednesday',
    });
    expect(result.status).toBe('Scheduled');
  });

  it('accepts optional staffId', () => {
    const withStaff = shiftSchema.parse({
      id: 10,
      name: 'Shift A',
      role: 'Manager',
      start: '09:00',
      end: '17:00',
      status: 'Completed',
      day: 'Friday',
      staffId: 99,
    });
    expect(withStaff.staffId).toBe(99);
  });

  it('allows staffId to be absent', () => {
    const result = shiftSchema.parse({
      id: 11,
      name: 'Shift B',
      role: 'Cook',
      start: '10:00',
      end: '18:00',
      status: 'Scheduled',
      day: 'Saturday',
    });
    expect(result.staffId).toBeUndefined();
  });

  it('rejects missing required fields', () => {
    expect(() =>
      shiftSchema.parse({ id: 1, name: 'Shift', role: 'Cook', start: '09:00' })
    ).toThrow();
  });
});

describe('scheduleShiftSchema', () => {
  it('parses a full schedule shift', () => {
    const result = scheduleShiftSchema.parse({
      id: '7',
      staffId: 3,
      staffName: 'Alice',
      role: 'Barista',
      day: 'Thursday',
      start: '07:00',
      end: '15:00',
      color: '#4ade80',
      published: true,
    });
    expect(result.id).toBe(7);
    expect(result.staffName).toBe('Alice');
    expect(result.color).toBe('#4ade80');
    expect(result.published).toBe(true);
  });

  it('allows color and published to be absent', () => {
    const result = scheduleShiftSchema.parse({
      id: 2,
      staffId: 5,
      staffName: 'Bob',
      role: 'Cook',
      day: 'Monday',
      start: '09:00',
      end: '17:00',
    });
    expect(result.color).toBeUndefined();
    expect(result.published).toBeUndefined();
  });
});

describe('availabilitySchema', () => {
  it('parses a recurring weekly pattern', () => {
    const result = availabilitySchema.parse({
      id: '1',
      staffId: 10,
      staffName: 'Carol',
      dayOfWeek: 0,
      isAvailable: true,
      startTime: '09:00',
      endTime: '17:00',
    });
    expect(result.id).toBe(1);
    expect(result.dayOfWeek).toBe(0);
    expect(result.specificDate).toBeUndefined();
    expect(result.isAvailable).toBe(true);
  });

  it('parses a single-date override', () => {
    const result = availabilitySchema.parse({
      id: 3,
      staffId: 4,
      staffName: 'Dave',
      specificDate: '2026-08-17',
      isAvailable: false,
      startTime: '09:00',
      endTime: '12:00',
    });
    expect(result.specificDate).toBe('2026-08-17');
    expect(result.dayOfWeek).toBeUndefined();
  });

  // The generated Go server tags every scalar omitempty, so a false
  // is_available never reaches the client. An absent key means false, and
  // reading it as undefined would render a blocked window as available.
  it('reads an absent isAvailable as false', () => {
    const result = availabilitySchema.parse({
      id: 5,
      staffId: 6,
      staffName: 'Erin',
      dayOfWeek: 2,
      startTime: '09:00',
      endTime: '17:00',
    });
    expect(result.isAvailable).toBe(false);
  });

  // dayOfWeek is Monday-first (0 = Monday through 6 = Sunday). Date.getDay()
  // would supply 0 for Sunday, so the bound is what catches that mix-up.
  it('rejects a dayOfWeek outside the Monday-first range', () => {
    expect(() =>
      availabilitySchema.parse({
        id: 7,
        staffId: 8,
        staffName: 'Frank',
        dayOfWeek: 7,
        isAvailable: true,
        startTime: '09:00',
        endTime: '17:00',
      })
    ).toThrow();
  });

  it('keeps a window that crosses midnight as given', () => {
    const result = availabilitySchema.parse({
      id: 9,
      staffId: 10,
      staffName: 'Grace',
      dayOfWeek: 4,
      isAvailable: true,
      startTime: '22:00',
      endTime: '06:00',
    });
    expect(result.startTime).toBe('22:00');
    expect(result.endTime).toBe('06:00');
  });
});

describe('swapRequestSchema', () => {
  it('parses a valid swap request', () => {
    const result = swapRequestSchema.parse({
      id: '2',
      requesterId: 1,
      requester: 'Eve',
      targetId: 2,
      target: 'Frank',
      shiftDay: 'Wednesday',
      reason: 'Doctor appointment',
      status: 'Pending',
    });
    expect(result.id).toBe(2);
    expect(result.status).toBe('Pending');
  });

  it('falls back to "Pending" for unrecognised status', () => {
    const result = swapRequestSchema.parse({
      id: 3,
      requesterId: 1,
      requester: 'Grace',
      targetId: 2,
      target: 'Hank',
      shiftDay: 'Thursday',
      reason: 'Family event',
      status: 'Disputed',
    });
    expect(result.status).toBe('Pending');
  });
});

describe('breakLogSchema', () => {
  it('parses a break log with all fields', () => {
    const result = breakLogSchema.parse({
      id: '5',
      staffId: 9,
      staffName: 'Iris',
      start: '2026-05-01T12:00:00',
      end: '2026-05-01T12:30:00',
      durationMin: 30,
      type: 'Lunch',
    });
    expect(result.id).toBe(5);
    expect(result.type).toBe('Lunch');
    expect(result.durationMin).toBe(30);
  });

  it('allows end and durationMin to be absent (break in progress)', () => {
    const result = breakLogSchema.parse({
      id: 6,
      staffId: 9,
      staffName: 'Iris',
      start: '2026-05-01T15:00:00',
      type: 'Short',
    });
    expect(result.end).toBeUndefined();
    expect(result.durationMin).toBeUndefined();
  });

  it('falls back to "Other" for unrecognised break type', () => {
    const result = breakLogSchema.parse({
      id: 7,
      staffId: 9,
      staffName: 'Iris',
      start: '2026-05-01T16:00:00',
      type: 'Nap',
    });
    expect(result.type).toBe('Other');
  });
});

describe('overtimeAlertSchema', () => {
  it('parses a valid overtime alert', () => {
    const result = overtimeAlertSchema.parse({
      id: '8',
      staffId: 12,
      staffName: 'Jake',
      hoursThisWeek: 45,
      threshold: 40,
      triggeredAt: '2026-04-28T18:00:00',
    });
    expect(result.id).toBe(8);
    expect(result.hoursThisWeek).toBe(45);
    expect(result.threshold).toBe(40);
  });

  it('rejects missing required fields', () => {
    expect(() =>
      overtimeAlertSchema.parse({ id: 1, staffId: 2, staffName: 'Joe', hoursThisWeek: 44 })
    ).toThrow();
  });
});
