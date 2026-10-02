import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../../shared/testUtils/i18nTestSetup';
import Products from './Products';

initTestI18n();

const mocks = vi.hoisted(() => ({
  useProductKPIs: vi.fn(),
  useBestSellers: vi.fn(),
  useSlowMovers: vi.fn(),
  useProductMargins: vi.fn(),
  useProductCategoryMix: vi.fn(),
  useInsightsSettings: vi.fn(),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('motion/react', () => ({
  motion: {
    div: ({ children, ...props }: React.PropsWithChildren<Record<string, unknown>>) => (
      <div {...props}>{children}</div>
    ),
  },
}));

vi.mock('../../shared/hooks/queries/useInsightsProducts', () => ({
  useProductKPIs: mocks.useProductKPIs,
  useBestSellers: mocks.useBestSellers,
  useSlowMovers: mocks.useSlowMovers,
  useProductMargins: mocks.useProductMargins,
  useProductCategoryMix: mocks.useProductCategoryMix,
}));

vi.mock('../../shared/hooks/queries/useInsightsSettings', () => ({
  useInsightsSettings: mocks.useInsightsSettings,
}));

describe('Products empty states', () => {
  const refetchBest = vi.fn();
  const refetchSlow = vi.fn();
  const refetchMargins = vi.fn();
  const refetchCategories = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useProductKPIs.mockReturnValue({
      data: { totalSkus: 0, activeSkus: 0, outOfStock: 0, avgMarginPct: 0 },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    });
    mocks.useBestSellers.mockReturnValue({
      data: { products: [] },
      isLoading: false,
      refetch: refetchBest,
    });
    mocks.useSlowMovers.mockReturnValue({
      data: { products: [] },
      isLoading: false,
      refetch: refetchSlow,
    });
    mocks.useProductMargins.mockReturnValue({
      data: { buckets: [] },
      isLoading: false,
      refetch: refetchMargins,
    });
    mocks.useProductCategoryMix.mockReturnValue({
      data: { categories: [] },
      isLoading: false,
      refetch: refetchCategories,
    });
    mocks.useInsightsSettings.mockReturnValue({ data: { defaultCurrency: 'USD' } });
  });

  it('shows a distinct title, description and working refresh action for best sellers', () => {
    render(<Products />);

    expect(screen.getByText('No best sellers yet')).toBeInTheDocument();
    expect(
      screen.getByText('Your top-selling products will appear here once orders start coming in.')
    ).toBeInTheDocument();

    fireEvent.click(screen.getAllByRole('button', { name: 'Refresh' })[0]);
    expect(refetchBest).toHaveBeenCalledTimes(1);
  });

  it('shows a distinct title, description and working refresh action for margins', () => {
    render(<Products />);

    expect(screen.getByText('No margin data yet')).toBeInTheDocument();
    expect(
      screen.getByText('Margin distribution will appear once product costs and sales are recorded.')
    ).toBeInTheDocument();

    fireEvent.click(screen.getAllByRole('button', { name: 'Refresh' })[1]);
    expect(refetchMargins).toHaveBeenCalledTimes(1);
  });

  it('shows a distinct title, description and working refresh action for slow movers', () => {
    render(<Products />);

    expect(screen.getByText('No slow movers yet')).toBeInTheDocument();
    expect(
      screen.getByText(
        'Products with low turnover will appear here once you have enough sales history.'
      )
    ).toBeInTheDocument();

    fireEvent.click(screen.getAllByRole('button', { name: 'Refresh' })[2]);
    expect(refetchSlow).toHaveBeenCalledTimes(1);
  });

  it('shows a distinct title, description and working refresh action for category mix', () => {
    render(<Products />);

    expect(screen.getByText('No category data yet')).toBeInTheDocument();
    expect(
      screen.getByText(
        'A breakdown by category will appear once products are assigned to categories and sold.'
      )
    ).toBeInTheDocument();

    fireEvent.click(screen.getAllByRole('button', { name: 'Refresh' })[3]);
    expect(refetchCategories).toHaveBeenCalledTimes(1);
  });
});
