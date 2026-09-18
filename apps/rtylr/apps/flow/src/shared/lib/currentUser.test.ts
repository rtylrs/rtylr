// @vitest-environment happy-dom
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

vi.mock('../services/flowApi', () => ({
  flowApi: {
    getProjectMembers: vi.fn(),
    getWorkspaceMembers: vi.fn(),
  },
}));

vi.mock('@voxire/api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@voxire/api')>();
  return { ...actual, unwrap: vi.fn((val: unknown) => val) };
});

import { useUserStore as useAuthUserStore } from '@voxire/auth';
import { findProjectRole, isTaskMine, useCanEditTask, useCanDeleteResource } from './currentUser';
import { flowApi } from '../services/flowApi';
import { useUserStore } from '../stores/userStore';
import { useFlowStore } from '../stores/flowStore';
import type { ApiProjectMember } from '../services/flowApi';

function makeProjectMember(overrides: Partial<ApiProjectMember>): ApiProjectMember {
  return {
    id: 'pm1',
    projectId: 'p1',
    userId: '7',
    name: 'Alice',
    email: 'alice@example.com',
    role: 'member',
    addedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

describe('findProjectRole', () => {
  it('returns the signed-in user role on the project', () => {
    const members = [
      makeProjectMember({ userId: '7', role: 'viewer' }),
      makeProjectMember({ id: 'pm2', userId: '8', role: 'admin' }),
    ];
    expect(findProjectRole(members, '7')).toBe('viewer');
    expect(findProjectRole(members, '8')).toBe('admin');
  });

  it('fails closed while the member list is still loading', () => {
    expect(findProjectRole(undefined, '7')).toBeUndefined();
  });

  it('fails closed when there is no signed-in user', () => {
    expect(findProjectRole([makeProjectMember({})], null)).toBeUndefined();
  });

  it('fails closed when the user is not a member of the project', () => {
    expect(findProjectRole([makeProjectMember({ userId: '7' })], '99')).toBeUndefined();
  });

  it('matches ids across string and number representations', () => {
    // The API layer can hand back a numeric id even though the contract types it
    // as a string, so a project role must not be missed on a type mismatch.
    const members = [{ ...makeProjectMember({ role: 'admin' }), userId: 7 as unknown as string }];
    expect(findProjectRole(members, '7')).toBe('admin');
  });
});

describe('useCanEditTask', () => {
  const createWrapper = () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return ({ children }: { children: React.ReactNode }) =>
      React.createElement(QueryClientProvider, { client }, children);
  };

  beforeEach(() => {
    vi.clearAllMocks();
    useUserStore.setState({
      user: { id: '7', name: 'Ada', email: 'ada@example.com' },
      token: 'tok',
      isAuthenticated: true,
    });
    useFlowStore.setState({ activeWorkspaceId: 'ws-1' });
    // No business-wide role by default - each test that needs the bypass sets
    // it explicitly, so a leftover role from a previous test can't leak in.
    useAuthUserStore.setState({ role: null });
  });

  it('grants edit from an explicit project role above viewer', async () => {
    (flowApi.getProjectMembers as ReturnType<typeof vi.fn>).mockResolvedValue({
      members: [makeProjectMember({ userId: '7', role: 'member' })],
    });
    (flowApi.getWorkspaceMembers as ReturnType<typeof vi.fn>).mockResolvedValue({ members: [] });

    const { result } = renderHook(() => useCanEditTask('p1'), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current).toBe(true));
  });

  it('denies edit when the explicit project role is viewer, even if the workspace role would allow it', async () => {
    (flowApi.getProjectMembers as ReturnType<typeof vi.fn>).mockResolvedValue({
      members: [makeProjectMember({ userId: '7', role: 'viewer' })],
    });
    (flowApi.getWorkspaceMembers as ReturnType<typeof vi.fn>).mockResolvedValue({
      members: [
        {
          id: 'wm1',
          workspaceId: 'ws-1',
          userId: '7',
          name: 'Ada',
          email: '',
          role: 'admin',
          joinedAt: '',
        },
      ],
    });

    const { result } = renderHook(() => useCanEditTask('p1'), { wrapper: createWrapper() });
    await waitFor(() => expect(flowApi.getProjectMembers).toHaveBeenCalled());
    await waitFor(() => expect(result.current).toBe(false));
  });

  it.each(['owner', 'admin', 'operator'] as const)(
    'does not grant edit from a legacy business-wide %s role without project scope',
    async (businessRole) => {
      useAuthUserStore.setState({ role: businessRole });
      (flowApi.getProjectMembers as ReturnType<typeof vi.fn>).mockResolvedValue({ members: [] });
      (flowApi.getWorkspaceMembers as ReturnType<typeof vi.fn>).mockResolvedValue({ members: [] });

      const { result } = renderHook(() => useCanEditTask('p1'), { wrapper: createWrapper() });
      await waitFor(() => expect(result.current).toBe(false));
    }
  );
});

describe('useCanDeleteResource', () => {
  const createWrapper = () => {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return ({ children }: { children: React.ReactNode }) =>
      React.createElement(QueryClientProvider, { client }, children);
  };

  beforeEach(() => {
    vi.clearAllMocks();
    useUserStore.setState({
      user: { id: '7', name: 'Ada', email: 'ada@example.com' },
      token: 'tok',
      isAuthenticated: true,
    });
    useFlowStore.setState({ activeWorkspaceId: 'ws-1' });
    useAuthUserStore.setState({ role: null });
  });

  it.each(['owner', 'admin', 'operator'] as const)(
    'does not grant delete from a legacy business-wide %s role without project scope',
    async (businessRole) => {
      useAuthUserStore.setState({ role: businessRole });
      (flowApi.getProjectMembers as ReturnType<typeof vi.fn>).mockResolvedValue({ members: [] });

      const { result } = renderHook(() => useCanDeleteResource('p1'), {
        wrapper: createWrapper(),
      });
      await waitFor(() => expect(result.current).toBe(false));
    }
  );

  it('does not bypass delete for a business-wide staff role with no project membership', async () => {
    useAuthUserStore.setState({ role: 'staff' });
    (flowApi.getProjectMembers as ReturnType<typeof vi.fn>).mockResolvedValue({ members: [] });

    const { result } = renderHook(() => useCanDeleteResource('p1'), { wrapper: createWrapper() });
    await waitFor(() => expect(flowApi.getProjectMembers).toHaveBeenCalled());
    expect(result.current).toBe(false);
  });
});

describe('isTaskMine', () => {
  // This predicate gates the My tasks count, the Today count and the whole
  // Today list. A strict === between a number and a string is false for every
  // task, so if either id arrives numeric it fails silently and completely:
  // the screens report zero rather than erroring. The comment on
  // findProjectRole already records that the API layer can hand back a
  // numeric id where the contract types it as a string.
  beforeEach(() => {
    useUserStore.setState({
      user: { id: '7', name: 'Ada', email: 'ada@example.com' },
      token: 'tok',
      isAuthenticated: true,
    });
  });

  it('matches when both ids are strings', () => {
    expect(isTaskMine('7')).toBe(true);
  });

  it('matches when the assignee id arrived as a number', () => {
    expect(isTaskMine(7 as unknown as string)).toBe(true);
  });

  it('matches when the stored id arrived as a number', () => {
    useUserStore.setState({
      user: { id: 7 as unknown as string, name: 'Ada', email: 'ada@example.com' },
      token: 'tok',
      isAuthenticated: true,
    });
    expect(isTaskMine('7')).toBe(true);
  });

  it('does not match a different person', () => {
    expect(isTaskMine('8')).toBe(false);
    expect(isTaskMine(8 as unknown as string)).toBe(false);
  });

  it('does not treat an unassigned task as mine', () => {
    expect(isTaskMine('')).toBe(false);
  });
});
