import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
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
// Stable arrays, as react-query hands back the same data between renders.
const plans = vi.hoisted(() => ({
  cycles: { data: [{ id: '4', name: 'Sprint 4' }] },
  milestones: { data: [{ id: '5', name: 'Beta launch' }] },
  labels: {
    data: [
      { id: 7, name: 'Design', color: '#f00', createdAt: '' },
      { id: 8, name: 'Backend', color: '#00f', createdAt: '' },
    ],
  },
}));
vi.mock('../cycles/useCycles', () => ({ useCycles: () => plans.cycles }));
vi.mock('../milestones/useMilestones', () => ({ useMilestones: () => plans.milestones }));
vi.mock('../tasks/useLabels', () => ({ useLabels: () => plans.labels }));
vi.mock('./useColumns', () => ({ useMoveTaskToColumn: () => mutation }));

const DAY = 24 * 60 * 60 * 1000;
const daysAgo = (n: number): string => new Date(Date.now() - n * DAY).toISOString();

const status = (id: string, name: string, sortOrder: number, bucket = 'active') => ({
  id,
  projectId: '2',
  name,
  bucket,
  legacyStatus: bucket === 'done' ? 'done' : 'todo',
  color: '#6366f1',
  sortOrder,
  isDefault: sortOrder === 0,
  createdAt: '',
});
vi.mock('./useStatuses', () => ({
  useProjectStatuses: () => ({
    isPending: false,
    isError: false,
    data: [status('9', 'To do', 0), status('11', 'Done', 1, 'done')],
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
const tasks = vi.hoisted(() => ({ data: [] as unknown[] }));
vi.mock('../tasks/useTasks', () => ({
  useTasks: () => ({ isLoading: false, isError: false, data: tasks.data }),
}));

const renderBoard = () =>
  render(
    <QueryClientProvider client={new QueryClient()}>
      <ProjectBoard />
    </QueryClientProvider>
  );

const cardIds = (): string[] =>
  [...document.querySelectorAll<HTMLElement>('[data-board-card]')].map(
    (el) => el.dataset.boardCard ?? ''
  );

const openSettings = (): HTMLElement => {
  fireEvent.click(screen.getByRole('button', { name: /open board settings/i }));
  return screen.getByRole('dialog', { name: /board settings/i });
};

const stored = () => JSON.parse(localStorage.getItem('flow:board:view:1:2') ?? 'null');

beforeAll(async () => {
  await initTestI18n();
});

beforeEach(() => {
  localStorage.clear();
  router.set({});
  tasks.data = [
    apiTask('1', 'Payment plan tracker', '9', { labels: ['7', '8'], cycleId: '4' }),
    apiTask('2', 'Document vault', '9', { labels: ['8'], milestoneId: '5', estimatedHours: 2 }),
    apiTask('3', 'Advisor chat', '9'),
    // Half a day, not one: a card closed exactly a day ago is already past the
    // one-day window by the time the board reads the clock.
    apiTask('4', 'Closed this morning', '11', { status: 'done', completedAt: daysAgo(0.5) }),
    apiTask('5', 'Closed ten days ago', '11', { status: 'done', completedAt: daysAgo(10) }),
    apiTask('6', 'Closed last month', '11', { status: 'done', completedAt: daysAgo(40) }),
  ];
});

describe('Board view settings', () => {
  it('opens on today board: 14 days of done cards, the usual card, fixed columns', () => {
    const { container } = renderBoard();

    expect(cardIds().sort()).toEqual(['1', '2', '3', '4', '5']);
    expect(screen.getByText('T-1')).toBeTruthy();
    expect(container.querySelector('[data-card-field]')).toBeNull();
    expect(container.querySelector('[data-board-width]')?.getAttribute('data-board-width')).toBe(
      'fixed'
    );
  });

  it('hides done cards after the chosen window, or never', () => {
    renderBoard();
    const panel = openSettings();

    fireEvent.click(within(panel).getByRole('radio', { name: /after 1 day/i }));
    expect(cardIds().sort()).toEqual(['1', '2', '3', '4']);

    fireEvent.click(within(panel).getByRole('radio', { name: /never/i }));
    expect(cardIds().sort()).toEqual(['1', '2', '3', '4', '5', '6']);
    expect(stored().hideDoneAfterDays).toBeNull();
  });

  it('turns card fields on and off for this person only', () => {
    const { container } = renderBoard();
    const panel = openSettings();

    fireEvent.click(within(panel).getByRole('switch', { name: /^key$/i }));
    fireEvent.click(within(panel).getByRole('switch', { name: /estimate/i }));
    fireEvent.click(within(panel).getByRole('switch', { name: /^cycle$/i }));

    expect(screen.queryByText('T-1')).toBeNull();
    expect(container.querySelector('[data-card-field="estimate"]')?.textContent).toBe('2h');
    expect(container.querySelector('[data-card-field="cycle"]')?.textContent).toBe('Sprint 4');
    expect(stored().cardFields).toMatchObject({ key: false, estimate: true, cycle: true });
    expect(localStorage.getItem('flow:board:view:5:2')).toBeNull();
  });

  it('groups by label with a card in each of its labels, counted once per column', () => {
    renderBoard();
    const panel = openSettings();
    fireEvent.click(within(panel).getByRole('radio', { name: /^label$/i }));
    fireEvent.click(within(panel).getByRole('button', { name: /close board settings/i }));

    const column = document.querySelector<HTMLElement>('[data-board-column="9"]')!;
    const lanes = [...column.querySelectorAll('.uppercase')].map((el) => el.textContent);
    expect(lanes).toEqual(['Design·1', 'Backend·2', 'No label·1']);
    expect(within(column).getAllByText('Payment plan tracker')).toHaveLength(2);
    expect(within(column).getByText('3')).toBeTruthy();
  });

  it('groups by cycle and by milestone from the toolbar', () => {
    renderBoard();
    fireEvent.click(screen.getByRole('button', { name: /group:/i }));
    fireEvent.click(screen.getByRole('button', { name: /^cycle$/i }));
    const column = document.querySelector<HTMLElement>('[data-board-column="9"]')!;
    expect([...column.querySelectorAll('.uppercase')].map((el) => el.textContent)).toEqual([
      'Sprint 4·1',
      'No cycle·2',
    ]);

    fireEvent.click(screen.getByRole('button', { name: /group:/i }));
    fireEvent.click(screen.getByRole('button', { name: /^milestone$/i }));
    expect([...column.querySelectorAll('.uppercase')].map((el) => el.textContent)).toEqual([
      'Beta launch·1',
      'No milestone·2',
    ]);
    expect(stored().groupBy).toBe('milestone');
  });

  it('lets the columns share the width', () => {
    const { container } = renderBoard();
    fireEvent.click(within(openSettings()).getByRole('radio', { name: /flexible/i }));

    expect(container.querySelector('[data-board-width]')?.getAttribute('data-board-width')).toBe(
      'flexible'
    );
    expect(document.querySelector('[data-board-column="9"]')?.className).toContain('flex-1');
  });

  it('reads back what this person chose on this project', () => {
    localStorage.setItem(
      'flow:board:view:1:2',
      JSON.stringify({ groupBy: 'none', hideDoneAfterDays: 30, columnWidth: 'fixed' })
    );
    renderBoard();
    expect(cardIds().sort()).toEqual(['1', '2', '3', '4', '5']);
  });
});
