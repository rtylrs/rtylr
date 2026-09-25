import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mocks.navigate,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const customer = {
  id: 1,
  name: 'Jane Doe',
  initials: 'JD',
  email: 'jane@example.com',
  phone: '555-1234',
  totalSpent: 480,
  visits: 3,
  loyaltyPoints: 120,
  loyaltyTier: 'Bronze',
  lastVisit: '2026-08-01',
  tags: [],
};

// Stable array references are required: the component's effect depends on
// array identity, and a mock that returns a new array on every render (as
// react-query would not, once settled) would loop forever.
const customersList: (typeof customer)[] = [];
const emptyList: (typeof customer)[] = [];

vi.mock('../../shared/hooks/queries/useCustomers', () => ({
  // Customers.tsx delegates search filtering to the query itself (the `search`
  // param), rather than filtering the returned list client-side, so the mock
  // has to honor that param to exercise the search-empty branch.
  useListCustomers: (params: { search?: string }) => ({
    data: params.search ? emptyList : customersList,
    isLoading: false,
    isFetching: false,
    isError: false,
  }),
  useUpdateCustomer: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/hooks/queries/useLoyalty', () => ({
  useLoyaltyConfig: () => ({ data: undefined }),
}));

import Customers from './Customers';

function renderCustomers(): void {
  render(<Customers />);
}

beforeEach(() => {
  mocks.navigate.mockClear();
  customersList.length = 0;
});

describe('Customers empty state', () => {
  it('shows a "new customer" call-to-action when there are no customers at all', () => {
    renderCustomers();

    expect(screen.getByText('crm.customers.empty')).toBeInTheDocument();
    expect(screen.getByText('crm.customers.createFirst')).toBeInTheDocument();

    // The header action and the empty-state action are the same action offered
    // twice, and both now go through crm.customers.new, so they share an
    // accessible name. Assert both are there rather than matching one by luck.
    const newCustomerButtons = screen.getAllByRole('button', { name: 'crm.customers.new' });
    expect(newCustomerButtons).toHaveLength(2);
    fireEvent.click(newCustomerButtons[1]);

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/customers/new' });
  });

  it('shows a "clear search" call-to-action when a search term matches nothing', () => {
    customersList.push(customer);
    renderCustomers();

    fireEvent.change(screen.getByPlaceholderText('crm.customers.searchPlaceholder'), {
      target: { value: 'nonexistent' },
    });

    expect(screen.getByText('crm.customers.emptySearchDescription')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'crm.actions.clearSearch' }));

    expect(screen.getByText('Jane Doe')).toBeInTheDocument();
  });
});
