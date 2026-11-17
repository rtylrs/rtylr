import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../../shared/testUtils/i18nTestSetup';
import Dashboard, { formatToday } from './Dashboard';

initTestI18n();

const mocks = vi.hoisted(() => ({
  useInsightsDashboard: vi.fn(),
  useRevenueSeries: vi.fn(),
  useInsightsSettings: vi.fn(),
  useWorkSummary: vi.fn(),
  useSalesSummary: vi.fn(),
  useInvoiceAging: vi.fn(),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('../../shared/hooks/queries/useInsightsDashboard', () => ({
  useInsightsDashboard: mocks.useInsightsDashboard,
  useWorkSummary: mocks.useWorkSummary,
  useSalesSummary: mocks.useSalesSummary,
}));

vi.mock('../../shared/hooks/queries/useInsightsFinancial', () => ({
  useInvoiceAging: mocks.useInvoiceAging,
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
    mocks.useWorkSummary.mockReturnValue({ data: undefined });
    mocks.useSalesSummary.mockReturnValue({ data: undefined });
    mocks.useInvoiceAging.mockReturnValue({ data: undefined });
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

    mocks.useWorkSummary.mockReturnValue({
      data: { totalTasks: 0, completedTasks: 0, overdueTasks: 0, projects: [] },
    });
    mocks.useInvoiceAging.mockReturnValue({ data: { buckets: [] } });
    mocks.useSalesSummary.mockReturnValue({
      data: {
        openPipelineValue: 0,
        openDealCount: 0,
        won: { value: 0, deltaPct: 0 },
        recentWonDealCount: 0,
      },
    });

    render(<Dashboard />);

    expect(screen.queryByText('No KPI data yet')).not.toBeInTheDocument();
    // A business without deals, Work tasks or unpaid invoices gets no empty cards.
    expect(screen.queryByText('Sales, projects and invoices')).not.toBeInTheDocument();
    expect(screen.queryByText('Open pipeline')).not.toBeInTheDocument();
    expect(screen.queryByText('Won this month')).not.toBeInTheDocument();
    expect(screen.queryByText('Tasks done')).not.toBeInTheDocument();
    expect(screen.queryByText('Unpaid invoices')).not.toBeInTheDocument();
  });

  it('shows the CRM pipeline and won deal cards for a business with no POS activity', () => {
    mocks.useInsightsDashboard.mockReturnValue({
      data: {
        revenue: { value: 0, deltaPct: 0 },
        customers: { value: 0, deltaPct: 0 },
        orders: { value: 0, deltaPct: 0 },
        margin: { value: 0, deltaPct: 0 },
      },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    });
    mocks.useSalesSummary.mockReturnValue({
      data: {
        openPipelineValue: 4500000,
        openDealCount: 12,
        won: { value: 300000, deltaPct: 50 },
        recentWonDealCount: 4,
      },
    });

    render(<Dashboard />);

    expect(screen.queryByText('No KPI data yet')).not.toBeInTheDocument();
    expect(
      screen.getByRole('heading', { name: 'Sales, projects and invoices' })
    ).toBeInTheDocument();
    const pipeline = screen.getByText('Open pipeline').closest('div')?.parentElement;
    expect(pipeline).toHaveTextContent('45,000');
    expect(pipeline).toHaveTextContent('12 open deals');
    const won = screen.getByText('Won this month').closest('div')?.parentElement;
    expect(won).toHaveTextContent('3,000');
    expect(won).toHaveTextContent('+50.0% vs prior');
  });

  it('keeps the won card when last month had wins but nothing is open', () => {
    mocks.useInsightsDashboard.mockReturnValue({
      data: {
        revenue: { value: 0, deltaPct: 0 },
        customers: { value: 0, deltaPct: 0 },
        orders: { value: 0, deltaPct: 0 },
        margin: { value: 0, deltaPct: 0 },
      },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    });
    mocks.useSalesSummary.mockReturnValue({
      data: {
        openPipelineValue: 0,
        openDealCount: 0,
        won: { value: 0, deltaPct: -100 },
        recentWonDealCount: 1,
      },
    });

    render(<Dashboard />);

    expect(screen.queryByText('No KPI data yet')).not.toBeInTheDocument();
    expect(screen.getByText('Open pipeline').closest('div')?.parentElement).toHaveTextContent(
      '0 open deals'
    );
    expect(screen.getByText('Won this month').closest('div')?.parentElement).toHaveTextContent(
      '-100.0% vs prior'
    );
  });

  it('shows Work and invoice figures for a business with no POS activity', () => {
    mocks.useInsightsDashboard.mockReturnValue({
      data: {
        revenue: { value: 0, deltaPct: 0 },
        customers: { value: 0, deltaPct: 0 },
        orders: { value: 0, deltaPct: 0 },
        margin: { value: 0, deltaPct: 0 },
      },
      isLoading: false,
      isError: false,
      refetch: vi.fn(),
    });
    mocks.useWorkSummary.mockReturnValue({
      data: {
        totalTasks: 61,
        completedTasks: 12,
        overdueTasks: 4,
        projects: [
          {
            id: '9',
            name: 'Investment Hub',
            totalTasks: 61,
            completedTasks: 12,
            progressPct: 19.7,
          },
        ],
      },
    });
    mocks.useInvoiceAging.mockReturnValue({
      data: {
        buckets: [
          { bucket: 'current', invoiceCount: 1, amountMinorUnits: 150000 },
          { bucket: '1-30', invoiceCount: 1, amountMinorUnits: 50000 },
        ],
      },
    });

    render(<Dashboard />);

    expect(screen.queryByText('No KPI data yet')).not.toBeInTheDocument();
    expect(screen.getByText('12 of 61', { selector: 'p' })).toBeInTheDocument();
    expect(screen.getByText('Overdue tasks').closest('div')?.parentElement).toHaveTextContent('4');
    expect(screen.getByText('Unpaid invoices').closest('div')?.parentElement).toHaveTextContent(
      '2,000'
    );
    expect(screen.getByText('Overdue invoices').closest('div')?.parentElement).toHaveTextContent(
      '500'
    );
    expect(screen.getByRole('progressbar', { name: 'Investment Hub' })).toHaveAttribute(
      'aria-valuenow',
      '20'
    );
  });
});

describe('formatToday', () => {
  it('formats the date in the reader language, not always English', () => {
    expect(formatToday('en')).toMatch(/[A-Za-z]/);
    expect(formatToday('ar')).toMatch(/[؀-ۿ]/);
  });
});
