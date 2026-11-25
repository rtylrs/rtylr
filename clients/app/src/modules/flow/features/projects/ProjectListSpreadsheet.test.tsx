import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { initTestI18n } from '../../shared/testing/i18n';
import type { ApiStatus, ApiTask, ApiTaskDetail } from '../../shared/services/flowApi';
import { useFlowStore } from '../../shared/stores/flowStore';
import ProjectList from './ProjectList';

const remote = vi.hoisted(() => ({
  tasks: [] as ApiTask[],
  detail: undefined as ApiTaskDetail | undefined,
  canEdit: true,
  updateTask: vi.fn(),
  setStatus: vi.fn(),
  setLabels: vi.fn(),
  openTask: vi.fn(),
  download: vi.fn(),
}));

vi.mock('../views/useSaveFiltersAsView', () => ({
  projectFiltersForView: (filters: unknown) => filters,
  useSaveFiltersAsView: () => ({ save: vi.fn(), pending: false }),
}));
// Filtering has its own suites; here every task shows.
vi.mock('./useProjectTaskFilters', () => ({
  useProjectTaskFilters: (_projectId: string, tasks: unknown[]) => ({
    state: {
      filters: {
        assignees: [],
        priorities: [],
        labels: [],
        due: [],
        statuses: [],
        cycles: [],
        milestones: [],
        epics: [],
        types: [],
        projects: [],
        query: '',
      },
      setFilters: vi.fn(),
      clear: vi.fn(),
      active: false,
      dim: false,
      setDim: vi.fn(),
    },
    fields: [],
    people: [],
    currentUserId: null,
    matchingIds: null,
    shownTasks: tasks,
    dimmedIds: null,
  }),
}));
vi.mock('../../shared/lib/useOpenTask', () => ({ useOpenTask: () => remote.openTask }));
vi.mock('@tanstack/react-router', () => ({
  useLocation: () => ({ pathname: '/flow/p/7/list' }),
  useParams: () => ({ projectId: '7' }),
  useNavigate: () => vi.fn(),
  Link: ({ children }: React.PropsWithChildren) => <>{children}</>,
}));
vi.mock('../../shared/lib/currentUser', () => ({
  useCanEditTask: () => remote.canEdit,
  useCanCreateTask: () => true,
  useCanAdministerProject: () => true,
}));
vi.mock('../../shared/stores/userStore', () => ({
  useUserStore: (selector: (state: unknown) => unknown) =>
    selector({ user: { id: 42, effectiveAccess: { isOwner: true, subscribedApps: ['flow'] } } }),
}));
vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('../../shared/lib/useProjectRealtime', () => ({ useProjectRealtime: () => undefined }));
vi.mock('./useCurrentProject', () => ({
  useCurrentProject: () => ({
    project: { id: '7', name: 'Remote project', color: '#6366f1' },
    isLoading: false,
    loadFailed: false,
    refetch: vi.fn(),
  }),
}));
vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock('./useStatuses', () => ({
  useProjectStatuses: () => ({
    data: STATUSES,
    isPending: false,
    isError: false,
    refetch: vi.fn(),
  }),
  useSetTaskStatus: () => ({ mutate: vi.fn(), mutateAsync: remote.setStatus }),
}));
vi.mock('../tasks/useTasks', () => ({
  useTasks: () => ({ data: remote.tasks, isPending: false, isError: false, refetch: vi.fn() }),
  useUpdateTask: () => ({ mutateAsync: remote.updateTask }),
  useTaskDetail: () => ({ data: remote.detail, isPending: !remote.detail, isError: false }),
}));
vi.mock('../tasks/useLabels', () => ({
  useLabels: () => ({
    data: [
      { id: 1, name: 'Bug', color: '#ef4444', createdAt: '' },
      { id: 2, name: 'UI', color: '#6366f1', createdAt: '' },
    ],
  }),
  useSetTaskLabels: () => ({ mutateAsync: remote.setLabels }),
}));
vi.mock('../../../../shared/lib/csv', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../../../../shared/lib/csv')>()),
  downloadCsv: remote.download,
}));

const status = (id: string, name: string, sortOrder: number): ApiStatus => ({
  id,
  projectId: '7',
  name,
  bucket: 'active',
  legacyStatus: 'todo',
  color: '#6366f1',
  sortOrder,
  isDefault: sortOrder === 0,
  createdAt: '',
});
const STATUSES = [status('8', 'Ideas', 0), status('9', 'Release review', 10)];

const apiTask = (id: string, title: string, extra: Partial<ApiTask> = {}): ApiTask => ({
  id,
  projectId: '7',
  sectionId: '9',
  statusId: '9',
  title,
  status: 'todo',
  priority: 'medium',
  assignee: '',
  assigneeInitials: '',
  dueDate: '',
  subtaskCount: 0,
  projectName: 'Remote project',
  identifier: `T-${id}`,
  ...extra,
});

const PREFS_KEY = 'flow:list:columns:42:7';

const renderList = () =>
  render(
    <QueryClientProvider client={new QueryClient()}>
      <ProjectList />
    </QueryClientProvider>
  );

/** The cell of one column in the row of the task with this title. */
function cell(title: string, column: string): HTMLElement {
  const row = screen
    .getByRole('button', { name: new RegExp(`Open task.*${title}`) })
    .closest('[role="row"]');
  const found = row?.querySelector<HTMLElement>(`[data-column="${column}"]`);
  if (!found) throw new Error(`no ${column} cell for ${title}`);
  return found;
}

const rowTitles = (): string[] =>
  screen.getAllByRole('button', { name: /Open task/ }).map((button) => button.textContent ?? '');

beforeEach(async () => {
  await initTestI18n();
  vi.clearAllMocks();
  localStorage.clear();
  remote.canEdit = true;
  remote.detail = undefined;
  remote.updateTask.mockResolvedValue(undefined);
  remote.setStatus.mockResolvedValue(undefined);
  remote.setLabels.mockResolvedValue(undefined);
  remote.tasks = [
    apiTask('1', 'Bravo task'),
    apiTask('2', 'Alpha task', { dueDate: '2026-12-01' }),
  ];
  useFlowStore.setState({
    members: [{ id: '3', name: 'Grace Hopper', email: '', initials: 'GH' }],
  });
});

describe('Project list as a spreadsheet', () => {
  it('starts with the columns the List always had and remembers a picked one', () => {
    const { unmount } = renderList();
    const headers = () =>
      screen.getAllByRole('columnheader').map((h) => h.getAttribute('data-column'));
    expect(headers()).toEqual([null, 'priority', 'title', 'assignee', 'dueDate', null]);

    fireEvent.click(screen.getByRole('button', { name: 'Columns' }));
    fireEvent.click(screen.getByRole('checkbox', { name: 'Status' }));
    expect(headers()).toContain('status');
    expect(within(cell('Bravo task', 'status')).getByText('Release review')).toBeInTheDocument();
    expect(JSON.parse(localStorage.getItem(PREFS_KEY)!).columns).toContain('status');

    unmount();
    renderList();
    expect(headers()).toContain('status');
  });

  it('sorts by a header click, ascending then descending, with empty values last', () => {
    renderList();
    expect(rowTitles()).toEqual(['Bravo task', 'Alpha task']);

    fireEvent.click(
      within(screen.getByRole('columnheader', { name: /Title/ })).getByRole('button')
    );
    expect(rowTitles()).toEqual(['Alpha task', 'Bravo task']);
    expect(screen.getByRole('columnheader', { name: /Title/ })).toHaveAttribute(
      'aria-sort',
      'ascending'
    );

    const due = within(screen.getByRole('columnheader', { name: /Due Date/ })).getByRole('button');
    fireEvent.click(due);
    fireEvent.click(due);
    // Descending still leaves the task with no due date at the bottom.
    expect(rowTitles()).toEqual(['Alpha task', 'Bravo task']);
    expect(JSON.parse(localStorage.getItem(PREFS_KEY)!).sort).toEqual({
      column: 'dueDate',
      direction: 'desc',
    });
  });

  it('edits the priority in place with Enter, through the task update', async () => {
    renderList();
    const priority = cell('Bravo task', 'priority');
    priority.focus();
    fireEvent.keyDown(priority, { key: 'Enter' });
    fireEvent.click(await screen.findByRole('menuitem', { name: /Urgent/ }));

    expect(remote.updateTask).toHaveBeenCalledWith({ id: '1', data: { priority: 'urgent' } });
    expect(remote.openTask).not.toHaveBeenCalled();
  });

  it('assigns from the cell and changes the status through the status endpoint', async () => {
    localStorage.setItem(PREFS_KEY, JSON.stringify({ columns: ['title', 'status', 'assignee'] }));
    renderList();

    fireEvent.click(cell('Bravo task', 'assignee'));
    fireEvent.click(await screen.findByRole('menuitem', { name: 'Grace Hopper' }));
    expect(remote.updateTask).toHaveBeenCalledWith({ id: '1', data: { assigneeId: '3' } });

    fireEvent.click(cell('Bravo task', 'status'));
    fireEvent.click(await screen.findByRole('menuitem', { name: 'Ideas' }));
    expect(remote.setStatus).toHaveBeenCalledWith({ taskId: '1', statusId: '8' });
  });

  it('drops a date edit on Escape and saves it on Enter', () => {
    renderList();
    const due = cell('Bravo task', 'dueDate');
    fireEvent.keyDown(due, { key: 'Enter' });
    fireEvent.change(screen.getByLabelText('Due Date'), { target: { value: '2026-12-24' } });
    fireEvent.keyDown(screen.getByLabelText('Due Date'), { key: 'Escape' });
    expect(remote.updateTask).not.toHaveBeenCalled();
    expect(document.activeElement).toBe(cell('Bravo task', 'dueDate'));

    fireEvent.keyDown(cell('Bravo task', 'dueDate'), { key: 'Enter' });
    fireEvent.change(screen.getByLabelText('Due Date'), { target: { value: '2026-12-24' } });
    fireEvent.keyDown(screen.getByLabelText('Due Date'), { key: 'Enter' });
    expect(remote.updateTask).toHaveBeenCalledWith({ id: '1', data: { dueDate: '2026-12-24' } });
  });

  it('saves an estimate and clears it when emptied', () => {
    localStorage.setItem(PREFS_KEY, JSON.stringify({ columns: ['title', 'estimate'] }));
    remote.tasks = [apiTask('1', 'Bravo task', { estimatedHours: 3 })];
    renderList();
    expect(cell('Bravo task', 'estimate')).toHaveTextContent('3h');

    fireEvent.keyDown(cell('Bravo task', 'estimate'), { key: 'Enter' });
    fireEvent.change(screen.getByLabelText('Estimate'), { target: { value: '' } });
    fireEvent.keyDown(screen.getByLabelText('Estimate'), { key: 'Enter' });
    expect(remote.updateTask).toHaveBeenCalledWith({ id: '1', data: { estimatedHours: 0 } });
  });

  it('toggles several labels in one opening of the menu', async () => {
    localStorage.setItem(PREFS_KEY, JSON.stringify({ columns: ['title', 'labels'] }));
    renderList();
    fireEvent.keyDown(cell('Bravo task', 'labels'), { key: 'Enter' });
    fireEvent.click(await screen.findByRole('menuitemcheckbox', { name: 'Bug' }));
    fireEvent.click(screen.getByRole('menuitemcheckbox', { name: 'UI' }));

    expect(remote.setLabels).toHaveBeenNthCalledWith(1, { task_id: 1, label_ids: [1] });
    expect(remote.setLabels).toHaveBeenNthCalledWith(2, { task_id: 1, label_ids: [1, 2] });
    expect(screen.getByRole('menuitemcheckbox', { name: 'Bug' })).toHaveAttribute(
      'aria-checked',
      'true'
    );
  });

  it('gives viewers read-only cells that open the task instead', () => {
    remote.canEdit = false;
    renderList();
    const priority = cell('Bravo task', 'priority');
    expect(priority).toHaveAttribute('aria-readonly', 'true');
    fireEvent.keyDown(priority, { key: 'Enter' });
    expect(screen.queryByRole('menu')).not.toBeInTheDocument();
    fireEvent.click(priority);
    expect(remote.openTask).toHaveBeenCalledWith('1');
  });

  it('opens a parent row to show its subtasks under it', () => {
    remote.tasks = [apiTask('1', 'Bravo task', { subtaskCount: 2 })];
    remote.detail = {
      task: remote.tasks[0],
      subtasks: [
        apiTask('11', 'Child one', { assigneeId: '3' }),
        apiTask('12', 'Child two', { archived: true }),
      ],
      comments: [],
      activity: [],
    } as unknown as ApiTaskDetail;
    renderList();
    const parentRow = cell('Bravo task', 'title').closest('[role="row"]')!;
    expect(parentRow).toHaveAttribute('aria-expanded', 'false');

    fireEvent.click(screen.getByRole('button', { name: 'Show 2 subtasks' }));
    expect(parentRow).toHaveAttribute('aria-expanded', 'true');
    const child = cell('Child one', 'title').closest('[role="row"]')!;
    expect(child).toHaveAttribute('aria-level', '2');
    // A subtask without a display name still shows its assignee.
    expect(
      within(child as HTMLElement).getByRole('img', { name: 'Grace Hopper' })
    ).toBeInTheDocument();
    expect(screen.queryByText('Child two')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Hide subtasks' }));
    expect(screen.queryByText('Child one')).not.toBeInTheDocument();
  });

  it('moves between cells with the arrow keys and keeps the column through a group header', () => {
    renderList();
    const start = cell('Bravo task', 'title');
    start.focus();
    fireEvent.keyDown(start, { key: 'ArrowRight' });
    expect(document.activeElement).toBe(cell('Bravo task', 'assignee'));

    fireEvent.keyDown(document.activeElement!, { key: 'ArrowDown' });
    expect(document.activeElement).toBe(cell('Alpha task', 'assignee'));

    fireEvent.keyDown(document.activeElement!, { key: 'ArrowUp' });
    fireEvent.keyDown(document.activeElement!, { key: 'ArrowUp' });
    // The group header is one wide cell...
    expect(document.activeElement).toHaveAccessibleName(/Release review/);
    fireEvent.keyDown(document.activeElement!, { key: 'ArrowUp' });
    // ...and the column header above it is the column the user was in.
    expect(document.activeElement).toHaveAttribute('data-column', 'assignee');
    expect(document.activeElement).toHaveAttribute('role', 'columnheader');
  });

  it('exports the rows on screen as CSV', () => {
    renderList();
    fireEvent.click(screen.getByRole('button', { name: 'Export CSV' }));

    expect(remote.download).toHaveBeenCalledOnce();
    const [name, csv] = remote.download.mock.calls[0] as [string, string];
    expect(name).toMatch(/^Remote project \d{4}-\d{2}-\d{2}\.csv$/);
    const lines = csv.trim().split('\r\n');
    expect(lines).toHaveLength(3);
    expect(lines[1]).toContain('"Bravo task"');
  });
});
