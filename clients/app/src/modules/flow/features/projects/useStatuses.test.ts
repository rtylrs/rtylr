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
    reorderStatuses: vi.fn(),
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
import {
  useCreateStatus,
  useDeleteStatus,
  useReorderStatuses,
  useUpdateStatus,
} from './useStatuses';

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

    await act(() => result.current.mutateAsync({ id: 's-1', moveToStatusId: 's-2' }));

    expect(spy).toHaveBeenCalledWith({ queryKey: ['my-tasks'] });
    expect(spy).toHaveBeenCalledWith({ queryKey: ['task'] });
    expect(flowApi.deleteStatus).toHaveBeenCalledWith('s-1', 's-2');
  });
});

const apiStatus = (id: string, sortOrder: number) => ({
  id,
  projectId: 'p-1',
  name: `S${id}`,
  bucket: 'active' as const,
  legacyStatus: 'todo',
  color: '#000',
  sortOrder,
  isDefault: false,
  createdAt: '',
});

describe('status reorder', () => {
  beforeEach(() => vi.clearAllMocks());

  it('sends the whole order in one request and shows it before the server answers', async () => {
    let resolve: (value: unknown) => void = () => undefined;
    vi.mocked(flowApi.reorderStatuses).mockReturnValue(
      new Promise((r) => {
        resolve = r;
      }) as never
    );
    const { client, wrapper } = createWrapper();
    client.setQueryData(['statuses', 'p-1'], [apiStatus('1', 0), apiStatus('2', 10)]);
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useReorderStatuses('p-1'), { wrapper });

    let done: Promise<unknown> = Promise.resolve();
    await act(async () => {
      done = result.current.mutateAsync(['2', '1']);
    });
    const optimistic = client.getQueryData<{ id: string; sortOrder: number }[]>([
      'statuses',
      'p-1',
    ]);
    expect(optimistic?.map((s) => [s.id, s.sortOrder])).toEqual([
      ['2', 0],
      ['1', 10],
    ]);

    await act(async () => {
      resolve({ statuses: [] });
      await done;
    });
    expect(flowApi.reorderStatuses).toHaveBeenCalledTimes(1);
    expect(flowApi.reorderStatuses).toHaveBeenCalledWith('p-1', ['2', '1']);
    expect(flowApi.updateStatus).not.toHaveBeenCalled();
    expect(spy).toHaveBeenCalledWith({ queryKey: ['statuses', 'p-1'] });
  });

  it('puts the previous order back when the server refuses', async () => {
    vi.mocked(flowApi.reorderStatuses).mockRejectedValue(new Error('400'));
    const { client, wrapper } = createWrapper();
    const before = [apiStatus('1', 0), apiStatus('2', 10)];
    client.setQueryData(['statuses', 'p-1'], before);
    // Keep the rolled-back value observable: the settle-time invalidation
    // would otherwise refetch through the mocked listStatuses.
    vi.spyOn(client, 'invalidateQueries').mockResolvedValue();
    const { result } = renderHook(() => useReorderStatuses('p-1'), { wrapper });

    await act(async () => {
      await result.current.mutateAsync(['2', '1']).catch(() => undefined);
    });

    expect(client.getQueryData(['statuses', 'p-1'])).toEqual(before);
  });

  it('places an inserted status with one reorder after the create', async () => {
    vi.mocked(flowApi.createStatus).mockResolvedValue({ status: apiStatus('9', 11) } as never);
    vi.mocked(flowApi.reorderStatuses).mockResolvedValue({ statuses: [] } as never);
    const { wrapper } = createWrapper();
    const { result } = renderHook(() => useCreateStatus('p-1'), { wrapper });

    await act(() =>
      result.current.mutateAsync({
        name: 'Triage',
        bucket: 'active',
        placeAmong: { ids: ['1', '2'], index: 1 },
      })
    );

    expect(flowApi.createStatus).toHaveBeenCalledWith(
      expect.not.objectContaining({ placeAmong: expect.anything() })
    );
    expect(flowApi.reorderStatuses).toHaveBeenCalledTimes(1);
    expect(flowApi.reorderStatuses).toHaveBeenCalledWith('p-1', ['1', '9', '2']);
    expect(flowApi.updateStatus).not.toHaveBeenCalled();
  });
});
