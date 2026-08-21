import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import Valuation from './Valuation';

const mocks = vi.hoisted(() => ({
  useValuation: vi.fn(),
  refetch: vi.fn(),
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

vi.mock('./useValuation', () => ({
  useValuation: mocks.useValuation,
}));

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

    const title = screen.getByText('valuation.empty.title');
    expect(title).toBeInTheDocument();
    expect(screen.getByText('valuation.empty.description')).toBeInTheDocument();

    const emptyState = title.parentElement as HTMLElement;
    fireEvent.click(within(emptyState).getByRole('button', { name: 'valuation.refresh' }));

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

    expect(screen.queryByText('valuation.empty.title')).not.toBeInTheDocument();
  });
});
