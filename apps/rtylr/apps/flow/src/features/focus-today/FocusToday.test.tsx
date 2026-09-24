import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ApiTask } from '../../shared/services/flowApi';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const navigateMock = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, to }: { children?: React.ReactNode; to: string }) => (
    <a href={to}>{children}</a>
  ),
  useNavigate: () => navigateMock,
}));

const myIssuesState = vi.hoisted(() => ({
  data: [] as ApiTask[],
  isLoading: false,
  isError: false,
  isPaused: false,
  failureCount: 0,
  refetch: vi.fn(),
}));
vi.mock('../tasks/useMyIssues', () => ({
  useMyIssues: () => myIssuesState,
}));

// activeWorkspaceId is '' in production: the workspace layer is gone. The old
// filter compared it against each task's workspaceId, which the API no longer
// sends, so a mock carrying 'ws-1' hid the defect in #1035.
// importOriginal, because rendering a task pulls PRIORITY_COLORS and friends
// out of this module; only the hook is replaced.
vi.mock('../../shared/stores/flowStore', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../../shared/stores/flowStore')>()),
  useFlowStore: () => ({ projects: [], activeWorkspaceId: '' }),
}));

// The existing tests below all render with no tasks, so this only affects the
// bucketing cases added at the end of the file.
vi.mock('../../shared/lib/currentUser', () => ({ isTaskMine: () => true }));

import { todayIso } from '../../shared/lib/lifeRhythm';
import FocusToday from './FocusToday';

beforeEach(() => {
  vi.clearAllMocks();
  myIssuesState.data = [];
  myIssuesState.isLoading = false;
  myIssuesState.isError = false;
  myIssuesState.isPaused = false;
  myIssuesState.failureCount = 0;
});

describe('FocusToday when My Tasks cannot load', () => {
  it('shows the load-failed state with retry, not an empty day', () => {
    myIssuesState.isError = true;
    myIssuesState.failureCount = 1;

    render(<FocusToday />);

    expect(screen.queryByText('common.focusToday.emptyTitle')).not.toBeInTheDocument();
    expect(screen.getByText('flow.loadFailed.title')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'flow.loadFailed.retry' }));
    expect(myIssuesState.refetch).toHaveBeenCalled();
  });

  it('treats a query paused after a failed attempt as failed', () => {
    myIssuesState.isPaused = true;
    myIssuesState.failureCount = 2;

    render(<FocusToday />);

    expect(screen.getByText('flow.loadFailed.title')).toBeInTheDocument();
  });

  it('does not claim an empty day while the first load is in flight', () => {
    myIssuesState.isLoading = true;

    render(<FocusToday />);

    expect(screen.queryByText('common.focusToday.emptyTitle')).not.toBeInTheDocument();
    expect(screen.queryByText('flow.loadFailed.title')).not.toBeInTheDocument();
  });
});

describe('FocusToday empty state', () => {
  it('shows the empty title/hint with a CTA to My Tasks', () => {
    render(<FocusToday />);

    expect(screen.getByText('common.focusToday.emptyTitle')).toBeInTheDocument();
    expect(screen.getByText('common.focusToday.emptyHint')).toBeInTheDocument();

    const cta = screen.getByRole('button', { name: 'flow.shell.myTasks' });
    fireEvent.click(cta);
    expect(navigateMock).toHaveBeenCalledWith({ to: '/my-issues' });
  });
});

describe('FocusToday lists what is due today (#1035)', () => {
  const isoToday = (): string => {
    const d = new Date();
    d.setHours(12, 0, 0, 0);
    return d.toISOString();
  };

  const task = (over: Partial<ApiTask> & { id: string; title: string }): ApiTask =>
    ({ status: 'todo', assigneeId: '9', dueDate: '', ...over }) as unknown as ApiTask;

  it('lists a task due today', () => {
    myIssuesState.data = [task({ id: '1', title: 'Due today thing', dueDate: isoToday() })];

    render(<FocusToday />);

    expect(screen.getByText('Due today thing')).toBeInTheDocument();
  });

  it('lists a pinned task with no due date', () => {
    myIssuesState.data = [task({ id: '2', title: 'Pinned thing', focusPin: true })];

    render(<FocusToday />);

    expect(screen.getByText('Pinned thing')).toBeInTheDocument();
  });

  it('leaves out a task that is neither pinned nor due today', () => {
    myIssuesState.data = [task({ id: '3', title: 'Some other week', dueDate: '2027-01-01' })];

    render(<FocusToday />);

    expect(screen.queryByText('Some other week')).not.toBeInTheDocument();
  });

  it('leaves out finished and archived work due today', () => {
    myIssuesState.data = [
      task({ id: '4', title: 'Done today', status: 'done', dueDate: isoToday() }),
      task({ id: '5', title: 'Cancelled today', status: 'cancelled', dueDate: isoToday() }),
      task({ id: '6', title: 'Archived today', dueDate: isoToday(), archived: true }),
      task({ id: '7', title: 'Open today', dueDate: isoToday() }),
    ];

    render(<FocusToday />);

    expect(screen.getByText('Open today')).toBeInTheDocument();
    expect(screen.queryByText('Done today')).not.toBeInTheDocument();
    expect(screen.queryByText('Cancelled today')).not.toBeInTheDocument();
    expect(screen.queryByText('Archived today')).not.toBeInTheDocument();
  });

  it('puts pinned work above work merely due today', () => {
    myIssuesState.data = [
      task({ id: '8', title: 'Just due', dueDate: isoToday() }),
      task({ id: '9', title: 'Actually pinned', focusPin: true }),
    ];

    const { container } = render(<FocusToday />);

    const text = container.textContent ?? '';
    expect(text.indexOf('Actually pinned')).toBeLessThan(text.indexOf('Just due'));
  });
});

describe('FocusToday separates pinned from due today (#1083)', () => {
  // The same helper the screen's filter uses. Building this from
  // toISOString() gives the UTC date, and lifeRhythm's todayIso() is local,
  // so the two disagree for the hours between local midnight and UTC
  // midnight: these cases passed when written in the evening and failed
  // after midnight local.
  const today = todayIso();

  function apiTask(over: Record<string, unknown> = {}): ApiTask {
    return {
      id: '1',
      projectId: 'p1',
      title: 'Fix the gate hinge',
      status: 'todo',
      priority: 'medium',
      assigneeId: 'u1',
      dueDate: '',
      archived: false,
      ...over,
    } as unknown as ApiTask;
  }

  beforeEach(() => {
    myIssuesState.data = [];
  });

  it('lists a task that is due today, which the subtitle already promised', () => {
    // The reported defect: the screen said "Nothing pinned for today" while
    // tasks carried today's due date, under a subtitle promising "anything
    // due today".
    myIssuesState.data = [apiTask({ dueDate: today })];
    render(<FocusToday />);

    expect(screen.getByText('Fix the gate hinge')).toBeInTheDocument();
    expect(screen.queryByText('common.focusToday.emptyTitle')).not.toBeInTheDocument();
  });

  it('puts pinned and due-today under their own headings with counts', () => {
    myIssuesState.data = [
      apiTask({ id: '1', title: 'Pinned one', focusPin: true }),
      apiTask({ id: '2', title: 'Due one', dueDate: today }),
      apiTask({ id: '3', title: 'Due two', dueDate: today }),
    ];
    render(<FocusToday />);

    expect(screen.getByText('common.focusToday.pinnedHeading')).toBeInTheDocument();
    expect(screen.getByText('common.focusToday.dueTodayHeading')).toBeInTheDocument();
    expect(screen.getByText('1')).toBeInTheDocument();
    expect(screen.getByText('2')).toBeInTheDocument();
  });

  it('counts a task that is both pinned and due today only once', () => {
    myIssuesState.data = [apiTask({ focusPin: true, dueDate: today })];
    render(<FocusToday />);

    expect(screen.getByText('common.focusToday.pinnedHeading')).toBeInTheDocument();
    expect(screen.queryByText('common.focusToday.dueTodayHeading')).not.toBeInTheDocument();
    expect(screen.getAllByText('Fix the gate hinge')).toHaveLength(1);
  });

  it('omits a heading whose group is empty', () => {
    myIssuesState.data = [apiTask({ dueDate: today })];
    render(<FocusToday />);

    expect(screen.queryByText('common.focusToday.pinnedHeading')).not.toBeInTheDocument();
    expect(screen.getByText('common.focusToday.dueTodayHeading')).toBeInTheDocument();
  });

  it('leaves a task due on another day out', () => {
    myIssuesState.data = [apiTask({ dueDate: '2020-01-01' })];
    render(<FocusToday />);
    expect(screen.getByText('common.focusToday.emptyTitle')).toBeInTheDocument();
  });
});
