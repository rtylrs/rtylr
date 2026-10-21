// @vitest-environment happy-dom
import React from 'react';
import { describe, it, expect, vi, afterEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';

const teamState = vi.hoisted(() => ({
  data: [] as Record<string, unknown>[],
  isLoading: false,
  isError: false,
  isPaused: false,
  failureCount: 0,
  refetch: vi.fn(),
}));
const navigateMock = vi.hoisted(() => vi.fn());

vi.mock('./useTeamWorkload', () => ({ useTeamWorkload: () => teamState }));
vi.mock('react-i18next', () => ({ useTranslation: () => ({ t: (k: string) => k }) }));
vi.mock('@tanstack/react-router', () => ({ useNavigate: () => navigateMock }));
vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children: React.ReactNode }) => <div>{children}</div> },
}));

import Team from './Team';

/**
 * An empty team and a failed request both left this screen showing a header
 * over blank space (#1072), with nothing to distinguish them and nothing to do.
 */
describe('Team', () => {
  afterEach(() => {
    teamState.data = [];
    teamState.isError = false;
    teamState.isPaused = false;
    teamState.failureCount = 0;
  });

  it('offers a retry when the request failed', () => {
    teamState.isError = true;
    teamState.failureCount = 1;
    render(<Team />);

    expect(screen.getByText('flow.loadFailed.title')).toBeInTheDocument();
    expect(screen.queryByText('flow.team.emptyTitle')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'flow.loadFailed.retry' }));
    expect(teamState.refetch).toHaveBeenCalled();
  });

  it('treats a query paused after failing as a failure', () => {
    teamState.isPaused = true;
    teamState.failureCount = 1;
    render(<Team />);
    expect(screen.getByText('flow.loadFailed.title')).toBeInTheDocument();
  });

  it('says the team is empty when it genuinely is', () => {
    render(<Team />);
    expect(screen.getByText('flow.team.emptyTitle')).toBeInTheDocument();
    expect(screen.queryByText('flow.loadFailed.title')).not.toBeInTheDocument();
  });
});
