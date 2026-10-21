// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/flowApi', () => ({
  flowApi: {
    getTasks: vi.fn(),
    getMyTasks: vi.fn(),
    getTaskDetail: vi.fn(),
    createTask: vi.fn(),
    updateTask: vi.fn(),
    deleteTask: vi.fn(),
    updateTaskStatus: vi.fn(),
  },
}));

vi.mock('../../shared/lib/currentUser', () => ({
  isTaskMine: vi.fn(() => false),
}));

vi.mock('../../shared/stores/flowStore', () => ({
  useFlowStore: vi.fn(() => ({ activeWorkspaceId: 'ws-1' })),
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import {
  useTasks,
  useMyTasks,
  useTaskDetail,
  useTaskDetails,
  useCreateTask,
  useUpdateTask,
  useDeleteTask,
  useUpdateTaskStatus,
  useBulkUpdateArchive,
} from './useTasks';
import { flowApi } from '../../shared/services/flowApi';
import { useFlowStore } from '../../shared/stores/flowStore';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useTasks (remote mode)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches tasks on mount', async () => {
    const data = [{ id: 't-1', title: 'Fix bug' }];
    (flowApi.getTasks as unknown as ReturnType<typeof vi.fn>).mockResolvedValue(data);
    const params = { projectId: 'p-1' };
    const { result } = renderHook(() => useTasks(params), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
    expect(flowApi.getTasks).toHaveBeenCalledWith({ ...params, page: 1, limit: 500 });
  });

  it('loads all project task pages and preserves failures on later pages', async () => {
    vi.mocked(flowApi.getTasks)
      .mockResolvedValueOnce({ tasks: [{ id: '1' }], pagination: { hasNext: true } })
      .mockResolvedValueOnce({ tasks: [{ id: '2' }], pagination: { hasNext: false } });
    const { result } = renderHook(() => useTasks({ projectId: '7' }), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual([{ id: '1' }, { id: '2' }]);
    expect(flowApi.getTasks).toHaveBeenNthCalledWith(2, { projectId: '7', page: 2, limit: 500 });
  });
  it('does not report a partial task list as successful when a later page fails', async () => {
    vi.mocked(flowApi.getTasks)
      .mockResolvedValueOnce({ tasks: [{ id: '1' }], pagination: { hasNext: true } })
      .mockRejectedValueOnce(new Error('API unavailable'));
    const { result } = renderHook(() => useTasks({ projectId: '7' }), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.data).toBeUndefined();
  });

  it('starts in loading state when projectId provided', () => {
    (flowApi.getTasks as unknown as ReturnType<typeof vi.fn>).mockReturnValue(
      new Promise(() => {})
    );
    const { result } = renderHook(() => useTasks({ projectId: 'p-1' }), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useMyTasks', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches current user tasks', async () => {
    const data = [{ id: 't-2', title: 'My task' }];
    (flowApi.getMyTasks as unknown as ReturnType<typeof vi.fn>).mockResolvedValue(data);
    const { result } = renderHook(() => useMyTasks(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
  });

  it('merges every page before Today filters the list', async () => {
    (flowApi.getMyTasks as unknown as ReturnType<typeof vi.fn>)
      .mockResolvedValueOnce({
        tasks: [{ id: 't-1' }],
        pagination: { page: 1, hasNext: true },
      })
      .mockResolvedValueOnce({
        tasks: [{ id: 't-2' }],
        pagination: { page: 2, hasNext: false },
      });
    const { result } = renderHook(() => useMyTasks(), { wrapper: createWrapper() });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual([{ id: 't-1' }, { id: 't-2' }]);
    expect(flowApi.getMyTasks).toHaveBeenNthCalledWith(1, { page: 1, limit: 500 });
    expect(flowApi.getMyTasks).toHaveBeenNthCalledWith(2, { page: 2, limit: 500 });
  });
});

describe('useTaskDetail', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches task detail by id', async () => {
    const task = { id: 't-5', title: 'Detail task' };
    (flowApi.getTaskDetail as unknown as ReturnType<typeof vi.fn>).mockResolvedValue(task);
    const { result } = renderHook(() => useTaskDetail('t-5'), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(task);
  });

  it('does not fetch when id is undefined', () => {
    const { result } = renderHook(() => useTaskDetail(undefined), {
      wrapper: createWrapper(),
    });
    expect(result.current.fetchStatus).toBe('idle');
    expect(flowApi.getTaskDetail).not.toHaveBeenCalled();
  });
});

describe('useTaskDetails', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // #01329: the business-wide list is one page, so BulkEditBar resolves the
  // selected ids it is missing by fetching their details individually.
  it('fetches the detail of every unresolved id', async () => {
    (flowApi.getTaskDetail as unknown as ReturnType<typeof vi.fn>).mockImplementation(
      (id: string) => Promise.resolve({ task: { id, projectId: 'p-9' } })
    );

    const { result } = renderHook(() => useTaskDetails(['t-1', 't-2']), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.every((q) => q.isSuccess)).toBe(true));
    expect(flowApi.getTaskDetail).toHaveBeenCalledWith('t-1');
    expect(flowApi.getTaskDetail).toHaveBeenCalledWith('t-2');
    expect(result.current.map((q) => q.data?.task.projectId)).toEqual(['p-9', 'p-9']);
  });
});

describe('useCreateTask', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls createTask service method', async () => {
    (flowApi.createTask as unknown as ReturnType<typeof vi.fn>).mockResolvedValue({ id: 't-new' });
    const { result } = renderHook(() => useCreateTask(), {
      wrapper: createWrapper(),
    });
    const input = { projectId: 'p-1', title: 'New task' };
    await act(() => result.current.mutateAsync(input));
    expect(flowApi.createTask).toHaveBeenCalledWith(input);
  });

  it('passes labels through to createTask unchanged', async () => {
    (flowApi.createTask as unknown as ReturnType<typeof vi.fn>).mockResolvedValue({ id: 't-new' });
    const { result } = renderHook(() => useCreateTask(), {
      wrapper: createWrapper(),
    });
    const input = { projectId: 'p-1', title: 'New task', labels: ['5'] };
    await act(() => result.current.mutateAsync(input));
    expect(flowApi.createTask).toHaveBeenCalledWith(input);
  });
});

describe('useUpdateTask', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls updateTask with id and data', async () => {
    (flowApi.updateTask as unknown as ReturnType<typeof vi.fn>).mockResolvedValue({});
    const { result } = renderHook(() => useUpdateTask(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync({ id: 't-1', data: { title: 'Updated' } }));
    expect(flowApi.updateTask).toHaveBeenCalledWith('t-1', { title: 'Updated' });
  });
});

describe('useDeleteTask', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls deleteTask with id', async () => {
    (flowApi.deleteTask as unknown as ReturnType<typeof vi.fn>).mockResolvedValue({ ok: true });
    const { result } = renderHook(() => useDeleteTask(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync('t-3'));
    expect(flowApi.deleteTask).toHaveBeenCalledWith('t-3');
  });
});

describe('useUpdateTaskStatus', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls updateTaskStatus with id and status', async () => {
    (flowApi.updateTaskStatus as unknown as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
    });
    const { result } = renderHook(() => useUpdateTaskStatus(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync({ id: 't-1', status: 'done' }));
    expect(flowApi.updateTaskStatus).toHaveBeenCalledWith('t-1', 'done');
  });
});

describe('useBulkUpdateArchive', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls updateTask with archived:true for every selected id', async () => {
    // Regression guard: BulkEditBar's Archive button used to call
    // useBulkUpdateStatus with status:'cancelled' - a different field
    // entirely. This hook must touch `archived`, not `status`.
    (flowApi.updateTask as unknown as ReturnType<typeof vi.fn>).mockResolvedValue({ ok: true });
    const { result } = renderHook(() => useBulkUpdateArchive(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync({ ids: ['t-1', 't-2'] }));
    expect(flowApi.updateTask).toHaveBeenCalledWith('t-1', { archived: true });
    expect(flowApi.updateTask).toHaveBeenCalledWith('t-2', { archived: true });
  });
});

// #1033: useTasks scoped itself by workspace and gated the query on
// `projectId || workspaceId`. With the workspace layer removed,
// activeWorkspaceId is permanently '', so every caller that passes no
// projectId was permanently disabled and rendered an empty state with no
// request issued.
describe('useTasks without a projectId', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // The real post-workspace-removal store value.
    (useFlowStore as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      activeWorkspaceId: '',
    });
  });

  it('queries the business-scoped list when no projectId is given', async () => {
    const data = [{ id: 't-1', title: 'Backlog item' }];
    (flowApi.getTasks as unknown as ReturnType<typeof vi.fn>).mockResolvedValue(data);

    const { result } = renderHook(() => useTasks({}), { wrapper: createWrapper() });

    await waitFor(() => expect(flowApi.getTasks).toHaveBeenCalledTimes(1));
    await waitFor(() => expect(result.current.data).toEqual(data));
  });

  it('does not send a workspaceId, which the list endpoint has no parameter for', async () => {
    (flowApi.getTasks as unknown as ReturnType<typeof vi.fn>).mockResolvedValue([]);

    renderHook(() => useTasks({}), { wrapper: createWrapper() });

    await waitFor(() => expect(flowApi.getTasks).toHaveBeenCalledTimes(1));
    const sent = (flowApi.getTasks as unknown as ReturnType<typeof vi.fn>).mock.calls[0][0];
    expect(sent).not.toHaveProperty('workspaceId');
  });

  it('queries for the archive shape too, which passes activeOnly and no project', async () => {
    (flowApi.getTasks as unknown as ReturnType<typeof vi.fn>).mockResolvedValue([]);

    renderHook(() => useTasks({ activeOnly: false }), { wrapper: createWrapper() });

    await waitFor(() => expect(flowApi.getTasks).toHaveBeenCalledTimes(1));
  });

  it('still passes a projectId through when one is given', async () => {
    (flowApi.getTasks as unknown as ReturnType<typeof vi.fn>).mockResolvedValue([]);

    renderHook(() => useTasks({ projectId: 'p-9' }), { wrapper: createWrapper() });

    await waitFor(() =>
      expect(flowApi.getTasks).toHaveBeenCalledWith(expect.objectContaining({ projectId: 'p-9' }))
    );
  });
});
