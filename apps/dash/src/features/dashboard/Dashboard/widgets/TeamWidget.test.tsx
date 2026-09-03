import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

const navigateMock = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => navigateMock,
}));

const teamData = vi.hoisted(() => ({
  data: { members: [] as { id: string }[] } as { members: { id: string }[] } | undefined,
}));
vi.mock('../../../../shared/hooks/queries/useTeam', () => ({
  useTeam: () => teamData,
}));

import TeamWidget from './TeamWidget';

beforeEach(() => {
  vi.clearAllMocks();
  teamData.data = { members: [] };
});

describe('TeamWidget empty state', () => {
  it('shows the illustration, title and an invite call-to-action when there are no members', () => {
    render(<TeamWidget />);

    expect(screen.getByText('dashboard.teamWidget.emptyTitle')).toBeInTheDocument();
    expect(
      screen.getByRole('button', { name: 'dashboard.teamWidget.addMember' })
    ).toBeInTheDocument();
  });

  it('navigates to the invite screen when the call-to-action is clicked', () => {
    render(<TeamWidget />);

    fireEvent.click(screen.getByRole('button', { name: 'dashboard.teamWidget.addMember' }));

    expect(navigateMock).toHaveBeenCalledWith({ to: '/team/invite' });
  });

  it('does not show the empty state when there are members', () => {
    teamData.data = { members: [{ id: '1' }] };
    render(<TeamWidget />);

    expect(screen.queryByText('dashboard.teamWidget.emptyTitle')).not.toBeInTheDocument();
  });
});
