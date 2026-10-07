import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  dashboardError: false,
  customersError: false,
  segmentsError: false,
  refetchCustomers: vi.fn(),
  refetchSegments: vi.fn(),
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mocks.navigate,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

// Stable references: Dashboard reads these straight from the query hooks with
// no local state, but keeping literal arrays declared once (rather than
// inline in the mock factory) keeps the "no at-risk customers" and "no
// campaigns" branches easy to toggle per test without new-array churn.
const refetchDashboard = vi.fn();
const dashboardData: { recentCampaigns: unknown[] } & Record<string, unknown> = {
  totalCustomers: 100,
  totalCustomersChange: '+5%',
  newThisMonth: 10,
  newThisMonthChange: '+2%',
  activeLoyaltyMembers: 40,
  activeLoyaltyMembersChange: '+1%',
  campaignReach: 500,
  campaignReachChange: '+3%',
  topCustomers: [],
  recentCampaigns: [],
};
const customersList: unknown[] = [];
const segmentsList: unknown[] = [];

vi.mock('../../shared/hooks/queries/useCrmDashboard', () => ({
  useCrmDashboard: () => ({
    data: mocks.dashboardError ? undefined : dashboardData,
    isLoading: false,
    isError: mocks.dashboardError,
    refetch: refetchDashboard,
  }),
}));

vi.mock('../../shared/hooks/queries/useCustomers', () => ({
  useListCustomers: () => ({
    data: mocks.customersError ? undefined : customersList,
    isError: mocks.customersError,
    refetch: mocks.refetchCustomers,
  }),
}));

vi.mock('../../shared/hooks/queries/useSegments', () => ({
  useListSegments: () => ({
    data: mocks.segmentsError ? undefined : segmentsList,
    isError: mocks.segmentsError,
    refetch: mocks.refetchSegments,
  }),
}));

import Dashboard from './Dashboard';

function renderDashboard(): void {
  render(<Dashboard />);
}

beforeEach(() => {
  mocks.navigate.mockClear();
  mocks.dashboardError = false;
  mocks.customersError = false;
  mocks.segmentsError = false;
  refetchDashboard.mockClear();
  mocks.refetchCustomers.mockClear();
  mocks.refetchSegments.mockClear();
  customersList.length = 0;
  segmentsList.length = 0;
  dashboardData.recentCampaigns = [];
});

describe('Dashboard read-only empty states', () => {
  it('uses the real segment id for the at-risk win-back campaign', () => {
    segmentsList.push({ id: '42', name: 'At Risk', customerCount: 0 });
    renderDashboard();

    fireEvent.click(screen.getByRole('button', { name: /Send win-back/i }));

    expect(mocks.navigate).toHaveBeenCalledWith({
      to: '/crm/campaigns/email',
      search: { segment: '42' },
    });
  });

  it('does not pass the symbolic at-risk key when no matching segment exists', () => {
    renderDashboard();

    fireEvent.click(screen.getByRole('button', { name: /Send win-back/i }));

    expect(mocks.navigate).toHaveBeenCalledWith({
      to: '/crm/campaigns/email',
      search: {},
    });
  });

  it('shows a "view customers" call-to-action when there are no at-risk customers', () => {
    renderDashboard();

    expect(screen.getByText('crm.dashboard.atRiskEmptyTitle')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'crm.actions.viewCustomers' }));

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/crm/customers' });
  });

  it('explains campaign performance is unavailable without offering a false retry', () => {
    renderDashboard();

    expect(
      screen.getByText('crm.dashboard.campaignPerformanceUnavailableTitle')
    ).toBeInTheDocument();
    expect(
      screen.getByText('crm.dashboard.campaignPerformanceUnavailableDescription')
    ).toBeInTheDocument();

    expect(screen.queryByRole('button', { name: 'crm.actions.refresh' })).not.toBeInTheDocument();
    expect(refetchDashboard).not.toHaveBeenCalled();
  });

  it('offers a retry when loading the dashboard actually fails', () => {
    mocks.dashboardError = true;

    renderDashboard();

    fireEvent.click(screen.getByRole('button', { name: 'crm.actions.refresh' }));

    expect(refetchDashboard).toHaveBeenCalledTimes(1);
  });

  it('does not show zero-valued customer KPIs when the customer list fails', () => {
    mocks.customersError = true;

    renderDashboard();

    expect(screen.getAllByText('—')).toHaveLength(2);
    expect(screen.getByRole('alert')).toHaveTextContent('crmErrors.generic');
    fireEvent.click(screen.getByRole('button', { name: 'crm.actions.refresh' }));
    expect(mocks.refetchCustomers).toHaveBeenCalledOnce();
  });
});
