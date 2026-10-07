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

vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ projectId: 'p1' }),
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

import ProjectDashboard from './ProjectDashboard';
import { useFlowStore } from '../../shared/stores/flowStore';

beforeEach(() => {
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
