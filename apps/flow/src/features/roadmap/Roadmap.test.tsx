// @vitest-environment happy-dom
import React from 'react';
import { describe, it, expect, vi, afterEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';

const roadmapState = vi.hoisted(() => ({
  data: [] as Record<string, unknown>[],
  isLoading: false,
  isError: false,
  isPaused: false,
  failureCount: 0,
  refetch: vi.fn(),
}));
const navigateMock = vi.hoisted(() => vi.fn());

vi.mock('./useRoadmap', () => ({ useRoadmap: () => roadmapState }));
vi.mock('react-i18next', () => ({ useTranslation: () => ({ t: (k: string) => k }) }));
vi.mock('@tanstack/react-router', () => ({
  Link: ({ children }: { children: React.ReactNode }) => <a href="/">{children}</a>,
  useNavigate: () => navigateMock,
}));
vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children: React.ReactNode }) => <div>{children}</div> },
}));

import Roadmap from './Roadmap';

/**
 * An empty roadmap and a failed request both left this screen showing a header
 * over blank space (#1072). Nothing distinguished "nothing is planned" from
 * "the server is down", and neither offered anything to do about it.
 */
describe('Roadmap', () => {
  afterEach(() => {
    roadmapState.data = [];
    roadmapState.isError = false;
    roadmapState.isPaused = false;
    roadmapState.failureCount = 0;
  });

  it('offers a retry when the request failed', () => {
    roadmapState.isError = true;
    roadmapState.failureCount = 1;
    render(<Roadmap />);

    expect(screen.getByText('flow.loadFailed.title')).toBeInTheDocument();
    expect(screen.queryByText('flow.roadmap.emptyTitle')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'flow.loadFailed.retry' }));
    expect(roadmapState.refetch).toHaveBeenCalled();
  });

  it('treats a query paused after failing as a failure', () => {
    // networkMode 'online' never settles these into isError.
    roadmapState.isPaused = true;
    roadmapState.failureCount = 1;
    render(<Roadmap />);
    expect(screen.getByText('flow.loadFailed.title')).toBeInTheDocument();
  });

  it('says the roadmap is empty when it genuinely is', () => {
    render(<Roadmap />);
    expect(screen.getByText('flow.roadmap.emptyTitle')).toBeInTheDocument();
    expect(screen.queryByText('flow.loadFailed.title')).not.toBeInTheDocument();
  });

  it('shows neither state once projects arrive', () => {
    roadmapState.data = [
      {
        id: 'p1',
        name: 'Web rebuild',
        color: '#123456',
        description: 'A description',
        memberAvatars: [],
        tasks: [],
        taskCount: 0,
        doneCount: 0,
      },
    ];
    render(<Roadmap />);
    expect(screen.queryByText('flow.roadmap.emptyTitle')).not.toBeInTheDocument();
    expect(screen.queryByText('flow.loadFailed.title')).not.toBeInTheDocument();
  });
});
