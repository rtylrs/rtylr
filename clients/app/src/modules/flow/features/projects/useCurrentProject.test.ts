// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach, type Mock } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/flowApi', () => ({
  flowApi: { getProjects: vi.fn() },
}));

vi.mock('../../shared/lib/clients', () => ({
  projectV2: { getProject: vi.fn() },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import { useCurrentProject } from './useCurrentProject';
import { flowApi } from '../../shared/services/flowApi';
import { projectV2 } from '../../shared/lib/clients';

const createWrapper = () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
};

const listed = { id: '1', name: 'Listed', color: '#111111', taskCount: 0, completedCount: 0 };

describe('useCurrentProject', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (flowApi.getProjects as unknown as Mock).mockResolvedValue({ projects: [listed] });
  });

  it('reads a listed project from the sidebar list without asking for it again', async () => {
    const { result } = renderHook(() => useCurrentProject('1'), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.project?.name).toBe('Listed'));
    expect(projectV2.getProject).not.toHaveBeenCalled();
  });

  // Ruling P8 keeps a private project out of a business admin's sidebar, so a
  // link to it has to be resolved by asking for the project itself.
  it('asks for an unlisted project directly, as an admin opening a private one by link', async () => {
    (projectV2.getProject as unknown as Mock).mockResolvedValue({
      project: { id: '8', name: 'Private', color: '#222222', ownerId: '5' },
    });
    const { result } = renderHook(() => useCurrentProject('8'), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.project?.name).toBe('Private'));
    expect(projectV2.getProject).toHaveBeenCalledWith({ id: '8' });
    expect(result.current.loadFailed).toBe(false);
    expect(result.current.isLoading).toBe(false);
  });

  it.each([403, 404])(
    'reads a %i on the direct lookup as not found, not as a failed load',
    async (status) => {
      (projectV2.getProject as unknown as Mock).mockRejectedValue(
        Object.assign(new Error('denied'), { statusCode: status })
      );
      const { result } = renderHook(() => useCurrentProject('9'), { wrapper: createWrapper() });
      await waitFor(() => expect(projectV2.getProject).toHaveBeenCalled());
      await waitFor(() => expect(result.current.isLoading).toBe(false));
      expect(result.current.project).toBeUndefined();
      expect(result.current.loadFailed).toBe(false);
    }
  );

  it('reports a server error on the direct lookup as a failed load', async () => {
    (projectV2.getProject as unknown as Mock).mockRejectedValue(
      Object.assign(new Error('boom'), { statusCode: 500 })
    );
    const { result } = renderHook(() => useCurrentProject('9'), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loadFailed).toBe(true));
    expect(result.current.project).toBeUndefined();
  });

  it('does not ask for the project directly while the list itself has failed', async () => {
    (flowApi.getProjects as unknown as Mock).mockRejectedValue(new Error('list down'));
    const { result } = renderHook(() => useCurrentProject('9'), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.loadFailed).toBe(true));
    expect(projectV2.getProject).not.toHaveBeenCalled();
  });
});
