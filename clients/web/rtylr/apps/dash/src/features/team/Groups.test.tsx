import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('motion/react', () => ({
  motion: {
    div: (props: React.ComponentProps<'div'>) => <div {...props} />,
  },
}));

const navigateMock = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => navigateMock,
}));

type Group = {
  id: number;
  name: string;
  color: string;
  memberCount: number;
  licensedAppIds: string[];
  permissionProfileAssignments: { appId: string; permissionProfileIds: string[] }[];
};

const groupsData = vi.hoisted(() => ({
  data: [] as Group[],
  isLoading: false,
}));

vi.mock('../../shared/hooks/queries/useGroups', () => ({
  useGroups: () => groupsData,
  useDeleteGroup: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

import Groups from './Groups';

beforeEach(() => {
  vi.clearAllMocks();
  groupsData.data = [];
  groupsData.isLoading = false;
});

describe('Groups list empty state', () => {
  it('shows the true-empty title, description and new-group call-to-action when there are no groups', () => {
    render(<Groups />);

    expect(screen.getByText('team.groups.noGroupsTitle')).toBeInTheDocument();
    expect(screen.getByText('team.groups.noGroupsDescription')).toBeInTheDocument();
  });

  it('navigates to the new-group screen when the true-empty call-to-action is clicked', () => {
    render(<Groups />);

    const emptyStateButtons = screen.getAllByRole('button', { name: 'team.groups.newGroup' });
    // The page header also has a "New group" button; click the one inside the empty state card.
    fireEvent.click(emptyStateButtons[emptyStateButtons.length - 1]);

    expect(navigateMock).toHaveBeenCalledWith({ to: '/team/groups/new' });
  });

  it('shows the search-empty title when a search matches no groups', () => {
    groupsData.data = [
      {
        id: 1,
        name: 'Managers',
        color: '#4583C4',
        memberCount: 2,
        licensedAppIds: [],
        permissionProfileAssignments: [],
      },
    ];
    render(<Groups />);

    fireEvent.change(screen.getByPlaceholderText('team.groups.searchPlaceholder'), {
      target: { value: 'no such group' },
    });

    expect(screen.getByText('team.groups.noGroupsMatch')).toBeInTheDocument();
    expect(screen.queryByText('team.groups.noGroupsTitle')).not.toBeInTheDocument();
  });

  it('does not show the empty state when groups match', () => {
    groupsData.data = [
      {
        id: 1,
        name: 'Managers',
        color: '#4583C4',
        memberCount: 2,
        licensedAppIds: [],
        permissionProfileAssignments: [],
      },
    ];
    render(<Groups />);

    expect(screen.queryByText('team.groups.noGroupsTitle')).not.toBeInTheDocument();
    expect(screen.getByText('Managers')).toBeInTheDocument();
  });
});
