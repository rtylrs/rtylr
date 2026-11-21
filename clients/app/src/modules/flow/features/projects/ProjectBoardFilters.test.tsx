import React from 'react';
import { act, fireEvent, render, screen, within } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import ProjectBoard from './ProjectBoard';
import { initTestI18n } from '../../shared/testing/i18n';

const mutation = vi.hoisted(() => ({ mutate: vi.fn(), isPending: false }));

// A tiny router: the board's filters live in `?filter=`, so the mock keeps one
// search object and re-renders whoever reads it when navigate replaces it.
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
    useParams: () => ({ projectId: '2' }),
    useLocation: () => ({ pathname: '/flow/p/2' }),
    useSearch: () => React.useSyncExternalStore(router.subscribe, router.get),
    useNavigate:
      () => (opts: { search: (prev: Record<string, unknown>) => Record<string, unknown> }) =>
        router.set(opts.search(router.get())),
    Link: ({ children, to, ...props }: React.PropsWithChildren<{ to: string }>) => (
      <a href={to} {...props}>
        {children}
      </a>
    ),
  };
});
vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('@voxire/config', () => ({
  useFeatureFlag: () => false,
  FLAG_KEYS: { APP_FLOW_PREVIEW_ROUTES: 'app.flow.previewRoutes' },
}));
vi.mock('../../shared/lib/currentUser', () => ({
  useCanCreateTask: () => true,
  useCanEditTask: () => true,
  useCanDeleteResource: () => true,
  useCanAdministerProject: () => true,
  useCurrentUserId: () => '1',
}));
vi.mock('../../shared/lib/useOpenTask', () => ({ useOpenTask: () => vi.fn() }));
vi.mock('../../shared/lib/useProjectRealtime', () => ({ useProjectRealtime: () => undefined }));
vi.mock('./useCurrentProject', () => ({
  useCurrentProject: () => ({
    project: { id: '2', name: 'Investment Hub', color: '#6366f1' },
    isLoading: false,
    loadFailed: false,
  }),
}));
vi.mock('./useProjectMembers', () => ({
  useProjectMembers: () => ({
    data: [
      { userId: '1', name: 'Abed Amouneh' },
      { userId: '2', name: 'Ahmad Tarabein' },
    ],
  }),
}));
vi.mock('../cycles/useCycles', () => ({ useCycles: () => ({ data: [] }) }));
vi.mock('../milestones/useMilestones', () => ({ useMilestones: () => ({ data: [] }) }));
// Cards render the blocked badge, whose relations hook reads the Flow API.
vi.mock('../tasks/useTaskRelations', () => ({
  useTaskRelations: () => ({ data: undefined, isPending: true }),
}));
vi.mock('../tasks/useLabels', () => ({
  useLabels: () => ({ data: [{ id: 7, name: 'Design', color: '#f00', createdAt: '' }] }),
}));
vi.mock('./useColumns', () => ({ useMoveTaskToColumn: () => mutation }));

const status = (id: string, name: string, sortOrder: number) => ({
  id,
  projectId: '2',
  name,
  bucket: 'active',
  legacyStatus: 'todo',
  color: '#6366f1',
  sortOrder,
  isDefault: sortOrder === 0,
  createdAt: '',
});
vi.mock('./useStatuses', () => ({
  useProjectStatuses: () => ({
    isPending: false,
    isError: false,
    data: [status('9', 'To do', 0), status('10', 'Doing', 1)],
  }),
  useCreateStatus: () => mutation,
  useUpdateStatus: () => mutation,
  useReorderStatuses: () => mutation,
  useDeleteStatus: () => mutation,
}));

const apiTask = (
  id: string,
  title: string,
  statusId: string,
  extra: Record<string, unknown> = {}
) => ({
  id,
  projectId: '2',
  statusId,
  title,
  status: 'todo',
  priority: 'medium',
  identifier: `T-${id}`,
  dueDate: '',
  subtaskCount: 0,
  ...extra,
});
vi.mock('../tasks/useTasks', () => ({
  useTasks: () => ({
    isLoading: false,
    isError: false,
    data: [
      apiTask('1', 'Payment plan tracker', '9', { assigneeId: 1, assigneeName: 'Abed Amouneh' }),
      apiTask('2', 'Document vault', '10', {
        assigneeId: 2,
        assigneeName: 'Ahmad Tarabein',
        labels: ['7'],
      }),
      apiTask('3', 'Advisor chat', '10', { priority: 'high' }),
    ],
  }),
}));

const renderBoard = () =>
  render(
    <QueryClientProvider client={new QueryClient()}>
      <ProjectBoard />
    </QueryClientProvider>
  );

const cardTitles = (): string[] =>
  [...document.querySelectorAll<HTMLElement>('[data-board-card]')]
    .filter((el) => !el.hasAttribute('data-filter-dimmed'))
    .map((el) => el.dataset.boardCard ?? '');

beforeAll(async () => {
  await initTestI18n();
});

beforeEach(() => {
  localStorage.clear();
  router.set({});
});

describe('Board filters', () => {
  it('searches title, identifier and label names, and says how many are left', () => {
    renderBoard();
    const search = screen.getByRole('textbox', { name: /search tasks by title/i });

    fireEvent.change(search, { target: { value: 'vault' } });
    expect(cardTitles()).toEqual(['2']);
    expect(screen.getByText('1 of 3 tasks visible')).toBeTruthy();

    fireEvent.change(search, { target: { value: 'T-3' } });
    expect(cardTitles()).toEqual(['3']);

    fireEvent.change(search, { target: { value: 'design' } });
    expect(cardTitles()).toEqual(['2']);
  });

  it('shows one person, then several, from the avatar strip', () => {
    renderBoard();
    const strip = screen.getByRole('group', { name: /show tasks by person/i });

    fireEvent.click(within(strip).getByRole('button', { name: /ahmad tarabein/i }));
    expect(cardTitles()).toEqual(['2']);

    fireEvent.click(within(strip).getByRole('button', { name: /unassigned/i }));
    expect(cardTitles().sort()).toEqual(['2', '3']);
  });

  it('makes Only mine and my avatar the same switch', () => {
    renderBoard();
    const mine = screen.getByRole('button', { name: /only mine/i });
    fireEvent.click(mine);
    expect(cardTitles()).toEqual(['1']);
    expect(mine.getAttribute('aria-pressed')).toBe('true');

    const strip = screen.getByRole('group', { name: /show tasks by person/i });
    fireEvent.click(within(strip).getByRole('button', { name: /ahmad tarabein/i }));
    expect(screen.getByRole('button', { name: /only mine/i }).getAttribute('aria-pressed')).toBe(
      'false'
    );
  });

  it('puts the filter in the URL and remembers it for the project', () => {
    renderBoard();
    fireEvent.click(screen.getByRole('button', { name: /only mine/i }));
    expect(router.get().filter).toEqual({ assignees: ['1'] });
    expect(JSON.parse(localStorage.getItem('flow:filters:project:2') ?? '{}')).toEqual({
      assignees: ['1'],
    });
  });

  it('opens filtered when the link carries a filter', () => {
    router.set({ filter: { priorities: ['high'] } });
    renderBoard();
    expect(cardTitles()).toEqual(['3']);
    expect(screen.getByRole('button', { name: /priority: high/i })).toBeTruthy();
  });

  it('reopens with the remembered filter when the URL has none', () => {
    localStorage.setItem('flow:filters:project:2', JSON.stringify({ priorities: ['high'] }));
    renderBoard();
    expect(cardTitles()).toEqual(['3']);
    expect(router.get().filter).toEqual({ priorities: ['high'] });
  });

  it('hides the columns a status filter leaves out', () => {
    router.set({ filter: { statuses: ['10'] } });
    renderBoard();
    expect(document.querySelector('[data-board-column="9"]')).toBeNull();
    expect(document.querySelector('[data-board-column="10"]')).not.toBeNull();
  });

  it('fades instead of hiding when dimming, and counts matches per column', () => {
    localStorage.setItem('flow:filters:dim', '1');
    router.set({ filter: { priorities: ['high'] } });
    renderBoard();
    expect(document.querySelectorAll('[data-board-card]')).toHaveLength(3);
    expect(document.querySelectorAll('[data-filter-dimmed]')).toHaveLength(2);
    const doing = document.querySelector('[data-board-column="10"]') as HTMLElement;
    expect(within(doing).getByText('1/2')).toBeTruthy();
  });

  it('keeps Q working under a non-modal dialog such as the cookie banner', () => {
    const banner = document.createElement('div');
    banner.setAttribute('role', 'dialog');
    banner.setAttribute('aria-modal', 'false');
    document.body.appendChild(banner);
    try {
      renderBoard();
      fireEvent.keyDown(document.body, { key: 'q' });
      expect(cardTitles()).toEqual(['1']);
    } finally {
      banner.remove();
    }
  });

  it('saves the current filters as a view named after the project', () => {
    router.set({ filter: { priorities: ['high'] } });
    renderBoard();
    fireEvent.click(screen.getByRole('button', { name: /save as view/i }));
    expect(saveAsView.save).toHaveBeenCalledWith(
      'Investment Hub: filtered',
      expect.objectContaining({ priorities: ['high'] })
    );
  });

  it('removes one value with its chip and everything with Shift+Escape', () => {
    router.set({ filter: { priorities: ['high'], labels: ['7'] } });
    renderBoard();
    expect(cardTitles()).toEqual([]);

    fireEvent.click(screen.getByRole('button', { name: /remove filter: label: design/i }));
    expect(cardTitles()).toEqual(['3']);

    act(() => {
      fireEvent.keyDown(document.body, { key: 'Escape', shiftKey: true });
    });
    expect(cardTitles().sort()).toEqual(['1', '2', '3']);
    expect(router.get().filter).toBeUndefined();
  });
});
