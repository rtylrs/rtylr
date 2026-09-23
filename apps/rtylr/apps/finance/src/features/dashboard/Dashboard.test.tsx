import React from 'react';
import { render, screen } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import Dashboard from './Dashboard';
import { initTestI18n } from '../../shared/testing/i18n';

const { dashboardMock } = vi.hoisted(() => ({
  dashboardMock: vi.fn(),
}));

vi.mock('./useDashboard', () => ({
  useFinanceDashboard: () => dashboardMock(),
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children?: React.ReactNode }) => <div>{children}</div> },
}));

vi.mock('@voxire/ui', () => ({
  ErrorBoundary: ({ children }: { children?: React.ReactNode }) => <>{children}</>,
}));

vi.mock('../../shared/components/FinanceKpiWidgets', () => ({ default: () => null }));
vi.mock('../../shared/components/OverdueInvoicesWidget', () => ({ default: () => null }));

beforeAll(async () => {
  await initTestI18n('en');
});

describe('Dashboard query states', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders a failure instead of default zero KPIs and empty-data messages', () => {
    dashboardMock.mockReturnValue({
      data: undefined,
      isPending: false,
      isError: true,
      isFetching: false,
      refetch: vi.fn(),
    });

    render(<Dashboard />);

    expect(screen.getByRole('alert')).toHaveTextContent('Something went wrong. Please try again.');
    expect(screen.queryByText('Total Revenue (MTD)')).not.toBeInTheDocument();
    expect(screen.queryByText('No chart data available')).not.toBeInTheDocument();
  });
});
