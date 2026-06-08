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
    expect(safe.post).toHaveBeenCalledWith('/hr/staff', data);
  });

  it('deleteStaff calls POST /hr/staff with delete action', () => {
    hrApi.deleteStaff(5);
    expect(safe.post).toHaveBeenCalledWith('/hr/staff', { id: 5, _action: 'delete' });
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
    expect(safe.post).toHaveBeenCalledWith('/hr/departments', data);
  });

  // Shifts
  it('listShifts calls GET /hr/shifts with params', () => {
    const params = { day: 'Monday' };
    hrApi.listShifts(params);
    expect(safe.get).toHaveBeenCalledWith('/hr/shifts', { params });
  });

  it('createShift calls POST /hr/shifts', () => {
    const data = { staffId: 1, start: '09:00', end: '17:00', day: 'Monday' };
    hrApi.createShift(data);
    expect(safe.post).toHaveBeenCalledWith('/hr/shifts', data);
  });

  // Payroll
  it('listPayroll calls GET /hr/payroll with params', () => {
    hrApi.listPayroll({ period: '2026-03' });
    expect(safe.get).toHaveBeenCalledWith('/hr/payroll', { params: { period: '2026-03' } });
  });

  it('runPayroll calls POST /hr/payroll/run', () => {
    hrApi.runPayroll('2026-03');
    expect(safe.post).toHaveBeenCalledWith('/hr/payroll/run', { period: '2026-03' });
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
    expect(safe.post).toHaveBeenCalledWith('/hr/time-off/approve', { id: 2, approved: true });
  });

  // Dashboard
  it('getDashboard calls GET /hr/dashboard', () => {
    hrApi.getDashboard();
    expect(safe.get).toHaveBeenCalledWith('/hr/dashboard');
  });
});
