import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../../shared/testUtils/i18nTestSetup';
import Revenue from './Revenue';

initTestI18n();

const mocks = vi.hoisted(() => ({
  useRevenueSeries: vi.fn(),
  useRevenueByApp: vi.fn(),
  useRevenueComparison: vi.fn(),
  useInsightsSettings: vi.fn(),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('../../shared/hooks/queries/useRevenue', () => ({
  useRevenueSeries: mocks.useRevenueSeries,
  useRevenueByApp: mocks.useRevenueByApp,
  useRevenueComparison: mocks.useRevenueComparison,
}));

vi.mock('../../shared/hooks/queries/useInsightsSettings', () => ({
  useInsightsSettings: mocks.useInsightsSettings,
}));

describe('Revenue empty states', () => {
  const refetchTrend = vi.fn();
  const refetchByApp = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useRevenueSeries.mockReturnValue({
      data: { series: [] },
      isLoading: false,
      isError: false,
      refetch: refetchTrend,
    });
    mocks.useRevenueByApp.mockReturnValue({
      data: { byApp: [] },
      isLoading: false,
      refetch: refetchByApp,
    });
    mocks.useRevenueComparison.mockReturnValue({
      data: { thisPeriod: 0, lastPeriod: 0, deltaPct: 0 },
      isLoading: false,
    });
    mocks.useInsightsSettings.mockReturnValue({ data: { defaultCurrency: 'USD' } });
  });

  it('shows a distinct title, description and working refresh action for the revenue trend', () => {
    render(<Revenue />);

    expect(screen.getByText('No revenue trend yet')).toBeInTheDocument();
    expect(
      screen.getByText('Revenue data will appear here once transactions are recorded.')
    ).toBeInTheDocument();

    fireEvent.click(screen.getAllByRole('button', { name: 'Refresh' })[0]);
    expect(refetchTrend).toHaveBeenCalledTimes(1);
  });

  it('shows a distinct title, description and working refresh action for the by-app breakdown', () => {
    render(<Revenue />);

    expect(screen.getByText('No revenue breakdown yet')).toBeInTheDocument();
    expect(
      screen.getByText(
        'A breakdown by app will appear once revenue comes in from more than one app.'
      )
    ).toBeInTheDocument();

    fireEvent.click(screen.getAllByRole('button', { name: 'Refresh' })[1]);
    expect(refetchByApp).toHaveBeenCalledTimes(1);
  });
});
