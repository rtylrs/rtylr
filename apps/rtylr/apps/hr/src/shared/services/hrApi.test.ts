import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ResultAsync } from 'neverthrow';

vi.mock('../lib/clients', () => ({
  hrTransport: {
    get: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: [] }), () => ({}))),
    post: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}))),
    put: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}))),
    delete: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}))),
  },
}));

import { hrApi } from './hrApi';
import { hrTransport } from '../lib/clients';

describe('hrApi', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // Staff
  it('createStaff calls POST /hr/staff', () => {
    const data = {
      name: 'Alice',
      role: 'Dev',
      department: 'Eng',
      email: 'a@test.com',
      employmentType: 'Full-time' as const,
    };
    hrApi.createStaff(data);
    expect(hrTransport.post).toHaveBeenCalledWith('/hr/staff', {
      firstName: 'Alice',
      lastName: 'Alice',
      email: 'a@test.com',
      role: 'Dev',
      employmentType: 'Full-time',
    });
  });

  it('deleteStaff calls DELETE /hr/staff', () => {
    hrApi.deleteStaff(5);
    expect(hrTransport.delete).toHaveBeenCalledWith('/hr/staff', { data: { id: 5 } });
  });

  it('clockIn calls POST /hr/staff/clock-in', () => {
    hrApi.clockIn(3);
    expect(hrTransport.post).toHaveBeenCalledWith('/hr/staff/clock-in', { staffId: 3 });
  });

  it('clockOut calls POST /hr/staff/clock-out', () => {
    hrApi.clockOut(3);
    expect(hrTransport.post).toHaveBeenCalledWith('/hr/staff/clock-out', { staffId: 3 });
  });

  // Shifts
  it('listShifts maps filters to the snake_case keys the API binds', () => {
    hrApi.listShifts({ day: '2026-06-01', staffId: 4101 });
    expect(hrTransport.get).toHaveBeenCalledWith('/hr/shifts', {
      params: { date: '2026-06-01', staff_id: 4101 },
    });
  });

  it('listShifts omits filters that were not supplied', () => {
    hrApi.listShifts();
    expect(hrTransport.get).toHaveBeenCalledWith('/hr/shifts', { params: {} });
  });

  it('createShift calls POST /hr/shifts with snake_case keys', () => {
    const data = { staffId: 1, start: '09:00', end: '17:00', day: '2026-06-01' };
    hrApi.createShift(data);
    expect(hrTransport.post).toHaveBeenCalledWith('/hr/shifts', {
      staff_id: 1,
      date: '2026-06-01',
      start_time: '09:00',
      end_time: '17:00',
    });
  });

  it('updateShift calls PUT /hr/shifts with snake_case keys', () => {
    hrApi.updateShift({ id: 7, start: '10:00', end: '18:00', day: '2026-06-02' });
    expect(hrTransport.put).toHaveBeenCalledWith('/hr/shifts', {
      id: 7,
      date: '2026-06-02',
      start_time: '10:00',
      end_time: '18:00',
    });
  });

  it('does not expose retired duplicate operations', () => {
    expect(hrApi).not.toHaveProperty('listStaff');
    expect(hrApi).not.toHaveProperty('listDepartments');
    expect(hrApi).not.toHaveProperty('listPayroll');
    expect(hrApi).not.toHaveProperty('runPayroll');
    expect(hrApi).not.toHaveProperty('markPayrollPaid');
    expect(hrApi).not.toHaveProperty('listTimeOff');
    expect(hrApi).not.toHaveProperty('submitTimeOff');
    expect(hrApi).not.toHaveProperty('approveTimeOff');
  });

  // Dashboard
  it('getDashboard calls GET /hr/dashboard/v1/dashboard', () => {
    hrApi.getDashboard();
    expect(hrTransport.get).toHaveBeenCalledWith('/hr/dashboard/v1/dashboard');
  });
});
