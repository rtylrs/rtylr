import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render as renderPlain, screen, within } from '@testing-library/react';
import path from 'node:path';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18nFrom } from '../../../../../test/i18nTestInstance';
import { useLeadFilters } from '../sales/data/leadFilters';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  activity: vi.fn(),
  leadStats: vi.fn(),
  dealStats: vi.fn(),
  board: vi.fn(),
  taskCounts: vi.fn(),
  taskList: vi.fn(),
}));

vi.mock('@tanstack/react-router', () => ({ useNavigate: () => mocks.navigate }));
vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

const day = (date: string, calls: number, reached: number) => ({
  date,
  calls,
  reached,
  leadsAdded: 0,
  dealsWon: 0,
});

let activity: Record<string, unknown> = {};
let byStatus: Array<{ status: string; count: number }> = [];

vi.mock('../sales/data/queries', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../sales/data/queries')>()),
  useReps: () => ({ data: [{ id: '5', name: 'Sam Test' }] }),
  useSalesActivity: (days: number, rep: string) => {
    mocks.activity(days, rep);
    return { data: activity, isPlaceholderData: false, isError: false, refetch: vi.fn() };
  },
  useLeadStats: (filters: { reps: string[] }) => {
    mocks.leadStats(filters.reps);
    return {
      data: { byStatus, followUpsOverdue: 2, followUpsToday: 1 },
      isPlaceholderData: false,
      isError: false,
    };
  },
  useDealStats: (pipelineId: string | undefined, rep: string) => {
    mocks.dealStats(rep);
    return {
      data: {
        openDeals: 3,
        openAmount: '12000',
        weightedAmount: '4800',
        averageDaysToWin: 0,
      },
    };
  },
  useDealBoard: (filters: { rep: string }) => {
    mocks.board(filters.rep);
    return {
      data: {
        pipeline: { name: 'Sales' },
        columns: [
          { stage: { id: '1', name: 'Discovery', kind: 'open' }, deals: 2, amount: '10000' },
          { stage: { id: '2', name: 'Proposal', kind: 'open' }, deals: 1, amount: '2000' },
          { stage: { id: '3', name: 'Won', kind: 'won' }, deals: 9, amount: '90000' },
        ],
      },
      isPlaceholderData: false,
      isError: false,
    };
  },
  useTaskCounts: (rep: string) => {
    mocks.taskCounts(rep);
    return { data: { overdue: 1, dueToday: 0, upcoming: 4 } };
  },
  useTaskList: (filters: { rep: string }) => {
    mocks.taskList(filters.rep);
    return { data: { pages: [{ tasks: [] }] }, isPending: false, isError: false };
  },
}));

vi.mock('../../shared/hooks/queries/useCrmDashboard', () => ({
  useCrmDashboard: () => ({
    data: {
      totalCustomers: 120,
      newThisMonth: 4,
      activeLoyaltyMembers: 30,
      topCustomers: [{ name: 'Rana Haddad', spend: '900', tier: 'Gold' }],
    },
    isError: false,
  }),
}));

import Dashboard from './Dashboard';

const render = (ui: React.ReactElement) =>
  renderPlain(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      {ui}
    </QueryClientProvider>
  );

beforeAll(async () => {
  // Real translations: an English assertion also proves the key resolves.
  await initTestI18nFrom(path.resolve(__dirname, '../../locales'));
});

beforeEach(() => {
  vi.clearAllMocks();
  activity = {
    days: [day('2026-10-19', 4, 1), day('2026-10-20', 0, 0), day('2026-10-21', 6, 3)],
    calls: 10,
    reached: 4,
    leadsAdded: 7,
    leadsQualified: 2,
    dealsWon: 2,
    dealsLost: 1,
    wonAmount: '5000',
    reps: [
      { repId: '5', calls: 10, reached: 4, leadsQualified: 1, dealsWon: 2, wonAmount: '5000' },
      { repId: '0', calls: 0, reached: 0, leadsQualified: 1, dealsWon: 0, wonAmount: '0' },
    ],
  };
  byStatus = [
    { status: 'new', count: 8 },
    { status: 'attempting', count: 3 },
    { status: 'qualified', count: 2 },
  ];
  useLeadFilters.getState().reset();
});

describe('CRM dashboard', () => {
  it('summarises the window: win rate, reach and what is due', () => {
    render(<Dashboard />);

    const strip = within(document.querySelector<HTMLElement>('.sales-stats')!);
    const stat = (label: string) => strip.getByText(label).closest('.sales-stat')!;
    // 2 won and 1 lost is a 67% win rate.
    expect(stat('Win rate')).toHaveTextContent('67%');
    expect(stat('Win rate')).toHaveTextContent('2 won · 1 lost');
    expect(stat('Calls')).toHaveTextContent('10');
    expect(stat('Calls')).toHaveTextContent('40% reached');
    expect(stat('Follow-ups due')).toHaveTextContent('3');
    expect(stat('Follow-ups due')).toHaveTextContent('2 overdue');
    expect(stat('Follow-ups due')).toHaveClass('is-warning');
    expect(stat('Customers')).toHaveTextContent('120');
    expect(stat('Customers')).toHaveTextContent('+4 this month');
    // Only open stages make up the pipeline.
    const pipeline = screen.getByRole('list', { name: 'Pipeline' });
    expect(within(pipeline).getAllByRole('listitem')).toHaveLength(2);
  });

  it('scopes every query to the chosen period and rep', () => {
    render(<Dashboard />);
    expect(mocks.activity).toHaveBeenLastCalledWith(30, 'all');

    fireEvent.click(screen.getByRole('radio', { name: '7 days' }));
    expect(mocks.activity).toHaveBeenLastCalledWith(7, 'all');

    fireEvent.change(screen.getByRole('combobox', { name: 'Rep' }), {
      target: { value: '5' },
    });
    expect(mocks.activity).toHaveBeenLastCalledWith(7, '5');
    expect(mocks.leadStats).toHaveBeenLastCalledWith(['5']);
    expect(mocks.dealStats).toHaveBeenLastCalledWith('5');
    expect(mocks.board).toHaveBeenLastCalledWith('5');
    expect(mocks.taskCounts).toHaveBeenLastCalledWith('5');
    expect(mocks.taskList).toHaveBeenLastCalledWith('5');

    fireEvent.change(screen.getByRole('combobox', { name: 'Rep' }), {
      target: { value: 'unassigned' },
    });
    // The leads filter spells unassigned "0".
    expect(mocks.leadStats).toHaveBeenLastCalledWith(['0']);
  });

  it('opens the leads list on the status or follow-ups clicked', () => {
    render(<Dashboard />);
    const funnel = screen.getByRole('list', { name: 'Leads by status' });
    fireEvent.click(within(funnel).getByRole('button', { name: /Attempting/ }));
    expect(useLeadFilters.getState().filters.statuses).toEqual(['attempting']);
    expect(mocks.navigate).toHaveBeenLastCalledWith({ to: '/crm/sales/leads' });

    fireEvent.click(screen.getByRole('button', { name: /Follow-ups due/ }));
    expect(useLeadFilters.getState().filters.followUp).toBe('due');
    expect(useLeadFilters.getState().filters.statuses).toEqual([
      'new',
      'attempting',
      'connected',
      'nurture',
    ]);
  });

  it('lists each rep, with unassigned work under its own row', () => {
    render(<Dashboard />);
    const team = screen.getByRole('region', { name: 'Team' });
    const rows = within(team).getAllByRole('row');
    expect(rows[1]).toHaveTextContent('Sam Test');
    expect(rows[1]).toHaveTextContent('40%');
    expect(rows[2]).toHaveTextContent('Unassigned');
    expect(rows[2]).toHaveTextContent('—');
  });

  it('reads the calls chart day by day from the keyboard, with a table behind it', () => {
    render(<Dashboard />);
    const chart = screen.getByRole('group', { name: /Calls per day/ });
    fireEvent.focus(chart);
    expect(screen.getByRole('status')).toHaveTextContent('6');
    fireEvent.keyDown(chart, { key: 'ArrowLeft' });
    fireEvent.keyDown(chart, { key: 'ArrowLeft' });
    expect(screen.getByRole('status')).toHaveTextContent('4');
    // A row per day after the header.
    const table = screen.getByRole('table', { name: 'Calls per day' });
    expect(within(table).getAllByRole('row')).toHaveLength(4);
  });

  it('points a quiet business at its first calls and leads', () => {
    activity = { ...activity, days: [day('2026-10-21', 0, 0)], calls: 0, reached: 0, reps: [] };
    byStatus = [];
    render(<Dashboard />);
    expect(screen.getByText('No calls logged in the last 30 days.')).toBeInTheDocument();
    expect(
      screen.getByText('No leads yet. Import a list or add your first one.')
    ).toBeInTheDocument();
    expect(
      screen.getByText('No calls, qualified leads or closed deals in the last 30 days.')
    ).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: /Import leads/ }));
    expect(mocks.navigate).toHaveBeenLastCalledWith({ to: '/crm/sales/leads/import' });
  });
});
