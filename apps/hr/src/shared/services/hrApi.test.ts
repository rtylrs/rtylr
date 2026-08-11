import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ResultAsync } from 'neverthrow';

vi.mock('../lib/api', () => ({
  safe: {
    get: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: [] }), () => ({}) as any)),
    post: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
    put: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
    delete: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
  },
}));

import { hrApi } from './hrApi';
import { safe } from '../lib/api';

describe('hrApi', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // Staff
  it('listStaff calls GET /hr/staff', () => {
    hrApi.listStaff();
    expect(safe.get).toHaveBeenCalledWith('/hr/staff');
  });

  it('createStaff calls POST /hr/staff', () => {
    const data = {
      name: 'Alice',
      role: 'Dev',
      department: 'Eng',
      email: 'a@test.com',
      employmentType: 'Full-time' as const,
    };
    hrApi.createStaff(data);
    expect(safe.post).toHaveBeenCalledWith('/hr/staff', {
      firstName: 'Alice',
      lastName: 'Alice',
      email: 'a@test.com',
      role: 'Dev',
      employmentType: 'Full-time',
    });
  });

  it('deleteStaff calls DELETE /hr/staff', () => {
    hrApi.deleteStaff(5);
    expect(safe.delete).toHaveBeenCalledWith('/hr/staff', { data: { id: 5 } });
  });

  it('clockIn calls POST /hr/staff/clock-in', () => {
    hrApi.clockIn(3);
    expect(safe.post).toHaveBeenCalledWith('/hr/staff/clock-in', { staffId: 3 });
  });

  it('clockOut calls POST /hr/staff/clock-out', () => {
    hrApi.clockOut(3);
    expect(safe.post).toHaveBeenCalledWith('/hr/staff/clock-out', { staffId: 3 });
  });

  // Departments
  it('listDepartments calls GET /hr/departments', () => {
    hrApi.listDepartments();
    expect(safe.get).toHaveBeenCalledWith('/hr/departments');
  });

  it('createDepartment calls POST /hr/departments', () => {
    const data = { name: 'Eng', description: 'Engineering', head: 'Bob' };
    hrApi.createDepartment(data);
    expect(safe.post).toHaveBeenCalledWith('/hr/departments', {
      name: 'Eng',
      description: 'Engineering',
    });
  });

  // Shifts
  it('listShifts maps filters to the snake_case keys the API binds', () => {
    hrApi.listShifts({ day: '2026-06-01', staffId: 4101 });
    expect(safe.get).toHaveBeenCalledWith('/hr/shifts', {
      params: { date: '2026-06-01', staff_id: 4101 },
    });
  });

  it('listShifts omits filters that were not supplied', () => {
    hrApi.listShifts();
    expect(safe.get).toHaveBeenCalledWith('/hr/shifts', { params: {} });
  });

  it('createShift calls POST /hr/shifts with snake_case keys', () => {
    const data = { staffId: 1, start: '09:00', end: '17:00', day: '2026-06-01' };
    hrApi.createShift(data);
    expect(safe.post).toHaveBeenCalledWith('/hr/shifts', {
      staff_id: 1,
      date: '2026-06-01',
      start_time: '09:00',
      end_time: '17:00',
    });
  });

  it('updateShift calls PUT /hr/shifts with snake_case keys', () => {
    hrApi.updateShift({ id: 7, start: '10:00', end: '18:00', day: '2026-06-02' });
    expect(safe.put).toHaveBeenCalledWith('/hr/shifts', {
      id: 7,
      date: '2026-06-02',
      start_time: '10:00',
      end_time: '18:00',
    });
  });

  // Payroll
  it('listPayroll calls GET /hr/payroll with params', () => {
    hrApi.listPayroll({ period: '2026-03' });
    expect(safe.get).toHaveBeenCalledWith('/hr/payroll', { params: { period: '2026-03' } });
  });

  it('runPayroll calls POST /hr/payroll/run', () => {
    hrApi.runPayroll('2026-03');
    expect(safe.post).toHaveBeenCalledWith('/hr/payroll/run', {
      periodStart: '2026-03',
      periodEnd: '2026-03',
      periodLabel: '2026-03',
    });
  });

  it('markPayrollPaid calls POST /hr/payroll/mark-paid', () => {
    hrApi.markPayrollPaid(7);
    expect(safe.post).toHaveBeenCalledWith('/hr/payroll/mark-paid', { id: 7 });
  });

  // Time Off
  it('listTimeOff calls GET /hr/time-off', () => {
    hrApi.listTimeOff();
    expect(safe.get).toHaveBeenCalledWith('/hr/time-off');
  });

  it('submitTimeOff calls POST /hr/time-off', () => {
    const data = { staffId: 1, startDate: '2026-04-01', endDate: '2026-04-05', reason: 'Vacation' };
    hrApi.submitTimeOff(data);
    expect(safe.post).toHaveBeenCalledWith('/hr/time-off', data);
  });

  it('approveTimeOff calls POST /hr/time-off/approve', () => {
    hrApi.approveTimeOff(2, true);
    expect(safe.post).toHaveBeenCalledWith('/hr/time-off/approve', {
      id: 2,
      status: 'approved',
    });
  });

  // Dashboard
  it('getDashboard calls GET /hr/dashboard/v1/dashboard', () => {
    hrApi.getDashboard();
    expect(safe.get).toHaveBeenCalledWith('/hr/dashboard/v1/dashboard');
  });
});
