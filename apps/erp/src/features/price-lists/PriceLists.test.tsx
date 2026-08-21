import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import PriceLists from './PriceLists';

const mocks = vi.hoisted(() => ({
  usePriceLists: vi.fn(),
  refetch: vi.fn(),
  role: 'admin' as string,
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('motion/react', () => ({
  motion: {
    div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div>,
  },
}));

vi.mock('@voxire/auth', () => ({
  useUserStore: (select: (state: { user: { role: string } }) => unknown) =>
    select({ user: { role: mocks.role } }),
}));

vi.mock('./usePriceLists', () => ({
  usePriceLists: mocks.usePriceLists,
  useCreatePriceList: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

describe('PriceLists empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.usePriceLists.mockReturnValue({ data: [], isLoading: false, refetch: mocks.refetch });
  });

  it('shows the create action and opens the create form for a role with erp:purchase_orders', () => {
    mocks.role = 'admin';
    render(<PriceLists />);

    const title = screen.getByText('priceLists.empty.title');
    expect(title).toBeInTheDocument();
    expect(screen.getByText('priceLists.empty.description')).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'priceLists.newPriceList' }));

    expect(screen.getByPlaceholderText('Currency (e.g. USD)')).toBeInTheDocument();
    expect(mocks.refetch).not.toHaveBeenCalled();
  });

  it('falls back to Refresh for a role without erp:purchase_orders', () => {
    mocks.role = 'staff';
    render(<PriceLists />);

    const title = screen.getByText('priceLists.empty.title');
    const emptyState = title.parentElement as HTMLElement;

    expect(
      within(emptyState).queryByRole('button', { name: 'priceLists.newPriceList' })
    ).not.toBeInTheDocument();

    fireEvent.click(within(emptyState).getByRole('button', { name: 'priceLists.refresh' }));

    expect(mocks.refetch).toHaveBeenCalledTimes(1);
    expect(screen.queryByPlaceholderText('Currency (e.g. USD)')).not.toBeInTheDocument();
  });

  it('does not show the empty state when price lists exist', () => {
    mocks.usePriceLists.mockReturnValue({
      data: [{ id: 'l1', name: 'Retail', currency: 'USD', isDefault: true }],
      isLoading: false,
      refetch: mocks.refetch,
    });

    render(<PriceLists />);

    expect(screen.queryByText('priceLists.empty.title')).not.toBeInTheDocument();
  });
});
