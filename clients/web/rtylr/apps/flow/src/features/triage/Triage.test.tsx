import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: { count?: number }) =>
      opts?.count === undefined ? key : `${key}:${opts.count}`,
  }),
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

vi.mock('@tanstack/react-router', () => ({ useNavigate: () => vi.fn() }));

vi.mock('../../shared/stores/flowStore', () => ({
  useFlowStore: () => ({ selectedTaskIds: [] }),
}));

vi.mock('../../shared/components/TaskRow', () => ({
  __esModule: true,
  default: ({ task }: { task: { title: string } }) => <div>{task.title}</div>,
}));

const tasksState = vi.hoisted(() => ({
  data: undefined as unknown[] | undefined,
  isLoading: false,
  isError: false,
  isPaused: false,
  failureCount: 0,
  refetch: vi.fn(),
}));

vi.mock('../tasks/useTasks', () => ({ useTasks: () => tasksState }));

import Triage from './Triage';

function apiTask(over: Record<string, unknown> = {}): Record<string, unknown> {
  return {
    id: '1',
    projectId: 'p1',
    workspaceId: 'w1',
    title: 'Fix the gate hinge',
    status: 'todo',
    priority: 'no_priority',
    dueDate: '',
    archived: false,
    ...over,
  };
}

beforeEach(() => {
  tasksState.data = undefined;
  tasksState.isLoading = false;
  tasksState.isError = false;
  tasksState.isPaused = false;
  tasksState.failureCount = 0;
  tasksState.refetch.mockClear();
});

describe('Triage does not claim inbox zero before the query ran (#1084)', () => {
  it('says nothing about the count while loading', () => {
    // The reported defect: apiTasks is undefined until the query resolves, so
    // needsTriage was empty and the screen said "Inbox zero" and "0 tasks
    // need attention" with three untriaged tasks in the project.
    tasksState.isLoading = true;
    render(<Triage />);

    expect(screen.queryByText('flow.triage.emptyTitle')).not.toBeInTheDocument();
    expect(screen.queryByText('flow.triage.subtitle:0')).not.toBeInTheDocument();
  });

  it('says nothing about the count when the query failed', () => {
    tasksState.isError = true;
    render(<Triage />);

    expect(screen.getByText('flow.loadFailed.title')).toBeInTheDocument();
    expect(screen.queryByText('flow.triage.emptyTitle')).not.toBeInTheDocument();
    expect(screen.queryByText('flow.triage.subtitle:0')).not.toBeInTheDocument();
  });

  it('retries the query from the failure state', () => {
    tasksState.isError = true;
    render(<Triage />);
    fireEvent.click(screen.getByText('flow.loadFailed.retry'));
    expect(tasksState.refetch).toHaveBeenCalledTimes(1);
  });

  it('claims inbox zero only once the query resolved empty', () => {
    tasksState.data = [];
    render(<Triage />);

    expect(screen.getByText('flow.triage.emptyTitle')).toBeInTheDocument();
    expect(screen.getByText('flow.triage.subtitle:0')).toBeInTheDocument();
  });

  it('lists the tasks that need triage and counts them', () => {
    tasksState.data = [
      apiTask(),
      apiTask({ id: '2', title: 'Order the timber', priority: 'high', dueDate: '2026-09-30' }),
      apiTask({ id: '3', title: 'Chase the invoice', status: 'backlog', priority: 'high' }),
    ];
    render(<Triage />);

    expect(screen.getByText('Fix the gate hinge')).toBeInTheDocument();
    expect(screen.getByText('Chase the invoice')).toBeInTheDocument();
    // Priority and a due date and not in the backlog, so nothing to triage.
    expect(screen.queryByText('Order the timber')).not.toBeInTheDocument();
    expect(screen.getByText('flow.triage.subtitle:2')).toBeInTheDocument();
  });

  it('explains what triage means whatever the state', () => {
    tasksState.isLoading = true;
    render(<Triage />);
    expect(screen.getByText('flow.triage.explanation')).toBeInTheDocument();
  });
});
