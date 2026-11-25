import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ApiTask } from '../../shared/services/flowApi';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en-US' } }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const navigateMock = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ projectId: 'p1' }),
  useNavigate: () => navigateMock,
}));

vi.mock('../../shared/lib/useProjectRealtime', () => ({
  useProjectRealtime: () => {},
}));

vi.mock('./ProjectViewTabs', () => ({
  default: ({ rightSlot }: { rightSlot?: React.ReactNode }) => <div>{rightSlot}</div>,
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
vi.mock('../tasks/useTasks', () => ({
  useTasks: () => tasksState,
}));
vi.mock('../milestones/useMilestones', () => ({
  useMilestones: () => ({ data: [{ id: 'm1', name: 'Phase one', targetDate: '2026-12-01' }] }),
}));
vi.mock('../../shared/lib/dateUtils', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../../shared/lib/dateUtils')>()),
  todayIso: () => '2026-11-25',
}));

import ProjectTimeline from './ProjectTimeline';
import { useFlowStore } from '../../shared/stores/flowStore';

beforeEach(() => {
  vi.clearAllMocks();
  tasksState.data = [];
  tasksState.isPending = false;
  tasksState.isLoading = false;
  tasksState.isError = false;
  tasksState.isPaused = false;
  tasksState.failureCount = 0;
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
  });
});

describe('ProjectTimeline: no scheduled tasks', () => {
  it('shows the empty title/description with a CTA to the project list', () => {
    render(<ProjectTimeline />);

    expect(screen.getByText('flow.projectTimeline.emptyTitle')).toBeInTheDocument();
    expect(screen.getByText('flow.projectTimeline.emptyDescription')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'flow.projectTimeline.goToBoardCta' }));
    expect(navigateMock).toHaveBeenCalledWith({
      to: '/flow/p/$projectId/list',
      params: { projectId: 'p1' },
    });
  });

  it('shows retry instead of the empty timeline when loading tasks fails', () => {
    tasksState.isError = true;
    render(<ProjectTimeline />);

    expect(screen.getByRole('alert')).toHaveTextContent('flow.loadFailed.description');
    expect(screen.queryByText('flow.projectTimeline.emptyTitle')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'flow.loadFailed.retry' }));
    expect(tasksState.refetch).toHaveBeenCalledOnce();
  });
});

function apiTask(over: Partial<ApiTask>): ApiTask {
  return {
    id: '1',
    projectId: 'p1',
    title: 'Task',
    status: 'todo',
    priority: 'medium',
    assignee: '',
    assigneeInitials: '',
    dueDate: '',
    subtaskCount: 0,
    projectName: 'Demo project',
    ...over,
  };
}

describe('ProjectTimeline: bars, groups and arrows (T-342)', () => {
  beforeEach(() => {
    tasksState.data = [
      apiTask({
        id: '1',
        title: 'Design',
        startDate: '2026-11-20',
        dueDate: '2026-11-22',
        assigneeId: '7',
        assigneeName: 'Amal',
        milestoneId: 'm1',
      }),
      apiTask({
        id: '2',
        title: 'Build',
        startDate: '2026-11-23',
        dueDate: '2026-11-28',
        assigneeId: '8',
        assigneeName: 'Zed',
        dependencies: ['1'],
      }),
      apiTask({ id: '3', title: 'Someday' }),
    ];
  });

  it('draws a bar from start to due, lists undated work, and links blockers', () => {
    const { container } = render(<ProjectTimeline />);
    expect(screen.getAllByRole('button', { name: 'flow.projectTimeline.barLabel' })).toHaveLength(
      2
    );
    expect(screen.getByText('flow.projectTimeline.unscheduled')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Someday' })).toBeInTheDocument();
    expect(container.querySelectorAll('path[marker-end]')).toHaveLength(1);
    expect(screen.getByRole('img', { name: 'Phase one' })).toBeInTheDocument();
  });

  it('groups rows by milestone with unplanned work last', async () => {
    render(<ProjectTimeline />);
    fireEvent.pointerDown(screen.getByRole('button', { name: /^flow.projectTimeline.groupBy/ }), {
      button: 0,
      ctrlKey: false,
    });
    fireEvent.click(
      await screen.findByRole('menuitem', { name: 'flow.projectTimeline.group.milestone' })
    );
    const headers = screen.getAllByText(/Phase one|flow.projectTimeline.noMilestone/);
    expect(headers.map((h) => h.textContent?.trim())).toEqual([
      'Phase one 1',
      'flow.projectTimeline.noMilestone 1',
    ]);
  });
});
