import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, beforeAll } from 'vitest';

import path from 'node:path';
import i18next from 'i18next';
import { initTestI18nFrom } from '../../../test/i18nTestInstance';

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
  isError: false,
  refetch: vi.fn(),
}));

vi.mock('../../shared/hooks/queries/useGroups', () => ({
  useGroups: () => groupsData,
  useDeleteGroup: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

import Groups from './Groups';

beforeAll(() => initTestI18nFrom(path.resolve(import.meta.dirname, '../../locales')));

beforeEach(() => {
  vi.clearAllMocks();
  groupsData.data = [];
  groupsData.isLoading = false;
  groupsData.isError = false;
});

describe('Groups list empty state', () => {
  it('shows a retryable load error instead of claiming there are no groups', () => {
    groupsData.isError = true;
    render(<Groups />);

    expect(screen.getByRole('alert')).toHaveTextContent('Unable to load groups');
    expect(screen.queryByText('No groups yet')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Refresh' }));
    expect(groupsData.refetch).toHaveBeenCalledOnce();
  });

  it('shows the true-empty title, description and new-group call-to-action when there are no groups', () => {
    render(<Groups />);

    expect(screen.getByText(i18next.t('team.groups.noGroupsTitle'))).toBeInTheDocument();
    expect(screen.getByText(i18next.t('team.groups.noGroupsDescription'))).toBeInTheDocument();
  });

  it('navigates to the new-group screen when the true-empty call-to-action is clicked', () => {
    render(<Groups />);

    const emptyStateButtons = screen.getAllByRole('button', {
      name: i18next.t('team.groups.newGroup'),
    });
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

    fireEvent.change(screen.getByPlaceholderText(i18next.t('team.groups.searchPlaceholder')), {
      target: { value: 'no such group' },
    });

    expect(screen.getByText(i18next.t('team.groups.noGroupsMatch'))).toBeInTheDocument();
    expect(screen.queryByText(i18next.t('team.groups.noGroupsTitle'))).not.toBeInTheDocument();
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

    expect(screen.queryByText(i18next.t('team.groups.noGroupsTitle'))).not.toBeInTheDocument();
    expect(screen.getByText('Managers')).toBeInTheDocument();
  });
});
