import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';
import { useUserStore } from '../stores/userStore';
import { useFlowStore } from '../stores/flowStore';
import {
  useCanCreateTask,
  useCanEditTask,
  useCanEditTaskIn,
  useCanDeleteResource,
  useCurrentProjectRole,
  isTaskMine,
  findProjectRole,
} from './currentUser';

const mocks = vi.hoisted(() => ({ members: vi.fn(), project: vi.fn() }));
vi.mock('../../features/projects/useProjectMembers', () => ({
  useProjectMembers: (id?: string) => ({ data: id ? mocks.members() : undefined }),
  projectMembersQueryOptions: () => ({ queryKey: ['members'], queryFn: mocks.members }),
}));
vi.mock('../../features/projects-v2/useAllProjects', () => ({
  useProject: (id?: string) => ({ data: id ? mocks.project() : undefined }),
}));
function wrapper({ children }: { children: React.ReactNode }) {
  return (
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      {children}
    </QueryClientProvider>
  );
}
beforeEach(() => {
  mocks.members.mockReturnValue([{ userId: '7', role: 'member' }]);
  mocks.project.mockReturnValue(undefined);
  useUserStore.setState({
    user: {
      id: 7,
      fullName: 'Ada',
      email: 'ada@example.test',
      effectiveAccess: createEffectiveAccess('flow', {
        permissions: [
          'flow:tasks:view',
          'flow:tasks:edit',
          'flow:projects:view',
          'flow:projects:edit',
        ],
      }),
    },
  });
});
describe('Flow authorization in Portal', () => {
  it.each([{ isOwner: true }, { isAdmin: true }])(
    'allows %j to administer any project without a membership',
    (role) => {
      mocks.members.mockReturnValue([]);
      useUserStore.setState({
        user: { id: 7, effectiveAccess: createEffectiveAccess('flow', role) },
      });
      const { result } = renderHook(
        () => ({ edit: useCanEditTask('p1'), remove: useCanDeleteResource('p1') }),
        { wrapper }
      );
      expect(result.current).toEqual({ edit: true, remove: true });
    }
  );
  it('requires a member to have task write access as well as a project role', () => {
    useUserStore.setState({
      user: {
        id: 7,
        effectiveAccess: createEffectiveAccess('flow', { permissions: ['flow:tasks:view'] }),
      },
    });
    const { result } = renderHook(() => useCanEditTask('p1'), { wrapper });
    expect(result.current).toBe(false);
  });
  it('keeps a project viewer read only even with business task write access', () => {
    mocks.members.mockReturnValue([{ userId: '7', role: 'viewer' }]);
    const { result } = renderHook(() => useCanEditTask('p1'), { wrapper });
    expect(result.current).toBe(false);
  });
  it('does not request project data without project read access', () => {
    mocks.members.mockClear();
    mocks.project.mockClear();
    useUserStore.setState({
      user: {
        id: 7,
        effectiveAccess: createEffectiveAccess('flow', { permissions: ['flow:tasks:view'] }),
      },
    });
    const { result } = renderHook(() => useCanEditTask('p1'), { wrapper });
    expect(result.current).toBe(false);
    expect(mocks.members).not.toHaveBeenCalled();
    expect(mocks.project).not.toHaveBeenCalled();
  });
  it('grants task edits to an authorized project member', () => {
    const { result } = renderHook(() => useCanEditTask('p1'), { wrapper });
    expect(result.current).toBe(true);
  });
  it('does not grant delete operations to a project member', () => {
    const { result } = renderHook(() => useCanDeleteResource('p1'), { wrapper });
    expect(result.current).toBe(false);
  });
  it('fails closed while project access is unresolved', () => {
    mocks.members.mockReturnValue(undefined);
    const { result } = renderHook(() => useCanEditTask('p1'), { wrapper });
    expect(result.current).toBe(false);
  });
  it('uses Portal identity for assigned task counts', () => {
    expect(isTaskMine('7')).toBe(true);
    expect(isTaskMine('8')).toBe(false);
    useUserStore.setState({ user: null });
    expect(isTaskMine('7')).toBe(false);
  });
  it('matches project collaborators across wire and stored id types', () => {
    expect(
      findProjectRole(
        [
          {
            id: '1',
            projectId: '1',
            userId: '7',
            name: 'Ada',
            email: 'ada@example.test',
            role: 'viewer',
            addedAt: '',
          },
        ],
        String(useUserStore.getState().user?.id)
      )
    ).toBe('viewer');
  });
});

describe('Flow project task creation', () => {
  it.each(['member', 'viewer'])('requires a writable project role for %s', (role) => {
    mocks.members.mockReturnValue([{ userId: '7', role }]);
    useUserStore.setState({
      user: {
        id: 7,
        effectiveAccess: createEffectiveAccess('flow', {
          permissions: ['flow:tasks:create', 'flow:projects:view'],
        }),
      },
    });
    const { result } = renderHook(() => useCanCreateTask('p1'), { wrapper });
    expect(result.current).toBe(role === 'member');
  });
});

// A department manager reaching a report's project, or a Flow Admin, opens a
// project without being on its roster. They read it, and the board offers no
// edit the server would refuse.
describe('Flow project opened without a roster seat', () => {
  beforeEach(() => {
    mocks.members.mockReturnValue([{ userId: '9', role: 'admin' }]);
    mocks.project.mockReturnValue({ id: 'p1', ownerId: '9', createdBy: '9' });
    useUserStore.setState({
      user: {
        id: 7,
        effectiveAccess: createEffectiveAccess('flow', {
          permissions: [
            'flow:projects:view',
            'flow:projects:edit',
            'flow:tasks:view',
            'flow:tasks:create',
            'flow:tasks:edit',
            'flow:team:manage',
          ],
        }),
      },
    });
  });

  it('reads as a viewer with no edit, create or delete controls', () => {
    const { result } = renderHook(
      () => ({
        role: useCurrentProjectRole('p1'),
        edit: useCanEditTask('p1'),
        create: useCanCreateTask('p1'),
        remove: useCanDeleteResource('p1'),
      }),
      { wrapper }
    );
    expect(result.current).toEqual({ role: 'viewer', edit: false, create: false, remove: false });
  });

  it('stays unresolved until the roster has loaded', () => {
    mocks.members.mockReturnValue(undefined);
    const { result } = renderHook(() => useCurrentProjectRole('p1'), { wrapper });
    expect(result.current).toBeUndefined();
  });

  it('stays unresolved when the project could not be opened', () => {
    mocks.project.mockReturnValue(undefined);
    const { result } = renderHook(() => useCurrentProjectRole('p1'), { wrapper });
    expect(result.current).toBeUndefined();
  });
});

// The Plan calendar decides each task's drag by its project. It must answer as
// useCanEditTask would for that project, because both read resolveProjectRole.
describe('Flow task edits across several projects', () => {
  let client: QueryClient;
  const shared = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
  // Waits for the roster read, so a false answer is the roster's, not a load.
  const rosterLoaded = () => waitFor(() => expect(client.getQueryData(['members'])).toBeDefined());

  beforeEach(() => {
    client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    useFlowStore.setState({ projects: [] });
  });

  it('lets a project member edit there', async () => {
    const { result } = renderHook(() => useCanEditTaskIn(['p1']), { wrapper });
    await waitFor(() => expect(result.current('p1')).toBe(true));
  });

  it('keeps a viewer read only', async () => {
    mocks.members.mockReturnValue([{ userId: '7', role: 'viewer' }]);
    const { result } = renderHook(() => useCanEditTaskIn(['p1']), { wrapper: shared });
    await rosterLoaded();
    expect(result.current('p1')).toBe(false);
  });

  it('treats the project owner off the roster as admin', async () => {
    mocks.members.mockReturnValue([{ userId: '9', role: 'admin' }]);
    useFlowStore.setState({
      projects: [{ id: 'p1', ownerId: '7' }] as unknown as ReturnType<
        typeof useFlowStore.getState
      >['projects'],
    });
    const { result } = renderHook(() => useCanEditTaskIn(['p1']), { wrapper });
    await waitFor(() => expect(result.current('p1')).toBe(true));
  });

  it('fails closed for someone off the roster who does not own the project', async () => {
    mocks.members.mockReturnValue([{ userId: '9', role: 'admin' }]);
    useFlowStore.setState({
      projects: [{ id: 'p1', ownerId: '9' }] as unknown as ReturnType<
        typeof useFlowStore.getState
      >['projects'],
    });
    const { result } = renderHook(() => useCanEditTaskIn(['p1']), { wrapper: shared });
    await rosterLoaded();
    expect(result.current('p1')).toBe(false);
  });

  it.each([{ isOwner: true }, { isAdmin: true }])('lets %j edit without a roster', (role) => {
    mocks.members.mockClear();
    useUserStore.setState({
      user: {
        id: 7,
        effectiveAccess: createEffectiveAccess('flow', {
          ...role,
          permissions: ['flow:tasks:edit', 'flow:projects:view'],
        }),
      },
    });
    const { result } = renderHook(() => useCanEditTaskIn(['p1']), { wrapper });
    expect(result.current('p1')).toBe(true);
    expect(mocks.members).not.toHaveBeenCalled();
  });

  it('refuses everything without task write access', () => {
    useUserStore.setState({
      user: {
        id: 7,
        effectiveAccess: createEffectiveAccess('flow', { permissions: ['flow:projects:view'] }),
      },
    });
    const { result } = renderHook(() => useCanEditTaskIn(['p1']), { wrapper });
    expect(result.current('p1')).toBe(false);
  });
});
