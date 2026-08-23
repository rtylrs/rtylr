// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/flowApi', () => ({
  flowApi: {
    getProjects: vi.fn(),
    createProject: vi.fn(),
    updateProject: vi.fn(),
    archiveProject: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import { useProjects, useCreateProject, useUpdateProject, useArchiveProject } from './useProjects';
import { flowApi } from '../../shared/services/flowApi';

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

describe('useProjects (remote mode)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // The list is business-scoped now, so there is no workspace to gate on and
  // nothing to wait for before the query may run. The two "does not fetch"
  // cases that used to live here guarded a workspace argument that no longer
  // exists.
  it('fetches the business projects without a workspace argument', async () => {
    const data = [{ id: 'p-1', name: 'Project Alpha' }];
    (flowApi.getProjects as any).mockResolvedValue(data);
    const { result } = renderHook(() => useProjects(), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
    expect(flowApi.getProjects).toHaveBeenCalledWith();
  });

  it('starts in loading state', () => {
    (flowApi.getProjects as any).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useProjects(), {
      wrapper: createWrapper(),
    });
    expect(result.current.isLoading).toBe(true);
  });
});

describe('useCreateProject', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls createProject service method', async () => {
    (flowApi.createProject as any).mockResolvedValue({ id: 'p-new' });
    const { result } = renderHook(() => useCreateProject(), {
      wrapper: createWrapper(),
    });
    const input = { workspaceId: 'ws-1', name: 'New Project' };
    await act(() => result.current.mutateAsync(input));
    expect(flowApi.createProject).toHaveBeenCalledWith(input);
  });
});

describe('useUpdateProject', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls updateProject with id and data', async () => {
    (flowApi.updateProject as any).mockResolvedValue({});
    const { result } = renderHook(() => useUpdateProject(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync({ id: 'p-1', data: { name: 'Renamed' } }));
    expect(flowApi.updateProject).toHaveBeenCalledWith('p-1', { name: 'Renamed' });
  });
});

describe('useArchiveProject', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls archiveProject with id', async () => {
    (flowApi.archiveProject as any).mockResolvedValue({ ok: true });
    const { result } = renderHook(() => useArchiveProject(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync('p-2'));
    expect(flowApi.archiveProject).toHaveBeenCalledWith('p-2');
  });
});
