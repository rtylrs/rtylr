import React from 'react';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  listAttendance: vi.fn(),
  attendanceClockIn: vi.fn(),
  attendanceClockOut: vi.fn(),
  get: vi.fn(),
  put: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => ({
    listAttendance: mocks.listAttendance,
    attendanceClockIn: mocks.attendanceClockIn,
    attendanceClockOut: mocks.attendanceClockOut,
  }),
  unwrap: (value: unknown) => value,
}));

vi.mock('@voxire/voxie/hr/attendance/v1/client', () => ({
  AttendanceServiceClient: class {},
}));

vi.mock('../../lib/clients', () => ({
  hrTransport: {
    get: mocks.get,
    put: mocks.put,
  },
}));

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

import {
  useAttendanceLocations,
  useClockIn,
  useClockOut,
  useGetMyAttendance,
} from './useAttendance';

function createWrapper(): React.FC<{ children: React.ReactNode }> {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return ({ children }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mocks.listAttendance.mockReset();
  mocks.attendanceClockIn.mockReset();
  mocks.attendanceClockOut.mockReset();
  mocks.get.mockReset();
  mocks.put.mockReset();
});

describe('attendance generated client hooks', () => {
  it('reads history from the registered generated attendance endpoint', async () => {
    const response = { records: [{ id: '1' }], total: 1, limit: 50, offset: 0 };
    mocks.listAttendance.mockResolvedValue(response);

    const { result } = renderHook(() => useGetMyAttendance({}), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mocks.listAttendance).toHaveBeenCalledWith({});
    expect(result.current.data).toEqual(response);
  });

  it('clocks in with the selected positive location and no legacy owner or employee ids', async () => {
    mocks.attendanceClockIn.mockResolvedValue({ record: { id: '2' } });

    const { result } = renderHook(() => useClockIn(), { wrapper: createWrapper() });
    result.current.mutate({ method: 'app', locationId: '7' });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mocks.attendanceClockIn).toHaveBeenCalledWith({
      method: 'app',
      locationId: '7',
      lat: undefined,
      lng: undefined,
      qrToken: undefined,
      pin: undefined,
    });
  });

  it('clocks out by the generated attendance record id', async () => {
    mocks.attendanceClockOut.mockResolvedValue({ record: { id: '2' } });

    const { result } = renderHook(() => useClockOut(), { wrapper: createWrapper() });
    result.current.mutate({ attendanceRecordId: '2' });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mocks.attendanceClockOut).toHaveBeenCalledWith({ attendanceRecordId: '2' });
  });

  it('resolves the default attendance location from the live directory endpoint', async () => {
    mocks.get.mockResolvedValue({
      locations: [
        { id: 7, name: 'HQ', is_default: false },
        { id: 8, name: 'Warehouse', is_default: true },
      ],
    });

    const { result } = renderHook(() => useAttendanceLocations(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mocks.get).toHaveBeenCalledWith('/hr/locations');
    expect(result.current.data?.find((location) => location.isDefault)).toEqual({
      id: '8',
      name: 'Warehouse',
      isDefault: true,
    });
  });
});
