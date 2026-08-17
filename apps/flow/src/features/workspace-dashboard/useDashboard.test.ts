// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/flowApi', () => ({
  flowApi: {
    getDashboard: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import { useFlowDashboard } from './useDashboard';
import { flowApi } from '../../shared/services/flowApi';
import { PERSONAL_WORKSPACE_ID } from '../../shared/lib/personalWorkspace';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useFlowDashboard (remote mode)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches dashboard data for a workspace', async () => {
    const data = { totalTasks: 50, completedTasks: 30 };
    (flowApi.getDashboard as any).mockResolvedValue(data);
    const { result } = renderHook(() => useFlowDashboard('12'), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
    expect(flowApi.getDashboard).toHaveBeenCalledWith('12');
  });

  it('does not fetch when workspaceId is undefined', () => {
    const { result } = renderHook(() => useFlowDashboard(undefined), {
      wrapper: createWrapper(),
    });
    expect(result.current.fetchStatus).toBe('idle');
    expect(flowApi.getDashboard).not.toHaveBeenCalled();
  });

  it('does not fetch while the workspace is the unresolved Personal sentinel', () => {
    const { result } = renderHook(() => useFlowDashboard(PERSONAL_WORKSPACE_ID), {
      wrapper: createWrapper(),
    });
    expect(result.current.fetchStatus).toBe('idle');
    expect(flowApi.getDashboard).not.toHaveBeenCalled();
  });

  it('starts in loading state', () => {
    (flowApi.getDashboard as any).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useFlowDashboard('12'), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });

  it('handles error state', async () => {
    (flowApi.getDashboard as any).mockRejectedValue(new Error('Fetch failed'));
    const { result } = renderHook(() => useFlowDashboard('12'), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toBeInstanceOf(Error);
  });
});
