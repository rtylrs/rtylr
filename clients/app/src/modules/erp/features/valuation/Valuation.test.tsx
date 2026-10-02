import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import Valuation from './Valuation';
import { initTestI18n } from '../../shared/testing/i18n';

const mocks = vi.hoisted(() => ({
  useValuation: vi.fn(),
  refetch: vi.fn(),
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

vi.mock('./useValuation', () => ({
  useValuation: mocks.useValuation,
}));

beforeAll(async () => {
  await initTestI18n('en');
});

describe('Valuation empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useValuation.mockReturnValue({
      data: { items: [], totalValue: 0 },
      isLoading: false,
      refetch: mocks.refetch,
    });
  });

  it('shows the illustration, empty copy, and refetches on CTA click', () => {
    render(<Valuation />);

    const title = screen.getByText('No valuation data available');
    expect(title).toBeInTheDocument();
    expect(
      screen.getByText('Valuation appears once inventory is tracked in a warehouse.')
    ).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'Refresh' }));

    expect(mocks.refetch).toHaveBeenCalledTimes(1);
  });

  it('does not show the empty state when categories have valuation data', () => {
    mocks.useValuation.mockReturnValue({
      data: {
        items: [
          {
            categoryName: 'Beverages',
            totalValue: 100,
            totalCost: 60,
            totalItems: 10,
          },
        ],
        totalValue: 100,
      },
      isLoading: false,
      refetch: mocks.refetch,
    });

    render(<Valuation />);

    expect(screen.queryByText('No valuation data available')).not.toBeInTheDocument();
  });
});
