import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, expect, it, vi, beforeEach } from 'vitest';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mocks.navigate,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const member = {
  id: 1,
  name: 'Jane Doe',
  initials: 'JD',
  points: 500,
  tier: 'Gold',
  visits: 4,
  joinedAt: '2026-01-01',
  lastActive: '2026-08-01',
};

// A stable reference is required: the component's effect depends on this
// array's identity, and a mock that returns a new array on every render (as
// react-query would not, once settled) would loop forever.
const membersList: (typeof member)[] = [];
const refetchMembers = vi.fn();

vi.mock('../../shared/hooks/queries/useLoyalty', () => ({
  useLoyaltyMembers: () => ({
    data: membersList,
    isLoading: false,
    isFetching: false,
    isError: false,
    refetch: refetchMembers,
  }),
}));

import LoyaltyMembers from './LoyaltyMembers';

function renderMembers(): void {
  render(<LoyaltyMembers />);
}

beforeEach(() => {
  mocks.navigate.mockClear();
  refetchMembers.mockClear();
  membersList.length = 0;
});

describe('LoyaltyMembers empty state', () => {
  it('shows a "view customers" call-to-action when there are no members at all', () => {
    renderMembers();

    expect(screen.getByText('crm.loyaltyMembers.trueEmptyDescription')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'crm.actions.viewCustomers' }));
    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/crm/customers' });
  });

  it('shows a "clear search" call-to-action when a search term matches nothing', () => {
    membersList.push(member);
    renderMembers();

    fireEvent.change(screen.getByPlaceholderText('crm.loyaltyMembers.searchPlaceholder'), {
      target: { value: 'nonexistent' },
    });

    expect(screen.getByText('crm.loyaltyMembers.searchEmptyDescription')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'crm.actions.clearSearch' }));
    expect(screen.getByText('Jane Doe')).toBeInTheDocument();
  });
});
