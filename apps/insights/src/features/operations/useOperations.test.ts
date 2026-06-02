import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/insightsService', () => ({
  insightsService: {
    getOperationsSummary: vi.fn(),
    getStaffUtilization: vi.fn(),
    getShiftCoverage: vi.fn(),
    getTimeOffTrends: vi.fn(),
    getDeptHeadcount: vi.fn(),
    getHrDashboard: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import {
  useOperationsSummary,
  useStaffUtilization,
  useShiftCoverage,
  useTimeOffTrends,
  useDeptHeadcount,
  useHrDashboard,
} from './useOperations';
import { insightsService } from '../../shared/services/insightsService';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useOperationsSummary', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches operations summary on mount', async () => {
    const data = { activeStaff: 38, shiftsThisWeek: 12, pendingTimeOff: 3, openLeaveRequests: 2 };
    (insightsService.getOperationsSummary as any).mockResolvedValue(data);
    const { result } = renderHook(() => useOperationsSummary(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });

  it('starts in loading state', () => {
    (insightsService.getOperationsSummary as any).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useOperationsSummary(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useStaffUtilization', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches staff utilization data', async () => {
    const data = { average: 0.78, staff: [] };
    (insightsService.getStaffUtilization as any).mockResolvedValue(data);
    const { result } = renderHook(() => useStaffUtilization(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });
});

describe('useShiftCoverage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches shift coverage data', async () => {
    const data = { coverage: 0.95, scheduled: 10, actual: 9 };
    (insightsService.getShiftCoverage as any).mockResolvedValue(data);
    const { result } = renderHook(() => useShiftCoverage(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });
});

describe('useTimeOffTrends', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches time-off trends data', async () => {
    const data = { pending: 5, approved: 12, denied: 2 };
    (insightsService.getTimeOffTrends as any).mockResolvedValue(data);
    const { result } = renderHook(() => useTimeOffTrends(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });
});

describe('useDeptHeadcount', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches department headcount data', async () => {
    const data = { department: 'Engineering', headcount: 15 };
    (insightsService.getDeptHeadcount as any).mockResolvedValue(data);
    const { result } = renderHook(() => useDeptHeadcount(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });
});

describe('useHrDashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches HR dashboard data', async () => {
    const data = { totalEmployees: 100, newHires: 5, attrition: 2, openPositions: 8 };
    (insightsService.getHrDashboard as any).mockResolvedValue(data);
    const { result } = renderHook(() => useHrDashboard(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });
});
