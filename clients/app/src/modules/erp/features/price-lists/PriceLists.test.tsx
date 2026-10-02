import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import PriceLists from './PriceLists';
import { initTestI18n } from '../../shared/testing/i18n';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';

const mocks = vi.hoisted(() => ({
  usePriceLists: vi.fn(),
  refetch: vi.fn(),
  effectiveAccess: undefined as unknown,
}));

// i18next is deliberately not mocked. An identity `t` makes an assertion and
// the component agree on a key, so neither notices the key never resolves
// (rtylr#1269). The real locale files mean the copy below also proves the lookup.

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('motion/react', () => ({
  motion: {
    div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div>,
  },
}));

vi.mock('../../../../shared/stores/userStore', () => ({
  useUserStore: Object.assign(
    (
      select: (state: { user: { effectiveAccess: unknown }; effectiveAccess: unknown }) => unknown
    ) =>
      select({
        user: { effectiveAccess: mocks.effectiveAccess },
        effectiveAccess: mocks.effectiveAccess,
      }),
    {
      getState: () => ({
        user: { effectiveAccess: mocks.effectiveAccess },
        effectiveAccess: mocks.effectiveAccess,
      }),
    }
  ),
}));

vi.mock('./usePriceLists', () => ({
  usePriceLists: mocks.usePriceLists,
  useCreatePriceList: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

beforeAll(async () => {
  await initTestI18n('en');
});

describe('PriceLists empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.effectiveAccess = createEffectiveAccess('erp', { isOwner: true });
    mocks.usePriceLists.mockReturnValue({ data: [], isLoading: false, refetch: mocks.refetch });
  });

  it('shows the create action and opens the create form with price-list management permission', () => {
    mocks.effectiveAccess = createEffectiveAccess('erp', {
      permissions: ['erp:price_lists:manage'],
    });
    render(<PriceLists />);

    const title = screen.getByText('No price lists');
    expect(title).toBeInTheDocument();
    expect(
      screen.getByText('Create your first price list to enable tier pricing')
    ).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'New Price List' }));

    expect(screen.getByPlaceholderText('Currency (e.g. USD)')).toBeInTheDocument();
    expect(mocks.refetch).not.toHaveBeenCalled();
  });

  it('falls back to Refresh without price-list management permission', () => {
    mocks.effectiveAccess = createEffectiveAccess('erp', { permissions: ['erp:price_lists:view'] });
    render(<PriceLists />);

    const title = screen.getByText('No price lists');
    const emptyState = title.parentElement as HTMLElement;

    expect(
      within(emptyState).queryByRole('button', { name: 'New Price List' })
    ).not.toBeInTheDocument();

    fireEvent.click(within(emptyState).getByRole('button', { name: 'Refresh' }));

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

    expect(screen.queryByText('No price lists')).not.toBeInTheDocument();
  });
});
