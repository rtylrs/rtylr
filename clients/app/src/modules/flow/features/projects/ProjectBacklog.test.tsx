// @vitest-environment happy-dom
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, within } from '@testing-library/react';

vi.mock('react-i18next', async (importOriginal) => ({
  ...(await importOriginal<typeof import('react-i18next')>()),
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, unknown>) =>
      opts && 'name' in opts ? `${key}:${String(opts.name)}` : key,
    i18n: { language: 'en' },
  }),
}));
vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ projectId: 'p1' }),
  useNavigate: () => vi.fn(),
}));
vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('./useProjectTaskFilters', () => ({
  useProjectTaskFilters: (_projectId: string, tasks: unknown[]) => ({
    state: {},
    fields: [],
    people: [],
    currentUserId: null,
    matchingIds: null,
    shownTasks: tasks,
    dimmedIds: null,
  }),
}));
vi.mock('../../shared/filters/FilterBar', () => ({ default: () => null }));
vi.mock('./ProjectViewTabs', () => ({
  default: ({ rightSlot }: { rightSlot?: React.ReactNode }) => <div>{rightSlot}</div>,
}));
vi.mock('../../shared/components/TaskRow', () => ({
  default: ({ task }: { task: { id: string; title: string } }) => (
    <button type="button" data-testid={`row-${task.id}`}>
      {task.title}
    </button>
  ),
}));
vi.mock('../../shared/lib/useProjectRealtime', () => ({ useProjectRealtime: () => undefined }));

const store = vi.hoisted(() => ({ setQuickCreateOpen: vi.fn() }));
vi.mock('../../shared/stores/flowStore', () => ({
  useFlowStore: Object.assign(
    (selector?: (state: object) => unknown) => (selector ? selector({}) : {}),
    { getState: () => store }
  ),
}));
vi.mock('../../shared/lib/currentUser', () => ({
  useCanEditTask: () => true,
  useCanCreateTask: () => true,
}));
vi.mock('./useCurrentProject', () => ({
  useCurrentProject: () => ({
    project: { id: 'p1', name: 'Web', color: '#000' },
    isLoading: false,
    loadFailed: false,
    refetch: vi.fn(),
  }),
}));
vi.mock('./useStatuses', () => ({ useProjectStatuses: () => ({ data: [] }) }));

const data = vi.hoisted(() => ({
  tasks: [] as Record<string, unknown>[],
  cycles: [] as Record<string, unknown>[],
}));
vi.mock('../tasks/useTasks', () => ({
  useTasks: () => ({ data: data.tasks, isPending: false, refetch: vi.fn() }),
}));

const mutations = vi.hoisted(() => ({
  move: vi.fn(),
  start: vi.fn(),
  complete: vi.fn(),
  create: vi.fn(),
}));
vi.mock('../cycles/useCycles', () => ({
  useCycles: () => ({ data: data.cycles, isPending: false, refetch: vi.fn() }),
  useCycleCapacity: (id?: string) => ({
    isPending: false,
    data:
      id === 'active'
        ? {
            leaveApplied: true,
            members: [{ userId: '7', name: 'Rana', availableMinutes: 600, committedMinutes: 570 }],
          }
        : { members: [] },
  }),
  useMoveCycleTask: () => ({ mutate: mutations.move }),
  useStartCycle: () => ({ mutate: mutations.start, isPending: false }),
  useCompleteCycle: () => ({ mutate: mutations.complete, isPending: false }),
  useCreateCycle: () => ({ mutate: mutations.create, isPending: false }),
}));

import ProjectBacklog from './ProjectBacklog';

const apiTask = (id: string, patch: Record<string, unknown> = {}) => ({
  id,
  projectId: 'p1',
  title: `Task ${id}`,
  status: 'todo',
  priority: 'medium',
  createdAt: `2026-11-0${id}T00:00:00Z`,
  ...patch,
});

function section(name: string): HTMLElement {
  return screen.getByRole('region', { name });
}

describe('ProjectBacklog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    data.cycles = [
      {
        id: 'active',
        name: 'Sprint 4',
        status: 'active',
        startDate: '2026-11-23',
        endDate: '2026-12-04',
        goal: 'Ship planning',
        carriedOver: [{ fromCycleId: 'old', fromCycleName: 'Sprint 3', taskIds: ['1'] }],
      },
      { id: 'next', name: 'Sprint 5', status: 'upcoming', startDate: '2026-12-07' },
      { id: 'old', name: 'Sprint 3', status: 'completed', startDate: '2026-11-09' },
    ];
    data.tasks = [
      apiTask('1', { cycleId: 'active', planRank: '1024', effortMinutes: 570 }),
      apiTask('2', { cycleId: 'active', planRank: '2048', status: 'done' }),
      apiTask('3', { cycleId: 'next', planRank: '1024' }),
      apiTask('4'),
      apiTask('5', { planRank: '-1024' }),
    ];
  });

  it('shows the active cycle, upcoming cycles and the backlog with their details', () => {
    render(<ProjectBacklog />);
    const regions = screen.getAllByRole('region');
    expect(regions.map((r) => within(r).getByRole('heading').textContent)).toEqual([
      'Sprint 4',
      'Sprint 5',
      'flow.backlog.unplanned',
    ]);
    const active = section('Sprint 4');
    expect(within(active).getByText('Ship planning')).toBeInTheDocument();
    expect(within(active).getByText('flow.backlog.carriedOver:Sprint 3')).toBeInTheDocument();
    // 570 of 600 minutes is 95%: amber.
    expect(active.querySelector('[data-capacity-tone]')?.getAttribute('data-capacity-tone')).toBe(
      'warn'
    );
    const backlogRows = within(section('flow.backlog.unplanned'))
      .getAllByTestId(/^row-/)
      .map((r) => r.textContent);
    expect(backlogRows).toEqual(['Task 5', 'Task 4']);
    expect(screen.queryByText('Sprint 3', { selector: 'h2' })).not.toBeInTheDocument();
  });

  it('does not start a second cycle while one is active', () => {
    render(<ProjectBacklog />);
    const start = within(section('Sprint 5')).getByRole('button', {
      name: 'flow.backlog.startCycle',
    });
    expect(start).toBeDisabled();
    expect(within(section('Sprint 5')).getByText('flow.backlog.startBlocked')).toBeInTheDocument();
  });

  it('starts an upcoming cycle with a length and a goal', () => {
    data.cycles = data.cycles.filter((c) => c.id !== 'active');
    render(<ProjectBacklog />);
    fireEvent.click(
      within(section('Sprint 5')).getByRole('button', { name: 'flow.backlog.startCycle' })
    );
    const dialog = screen.getByRole('dialog');
    fireEvent.change(within(dialog).getByLabelText('flow.backlog.start.goal'), {
      target: { value: 'Close the epic' },
    });
    fireEvent.click(within(dialog).getByRole('button', { name: 'flow.backlog.start.confirm' }));
    expect(mutations.start).toHaveBeenCalledWith(
      { id: 'next', lengthWeeks: 2, goal: 'Close the epic' },
      expect.anything()
    );
  });

  it('completes a cycle sending unfinished tasks to the next cycle by default', () => {
    render(<ProjectBacklog />);
    fireEvent.click(
      within(section('Sprint 4')).getByRole('button', { name: 'flow.backlog.completeCycle' })
    );
    const dialog = screen.getByRole('dialog');
    // Task 2 is done, so only task 1 needs a decision.
    expect(within(dialog).getByText('Task 1')).toBeInTheDocument();
    expect(within(dialog).queryByText('Task 2')).not.toBeInTheDocument();
    fireEvent.click(within(dialog).getByRole('button', { name: 'flow.backlog.complete.confirm' }));
    expect(mutations.complete).toHaveBeenCalledWith(
      { id: 'active', decisions: [{ taskId: '1', decision: 'move', toCycleId: 'next' }] },
      expect.anything()
    );
  });

  it('moves a task one place with Alt and an arrow key', () => {
    render(<ProjectBacklog />);
    const row = within(section('Sprint 4')).getByTestId('row-1');
    fireEvent.keyDown(row, { key: 'ArrowDown', altKey: true });
    expect(mutations.move).toHaveBeenCalledWith({
      taskId: '1',
      cycleId: 'active',
      afterTaskId: '2',
      optimisticRank: 3072,
    });
  });

  it('creates a task straight into a section through quick create', () => {
    render(<ProjectBacklog />);
    fireEvent.click(
      within(section('Sprint 5')).getByRole('button', { name: /flow.backlog.create$/ })
    );
    expect(store.setQuickCreateOpen).toHaveBeenCalledWith(true, undefined, 'next');
    fireEvent.click(
      within(section('flow.backlog.unplanned')).getByRole('button', {
        name: /flow.backlog.create$/,
      })
    );
    expect(store.setQuickCreateOpen).toHaveBeenLastCalledWith(true, undefined, undefined);
  });

  it('creates the next cycle after the last open one', () => {
    data.cycles = [
      {
        id: 'next',
        name: 'Sprint 5',
        status: 'upcoming',
        startDate: '2026-12-07',
        endDate: '2026-12-18',
      },
    ];
    render(<ProjectBacklog />);
    fireEvent.click(screen.getByRole('button', { name: /flow.backlog.createCycle/ }));
    expect(mutations.create).toHaveBeenCalledWith(
      expect.objectContaining({ projectId: 'p1', startDate: '2026-12-19', endDate: '2027-01-01' }),
      expect.anything()
    );
  });
});
