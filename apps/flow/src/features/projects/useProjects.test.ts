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
    deleteProject: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import { useProjects, useCreateProject, useUpdateProject, useDeleteProject } from './useProjects';
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

  it('fetches projects for a workspace', async () => {
    const data = [{ id: 'p-1', name: 'Project Alpha' }];
    (flowApi.getProjects as any).mockResolvedValue(data);
    const { result } = renderHook(() => useProjects('ws-1'), {
      wrapper: createWrapper(),
    });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(data);
    expect(flowApi.getProjects).toHaveBeenCalledWith('ws-1');
  });

  it('does not fetch when workspaceId is undefined', () => {
    const { result } = renderHook(() => useProjects(undefined), {
      wrapper: createWrapper(),
    });
    expect(result.current.fetchStatus).toBe('idle');
    expect(flowApi.getProjects).not.toHaveBeenCalled();
  });

  it('starts in loading state', () => {
    (flowApi.getProjects as any).mockReturnValue(new Promise(() => {}));
    const { result } = renderHook(() => useProjects('ws-1'), {
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

describe('useDeleteProject', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls deleteProject with id', async () => {
    (flowApi.deleteProject as any).mockResolvedValue({ ok: true });
    const { result } = renderHook(() => useDeleteProject(), {
      wrapper: createWrapper(),
    });
    await act(() => result.current.mutateAsync('p-2'));
    expect(flowApi.deleteProject).toHaveBeenCalledWith('p-2');
  });
});
