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
vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

const tasksState = vi.hoisted(() => ({
  data: [] as unknown[],
  isLoading: false,
  isError: false,
  isPaused: false,
  failureCount: 0,
  refetch: vi.fn(),
}));

// Both hooks come from the same module, so they mock together.
vi.mock('../tasks/useTasks', () => ({
  useTasks: () => tasksState,
  useUpdateTask: () => ({ mutate: vi.fn(), isPending: false }),
}));

import Archive from './Archive';

function archivedTask(over: Record<string, unknown> = {}): Record<string, unknown> {
  return {
    id: '1',
    projectId: 'p1',
    title: 'Old gate hinge',
    status: 'done',
    priority: 'medium',
    archived: true,
    dueDate: '',
    ...over,
  };
}

beforeEach(() => {
  tasksState.data = [];
  tasksState.isLoading = false;
  tasksState.isError = false;
  tasksState.isPaused = false;
  tasksState.failureCount = 0;
  tasksState.refetch.mockClear();
});

describe('Archive screen (#1089)', () => {
  it('explains that archiving is not deleting', () => {
    render(<Archive />);
    expect(screen.getByText('flowArchive.explanation')).toBeInTheDocument();
  });

  it('claims no count while loading', () => {
    // apiTasks defaults to [] until the query resolves, so the screen said
    // "0 archived tasks" and "Nothing archived" before it had fetched.
    tasksState.isLoading = true;
    render(<Archive />);

    expect(screen.queryByText('flowArchive.count:0')).not.toBeInTheDocument();
    expect(screen.queryByText('flowArchive.emptyTitle')).not.toBeInTheDocument();
  });

  it('claims no count when the request failed', () => {
    tasksState.isError = true;
    render(<Archive />);

    expect(screen.getByText('flow.loadFailed.title')).toBeInTheDocument();
    expect(screen.queryByText('flowArchive.count:0')).not.toBeInTheDocument();
    expect(screen.queryByText('flowArchive.emptyTitle')).not.toBeInTheDocument();
  });

  it('retries from the failure state', () => {
    tasksState.isError = true;
    render(<Archive />);
    fireEvent.click(screen.getByText('flow.loadFailed.retry'));
    expect(tasksState.refetch).toHaveBeenCalledTimes(1);
  });

  it('reports a true count once the query resolved', () => {
    tasksState.data = [archivedTask(), archivedTask({ id: '2', title: 'Old invoice' })];
    render(<Archive />);

    expect(screen.getByText('flowArchive.count:2')).toBeInTheDocument();
    expect(screen.getByText('Old gate hinge')).toBeInTheDocument();
  });

  it('says the archive is empty only when it really is', () => {
    tasksState.data = [archivedTask({ archived: false })];
    render(<Archive />);

    expect(screen.getByText('flowArchive.emptyTitle')).toBeInTheDocument();
    expect(screen.getByText('flowArchive.count:0')).toBeInTheDocument();
  });

  it('dims the archived content but not the restore control', () => {
    tasksState.data = [archivedTask()];
    const { container } = render(<Archive />);

    const dimmed = container.querySelector('.opacity-\\[0\\.62\\]');
    expect(dimmed).not.toBeNull();
    expect(dimmed?.textContent).toContain('Old gate hinge');
    // Restore sits outside the dimmed wrapper, so it stays legible.
    expect(dimmed?.textContent).not.toContain('flowArchive.restore');
    expect(screen.getByText(/flowArchive\.restore/)).toBeInTheDocument();
  });
});
