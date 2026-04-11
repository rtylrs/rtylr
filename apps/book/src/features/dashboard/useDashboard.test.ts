import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/bookService', () => ({
  bookService: {
    getDashboardStats: vi.fn(),
    getTodayAppointments: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import { useDashboardStats, useTodayAppointments } from './useDashboard';
import { bookService } from '../../shared/services/bookService';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useDashboardStats', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches dashboard stats', async () => {
    const stats = { totalAppointments: 42, revenue: 1200 };
    (bookService.getDashboardStats as any).mockResolvedValue(stats);
    const { result } = renderHook(() => useDashboardStats(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(stats);
  });

  it('starts in loading state', () => {
    (bookService.getDashboardStats as any).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useDashboardStats(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useTodayAppointments', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches today appointments', async () => {
    const appts = [{ id: 1, time: '10:00' }];
    (bookService.getTodayAppointments as any).mockResolvedValue(appts);
    const { result } = renderHook(() => useTodayAppointments(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(appts);
  });

  it('starts in loading state', () => {
    (bookService.getTodayAppointments as any).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useTodayAppointments(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });
});
