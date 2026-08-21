import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
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
    data: dashboardData,
    isLoading: false,
    refetch: refetchDashboard,
  }),
}));

vi.mock('../../shared/hooks/queries/useCustomers', () => ({
  useListCustomers: () => ({ data: customersList }),
}));

vi.mock('../../shared/hooks/queries/useSegments', () => ({
  useListSegments: () => ({ data: segmentsList }),
}));

import Dashboard from './Dashboard';

function renderDashboard(): void {
  render(<Dashboard />);
}

beforeEach(() => {
  mocks.navigate.mockClear();
  refetchDashboard.mockClear();
  customersList.length = 0;
  dashboardData.recentCampaigns = [];
});

describe('Dashboard read-only empty states', () => {
  it('shows a "view customers" call-to-action when there are no at-risk customers', () => {
    renderDashboard();

    expect(screen.getByText('crm.dashboard.atRiskEmptyTitle')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'crm.actions.viewCustomers' }));

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/customers' });
  });

  it('shows a "refresh" call-to-action when campaign performance is unavailable', () => {
    renderDashboard();

    expect(
      screen.getByText('crm.dashboard.campaignPerformanceUnavailableTitle')
    ).toBeInTheDocument();
    expect(
      screen.getByText('crm.dashboard.campaignPerformanceUnavailableDescription')
    ).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'crm.actions.refresh' }));

    expect(refetchDashboard).toHaveBeenCalledTimes(1);
  });
});
