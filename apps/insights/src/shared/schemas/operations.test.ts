import { describe, it, expect } from 'vitest';
import {
  operationsSummarySchema,
  staffUtilizationSchema,
  shiftCoverageSchema,
  timeOffTrendSchema,
  deptHeadcountSchema,
  hrDashboardSchema,
} from './operations';

describe('operationsSummarySchema', () => {
  it('parses a full operations summary', () => {
    const data = {
      activeStaff: 45,
      shiftsThisWeek: 120,
      pendingTimeOff: 3,
      openLeaveRequests: 2,
    };
    const result = operationsSummarySchema.parse(data);
    expect(result.activeStaff).toBe(45);
    expect(result.openLeaveRequests).toBe(2);
  });

  it('accepts an empty object (all fields optional)', () => {
    expect(() => operationsSummarySchema.parse({})).not.toThrow();
  });
});

describe('staffUtilizationSchema', () => {
  it('coerces numeric staffId to string', () => {
    const result = staffUtilizationSchema.parse({ staffId: 7, name: 'Alice' });
    expect(result.staffId).toBe('7');
  });

  it('accepts the "data" and "staff" array fields', () => {
    const result = staffUtilizationSchema.parse({ data: [], staff: [] });
    expect(Array.isArray(result.data)).toBe(true);
    expect(Array.isArray(result.staff)).toBe(true);
  });

  it('accepts an empty object', () => {
    expect(() => staffUtilizationSchema.parse({})).not.toThrow();
  });
});

describe('shiftCoverageSchema', () => {
  it('parses a shift coverage object', () => {
    const data = {
      date: '2026-01-10',
      coverage: 0.92,
      coveredPercent: 0.92,
      scheduled: 8,
      actual: 7,
    };
    const result = shiftCoverageSchema.parse(data);
    expect(result.coverage).toBe(0.92);
    expect(result.scheduled).toBe(8);
  });

  it('accepts an empty object', () => {
    expect(() => shiftCoverageSchema.parse({})).not.toThrow();
  });
});

describe('timeOffTrendSchema', () => {
  it('parses a time-off trend object', () => {
    const data = { period: '2026-01', requests: 10, pending: 2, approved: 7, denied: 1 };
    const result = timeOffTrendSchema.parse(data);
    expect(result.approved).toBe(7);
    expect(result.denied).toBe(1);
  });

  it('accepts both "data" and "recent" array fields', () => {
    const result = timeOffTrendSchema.parse({ data: [], recent: [] });
    expect(Array.isArray(result.data)).toBe(true);
    expect(Array.isArray(result.recent)).toBe(true);
  });

  it('accepts an empty object', () => {
    expect(() => timeOffTrendSchema.parse({})).not.toThrow();
  });
});

describe('deptHeadcountSchema', () => {
  it('parses a department headcount entry', () => {
    const data = { department: 'Sales', headcount: 12, change: 2 };
    const result = deptHeadcountSchema.parse(data);
    expect(result.department).toBe('Sales');
    expect(result.headcount).toBe(12);
  });

  it('accepts an empty object', () => {
    expect(() => deptHeadcountSchema.parse({})).not.toThrow();
  });
});

describe('hrDashboardSchema', () => {
  it('parses an HR dashboard object', () => {
    const data = { totalEmployees: 80, newHires: 5, attrition: 2, openPositions: 4 };
    const result = hrDashboardSchema.parse(data);
    expect(result.totalEmployees).toBe(80);
    expect(result.openPositions).toBe(4);
  });

  it('accepts an empty object', () => {
    expect(() => hrDashboardSchema.parse({})).not.toThrow();
  });
});
