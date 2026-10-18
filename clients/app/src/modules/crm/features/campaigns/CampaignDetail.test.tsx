import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  refetchCampaign: vi.fn(),
  refetchStats: vi.fn(),
  campaign: {
    data: undefined as { id: number; name: string; subject?: string; status: string } | undefined,
    error: new Error('not found') as unknown,
    isError: false,
    isPending: false,
    isSuccess: false,
    isPaused: false,
    failureCount: 0,
    refetch: vi.fn(),
  },
  stats: {
    data: undefined as unknown,
    isError: false,
    isPending: false,
    isPaused: false,
    failureCount: 0,
    refetch: vi.fn(),
  },
}));

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, to }: { children?: React.ReactNode; to: string }) => (
    <a href={to}>{children}</a>
  ),
  useParams: () => ({ id: '999' }),
  useNavigate: () => mocks.navigate,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('../../shared/hooks/queries/useCampaigns', () => ({
  useCampaignAudience: () => ({ data: undefined, isPending: true, isError: false }),
  useSendCampaign: () => ({ mutate: vi.fn(), isPending: false }),
  useGetCampaign: () => ({ ...mocks.campaign, refetch: mocks.refetchCampaign }),
  useCampaignStats: () => ({ ...mocks.stats, refetch: mocks.refetchStats }),
}));

import CampaignDetail from './CampaignDetail';

function renderDetail(): void {
  render(<CampaignDetail />);
}

beforeEach(() => {
  mocks.navigate.mockClear();
  mocks.refetchCampaign.mockClear();
  mocks.refetchStats.mockClear();
  Object.assign(mocks.campaign, {
    data: undefined,
    error: new Error('not found'),
    isError: false,
    isPending: false,
    isSuccess: false,
    isPaused: false,
    failureCount: 0,
  });
  Object.assign(mocks.stats, {
    data: undefined,
    isError: false,
    isPending: false,
    isPaused: false,
    failureCount: 0,
  });
});

describe('CampaignDetail not-found state', () => {
  it('shows a "back to campaigns" call-to-action when the campaign failed to load', () => {
    renderDetail();

    expect(screen.getByText('crm.campaignDetail.notFoundTitle')).toBeInTheDocument();
    expect(screen.getByText('crm.campaignDetail.notFoundDescription')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'crm.campaignDetail.backToCampaigns' }));

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/crm/campaigns' });
  });

  it('shows retry when a campaign request fails while paused offline', () => {
    Object.assign(mocks.campaign, {
      error: new Error('network unavailable'),
      isPending: true,
      isPaused: true,
      failureCount: 1,
    });

    renderDetail();

    expect(screen.getByText('crmErrors.generic')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'crm.actions.retry' }));
    expect(mocks.refetchCampaign).toHaveBeenCalledTimes(1);
  });

  it('shows retry instead of an empty-stats message when stats fail while paused offline', () => {
    Object.assign(mocks.campaign, {
      data: { id: 999, name: 'Launch', subject: 'Hello', status: 'active' },
      error: undefined,
      isSuccess: true,
    });
    Object.assign(mocks.stats, {
      error: new Error('network unavailable'),
      isPending: true,
      isPaused: true,
      failureCount: 1,
    });

    renderDetail();

    expect(screen.getByText('crmErrors.refreshAndTryAgain')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'crm.actions.retry' }));
    expect(mocks.refetchStats).toHaveBeenCalledTimes(1);
    expect(screen.queryByText('No campaign statistics are available yet.')).not.toBeInTheDocument();
  });
});
