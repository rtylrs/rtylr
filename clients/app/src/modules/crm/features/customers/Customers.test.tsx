import React from 'react';
import path from 'node:path';
import { initTestI18nFrom } from '../../../../../test/i18nTestInstance';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';
vi.unmock('react-i18next');
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  access: {} as unknown,
  failed: false,
  pending: false,
  retry: vi.fn(),
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mocks.navigate,
}));

vi.mock('../../../../shared/stores/userStore', () => ({
  useUserStore: { getState: () => ({ user: { effectiveAccess: mocks.access } }) },
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
    isPending: mocks.pending,
    isPaused: mocks.pending,
    failureCount: 0,
    isFetching: false,
    isError: mocks.failed,
    refetch: mocks.retry,
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

beforeEach(async () => {
  await initTestI18nFrom(path.resolve(__dirname, '../../locales'));
  mocks.access = createEffectiveAccess('crm', { isOwner: true });
  mocks.failed = false;
  mocks.pending = false;
  mocks.retry.mockClear();
  mocks.navigate.mockClear();
  customersList.length = 0;
});

describe('Customers empty state', () => {
  it('shows a "new customer" call-to-action when there are no customers at all', () => {
    renderCustomers();

    expect(screen.getByText('No customers found')).toBeInTheDocument();
    expect(screen.getByText('Create your first customer to get started.')).toBeInTheDocument();

    // The header action and the empty-state action are the same action offered
    // twice, and both now go through crm.customers.new, so they share an
    // accessible name. Assert both are there rather than matching one by luck.
    const newCustomerButtons = screen.getAllByRole('button', { name: 'New customer' });
    expect(newCustomerButtons).toHaveLength(2);
    fireEvent.click(newCustomerButtons[1]);

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/crm/customers/new' });
  });

  it('shows a "clear search" call-to-action when a search term matches nothing', () => {
    customersList.push(customer);
    renderCustomers();

    fireEvent.change(screen.getByPlaceholderText('Search customers...'), {
      target: { value: 'nonexistent' },
    });

    expect(screen.getByText('No customers match "nonexistent"')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Clear search' }));

    expect(screen.getByText('Jane Doe')).toBeInTheDocument();
  });
});

describe('Restricted member customer list', () => {
  it('hides creation, importing, exporting and bulk edits from a read-only member', () => {
    mocks.access = createEffectiveAccess('crm', { permissions: ['crm:customers:view'] });
    customersList.push(customer);
    renderCustomers();
    expect(screen.getByText('Jane Doe')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'New customer' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Import' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Export CSV' })).not.toBeInTheDocument();
    for (const checkbox of screen.getAllByRole('checkbox')) expect(checkbox).toBeDisabled();
  });
  it('offers retry instead of an empty list when loading fails', () => {
    mocks.failed = true;
    renderCustomers();
    expect(screen.getByRole('alert')).toHaveTextContent('Unable to load customers');
    fireEvent.click(screen.getByRole('button', { name: 'Try again' }));
    expect(mocks.retry).toHaveBeenCalledOnce();
  });

  it('keeps the list in a loading state when the first request is paused offline', () => {
    mocks.pending = true;
    renderCustomers();

    expect(screen.queryByText('No customers found')).not.toBeInTheDocument();
  });
});
