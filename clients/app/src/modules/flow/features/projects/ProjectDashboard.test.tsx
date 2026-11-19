import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { afterEach, describe, it, expect, vi, beforeEach } from 'vitest';
import type { ApiTask } from '../../shared/services/flowApi';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en' } }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

// Summary links render as anchors carrying the filter they would open.
vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ projectId: 'p1' }),
  Link: ({
    children,
    search,
    params: _params,
    to,
    ...rest
  }: {
    children: React.ReactNode;
    search?: { filter?: unknown };
    params?: unknown;
    to: string;
  }) => (
    <a href={to} data-filter={JSON.stringify(search?.filter ?? null)} {...rest}>
      {children}
    </a>
  ),
}));

const summaryState = vi.hoisted(() => ({
  statuses: [] as Array<{ id: string; bucket: string; legacyStatus: string }>,
  cycles: [] as Array<{ id: number; name: string; status: string }>,
}));
vi.mock('./useStatuses', () => ({
  useProjectStatuses: () => ({ data: summaryState.statuses }),
}));
vi.mock('../cycles/useCycles', () => ({
  useCycles: () => ({ data: summaryState.cycles }),
}));

vi.mock('../../shared/lib/useProjectRealtime', () => ({
  useProjectRealtime: () => {},
}));

vi.mock('./ProjectViewTabs', () => ({
  default: () => null,
}));

vi.mock('./useProjects', () => ({
  useProjects: () => ({ data: [], isLoading: false, refetch: vi.fn() }),
}));

// An unlisted project is asked for directly; these screens are tested with the
// project in the store, so the direct lookup never has anything to say.
vi.mock('../projects-v2/useAllProjects', () => ({
  useProject: () => ({
    data: undefined,
    isPending: false,
    isError: false,
    isPaused: false,
    failureCount: 0,
    error: null,
    refetch: vi.fn(),
  }),
}));

const tasksState = vi.hoisted(() => ({
  data: [] as ApiTask[],
  isPending: false,
  isLoading: false,
  isError: false,
  isPaused: false,
  failureCount: 0,
  refetch: vi.fn(),
}));
vi.mock('../../shared/hooks/queries/useProjectLabourCost', () => ({
  useProjectLabourCost: vi.fn(() => ({ data: null })),
}));

vi.mock('../tasks/useTasks', () => ({
  useTasks: () => tasksState,
}));

// The Insights project dashboard panel (T-567) is tested on its own. Here it
// stands in as unavailable, so the client-side tiles it falls back to render,
// unless a test says the server panel answered.
const panelState = vi.hoisted(() => ({ available: false }));
vi.mock('../../../insights/shared/components/ProjectDashboardPanel', () => ({
  default: ({
    projectId,
    unavailableFallback,
  }: {
    projectId: string;
    unavailableFallback?: React.ReactNode;
  }) => (panelState.available ? <p>Insights panel for {projectId}</p> : <>{unavailableFallback}</>),
}));

import ProjectDashboard from './ProjectDashboard';
import { useFlowStore } from '../../shared/stores/flowStore';

beforeEach(() => {
  panelState.available = false;
  summaryState.statuses = [];
  summaryState.cycles = [];
  tasksState.data = [];
  tasksState.isPending = false;
  tasksState.isLoading = false;
  tasksState.isError = false;
  tasksState.isPaused = false;
  tasksState.failureCount = 0;
  tasksState.refetch.mockClear();
  useFlowStore.setState({
    projects: [
      {
        id: 'p1',
        name: 'Demo project',
        color: '#3b82f6',
        taskCount: 0,
        completedCount: 0,
        memberAvatars: [],
      },
    ],
    quickCreateOpen: false,
  });
});

// A due date is a calendar day. `new Date(dueDate) < new Date()` parses the
// bare date as UTC midnight, so at 09:00 local (UTC+3) a task due *today* was
// already 6 hours past midnight and counted as overdue (rtylr#1077).
describe('ProjectDashboard: overdue counting', () => {
  const todayAt = (isoTime: string): void => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date(isoTime));
  };

  afterEach(() => {
    vi.useRealTimers();
  });

  // The stat tiles render label then value; the overdue tile is the fourth, so
  // read the value that follows its label.
  function overdueTile(): string {
    const tiles = screen.getAllByText('flow.projectDashboard.overdue');
    const tile = tiles[0].parentElement as HTMLElement;
    return tile.textContent?.replace('flow.projectDashboard.overdue', '').trim() ?? '';
  }

  it('does not count a task due today as overdue', () => {
    todayAt('2026-05-02T09:00:00+03:00');
    tasksState.data = [
      { id: '1', title: 'Due today', status: 'todo', dueDate: '2026-05-02' } as ApiTask,
    ];
    render(<ProjectDashboard />);

    expect(overdueTile()).toBe('0');
  });

  it('counts a task due yesterday as overdue, and never a done one', () => {
    todayAt('2026-05-02T09:00:00+03:00');
    tasksState.data = [
      { id: '1', title: 'Open and late', status: 'todo', dueDate: '2026-05-01' } as ApiTask,
      { id: '2', title: 'Done and late', status: 'done', dueDate: '2026-05-01' } as ApiTask,
    ];
    render(<ProjectDashboard />);

    expect(overdueTile()).toBe('1');
  });
});

describe('ProjectDashboard: no tasks yet', () => {
  it('shows the empty title/description with an Add task CTA in the status breakdown card', () => {
    render(<ProjectDashboard />);

    expect(screen.getAllByText('flow.projectDashboard.noTasksYet').length).toBeGreaterThan(0);
    expect(screen.getByText('flow.projectDashboard.noTasksYetDescription')).toBeInTheDocument();

    const [firstCta] = screen.getAllByRole('button', {
      name: 'flow.projectDashboard.addTaskCta',
    });
    fireEvent.click(firstCta);
    expect(useFlowStore.getState().quickCreateOpen).toBe(true);
  });

  it('shows retry instead of zero task totals when loading tasks fails', () => {
    tasksState.isError = true;
    render(<ProjectDashboard />);

    expect(screen.getByRole('alert')).toHaveTextContent('flow.loadFailed.description');
    expect(screen.queryByText('flow.projectDashboard.noTasksYet')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'flow.loadFailed.retry' }));
    expect(tasksState.refetch).toHaveBeenCalledOnce();
  });

  it('shows the same empty title in the by-assignee card', () => {
    render(<ProjectDashboard />);

    const emptyTitles = screen.getAllByText('flow.projectDashboard.noTasksYet');
    expect(emptyTitles.length).toBe(2);
  });
});

describe('ProjectDashboard: Insights project dashboard', () => {
  it('shows the server panel for this project instead of the client tiles', () => {
    panelState.available = true;
    render(<ProjectDashboard />);

    expect(screen.getByText('Insights panel for p1')).toBeInTheDocument();
    expect(screen.queryByText('flow.projectDashboard.totalTasks')).not.toBeInTheDocument();
    // The status and assignee cards stay below the panel.
    expect(screen.getByText('flow.projectDashboard.statusBreakdown')).toBeInTheDocument();
    expect(screen.getByText('flow.projectDashboard.byAssignee')).toBeInTheDocument();
  });

  it('falls back to the client tiles when the panel is unavailable', () => {
    render(<ProjectDashboard />);
    expect(screen.queryByText(/Insights panel/)).not.toBeInTheDocument();
    expect(screen.getByText('flow.projectDashboard.totalTasks')).toBeInTheDocument();
  });
});

// The task list carries the assignee's display name in assigneeName; the legacy
// `assignee` field is empty, so every task used to land under Unassigned.
describe('ProjectDashboard: by assignee', () => {
  it('groups tasks by assignee name and labels the rest as unassigned', () => {
    tasksState.data = [
      {
        id: '1',
        title: 'A',
        status: 'todo',
        assignee: '',
        assigneeName: 'Abed Amouneh',
      } as ApiTask,
      {
        id: '2',
        title: 'B',
        status: 'todo',
        assignee: '',
        assigneeName: 'Abed Amouneh',
      } as ApiTask,
      { id: '3', title: 'C', status: 'todo', assignee: '', assigneeName: '' } as ApiTask,
    ];
    render(<ProjectDashboard />);

    expect(screen.getByText('Abed Amouneh')).toBeInTheDocument();
    expect(screen.getByText('flow.projectDashboard.unassigned')).toBeInTheDocument();
    expect(screen.queryByText('Unassigned')).not.toBeInTheDocument();
  });
});

describe('ProjectDashboard: project summary (T-591)', () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  beforeEach(() => {
    vi.useFakeTimers({ toFake: ['Date'] });
    vi.setSystemTime(new Date('2026-11-18T12:00:00Z'));
    summaryState.statuses = [
      { id: '10', bucket: 'active', legacyStatus: 'todo' },
      { id: '12', bucket: 'done', legacyStatus: 'done' },
    ];
    summaryState.cycles = [{ id: 7, name: 'Sprint 7', status: 'active' }];
    tasksState.data = [
      {
        id: '1',
        title: 'Late',
        status: 'todo',
        statusId: '10',
        dueDate: '2026-11-10',
        assigneeId: 'u1',
        assigneeName: 'Ahmed',
        estimatedHours: 4,
        actualHours: 1,
        createdAt: '2026-11-17T12:00:00Z',
        updatedAt: '2026-11-17T12:00:00Z',
      },
      {
        id: '2',
        title: 'Nobody on it',
        status: 'todo',
        statusId: '10',
        cycleId: 7,
        createdAt: '2026-10-01T12:00:00Z',
        updatedAt: '2026-10-01T12:00:00Z',
      },
    ] as unknown as ApiTask[];
  });

  const link = (el: HTMLElement): HTMLAnchorElement => el.closest('a') as HTMLAnchorElement;

  it('opens the list filtered to what each recent card counts', () => {
    render(<ProjectDashboard />);
    const created = document.querySelector('[data-summary-count="created"]') as HTMLElement;
    expect(created.textContent).toBe('1');
    expect(link(created).getAttribute('href')).toBe('/flow/p/$projectId/list');
    expect(JSON.parse(link(created).dataset.filter ?? 'null')).toEqual({
      changed: ['created_7d'],
    });
  });

  it('lists overdue and unassigned work in the active cycle under Needs attention', () => {
    render(<ProjectDashboard />);
    expect(screen.getByText('flow.projectDashboard.summary.needsAttention')).toBeInTheDocument();
    const overdue = document.querySelector('[data-attention-count="overdue"]') as HTMLElement;
    expect(overdue.textContent).toBe('1');
    expect(JSON.parse(link(overdue).dataset.filter ?? 'null')).toEqual({
      due: ['overdue'],
      statuses: ['10'],
    });
    const gap = document.querySelector('[data-attention-count="cycle:7"]') as HTMLElement;
    expect(gap.textContent).toBe('1');
    expect(JSON.parse(link(gap).dataset.filter ?? 'null')).toEqual({
      assignees: ['none'],
      cycles: ['7'],
      statuses: ['10'],
    });
  });

  it('shows logged hours against the estimate and each person their open work', () => {
    render(<ProjectDashboard />);
    expect(screen.getByText('flow.projectDashboard.summary.hoursTitle')).toBeInTheDocument();
    const person = document.querySelector('[data-workload-detail="u1"]') as HTMLElement;
    expect(JSON.parse(link(person).dataset.filter ?? 'null')).toEqual({
      assignees: ['u1'],
      statuses: ['10'],
    });
  });

  it('hides the hours card when nothing is estimated', () => {
    tasksState.data = tasksState.data.map((t) => ({ ...t, estimatedHours: undefined }));
    render(<ProjectDashboard />);
    expect(screen.queryByText('flow.projectDashboard.summary.hoursTitle')).not.toBeInTheDocument();
  });
});
