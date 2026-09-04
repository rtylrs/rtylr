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

// t interpolates here so the assertions can see the date itself rather than just
// the key. Asserting the key alone would pass even if {{date}} never resolved,
// which is precisely what ruling P3 is about.
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, unknown>) =>
      opts?.date ? `${key}|${String(opts.date)}` : key,
  }),
}));

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((val) => val),
}));

import { useProjects, useCreateProject, useUpdateProject, useArchiveProject } from './useProjects';
import { flowApi } from '../../shared/services/flowApi';
import { toast } from 'sonner';

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

  // The New project modal collects a description and a visibility, and the
  // hook's parameter type used to omit description, so it had nowhere to send
  // it (#1060).
  it('passes the description and visibility straight through', async () => {
    (flowApi.createProject as any).mockResolvedValue({ id: 'p-new' });
    const { result } = renderHook(() => useCreateProject(), {
      wrapper: createWrapper(),
    });

    await act(() =>
      result.current.mutateAsync({
        name: 'QC Delta Two',
        color: '#6366f1',
        description: 'DESC SHOULD PERSIST',
        visibility: 'public',
      })
    );

    expect(flowApi.createProject).toHaveBeenCalledWith({
      name: 'QC Delta Two',
      color: '#6366f1',
      description: 'DESC SHOULD PERSIST',
      visibility: 'public',
    });
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

  // Ruling P3: the archive confirmation must state the scheduled permanent
  // deletion date explicitly. The date is only knowable after the archive,
  // because the server stamps purge_at from the business's retention setting -
  // which is why the confirmation carries it rather than the pre-archive dialog.
  it('states the scheduled deletion date in the confirmation', async () => {
    (flowApi.archiveProject as any).mockResolvedValue({
      project: { id: 'p-2', purgeAt: '2026-09-23T00:00:00Z' },
    });
    const { result } = renderHook(() => useArchiveProject(), {
      wrapper: createWrapper(),
    });

    await act(() => result.current.mutateAsync('p-2'));

    expect(toast.success).toHaveBeenCalledWith('toasts.project.archivedUntil|2026-09-23');
  });

  // The response previously typed away the project, so a build that still does
  // must not print "undefined" at the user. Undated message, no crash.
  it('falls back to the undated message when no purge date comes back', async () => {
    (flowApi.archiveProject as any).mockResolvedValue({ project: { id: 'p-2' } });
    const { result } = renderHook(() => useArchiveProject(), {
      wrapper: createWrapper(),
    });

    await act(() => result.current.mutateAsync('p-2'));

    expect(toast.success).toHaveBeenCalledWith('toasts.project.archived');
  });
});
