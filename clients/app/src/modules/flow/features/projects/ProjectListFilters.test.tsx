import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import ProjectList from './ProjectList';
import { initTestI18n } from '../../shared/testing/i18n';

// The list reads the same `?filter=` as the board: one search object that
// navigate replaces and useSearch re-reads.
const router = vi.hoisted(() => {
  let search: Record<string, unknown> = {};
  const listeners = new Set<() => void>();
  return {
    get: () => search,
    set: (next: Record<string, unknown>) => {
      search = next;
      listeners.forEach((l) => l());
    },
    subscribe: (l: () => void) => {
      listeners.add(l);
      return () => listeners.delete(l);
    },
  };
});

const saveAsView = vi.hoisted(() => ({ save: vi.fn(), pending: false }));
vi.mock('../views/useSaveFiltersAsView', () => ({
  projectFiltersForView: (filters: unknown) => filters,
  useSaveFiltersAsView: () => saveAsView,
}));
vi.mock('@tanstack/react-router', async () => {
  const React = await import('react');
  return {
    useParams: () => ({ projectId: '7' }),
    useLocation: () => ({ pathname: '/flow/p/7/list' }),
    useSearch: () => React.useSyncExternalStore(router.subscribe, router.get),
    useNavigate:
      () => (opts: { search: (prev: Record<string, unknown>) => Record<string, unknown> }) =>
        router.set(opts.search(router.get())),
    Link: ({ children }: React.PropsWithChildren) => <>{children}</>,
  };
});
vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('@voxire/config', () => ({
  useFeatureFlag: () => false,
  FLAG_KEYS: { APP_FLOW_PREVIEW_ROUTES: 'app.flow.previewRoutes' },
}));
vi.mock('../../shared/lib/currentUser', () => ({
  useCanEditTask: () => true,
  useCanCreateTask: () => true,
  useCanAdministerProject: () => true,
  useCurrentUserId: () => '1',
}));
vi.mock('../../shared/stores/userStore', () => ({
  useUserStore: (selector: (state: unknown) => unknown) =>
    selector({ user: { effectiveAccess: { isOwner: true, subscribedApps: ['flow'] } } }),
}));
vi.mock('../../shared/lib/useOpenTask', () => ({ useOpenTask: () => vi.fn() }));
vi.mock('../../shared/lib/useProjectRealtime', () => ({ useProjectRealtime: () => undefined }));
vi.mock('./useCurrentProject', () => ({
  useCurrentProject: () => ({
    project: { id: '7', name: 'Investment Hub', color: '#6366f1' },
    isLoading: false,
    loadFailed: false,
  }),
}));
vi.mock('./useProjectMembers', () => ({
  useProjectMembers: () => ({ data: [{ userId: '1', name: 'Abed Amouneh' }] }),
}));
vi.mock('../cycles/useCycles', () => ({ useCycles: () => ({ data: [] }) }));
vi.mock('../milestones/useMilestones', () => ({ useMilestones: () => ({ data: [] }) }));
vi.mock('../tasks/useLabels', () => ({ useLabels: () => ({ data: [] }) }));
vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock('./useStatuses', () => ({
  useProjectStatuses: () => ({
    isPending: false,
    isError: false,
    data: [
      {
        id: '8',
        projectId: '7',
        name: 'To do',
        bucket: 'backlog',
        legacyStatus: 'todo',
        color: '#6366f1',
        sortOrder: 0,
        isDefault: true,
        createdAt: '',
      },
    ],
  }),
  useSetTaskStatus: () => ({ mutate: vi.fn() }),
}));
vi.mock('../tasks/useTasks', () => ({
  useTasks: () => ({
    isLoading: false,
    isPending: false,
    isError: false,
    data: [
      {
        id: '1',
        projectId: '7',
        statusId: '8',
        title: 'Payment plan tracker',
        status: 'todo',
        priority: 'high',
        assigneeId: 1,
        assigneeName: 'Abed Amouneh',
        identifier: 'T-1',
      },
      {
        id: '2',
        projectId: '7',
        statusId: '8',
        title: 'Document vault',
        status: 'todo',
        priority: 'low',
        identifier: 'T-2',
      },
    ],
  }),
}));

const renderList = () =>
  render(
    <QueryClientProvider client={new QueryClient()}>
      <ProjectList />
    </QueryClientProvider>
  );

beforeAll(async () => {
  await initTestI18n();
});

beforeEach(() => {
  localStorage.clear();
  sessionStorage.clear();
  router.set({});
});

describe('Project list filters', () => {
  it('shares the filter bar with the board and narrows the rows', () => {
    renderList();
    expect(screen.getByText('Payment plan tracker')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: /only mine/i }));
    expect(screen.queryByText('Document vault')).toBeNull();
    expect(screen.getByText('1 of 2 tasks visible')).toBeTruthy();
  });

  it('opens with the project filter the board left behind in the same visit', () => {
    sessionStorage.setItem('flow:filters:project:7', JSON.stringify({ priorities: ['low'] }));
    renderList();
    expect(screen.queryByText('Payment plan tracker')).toBeNull();
    expect(screen.getByText('Document vault')).toBeTruthy();
  });

  it('says nothing matches instead of calling the project empty', () => {
    router.set({ filter: { query: 'no such task' } });
    renderList();
    expect(screen.getByText('No tasks match these filters.')).toBeTruthy();
  });

  it('fades rows that do not match while dimming', () => {
    localStorage.setItem('flow:filters:dim', '1');
    router.set({ filter: { priorities: ['high'] } });
    renderList();
    expect(document.querySelectorAll('[data-filter-dimmed]')).toHaveLength(1);
  });
});
