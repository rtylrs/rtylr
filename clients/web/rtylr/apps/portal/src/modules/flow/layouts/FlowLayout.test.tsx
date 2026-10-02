import React from 'react';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createEffectiveAccess } from '../../../../../../test/effectiveAccess';
import { initTestI18n } from '../shared/testing/i18n';
import { useUserStore } from '../../../shared/stores/userStore';
import { useFlowStore } from '../shared/stores/flowStore';
import FlowLayout from './FlowLayout';

const mocks = vi.hoisted(() => ({
  projects: [] as Array<{
    id: string;
    name: string;
    color: string;
    taskCount: number;
    completedCount: number;
    memberAvatars: string[];
  }>,
  members: [] as Array<{ id: string; name: string; email: string; isActive: boolean }>,
  capacity: {} as Record<string, number>,
  realtime: vi.fn(),
  create: vi.fn(),
}));
vi.mock('@tanstack/react-router', () => ({
  Link: ({
    children,
    to,
    params,
    ...rest
  }: React.PropsWithChildren<{ to: string; params?: { userId?: string } }>) => (
    <a href={to.replace('$userId', params?.userId ?? '')} {...rest}>
      {children}
    </a>
  ),
  useLocation: () => ({ pathname: '/flow/my-issues' }),
  useParams: () => ({}),
}));
vi.mock('@voxire/config', async (original) => ({
  ...(await original<typeof import('@voxire/config')>()),
  useFeatureFlag: () => false,
}));
vi.mock('../features/projects/useProjects', () => ({
  useProjects: () => ({ data: mocks.projects }),
  useCreateProject: () => ({ mutate: mocks.create, isPending: false }),
}));
vi.mock('../features/tasks/useMyTasksProject', () => ({
  useMyTasksProject: () => ({ data: { projectId: 'personal' } }),
}));
vi.mock('../shared/hooks/queries/useFlowTeam', () => ({
  useFlowTeam: () => ({ data: mocks.members }),
}));
vi.mock('../shared/hooks/queries/useMemberCapacity', () => ({
  useMemberCapacityHours: () => ({ data: mocks.capacity }),
}));
vi.mock('../features/inbox/useNotifications', () => ({ useNotifications: () => ({ data: [] }) }));
vi.mock('../shared/hooks/useKeyboardShortcuts', () => ({ useKeyboardShortcuts: () => undefined }));
vi.mock('../shared/lib/useLayoutRealtime', () => ({ useLayoutRealtime: mocks.realtime }));
vi.mock('../shared/lib/realtime', () => ({
  useRealtimeReconnectRefetch: () => undefined,
  useRealtimeStatus: () => 'connected',
}));
vi.mock('../shared/lib/lifePreferences', () => ({
  useLifePreferences: () => ({ prefs: { quietStart: '00:00', quietEnd: '00:00', digestHour: 8 } }),
}));
vi.mock('../shared/lib/sidebarPreferences', async (original) => ({
  ...(await original<typeof import('../shared/lib/sidebarPreferences')>()),
  useSidebarPreferences: () => ({
    preferences: { version: 1, items: [], groups: [], groupOrder: [], topLevelItemIds: [] },
    update: vi.fn(),
  }),
}));
vi.mock('../shared/components/SidebarNavigation', () => ({
  SidebarNavigation: ({
    onNewProject,
    canCreateProject,
  }: {
    onNewProject: () => void;
    canCreateProject: boolean;
  }) => (canCreateProject ? <button onClick={onNewProject}>New project</button> : null),
}));
vi.mock('../shared/components/TrialBanner', () => ({ default: () => null }));
vi.mock('../shared/components/QuickCreate', () => ({ default: () => null }));
vi.mock('../shared/components/BulkEditBar', () => ({ default: () => null }));
vi.mock('../shared/components/TaskDrawer', () => ({ default: () => null }));

beforeEach(async () => {
  vi.clearAllMocks();
  await initTestI18n();
  mocks.projects = [];
  mocks.members = [];
  useUserStore.setState({
    user: {
      id: 7,
      businessId: 42,
      fullName: 'Ada',
      effectiveAccess: createEffectiveAccess('flow', { isOwner: true }),
    },
  });
  useFlowStore.setState({
    projects: [],
    members: [],
    tasks: [],
    quickCreateOpen: false,
    projectCreateOpen: false,
  });
});
describe('Flow module layout', () => {
  it('uses the selected Portal business and owner identity for realtime', () => {
    render(
      <FlowLayout>
        <p>Flow page</p>
      </FlowLayout>
    );
    expect(mocks.realtime).toHaveBeenCalledWith('42', '7', true);
    expect(screen.getByRole('link', { name: 'RTYLR' })).toHaveAttribute('href', '/');
    expect(screen.getByRole('link', { name: 'Profile' })).toHaveAttribute('href', '/flow/u/7');
  });
  it('clears stale projects when the current list is empty', () => {
    useFlowStore.setState({
      projects: [
        {
          id: 'old',
          name: 'Old',
          color: '#111',
          taskCount: 1,
          completedCount: 0,
          memberAvatars: [],
          health: 'on_track',
          progress: 0,
          startDate: '',
          targetDate: '',
          lead: '',
        },
      ],
    });
    render(
      <FlowLayout>
        <p>Flow page</p>
      </FlowLayout>
    );
    expect(useFlowStore.getState().projects).toEqual([]);
  });
  it('hides creation controls for a read-only member', () => {
    useUserStore.setState({
      user: {
        id: 7,
        businessId: 42,
        effectiveAccess: createEffectiveAccess('flow', { permissions: ['flow:tasks:view'] }),
      },
    });
    render(
      <FlowLayout>
        <p>Flow page</p>
      </FlowLayout>
    );
    expect(screen.queryByRole('button', { name: 'New task' })).toBeNull();
    expect(screen.queryByRole('button', { name: 'New project' })).toBeNull();
  });
  it('creates a project with the edited fields and visibility', () => {
    render(
      <FlowLayout>
        <p>Flow page</p>
      </FlowLayout>
    );
    fireEvent.click(screen.getByRole('button', { name: 'New project' }));
    const inputs = screen.getAllByRole('textbox');
    fireEvent.change(inputs[0], { target: { value: '  Launch  ' } });
    fireEvent.change(inputs[1], { target: { value: 'Release planning' } });
    fireEvent.click(screen.getByRole('radio', { name: 'Public' }));
    fireEvent.click(screen.getByRole('button', { name: 'Create', exact: true }));
    expect(mocks.create).toHaveBeenCalledWith(
      expect.objectContaining({
        name: 'Launch',
        description: 'Release planning',
        visibility: 'public',
      }),
      expect.any(Object)
    );
  });
});
