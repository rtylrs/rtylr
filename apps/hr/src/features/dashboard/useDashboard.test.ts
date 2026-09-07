import { describe, it, expect, vi, beforeEach, type Mock } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/hrApi', () => ({
  hrApi: {
    getDashboard: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import { useHRDashboard } from './useDashboard';
import { hrApi } from '../../shared/services/hrApi';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useHRDashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches HR dashboard data', async () => {
    const data = {
      totalStaff: 25,
      onShiftToday: 18,
      openPositions: 3,
      monthlyPayroll: 50000,
      recentActivity: [],
      todayAtGlance: {
        shiftsScheduled: 10,
        newApplications: 2,
        pendingApprovals: 1,
      },
    };
    (hrApi.getDashboard as unknown as Mock).mockResolvedValue(data);
    const { result } = renderHook(() => useHRDashboard(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });

  it('starts in loading state', () => {
    (hrApi.getDashboard as unknown as Mock).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useHRDashboard(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });

  it('handles error state', async () => {
    (hrApi.getDashboard as unknown as Mock).mockRejectedValue(new Error('Server error'));
    const { result } = renderHook(() => useHRDashboard(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toBeInstanceOf(Error);
  });
});
