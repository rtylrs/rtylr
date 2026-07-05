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

import { hrService } from './hrService';
import { safe } from '../lib/api';

describe('hrService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // Staff
  it('getStaff calls GET /hr/staff', () => {
    hrService.getStaff();
    expect(safe.get).toHaveBeenCalledWith('/hr/staff');
  });

  it('createStaff calls POST /hr/staff', () => {
    const data = { name: 'Jane' };
    hrService.createStaff(data);
    expect(safe.post).toHaveBeenCalledWith('/hr/staff', data);
  });

  it('updateStaff calls PUT /hr/staff with id', () => {
    hrService.updateStaff(1, { name: 'Updated' });
    expect(safe.put).toHaveBeenCalledWith('/hr/staff', { id: 1, name: 'Updated' });
  });

  it('deleteStaff calls DELETE /hr/staff with id', () => {
    hrService.deleteStaff(1);
    expect(safe.delete).toHaveBeenCalledWith('/hr/staff', { data: { id: 1 } });
  });

  it('getDashboard calls GET /api/hr/dashboard/v1/dashboard', () => {
    hrService.getDashboard();
    expect(safe.get).toHaveBeenCalledWith('/api/hr/dashboard/v1/dashboard');
  });

  // Departments
  it('getDepartments calls GET /hr/departments', () => {
    hrService.getDepartments();
    expect(safe.get).toHaveBeenCalledWith('/hr/departments');
  });

  it('createDepartment calls POST /hr/departments', () => {
    const data = { name: 'Sales' };
    hrService.createDepartment(data);
    expect(safe.post).toHaveBeenCalledWith('/hr/departments', data);
  });

  it('updateDepartment calls PUT /hr/departments with id', () => {
    hrService.updateDepartment(2, { name: 'Updated' });
    expect(safe.put).toHaveBeenCalledWith('/hr/departments', { id: 2, name: 'Updated' });
  });

  // Shifts
  it('getShifts calls GET /hr/shifts', () => {
    hrService.getShifts();
    expect(safe.get).toHaveBeenCalledWith('/hr/shifts');
  });

  it('createShift calls POST /hr/shifts', () => {
    const data = { staffId: 1 };
    hrService.createShift(data);
    expect(safe.post).toHaveBeenCalledWith('/hr/shifts', data);
  });

  it('clockIn calls POST /hr/staff/clock-in', () => {
    const data = { staffId: 1 };
    hrService.clockIn(data);
    expect(safe.post).toHaveBeenCalledWith('/hr/staff/clock-in', data);
  });

  it('clockOut calls POST /hr/staff/clock-out', () => {
    const data = { staffId: 1 };
    hrService.clockOut(data);
    expect(safe.post).toHaveBeenCalledWith('/hr/staff/clock-out', data);
  });

  // Time Off
  it('getTimeOff calls GET /hr/time-off', () => {
    hrService.getTimeOff();
    expect(safe.get).toHaveBeenCalledWith('/hr/time-off');
  });

  it('submitTimeOff calls POST /hr/time-off', () => {
    const data = { staffId: 1, reason: 'Sick' };
    hrService.submitTimeOff(data);
    expect(safe.post).toHaveBeenCalledWith('/hr/time-off', data);
  });

  it('approveTimeOff calls POST /hr/time-off/approve', () => {
    const data = { id: 1, approved: true };
    hrService.approveTimeOff(data);
    expect(safe.post).toHaveBeenCalledWith('/hr/time-off/approve', data);
  });

  // Payroll
  it('getPayroll calls GET /hr/payroll', () => {
    hrService.getPayroll();
    expect(safe.get).toHaveBeenCalledWith('/hr/payroll');
  });

  it('runPayroll calls POST /hr/payroll/run', () => {
    const data = { period: '2026-03' };
    hrService.runPayroll(data);
    expect(safe.post).toHaveBeenCalledWith('/hr/payroll/run', data);
  });

  it('markPayrollPaid calls POST /hr/payroll/mark-paid', () => {
    const data = { id: 1 };
    hrService.markPayrollPaid(data);
    expect(safe.post).toHaveBeenCalledWith('/hr/payroll/mark-paid', data);
  });
});
