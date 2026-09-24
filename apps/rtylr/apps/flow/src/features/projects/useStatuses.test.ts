// @vitest-environment happy-dom
import React from 'react';
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

vi.mock('../../shared/services/flowApi', () => ({
  flowApi: {
    listStatuses: vi.fn(),
    createStatus: vi.fn(),
    updateStatus: vi.fn(),
    deleteStatus: vi.fn(),
    setTaskStatus: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((value) => value),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

import { flowApi } from '../../shared/services/flowApi';
import { useDeleteStatus, useUpdateStatus } from './useStatuses';

function createWrapper() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return {
    client,
    wrapper: ({ children }: { children: React.ReactNode }) =>
      React.createElement(QueryClientProvider, { client }, children),
  };
}

describe('status mutation cache invalidation', () => {
  beforeEach(() => vi.clearAllMocks());

  it('invalidates task projections after a status update', async () => {
    vi.mocked(flowApi.updateStatus).mockResolvedValue({ status: { id: 's-1' } } as never);
    const { client, wrapper } = createWrapper();
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useUpdateStatus('p-1'), { wrapper });

    await act(() =>
      result.current.mutateAsync({
        id: 's-1',
        name: 'Blocked',
        bucket: 'active',
        color: '#000',
      })
    );

    expect(spy).toHaveBeenCalledWith({ queryKey: ['statuses', 'p-1'] });
    expect(spy).toHaveBeenCalledWith({ queryKey: ['tasks'] });
    expect(spy).toHaveBeenCalledWith({ queryKey: ['my-tasks'] });
    expect(spy).toHaveBeenCalledWith({ queryKey: ['task'] });
  });

  it('invalidates detail and My Tasks caches after a status delete', async () => {
    vi.mocked(flowApi.deleteStatus).mockResolvedValue({ reassignedTaskCount: 1 } as never);
    const { client, wrapper } = createWrapper();
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useDeleteStatus('p-1'), { wrapper });

    await act(() => result.current.mutateAsync('s-1'));

    expect(spy).toHaveBeenCalledWith({ queryKey: ['my-tasks'] });
    expect(spy).toHaveBeenCalledWith({ queryKey: ['task'] });
  });
});
