import React from 'react';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';
import { useUserStore } from '../../shared/stores/userStore';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const myIssuesData = vi.hoisted(() => ({
  data: [] as unknown[],
  isLoading: false,
  isError: false,
  isPending: false,
  isPaused: false,
  failureCount: 0,
  refetch: vi.fn(),
}));
vi.mock('./useMyIssues', () => ({
  useMyIssues: () => myIssuesData,
}));

const setQuickCreateOpen = vi.hoisted(() => vi.fn());

vi.mock('../../shared/stores/flowStore', () => ({
  useFlowStore: () => ({ selectedTaskIds: [], setQuickCreateOpen }),
}));

vi.mock('../../shared/components/TaskRow', () => ({
  TaskRow: ({ task }: { task: { title: string } }) => <div>{task.title}</div>,
}));

import MyIssues from './MyIssues';

beforeEach(() => {
  vi.clearAllMocks();
  useUserStore.setState({
    user: { id: 7, effectiveAccess: createEffectiveAccess('flow', { isOwner: true }) },
  });
  myIssuesData.data = [];
  myIssuesData.isLoading = false;
  // The query-failure flags belong here, not in one describe's own hook.
  // They were reset inside the failure describe only, so `--sequence.shuffle`
  // could run a failure test before an empty-state test and the empty state
  // rendered the failed state instead (#1194).
  myIssuesData.isError = false;
  myIssuesData.isPending = false;
  myIssuesData.isPaused = false;
  myIssuesData.failureCount = 0;
});

describe('MyIssues empty state', () => {
  it('does not claim the task list is empty before a paused first request resolves', () => {
    myIssuesData.data = [];
    myIssuesData.isPending = true;
    myIssuesData.isPaused = true;
    render(<MyIssues />);

    expect(screen.getByText('flow.myIssues.loading')).toBeInTheDocument();
    expect(screen.queryByText('flow.myIssues.emptyTitle')).not.toBeInTheDocument();
  });

  it('shows the board-level CTA and no status columns when there are no tasks', () => {
    myIssuesData.data = [];
    render(<MyIssues />);

    // The five status columns are gone: the screen groups by time bucket now,
    // which is what myTasksBuckets has always implemented (#1035).
    expect(screen.queryByText('flow.taskDetail.backlog')).not.toBeInTheDocument();
    expect(screen.queryByText('flow.taskDetail.inProgress')).not.toBeInTheDocument();
    expect(screen.getByText('flow.myIssues.emptyTitle')).toBeInTheDocument();
    expect(screen.getByText('flow.myIssues.emptyDescription')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'flow.board.addTaskLabel' })).toBeInTheDocument();
  });

  it('opens the full task form from the board CTA', () => {
    myIssuesData.data = [];
    render(<MyIssues />);

    fireEvent.click(screen.getByRole('button', { name: 'flow.board.addTaskLabel' }));

    expect(setQuickCreateOpen).toHaveBeenCalledWith(true);
  });

  it('does not show the board-level welcome when there are tasks', () => {
    myIssuesData.data = [{ id: '1', title: 'Ship the thing', status: 'todo' }];
    render(<MyIssues />);

    expect(screen.queryByText('flow.myIssues.emptyTitle')).not.toBeInTheDocument();
    expect(
      screen.queryByRole('button', { name: 'flow.board.addTaskLabel' })
    ).not.toBeInTheDocument();
  });

  it('does not render the old inline quick-add field', () => {
    render(<MyIssues />);

    expect(
      screen.queryByPlaceholderText('flow.myIssues.quickAddPlaceholder')
    ).not.toBeInTheDocument();
  });
});

describe('MyIssues time buckets (#1035)', () => {
  const iso = (offsetDays: number): string => {
    const d = new Date();
    d.setHours(12, 0, 0, 0);
    d.setDate(d.getDate() + offsetDays);
    return d.toISOString();
  };

  it('gives overdue work its own section above the rest', () => {
    myIssuesData.data = [
      { id: '1', title: 'Overdue thing', status: 'todo', dueDate: iso(-15) },
      { id: '2', title: 'Due today', status: 'todo', dueDate: iso(0) },
      { id: '3', title: 'Due this week', status: 'todo', dueDate: iso(3) },
      { id: '4', title: 'Due much later', status: 'todo', dueDate: iso(60) },
    ];

    render(<MyIssues />);

    // Late is its own section now. It used to be folded into Today, which is
    // what left a fifteen-day-overdue task with no promotion of any kind.
    expect(screen.getByText('flow.myIssues.buckets.late')).toBeInTheDocument();
    expect(screen.getByText('flow.myIssues.buckets.today')).toBeInTheDocument();
    expect(screen.getByText('flow.myIssues.buckets.thisWeek')).toBeInTheDocument();
    expect(screen.getByText('flow.myIssues.buckets.later')).toBeInTheDocument();
    expect(screen.getByText('Overdue thing')).toBeInTheDocument();
    expect(screen.getByText('Due today')).toBeInTheDocument();
  });

  it('puts Late first, before Today', () => {
    myIssuesData.data = [
      { id: '2', title: 'Due today', status: 'todo', dueDate: iso(0) },
      { id: '1', title: 'Overdue thing', status: 'todo', dueDate: iso(-15) },
    ];

    render(<MyIssues />);

    const headings = screen.getAllByRole('button', { expanded: true }).map((b) => b.textContent);
    const late = headings.findIndex((h) => h?.includes('buckets.late'));
    const today = headings.findIndex((h) => h?.includes('buckets.today'));
    expect(late).toBeGreaterThanOrEqual(0);
    expect(late).toBeLessThan(today);
  });

  it('buckets a task with no due date by that fact', () => {
    myIssuesData.data = [
      { id: '5', title: 'No date, just assigned', status: 'todo', dueDate: '', createdAt: iso(-1) },
    ];

    render(<MyIssues />);

    expect(screen.getByText('flow.myIssues.buckets.noDate')).toBeInTheDocument();
  });

  it('opens Late and Today expanded, so overdue work needs no interaction', () => {
    myIssuesData.data = [
      { id: '1', title: 'Overdue thing', status: 'todo', dueDate: iso(-15) },
      { id: '5', title: 'No date', status: 'todo', dueDate: '' },
    ];

    render(<MyIssues />);

    const late = screen.getAllByRole('button').find((b) => b.textContent?.includes('buckets.late'));
    expect(late).toHaveAttribute('aria-expanded', 'true');
    // The panel itself, via aria-controls, rather than `toBeVisible`: the
    // section is a motion.section with initial opacity 0, and jest-dom counts
    // that as not visible even though a browser animates it straight in.
    const latePanel = document.getElementById(late!.getAttribute('aria-controls')!);
    expect(latePanel).not.toHaveAttribute('hidden');
    expect(latePanel!.textContent).toContain('Overdue thing');

    // No date is the one bucket with no claim on attention, so it starts closed.
    const noDate = screen
      .getAllByRole('button')
      .find((b) => b.textContent?.includes('buckets.noDate'));
    expect(noDate).toHaveAttribute('aria-expanded', 'false');
  });

  it('collapses and reopens a section from its heading', () => {
    myIssuesData.data = [{ id: '1', title: 'Overdue thing', status: 'todo', dueDate: iso(-15) }];

    render(<MyIssues />);

    const heading = screen
      .getAllByRole('button')
      .find((b) => b.textContent?.includes('buckets.late'))!;
    const panel = document.getElementById(heading.getAttribute('aria-controls')!)!;
    expect(panel).not.toHaveAttribute('hidden');

    fireEvent.click(heading);
    expect(heading).toHaveAttribute('aria-expanded', 'false');
    expect(panel).toHaveAttribute('hidden');

    fireEvent.click(heading);
    expect(heading).toHaveAttribute('aria-expanded', 'true');
    expect(panel).not.toHaveAttribute('hidden');
  });

  it('renders no heading for an empty bucket', () => {
    myIssuesData.data = [{ id: '6', title: 'Only one', status: 'todo', dueDate: iso(0) }];

    render(<MyIssues />);

    expect(screen.getByText('flow.myIssues.buckets.today')).toBeInTheDocument();
    expect(screen.queryByText('flow.myIssues.buckets.later')).not.toBeInTheDocument();
    expect(screen.queryByText('flow.myIssues.buckets.thisWeek')).not.toBeInTheDocument();
  });

  it('leaves finished work out of the buckets', () => {
    myIssuesData.data = [
      { id: '7', title: 'Finished', status: 'done', dueDate: iso(0) },
      { id: '8', title: 'Abandoned', status: 'cancelled', dueDate: iso(0) },
      { id: '9', title: 'Still open', status: 'todo', dueDate: iso(0) },
    ];

    render(<MyIssues />);

    expect(screen.getByText('Still open')).toBeInTheDocument();
    expect(screen.queryByText('Finished')).not.toBeInTheDocument();
    expect(screen.queryByText('Abandoned')).not.toBeInTheDocument();
  });

  it('translates every bucket heading rather than printing English', () => {
    myIssuesData.data = [{ id: '10', title: 'Anything', status: 'todo', dueDate: iso(0) }];

    render(<MyIssues />);

    // MY_TASK_BUCKET_LABEL held literal English while nothing imported it.
    expect(screen.queryByText('Today')).not.toBeInTheDocument();
    expect(screen.getByText('flow.myIssues.buckets.today')).toBeInTheDocument();
  });
});

describe('MyIssues: a failed load is not an empty list', () => {
  beforeEach(() => {
    myIssuesData.data = [];
    myIssuesData.isLoading = false;
    myIssuesData.isError = false;
    myIssuesData.isPaused = false;
    myIssuesData.failureCount = 0;
  });

  it('shows the failure state instead of "nice work" when the request fails', () => {
    // The reported case: with every request forced to 500 this screen said
    // "Nothing is on your plate right now. Nice work!", reporting success
    // while the server was down.
    myIssuesData.isError = true;
    render(<MyIssues />);

    expect(screen.getByText('flow.loadFailed.title')).toBeInTheDocument();
    expect(screen.queryByText('flow.myIssues.emptyTitle')).not.toBeInTheDocument();
  });

  it('retries the query from the failure state', () => {
    myIssuesData.isError = true;
    render(<MyIssues />);

    fireEvent.click(screen.getByText('flow.loadFailed.retry'));
    expect(myIssuesData.refetch).toHaveBeenCalledTimes(1);
  });

  it('does not assert a task count while the load has failed', () => {
    myIssuesData.isError = true;
    render(<MyIssues />);
    expect(screen.queryByText('flow.myIssues.taskCount')).not.toBeInTheDocument();
  });

  it('still shows the empty state when the request succeeded and there is nothing', () => {
    render(<MyIssues />);
    expect(screen.getByText('flow.myIssues.emptyTitle')).toBeInTheDocument();
    expect(screen.queryByText('flow.loadFailed.title')).not.toBeInTheDocument();
  });

  it('treats a paused query with a failure behind it as a failure', () => {
    // Offline: react-query pauses rather than erroring, so isError alone
    // leaves an offline user looking at "Nice work!".
    myIssuesData.isPaused = true;
    myIssuesData.failureCount = 1;
    render(<MyIssues />);
    expect(screen.getByText('flow.loadFailed.title')).toBeInTheDocument();
  });
});
