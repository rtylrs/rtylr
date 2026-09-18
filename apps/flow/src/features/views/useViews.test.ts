// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach, type Mock } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import type { Task, SavedView } from '../../shared/stores/flowStore';

vi.mock('../../shared/services/flowApi', () => ({
  flowApi: {
    getViews: vi.fn(),
    createView: vi.fn(),
    updateView: vi.fn(),
    deleteView: vi.fn(),
  },
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

vi.mock('../../shared/services/i18next', () => ({
  default: { t: (key: string) => key },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

vi.mock('../../shared/stores/flowStore', async () => {
  const actual = await vi.importActual<typeof import('../../shared/stores/flowStore')>(
    '../../shared/stores/flowStore'
  );
  return {
    ...actual,
    useFlowStore: vi.fn(() => ({ activeWorkspaceId: 'ws-store' })),
  };
});

import {
  useViews,
  useCreateView,
  useUpdateView,
  useDeleteView,
  useFilteredTasks,
} from './useViews';
import { flowApi } from '../../shared/services/flowApi';
import { useFlowStore } from '../../shared/stores/flowStore';
import { toast } from 'sonner';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return {
    wrapper: ({ children }: { children: React.ReactNode }) =>
      React.createElement(QueryClientProvider, { client }, children),
    client,
  };
};

describe('useViews (remote mode)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (useFlowStore as unknown as Mock).mockReturnValue({ activeWorkspaceId: 'ws-store' });
  });

  // Saved views are business-shared, so the list takes no scope at all. These
  // two cases used to assert a store fallback and an explicit override.
  it('asks for views without a workspace scope', async () => {
    (flowApi.getViews as unknown as Mock).mockResolvedValue({ views: [] });
    const { wrapper } = createWrapper();
    const { result } = renderHook(() => useViews(), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(flowApi.getViews).toHaveBeenCalledWith();
  });

  it('unwraps the enveloped { views: [...] } shape', async () => {
    const views = [{ id: 'v-1', name: 'My view' }];
    (flowApi.getViews as unknown as Mock).mockResolvedValue({ views });
    const { wrapper } = createWrapper();
    const { result } = renderHook(() => useViews(), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(views);
  });

  it('accepts a raw array response', async () => {
    const views = [{ id: 'v-2', name: 'Another view' }];
    (flowApi.getViews as unknown as Mock).mockResolvedValue(views);
    const { wrapper } = createWrapper();
    const { result } = renderHook(() => useViews(), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(views);
  });
});

describe('useCreateView', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (useFlowStore as unknown as Mock).mockReturnValue({ activeWorkspaceId: 'ws-store' });
  });

  // Creating a view no longer stamps a workspace on it, so the payload reaches
  // the API exactly as the caller wrote it. The two cases this replaces covered
  // the store default and an explicit override of it.
  it('sends the payload through untouched and invalidates views', async () => {
    (flowApi.createView as unknown as Mock).mockResolvedValue({ id: 'v-new' });
    const { wrapper, client } = createWrapper();
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useCreateView(), { wrapper });
    const input = { name: 'New view', icon: 'star' };
    await act(() => result.current.mutateAsync(input));
    expect(flowApi.createView).toHaveBeenCalledWith(input);
    expect(spy).toHaveBeenCalledWith({ queryKey: ['views'] });
  });

  it('shows an error toast and does not invalidate when the create call fails', async () => {
    (flowApi.createView as unknown as Mock).mockRejectedValue(new Error('boom'));
    const { wrapper, client } = createWrapper();
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useCreateView(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ name: 'New view' }).catch(() => {});
    });
    expect(toast.error).toHaveBeenCalledWith('errors.generic');
    expect(spy).not.toHaveBeenCalled();
  });
});

describe('useUpdateView', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls updateView with id and data and invalidates views on success', async () => {
    (flowApi.updateView as unknown as Mock).mockResolvedValue({});
    const { wrapper, client } = createWrapper();
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useUpdateView(), { wrapper });
    await act(() => result.current.mutateAsync({ id: 'v-1', data: { name: 'Renamed' } }));
    expect(flowApi.updateView).toHaveBeenCalledWith('v-1', { name: 'Renamed' });
    expect(spy).toHaveBeenCalledWith({ queryKey: ['views'] });
  });

  // The v2 update is a REPLACEMENT: onion#906 made the handler honour every field
  // it receives, and an unsent field arrives as "". So a patch has to be merged
  // over the cached view before it goes out, or renaming a view blanks its
  // filters, icon and sort.
  it('merges a patch over the cached view so the update does not blank fields', async () => {
    (flowApi.updateView as unknown as Mock).mockResolvedValue({});
    const { wrapper, client } = createWrapper();
    client.setQueryData(
      ['views'],
      [{ id: 'v-1', name: 'Old', icon: 'flame', filters: { status: ['todo'] }, sortBy: 'priority' }]
    );
    const { result } = renderHook(() => useUpdateView(), { wrapper });

    await act(() => result.current.mutateAsync({ id: 'v-1', data: { name: 'Renamed' } }));

    expect(flowApi.updateView).toHaveBeenCalledWith('v-1', {
      id: 'v-1',
      name: 'Renamed',
      icon: 'flame',
      filters: { status: ['todo'] },
      sortBy: 'priority',
    });
  });

  // The toggle's field has to reach the wire. It was previously dropped: the
  // backend chain had landed on flow/view/v1 while the client is on v2.
  it('carries visibility through to the update call', async () => {
    (flowApi.updateView as unknown as Mock).mockResolvedValue({});
    const { wrapper, client } = createWrapper();
    client.setQueryData(['views'], [{ id: 'v-1', name: 'Mine', icon: 'eye' }]);
    const { result } = renderHook(() => useUpdateView(), { wrapper });

    await act(() => result.current.mutateAsync({ id: 'v-1', data: { visibility: 'private' } }));

    expect(flowApi.updateView).toHaveBeenCalledWith('v-1', {
      id: 'v-1',
      name: 'Mine',
      icon: 'eye',
      visibility: 'private',
    });
  });

  it('shows an error toast and does not invalidate when the update call fails', async () => {
    (flowApi.updateView as unknown as Mock).mockRejectedValue(new Error('boom'));
    const { wrapper, client } = createWrapper();
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useUpdateView(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ id: 'v-1', data: { name: 'Renamed' } }).catch(() => {});
    });
    expect(toast.error).toHaveBeenCalledWith('errors.generic');
    expect(spy).not.toHaveBeenCalled();
  });
});

describe('useDeleteView', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls deleteView with id and invalidates views on success', async () => {
    (flowApi.deleteView as unknown as Mock).mockResolvedValue({ ok: true });
    const { wrapper, client } = createWrapper();
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useDeleteView(), { wrapper });
    await act(() => result.current.mutateAsync('v-1'));
    expect(flowApi.deleteView).toHaveBeenCalledWith('v-1');
    expect(spy).toHaveBeenCalledWith({ queryKey: ['views'] });
  });

  it('shows an error toast and does not invalidate when the delete call fails', async () => {
    (flowApi.deleteView as unknown as Mock).mockRejectedValue(new Error('boom'));
    const { wrapper, client } = createWrapper();
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useDeleteView(), { wrapper });
    await act(async () => {
      await result.current.mutateAsync('v-1').catch(() => {});
    });
    expect(toast.error).toHaveBeenCalledWith('errors.generic');
    expect(spy).not.toHaveBeenCalled();
  });
});

function makeTask(overrides: Partial<Task> = {}): Task {
  return {
    id: 't-1',
    projectId: 'p-1',
    workspaceId: 'ws-1',
    title: 'Task',
    status: 'todo',
    priority: 'medium',
    assignee: 'Jane',
    assigneeInitials: 'J',
    assigneeId: 'u-1',
    dueDate: '2026-01-01',
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  } as Task;
}

function makeView(overrides: Partial<SavedView> = {}): SavedView {
  return {
    id: 'v-1',
    name: 'My view',
    icon: 'star',
    filters: {},
    ...overrides,
  } as SavedView;
}

describe('useFilteredTasks', () => {
  it('returns the tasks unchanged when there is no view', () => {
    const tasks = [makeTask({ id: 't-1' }), makeTask({ id: 't-2' })];
    const { result } = renderHook(() => useFilteredTasks(null, tasks));
    expect(result.current).toEqual(tasks);
  });

  it('filters by status', () => {
    const tasks = [
      makeTask({ id: 't-1', status: 'todo' }),
      makeTask({ id: 't-2', status: 'done' }),
    ];
    const view = makeView({ filters: { status: ['done'] } });
    const { result } = renderHook(() => useFilteredTasks(view, tasks));
    expect(result.current.map((t) => t.id)).toEqual(['t-2']);
  });

  it('filters by priority', () => {
    const tasks = [
      makeTask({ id: 't-1', priority: 'low' }),
      makeTask({ id: 't-2', priority: 'urgent' }),
    ];
    const view = makeView({ filters: { priority: ['urgent'] } });
    const { result } = renderHook(() => useFilteredTasks(view, tasks));
    expect(result.current.map((t) => t.id)).toEqual(['t-2']);
  });

  it('filters by assigneeId', () => {
    const tasks = [
      makeTask({ id: 't-1', assigneeId: 'u-1' }),
      makeTask({ id: 't-2', assigneeId: 'u-2' }),
    ];
    const view = makeView({ filters: { assigneeId: 'u-2' } });
    const { result } = renderHook(() => useFilteredTasks(view, tasks));
    expect(result.current.map((t) => t.id)).toEqual(['t-2']);
  });

  it('filters by projectId', () => {
    const tasks = [
      makeTask({ id: 't-1', projectId: 'p-1' }),
      makeTask({ id: 't-2', projectId: 'p-2' }),
    ];
    const view = makeView({ filters: { projectId: 'p-2' } });
    const { result } = renderHook(() => useFilteredTasks(view, tasks));
    expect(result.current.map((t) => t.id)).toEqual(['t-2']);
  });

  it('filters by labels', () => {
    const tasks = [
      makeTask({ id: 't-1', labelIds: ['l-1'] } as Partial<Task>),
      makeTask({ id: 't-2', labelIds: ['l-2'] } as Partial<Task>),
    ];
    const view = makeView({ filters: { labels: ['l-2'] } });
    const { result } = renderHook(() => useFilteredTasks(view, tasks));
    expect(result.current.map((t) => t.id)).toEqual(['t-2']);
  });

  it('sorts by priority ascending (urgent first)', () => {
    const tasks = [
      makeTask({ id: 't-low', priority: 'low' }),
      makeTask({ id: 't-urgent', priority: 'urgent' }),
      makeTask({ id: 't-medium', priority: 'medium' }),
    ];
    const view = makeView({ sortBy: 'priority', sortDir: 'asc' });
    const { result } = renderHook(() => useFilteredTasks(view, tasks));
    expect(result.current.map((t) => t.id)).toEqual(['t-urgent', 't-medium', 't-low']);
  });

  it('sorts by priority descending when sortDir is desc', () => {
    const tasks = [
      makeTask({ id: 't-low', priority: 'low' }),
      makeTask({ id: 't-urgent', priority: 'urgent' }),
    ];
    const view = makeView({ sortBy: 'priority', sortDir: 'desc' });
    const { result } = renderHook(() => useFilteredTasks(view, tasks));
    expect(result.current.map((t) => t.id)).toEqual(['t-low', 't-urgent']);
  });

  it('sorts by dueDate ascending', () => {
    const tasks = [
      makeTask({ id: 't-later', dueDate: '2026-05-01' }),
      makeTask({ id: 't-earlier', dueDate: '2026-01-01' }),
    ];
    const view = makeView({ sortBy: 'dueDate', sortDir: 'asc' });
    const { result } = renderHook(() => useFilteredTasks(view, tasks));
    expect(result.current.map((t) => t.id)).toEqual(['t-earlier', 't-later']);
  });
});
