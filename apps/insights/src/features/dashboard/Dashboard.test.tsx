import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../../shared/testUtils/i18nTestSetup';
import Dashboard from './Dashboard';

initTestI18n();

const mocks = vi.hoisted(() => ({
  useInsightsDashboard: vi.fn(),
  useRevenueSeries: vi.fn(),
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

vi.mock('../../shared/hooks/queries/useInsightsDashboard', () => ({
  useInsightsDashboard: mocks.useInsightsDashboard,
}));

vi.mock('../../shared/hooks/queries/useRevenue', () => ({
  useRevenueSeries: mocks.useRevenueSeries,
}));

vi.mock('../../shared/hooks/queries/useInsightsSettings', () => ({
  useInsightsSettings: mocks.useInsightsSettings,
}));

describe('Dashboard empty state', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.useRevenueSeries.mockReturnValue({ data: { series: [] } });
    mocks.useInsightsSettings.mockReturnValue({ data: { defaultCurrency: 'USD' } });
  });

  it('shows the illustration, real title/description and a refresh action when every KPI is zero', () => {
    const refetch = vi.fn();
    mocks.useInsightsDashboard.mockReturnValue({
      data: {
        revenue: { value: 0, deltaPct: 0 },
        customers: { value: 0, deltaPct: 0 },
        orders: { value: 0, deltaPct: 0 },
        margin: { value: 0, deltaPct: 0 },
      },
      isLoading: false,
      isError: false,
      refetch,
    });

    render(<Dashboard />);

    expect(screen.getByText('No KPI data yet')).toBeInTheDocument();
    expect(
      screen.getByText('Metrics will appear here once your business starts generating activity.')
    ).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Refresh' }));
    expect(refetch).toHaveBeenCalledTimes(1);
  });

  it('does not show the empty state once any KPI is non-zero', () => {
    mocks.useInsightsDashboard.mockReturnValue({
      data: {
        revenue: { value: 12000, deltaPct: 5 },
        customers: { value: 0, deltaPct: 0 },
        orders: { value: 0, deltaPct: 0 },
        margin: { value: 0, deltaPct: 0 },
      },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    });

    render(<Dashboard />);

    expect(screen.queryByText('No KPI data yet')).not.toBeInTheDocument();
  });
});
