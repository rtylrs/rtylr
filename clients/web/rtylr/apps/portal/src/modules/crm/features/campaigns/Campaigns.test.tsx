import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createEffectiveAccess } from '../../../../../../../test/effectiveAccess';

const mocks = vi.hoisted(() => ({
  effectiveAccess: null as unknown,
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('../../../../shared/stores/userStore', () => {
  const state = () => ({
    user: { effectiveAccess: mocks.effectiveAccess },
    effectiveAccess: mocks.effectiveAccess,
  });
  return {
    useUserStore: Object.assign(
      (selector: (value: ReturnType<typeof state>) => unknown) => selector(state()),
      { getState: state }
    ),
  };
});

const draftCampaign = {
  id: 1,
  ownerId: 1,
  name: 'Spring promo',
  type: 'email',
  status: 'draft',
  audienceSegmentId: null,
  subject: 'Spring is here',
  body: 'Hello',
  audienceCount: 100,
  sentCount: 0,
  openCount: 0,
  scheduledAt: null,
  sentAt: null,
  createdAt: '2026-01-01',
  updatedAt: '2026-01-01',
};

// A stable array reference is required: the component's effect depends on this
// array's identity, and a mock that returns a new array on every render (as
// react-query would not, once settled) would loop forever.
const campaignsList = [draftCampaign];

const refetchCampaigns = vi.fn();

vi.mock('../../shared/hooks/queries/useCampaigns', () => ({
  useListCampaigns: () => ({
    data: campaignsList,
    isLoading: false,
    isFetching: false,
    isError: false,
    refetch: refetchCampaigns,
  }),
  useUpdateCampaignStatus: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteCampaign: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/hooks/queries/useSegments', () => ({
  useListSegments: () => ({ data: [] }),
}));

import Campaigns from './Campaigns';

function renderCampaigns(): void {
  render(<Campaigns />);
}

/**
 * File-level, deliberately, not inside a describe.
 *
 * campaignsList is module state and two tests in "Campaigns empty state" empty
 * it in place (`campaignsList.length = 0`). Only that describe had a beforeEach
 * to put it back, so the gating describe below depended on running first. With
 * --sequence.shuffle this file failed roughly 1 run in 5: an emptied list means
 * no campaign row, so the edit and activate controls the gating tests look for
 * are absent and the failure reads as a permissions bug rather than a fixture
 * leak.
 *
 * effective access is reset too so permission changes do not leak between tests.
 */
beforeEach(() => {
  refetchCampaigns.mockClear();
  mocks.effectiveAccess = createEffectiveAccess('crm', {
    permissions: ['crm:campaigns:create', 'crm:campaigns:manage'],
  });
  campaignsList.length = 0;
  campaignsList.push(draftCampaign);
});

describe('Campaigns send gating', () => {
  it('hides the Activate (send) control for an operator, who lacks crm:campaigns:send', () => {
    mocks.effectiveAccess = createEffectiveAccess('crm', {
      permissions: ['crm:campaigns:create', 'crm:campaigns:manage'],
    });
    renderCampaigns();
    expect(screen.queryByRole('button', { name: /activate/i })).not.toBeInTheDocument();
    // Operator still has crm:campaigns, so edit/delete controls remain visible.
    expect(screen.getByRole('button', { name: /edit/i })).toBeInTheDocument();
  });

  it('shows the Activate (send) control for admin, who has crm:campaigns:send directly', () => {
    mocks.effectiveAccess = createEffectiveAccess('crm', { isAdmin: true });
    renderCampaigns();
    expect(screen.getByRole('button', { name: /activate/i })).toBeInTheDocument();
  });

  it('hides all management controls for staff, who has no crm:campaigns access', () => {
    mocks.effectiveAccess = createEffectiveAccess('crm', { permissions: [] });
    renderCampaigns();
    expect(screen.queryByRole('button', { name: /activate/i })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /edit/i })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /delete campaign/i })).not.toBeInTheDocument();
  });
});

describe('Campaigns empty state', () => {
  it('shows a "view all campaigns" call-to-action when the selected status filter has no matches', () => {
    mocks.effectiveAccess = createEffectiveAccess('crm', { isAdmin: true });
    renderCampaigns();

    fireEvent.click(screen.getByRole('button', { name: /^Paused/ }));

    expect(screen.getByText('crm.campaigns.emptyStatus')).toBeInTheDocument();
    const viewAll = screen.getByRole('button', { name: 'crm.campaigns.viewAllCampaigns' });
    expect(viewAll).toBeInTheDocument();

    fireEvent.click(viewAll);
    expect(screen.getByText(draftCampaign.name)).toBeInTheDocument();
  });

  it('shows a "new email campaign" call-to-action when there are no campaigns at all and the user can manage campaigns', () => {
    mocks.effectiveAccess = createEffectiveAccess('crm', { isAdmin: true });
    campaignsList.length = 0;
    renderCampaigns();

    expect(screen.getByText('crm.campaigns.emptyAllTitle')).toBeInTheDocument();
    expect(
      screen.getByRole('button', { name: 'crm.campaigns.newEmailCampaign' })
    ).toBeInTheDocument();
  });

  it('falls back to a refresh call-to-action when there are no campaigns and the user cannot manage them', () => {
    mocks.effectiveAccess = createEffectiveAccess('crm', { permissions: [] });
    campaignsList.length = 0;
    renderCampaigns();

    expect(screen.getByText('crm.campaigns.emptyAllTitle')).toBeInTheDocument();
    const refresh = screen.getByRole('button', { name: 'crm.actions.refresh' });
    fireEvent.click(refresh);
    expect(refetchCampaigns).toHaveBeenCalledTimes(1);
  });
});
