import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  pages: new Map<string, unknown[]>(),
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mocks.navigate,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

function customerRow(id: number) {
  return {
    id,
    name: `Customer ${id}`,
    initials: 'CU',
    email: `customer${id}@example.com`,
    phone: '555-1234',
    totalSpent: 100,
    visits: 1,
    loyaltyPoints: 10,
    loyaltyTier: 'Bronze',
    lastVisit: '2026-08-01',
    tags: [],
  };
}

// Stable array references per query: the component's effect depends on the
// array identity, so returning a fresh array on every render would loop.
function pageRows(search: string, offset: number): unknown[] {
  const key = `${search}:${offset}`;
  const existing = mocks.pages.get(key);
  if (existing) return existing;
  const rows: unknown[] = [];
  mocks.pages.set(key, rows);
  return rows;
}

vi.mock('../../shared/hooks/queries/useCustomers', () => ({
  // Honors search/offset like the real query so pagination state can be
  // exercised; an unseeded key is an empty page.
  useListCustomers: (params: { search?: string; offset: number }) => ({
    data: pageRows(params.search ?? '', params.offset),
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

const loadMore = () => screen.queryByRole('button', { name: 'Load more customers' });

beforeEach(() => {
  mocks.navigate.mockClear();
  mocks.pages.clear();
  mocks.pages.set(
    ':0',
    Array.from({ length: 100 }, (_, i) => customerRow(i + 1))
  );
  mocks.pages.set(':100', []);
  mocks.pages.set(
    'jo:0',
    Array.from({ length: 100 }, (_, i) => customerRow(i + 1000))
  );
});

describe('Customers pagination', () => {
  it('offers Load more again after a search change when the previous page was exhausted', () => {
    render(<Customers />);

    expect(loadMore()).toBeInTheDocument();

    // The exact page-size multiple refreshes to an empty extra page, which
    // hides the button for the current query.
    fireEvent.click(loadMore() as HTMLElement);
    expect(loadMore()).not.toBeInTheDocument();

    fireEvent.change(screen.getByPlaceholderText('crm.customers.searchPlaceholder'), {
      target: { value: 'jo' },
    });

    expect(loadMore()).toBeInTheDocument();
  });

  it('offers Load more again after clearing a search when the previous page was exhausted', () => {
    render(<Customers />);

    fireEvent.click(loadMore() as HTMLElement);
    expect(loadMore()).not.toBeInTheDocument();

    // Clearing through the empty-state action must reset the same state the
    // search input resets.
    fireEvent.change(screen.getByPlaceholderText('crm.customers.searchPlaceholder'), {
      target: { value: 'nomatch' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'crm.actions.clearSearch' }));

    expect(loadMore()).toBeInTheDocument();
  });
});
