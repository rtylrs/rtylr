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
  useParams: () => ({ projectId: 'p1' }),
  useNavigate: () => navigateMock,
}));

vi.mock('../../shared/lib/useProjectRealtime', () => ({
  useProjectRealtime: () => {},
}));

vi.mock('./ProjectViewTabs', () => ({
  default: () => null,
}));

vi.mock('./useProjects', () => ({
  useProjects: () => ({ data: [], isLoading: false }),
}));

const tasksState = vi.hoisted(() => ({ data: [] as ApiTask[] }));
vi.mock('../tasks/useTasks', () => ({
  useTasks: () => tasksState,
}));

import ProjectTimeline from './ProjectTimeline';
import { useFlowStore } from '../../shared/stores/flowStore';

beforeEach(() => {
  vi.clearAllMocks();
  tasksState.data = [];
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
  it('shows the empty title/description with a CTA to the project board', () => {
    render(<ProjectTimeline />);

    expect(screen.getByText('flow.projectTimeline.emptyTitle')).toBeInTheDocument();
    expect(screen.getByText('flow.projectTimeline.emptyDescription')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'flow.projectTimeline.goToBoardCta' }));
    expect(navigateMock).toHaveBeenCalledWith({
      to: '/flow/p/$projectId/board',
      params: { projectId: 'p1' },
    });
  });
});
