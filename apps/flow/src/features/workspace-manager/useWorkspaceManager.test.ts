// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../../shared/services/flowApi', () => ({
  flowApi: {
    getWorkspaceMembers: vi.fn(),
    addWorkspaceMember: vi.fn(),
    removeWorkspaceMember: vi.fn(),
    updateWorkspaceMemberRole: vi.fn(),
  },
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

// Keep the real toFiniteNumber - the wsId === 0 / uid === 0 guards in the
// hooks below depend on its actual coercion behavior, not a stand-in.
vi.mock('@voxire/api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@voxire/api')>();
  return { ...actual, unwrap: vi.fn((val: unknown) => val) };
});

import {
  useWorkspaceMembers,
  useAddWorkspaceMember,
  useRemoveWorkspaceMember,
  useUpdateWorkspaceMemberRole,
} from './useWorkspaceManager';
import { flowApi } from '../../shared/services/flowApi';
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

describe('useWorkspaceMembers', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('unwraps the enveloped { members: [...] } shape', async () => {
    const members = [{ id: 'm-1', name: 'Jane' }];
    (flowApi.getWorkspaceMembers as any).mockResolvedValue({ members });
    const { wrapper } = createWrapper();
    const { result } = renderHook(() => useWorkspaceMembers('ws-1'), { wrapper });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toEqual(members);
    expect(flowApi.getWorkspaceMembers).toHaveBeenCalledWith('ws-1');
  });

  it('does not fetch when workspaceId is undefined', () => {
    const { wrapper } = createWrapper();
    const { result } = renderHook(() => useWorkspaceMembers(undefined), { wrapper });
    expect(result.current.fetchStatus).toBe('idle');
    expect(flowApi.getWorkspaceMembers).not.toHaveBeenCalled();
  });
});

describe('useAddWorkspaceMember', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls addWorkspaceMember and invalidates the workspace member list on success', async () => {
    (flowApi.addWorkspaceMember as any).mockResolvedValue({ ok: true });
    const { wrapper, client } = createWrapper();
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useAddWorkspaceMember('ws-1'), { wrapper });
    const input = { workspace_id: 1, user_id: 2, role: 'member' };
    await act(() => result.current.mutateAsync(input));
    expect(flowApi.addWorkspaceMember).toHaveBeenCalledWith(input);
    expect(spy).toHaveBeenCalledWith({ queryKey: ['workspace-members', 'ws-1'] });
  });

  it('rejects and never calls the API when workspace_id is not a valid finite number', async () => {
    const { wrapper, client } = createWrapper();
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useAddWorkspaceMember('ws-1'), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ workspace_id: 0, user_id: 2 }).catch(() => {});
    });
    expect(flowApi.addWorkspaceMember).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith('Invalid workspace ID');
    expect(spy).not.toHaveBeenCalled();
  });

  it('rejects and never calls the API when user_id is not a valid finite number', async () => {
    const { wrapper } = createWrapper();
    const { result } = renderHook(() => useAddWorkspaceMember('ws-1'), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ workspace_id: 1, user_id: 0 }).catch(() => {});
    });
    expect(flowApi.addWorkspaceMember).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith('Invalid user ID');
  });

  it('shows an error toast and does not invalidate when the API call fails', async () => {
    (flowApi.addWorkspaceMember as any).mockRejectedValue(new Error('boom'));
    const { wrapper, client } = createWrapper();
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useAddWorkspaceMember('ws-1'), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ workspace_id: 1, user_id: 2 }).catch(() => {});
    });
    expect(toast.error).toHaveBeenCalledWith('Failed to add workspace member');
    expect(spy).not.toHaveBeenCalled();
  });
});

describe('useRemoveWorkspaceMember', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls removeWorkspaceMember and invalidates the workspace member list on success', async () => {
    (flowApi.removeWorkspaceMember as any).mockResolvedValue({ ok: true });
    const { wrapper, client } = createWrapper();
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useRemoveWorkspaceMember('ws-1'), { wrapper });
    const input = { workspace_id: 1, user_id: 2 };
    await act(() => result.current.mutateAsync(input));
    expect(flowApi.removeWorkspaceMember).toHaveBeenCalledWith(input);
    expect(spy).toHaveBeenCalledWith({ queryKey: ['workspace-members', 'ws-1'] });
  });

  it('rejects and never calls the API when workspace_id is not a valid finite number', async () => {
    const { wrapper } = createWrapper();
    const { result } = renderHook(() => useRemoveWorkspaceMember('ws-1'), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ workspace_id: 0, user_id: 2 }).catch(() => {});
    });
    expect(flowApi.removeWorkspaceMember).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith('Invalid workspace ID');
  });

  it('rejects and never calls the API when user_id is not a valid finite number', async () => {
    const { wrapper } = createWrapper();
    const { result } = renderHook(() => useRemoveWorkspaceMember('ws-1'), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ workspace_id: 1, user_id: 0 }).catch(() => {});
    });
    expect(flowApi.removeWorkspaceMember).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith('Invalid user ID');
  });

  it('shows an error toast and does not invalidate when the API call fails', async () => {
    (flowApi.removeWorkspaceMember as any).mockRejectedValue(new Error('boom'));
    const { wrapper, client } = createWrapper();
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useRemoveWorkspaceMember('ws-1'), { wrapper });
    await act(async () => {
      await result.current.mutateAsync({ workspace_id: 1, user_id: 2 }).catch(() => {});
    });
    expect(toast.error).toHaveBeenCalledWith('Failed to remove workspace member');
    expect(spy).not.toHaveBeenCalled();
  });
});

describe('useUpdateWorkspaceMemberRole', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('calls updateWorkspaceMemberRole and invalidates the workspace member list on success', async () => {
    (flowApi.updateWorkspaceMemberRole as any).mockResolvedValue({ ok: true });
    const { wrapper, client } = createWrapper();
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useUpdateWorkspaceMemberRole('ws-1'), { wrapper });
    const input = { workspace_id: 1, user_id: 2, role: 'admin' };
    await act(() => result.current.mutateAsync(input));
    expect(flowApi.updateWorkspaceMemberRole).toHaveBeenCalledWith(input);
    expect(spy).toHaveBeenCalledWith({ queryKey: ['workspace-members', 'ws-1'] });
  });

  it('rejects and never calls the API when workspace_id is not a valid finite number', async () => {
    const { wrapper } = createWrapper();
    const { result } = renderHook(() => useUpdateWorkspaceMemberRole('ws-1'), { wrapper });
    await act(async () => {
      await result.current
        .mutateAsync({ workspace_id: 0, user_id: 2, role: 'admin' })
        .catch(() => {});
    });
    expect(flowApi.updateWorkspaceMemberRole).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith('Invalid workspace ID');
  });

  it('rejects and never calls the API when user_id is not a valid finite number', async () => {
    const { wrapper } = createWrapper();
    const { result } = renderHook(() => useUpdateWorkspaceMemberRole('ws-1'), { wrapper });
    await act(async () => {
      await result.current
        .mutateAsync({ workspace_id: 1, user_id: 0, role: 'admin' })
        .catch(() => {});
    });
    expect(flowApi.updateWorkspaceMemberRole).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith('Invalid user ID');
  });

  it('shows an error toast and does not invalidate when the API call fails', async () => {
    (flowApi.updateWorkspaceMemberRole as any).mockRejectedValue(new Error('boom'));
    const { wrapper, client } = createWrapper();
    const spy = vi.spyOn(client, 'invalidateQueries');
    const { result } = renderHook(() => useUpdateWorkspaceMemberRole('ws-1'), { wrapper });
    await act(async () => {
      await result.current
        .mutateAsync({ workspace_id: 1, user_id: 2, role: 'admin' })
        .catch(() => {});
    });
    expect(toast.error).toHaveBeenCalledWith('Failed to update workspace member role');
    expect(spy).not.toHaveBeenCalled();
  });
});
