import React from 'react';
import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  get: vi.fn(),
  leaveClient: {
    listLeaveTypes: vi.fn(),
    createLeaveType: vi.fn(),
    listLeaveRequests: vi.fn(),
    createLeaveRequest: vi.fn(),
    approveLeave: vi.fn(),
    listLeaveBalances: vi.fn(),
  },
}));

vi.mock('../../lib/clients', () => ({ hrTransport: { get: mocks.get } }));
vi.mock('@voxire/api', () => ({
  createOneKitClient: () => mocks.leaveClient,
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
  unwrap: (value: unknown) => value,
}));
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));
vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

import { useGetLeaveCalendar, useApproveLeave } from './useLeave';

const createWrapper = (): React.FC<{ children: React.ReactNode }> => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }) => React.createElement(QueryClientProvider, { client }, children);
};

describe('useApproveLeave', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.leaveClient.approveLeave.mockResolvedValue({ success: true });
  });

  it('forwards a required rejection reason to the generated client', async () => {
    const { result } = renderHook(() => useApproveLeave(), { wrapper: createWrapper() });

    await act(async () => {
      await result.current.mutateAsync({
        requestId: 17,
        body: { approved: false, rejectionReason: 'Insufficient balance' },
      });
    });

    expect(mocks.leaveClient.approveLeave).toHaveBeenCalledWith({
      requestId: '17',
      body: { approved: false, rejectionReason: 'Insufficient balance' },
    });
  });

  it('rejects an empty reason before making the request', async () => {
    const { result } = renderHook(() => useApproveLeave(), { wrapper: createWrapper() });

    await expect(
      result.current.mutateAsync({
        requestId: 17,
        body: { approved: false, rejectionReason: '  ' },
      })
    ).rejects.toThrow('rejection reason');

    expect(mocks.leaveClient.approveLeave).not.toHaveBeenCalled();
  });
});

describe('useGetLeaveCalendar', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.get.mockResolvedValue({ requests: [] });
  });

  it('translates the selected month into the live calendar date range', async () => {
    const { result } = renderHook(
      () =>
        useGetLeaveCalendar({
          month: 2,
          year: 2024,
          departmentId: 9,
          locationId: 4,
        }),
      { wrapper: createWrapper() }
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(mocks.get).toHaveBeenCalledWith('/hr/leave/calendar', {
      params: {
        dateFrom: '2024-02-01',
        dateTo: '2024-02-29',
        locationId: 4,
      },
    });
  });
});
