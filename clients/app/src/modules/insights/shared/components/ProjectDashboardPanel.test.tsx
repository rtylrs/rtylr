import React from 'react';
import { render, screen, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../testUtils/i18nTestSetup';
import ProjectDashboardPanel, { formatDueDate } from './ProjectDashboardPanel';
import type { ProjectDashboardResponse } from '../hooks/queries/types';

initTestI18n();

const mocks = vi.hoisted(() => ({
  useProjectDashboard: vi.fn(),
  useInsightsSettings: vi.fn(),
}));

vi.mock('../hooks/queries/useInsightsDashboard', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../hooks/queries/useInsightsDashboard')>();
  return {
    requestErrorStatus: actual.requestErrorStatus,
    useProjectDashboard: mocks.useProjectDashboard,
  };
});

vi.mock('../hooks/queries/useInsightsSettings', () => ({
  useInsightsSettings: mocks.useInsightsSettings,
}));

vi.mock('@tanstack/react-router', () => ({
  Link: ({
    children,
    to,
    params,
    search,
    className,
  }: React.PropsWithChildren<{
    to: string;
    params?: Record<string, string>;
    search?: Record<string, string>;
    className?: string;
  }>) => {
    const path = Object.entries(params ?? {}).reduce(
      (out, [key, value]) => out.replace(`$${key}`, value),
      to
    );
    const query = new URLSearchParams(search ?? {}).toString();
    return (
      <a href={query ? `${path}?${query}` : path} className={className}>
        {children}
      </a>
    );
  },
}));

const dashboard: ProjectDashboardResponse = {
  project: { id: '8', name: 'Website', status: 'active', visibility: 'public' },
  totalTasks: 8,
  completedTasks: 2,
  progressPct: 25,
  overdueTaskCount: 7,
  overdueTasks: [
    { id: '51', title: 'Fix footer', dueDate: '2026-10-01', assigneeId: '7' },
    { id: '52', title: 'Write copy', dueDate: '2026-10-02', assigneeId: '' },
  ],
  hoursThisMonth: { value: 2400, deltaPct: 50 },
  priorMinutes: 1600,
  allTimeMinutes: 9000,
};

const card = (label: string): HTMLElement =>
  screen.getByText(label, { selector: 'span' }).closest('div')!.parentElement!;

function loaded(data: ProjectDashboardResponse) {
  mocks.useProjectDashboard.mockReturnValue({
    data,
    isLoading: false,
    isError: false,
    error: null,
    refetch: vi.fn(),
  });
}

function failed(status: number) {
  mocks.useProjectDashboard.mockReturnValue({
    data: undefined,
    isLoading: false,
    isError: true,
    error: Object.assign(new Error('failed'), { status }),
    refetch: vi.fn(),
  });
}

beforeEach(() => {
  vi.clearAllMocks();
  mocks.useInsightsSettings.mockReturnValue({ data: { defaultCurrency: 'USD' } });
});

describe('ProjectDashboardPanel', () => {
  it('shows progress, overdue work and hours for the project', () => {
    loaded(dashboard);
    render(<ProjectDashboardPanel projectId="8" />);

    expect(mocks.useProjectDashboard).toHaveBeenCalledWith('8');
    expect(card('Progress')).toHaveTextContent('25%');
    expect(card('Progress')).toHaveTextContent('2 of 8 tasks done');
    expect(screen.getByRole('progressbar', { name: 'Progress' })).toHaveAttribute(
      'aria-valuenow',
      '25'
    );
    expect(card('Overdue tasks')).toHaveTextContent('7');
    expect(card('Hours logged (this month)')).toHaveTextContent('40 h');
    expect(card('Hours logged (this month)')).toHaveTextContent('+50.0% vs prior');
    expect(card('Hours logged (all time)')).toHaveTextContent('150 h');
  });

  it('links each overdue task to the task in Work and counts the rest', () => {
    loaded(dashboard);
    render(<ProjectDashboardPanel projectId="8" />);

    const list = screen.getByRole('region', { name: 'Overdue tasks' });
    const link = within(list).getByRole('link', { name: /Fix footer/ });
    expect(link).toHaveAttribute('href', '/flow/p/8/list?openTask=51');
    expect(link).toHaveTextContent(`Due ${formatDueDate('2026-10-01', 'en')}`);
    expect(within(list).getByText('and 5 more overdue tasks')).toBeInTheDocument();
  });

  it('says so when nothing is overdue', () => {
    loaded({ ...dashboard, overdueTaskCount: 0, overdueTasks: [] });
    render(<ProjectDashboardPanel projectId="8" />);
    expect(screen.getByText('Nothing is overdue.')).toBeInTheDocument();
  });

  it('hides the invoice section when the viewer may not see invoices', () => {
    loaded(dashboard);
    render(<ProjectDashboardPanel projectId="8" />);
    expect(screen.queryByRole('heading', { name: 'Invoices' })).not.toBeInTheDocument();
    expect(screen.queryByText('Invoiced')).not.toBeInTheDocument();
  });

  it('shows the project invoices in the insights currency', () => {
    loaded({
      ...dashboard,
      invoices: {
        invoiceCount: 4,
        invoiced: 200000,
        paid: 120000,
        outstanding: 80000,
        overdue: 30000,
        overdueCount: 1,
      },
    });
    render(<ProjectDashboardPanel projectId="8" />);

    expect(screen.getByRole('heading', { name: 'Invoices' })).toBeInTheDocument();
    expect(card('Invoiced')).toHaveTextContent('2,000');
    expect(card('Invoiced')).toHaveTextContent('4 invoices');
    expect(card('Paid')).toHaveTextContent('1,200');
    expect(card('Outstanding')).toHaveTextContent('800');
    expect(card('Overdue')).toHaveTextContent('300');
    expect(card('Overdue')).toHaveTextContent('1 overdue invoice');
  });

  it('says when no invoice is tagged with the project yet', () => {
    loaded({
      ...dashboard,
      invoices: {
        invoiceCount: 0,
        invoiced: 0,
        paid: 0,
        outstanding: 0,
        overdue: 0,
        overdueCount: 0,
      },
    });
    render(<ProjectDashboardPanel projectId="8" />);
    expect(screen.getByText('No invoices are tagged with this project yet.')).toBeInTheDocument();
    expect(screen.queryByText('Invoiced')).not.toBeInTheDocument();
  });

  it('treats a private project like a missing one', () => {
    failed(404);
    render(<ProjectDashboardPanel projectId="8" />);
    expect(screen.getByText('Project not available')).toBeInTheDocument();
    expect(
      screen.getByText('This project does not exist, or it is private and you are not a member.')
    ).toBeInTheDocument();
  });

  it('explains a missing Business Overview permission', () => {
    failed(403);
    render(<ProjectDashboardPanel projectId="8" />);
    expect(
      screen.getByText('You need access to the Business Overview to see project dashboards.')
    ).toBeInTheDocument();
  });

  it.each([403, 404])('renders the fallback instead on a %s when one is given', (status) => {
    failed(status);
    render(<ProjectDashboardPanel projectId="8" unavailableFallback={<p>Task tiles</p>} />);
    expect(screen.getByText('Task tiles')).toBeInTheDocument();
    expect(screen.queryByText('Project not available')).not.toBeInTheDocument();
  });

  it('offers a retry on any other failure', () => {
    failed(500);
    render(<ProjectDashboardPanel projectId="8" unavailableFallback={<p>Task tiles</p>} />);
    expect(screen.queryByText('Task tiles')).not.toBeInTheDocument();
    expect(screen.getByText('Something went wrong. Please try again.')).toBeInTheDocument();
  });

  it('shows placeholders while loading', () => {
    mocks.useProjectDashboard.mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
      error: null,
      refetch: vi.fn(),
    });
    const { container } = render(<ProjectDashboardPanel projectId="8" />);
    expect(container.querySelector('[aria-busy="true"]')).toBeInTheDocument();
  });
});

describe('formatDueDate', () => {
  it('keeps the calendar day and Latin digits in every language', () => {
    expect(formatDueDate('2026-10-01', 'en')).toBe('Oct 1, 2026');
    expect(formatDueDate('2026-10-01', 'ar')).toMatch(/1/);
    expect(formatDueDate('2026-10-01', 'ar')).not.toMatch(/[٠-٩]/);
  });

  it('returns an unparseable date unchanged', () => {
    expect(formatDueDate('soon', 'en')).toBe('soon');
  });
});
