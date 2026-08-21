import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

const mockStaffClient = vi.hoisted(() => ({
  listStaff: vi.fn(),
}));
const mockHrTransport = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
  createOneKitClient: (Client: { name: string }) => {
    if (Client.name === 'StaffServiceClient') return mockStaffClient;
    throw new Error(`unexpected client ${Client.name}`);
  },
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
}));

vi.mock('../../lib/clients', () => ({ hrTransport: mockHrTransport }));

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

import { useAvailability, useSetAvailability } from './useAvailability';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

const STAFF_RESPONSE = {
  staff: [{ id: '10', firstName: 'Carol', lastName: 'Diaz' }],
};

describe('useAvailability', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockStaffClient.listStaff.mockResolvedValue(STAFF_RESPONSE);
  });

  // The old call was GET /hr/shifts?view=availability. No handler has ever read
  // a view parameter, so it returned a shift list that the array guard then
  // discarded, leaving a permanently empty table.
  it('reads the availability endpoint with explicit pagination', async () => {
    mockHrTransport.get.mockResolvedValue({ entries: [], total: 0, limit: 200, offset: 0 });

    const { result } = renderHook(() => useAvailability(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mockHrTransport.get).toHaveBeenCalledWith('/hr/shifts/availability', {
      params: { limit: 200, offset: 0 },
    });
    expect(mockHrTransport.get).not.toHaveBeenCalledWith(
      '/hr/shifts',
      expect.objectContaining({ params: expect.objectContaining({ view: 'availability' }) })
    );
  });

  it('joins the staff name onto a recurring weekly window', async () => {
    mockHrTransport.get.mockResolvedValue({
      entries: [
        {
          id: '1',
          staffId: '10',
          dayOfWeek: 0,
          startTime: '09:00',
          endTime: '17:00',
          isAvailable: true,
        },
      ],
      total: 1,
      limit: 200,
      offset: 0,
    });

    const { result } = renderHook(() => useAvailability(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual([
      {
        id: 1,
        staffId: 10,
        staffName: 'Carol Diaz',
        dayOfWeek: 0,
        specificDate: undefined,
        isAvailable: true,
        startTime: '09:00',
        endTime: '17:00',
      },
    ]);
  });

  // The generated Go server tags every scalar omitempty, so a false
  // is_available never arrives. Treating the absent key as undefined would show
  // a deliberately blocked window as available.
  it('reads an absent isAvailable as blocked rather than available', async () => {
    mockHrTransport.get.mockResolvedValue({
      entries: [
        {
          id: '2',
          staffId: '10',
          specificDate: '2026-08-17',
          startTime: '09:00',
          endTime: '12:00',
        },
      ],
      total: 1,
      limit: 200,
      offset: 0,
    });

    const { result } = renderHook(() => useAvailability(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.[0].isAvailable).toBe(false);
    expect(result.current.data?.[0].specificDate).toBe('2026-08-17');
  });

  it('falls back to a staff id label when the join misses', async () => {
    mockHrTransport.get.mockResolvedValue({
      entries: [
        {
          id: '3',
          staffId: '99',
          dayOfWeek: 3,
          startTime: '09:00',
          endTime: '17:00',
          isAvailable: true,
        },
      ],
      total: 1,
      limit: 200,
      offset: 0,
    });

    const { result } = renderHook(() => useAvailability(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.[0].staffName).toBe('Staff #99');
  });
});

describe('useSetAvailability', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // The old call was PUT /hr/shifts with a shift-shaped body: it sent the
  // weekday under a `date` key and never sent is_available at all.
  it('posts the contract field set for a recurring window', async () => {
    mockHrTransport.post.mockResolvedValue({ entry: {} });

    const { result } = renderHook(() => useSetAvailability(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync({
        staffId: 4,
        dayOfWeek: 2,
        startTime: '09:00',
        endTime: '17:00',
        isAvailable: false,
      });
    });

    expect(mockHrTransport.post).toHaveBeenCalledWith('/hr/shifts/availability', {
      id: undefined,
      staffId: '4',
      dayOfWeek: 2,
      specificDate: undefined,
      startTime: '09:00',
      endTime: '17:00',
      isAvailable: false,
    });
  });

  it('sends the id as a string when replacing an existing window', async () => {
    mockHrTransport.post.mockResolvedValue({ entry: {} });

    const { result } = renderHook(() => useSetAvailability(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync({
        id: 7,
        staffId: 4,
        specificDate: '2026-08-17',
        startTime: '09:00',
        endTime: '12:00',
      });
    });

    expect(mockHrTransport.post).toHaveBeenCalledWith(
      '/hr/shifts/availability',
      expect.objectContaining({ id: '7', specificDate: '2026-08-17', dayOfWeek: undefined })
    );
  });
});
