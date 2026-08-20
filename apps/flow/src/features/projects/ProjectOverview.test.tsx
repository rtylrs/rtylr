import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ApiMilestone } from '../../shared/services/flowApi';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const navigateMock = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ workspaceId: 'w1', projectId: 'p1' }),
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
  useFlowStore.setState({
    projects: [
      {
        id: 'p1',
        workspaceId: 'w1',
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
    expect(navigateMock).toHaveBeenCalledWith({ to: '/milestones' });
  });
});

describe('ProjectOverview: no recent comments', () => {
  it('shows the empty title with a CTA to the project board', () => {
    render(<ProjectOverview />);

    expect(screen.getByText('flow.projectOverview.noComments')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'flow.projectOverview.goToBoardCta' }));
    expect(navigateMock).toHaveBeenCalledWith({
      to: '/w/$workspaceId/p/$projectId/board',
      params: { workspaceId: 'w1', projectId: 'p1' },
    });
  });
});
