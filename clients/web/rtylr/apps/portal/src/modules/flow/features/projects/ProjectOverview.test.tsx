import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import type { ApiMilestone } from '../../shared/services/flowApi';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
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
  default: () => null,
}));

const overviewState = vi.hoisted(() => ({
  data: {
    stats: { total: 0, done: 0, inProgress: 0, overdue: 0, progress: 0 },
    milestones: [] as ApiMilestone[],
  } as {
    stats: { total: number; done: number; inProgress: number; overdue: number; progress: number };
    milestones: ApiMilestone[];
  } | null,
  isLoading: false,
  isError: false,
  isPaused: false,
  failureCount: 0,
  refetch: vi.fn(),
}));
vi.mock('./useProjectOverview', () => ({
  useProjectOverview: () => overviewState,
}));

import ProjectOverview from './ProjectOverview';
import { useFlowStore } from '../../shared/stores/flowStore';

beforeEach(() => {
  vi.clearAllMocks();
  overviewState.data = {
    stats: { total: 0, done: 0, inProgress: 0, overdue: 0, progress: 0 },
    milestones: [],
  };
  overviewState.isLoading = false;
  // The query-failure flags belong here, not in one describe's own hook.
  // They were reset inside the failure describe only, so `--sequence.shuffle`
  // could run a failure test before an empty-state test and the empty state
  // rendered the failed state instead (#1194).
  overviewState.isError = false;
  overviewState.isPaused = false;
  overviewState.failureCount = 0;
  useFlowStore.setState({
    projects: [
      {
        id: 'p1',
        name: 'Demo project',
        color: '#3b82f6',
        taskCount: 0,
        completedCount: 0,
        memberAvatars: [],
        health: 'on_track',
      },
    ],
    members: [],
  });
});

describe('ProjectOverview: no milestones', () => {
  it('shows the empty title with a CTA to Milestones', () => {
    render(<ProjectOverview />);

    expect(screen.getByText('flow.projectOverview.noMilestones')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'flow.projectOverview.goToMilestonesCta' }));
    expect(navigateMock).toHaveBeenCalledWith({ to: '/flow/milestones' });
  });
});

describe('ProjectOverview: no recent comments', () => {
  it('shows the empty title with a CTA to the project board', () => {
    render(<ProjectOverview />);

    expect(screen.getByText('flow.projectOverview.noComments')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'flow.projectOverview.goToBoardCta' }));
    expect(navigateMock).toHaveBeenCalledWith({
      to: '/flow/p/$projectId/board',
      params: { projectId: 'p1' },
    });
  });
});

describe('ProjectOverview when the request fails', () => {
  afterEach(() => {
    overviewState.isError = false;
    overviewState.isPaused = false;
    overviewState.failureCount = 0;
  });

  it('offers a retry instead of reporting the project deleted', () => {
    // A failed overview leaves `data` undefined exactly as a missing project
    // does, so a 500 used to say the project was gone.
    overviewState.isError = true;
    overviewState.failureCount = 1;
    render(<ProjectOverview />);

    expect(screen.getByText('flow.loadFailed.title')).toBeInTheDocument();
    expect(screen.queryByText('flow.project.notFound')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'flow.loadFailed.retry' }));
    expect(overviewState.refetch).toHaveBeenCalled();
  });

  it('reports a paused query that already failed as a failure', () => {
    overviewState.isPaused = true;
    overviewState.failureCount = 1;
    render(<ProjectOverview />);
    expect(screen.getByText('flow.loadFailed.title')).toBeInTheDocument();
  });
});
