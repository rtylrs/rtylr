// @vitest-environment happy-dom
import React from 'react';
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

vi.mock('../../shared/services/flowApi', () => ({
  flowApi: {
    getProjectColumns: vi.fn(),
    moveTaskToColumn: vi.fn(async () => ({ message: 'Task moved' })),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((value) => value),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn() },
}));

import { flowApi } from '../../shared/services/flowApi';
import { useMoveTaskToColumn, useProjectColumns } from './useColumns';

function wrapper({ children }: { children: React.ReactNode }) {
  return React.createElement(
    QueryClientProvider,
    { client: new QueryClient({ defaultOptions: { queries: { retry: false } } }) },
    children
  );
}

describe('useProjectColumns', () => {
  beforeEach(() => vi.clearAllMocks());

  it('loads every page so later columns are available to the board', async () => {
    vi.mocked(flowApi.getProjectColumns)
      .mockResolvedValueOnce({
        columns: [{ id: '1', name: 'To Do' }],
        pagination: { page: 1, hasNext: true },
      } as never)
      .mockResolvedValueOnce({
        columns: [{ id: '2', name: 'Done' }],
        pagination: { page: 2, hasNext: false },
      } as never);

    const { result } = renderHook(() => useProjectColumns('p-1'), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toHaveLength(2);
    expect(flowApi.getProjectColumns).toHaveBeenNthCalledWith(1, 'p-1', {
      page: 1,
      limit: 500,
    });
    expect(flowApi.getProjectColumns).toHaveBeenNthCalledWith(2, 'p-1', {
      page: 2,
      limit: 500,
    });
  });
});

// Dropping a card on a column tied to a status changes the task's status on
// the server, so everything that shows status has to refetch, not only the
// task lists: My tasks and Today, and the projects' done counts.
describe('useMoveTaskToColumn', () => {
  it('refreshes every view a status change shows up in', async () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useMoveTaskToColumn(), {
      wrapper: ({ children }: { children: React.ReactNode }) =>
        React.createElement(QueryClientProvider, { client }, children),
    });
    await act(async () => {
      await result.current.mutateAsync({ taskId: '7', columnId: '3' });
    });
    const keys = spy.mock.calls.map((call) => JSON.stringify(call[0]?.queryKey));
    for (const key of [['tasks'], ['task', '7'], ['my-tasks'], ['projects'], ['projects-v2']]) {
      expect(keys).toContain(JSON.stringify(key));
    }
  });
});
