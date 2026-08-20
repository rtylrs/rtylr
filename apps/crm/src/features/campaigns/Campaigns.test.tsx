import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  role: 'operator' as string,
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('@voxire/auth', () => ({
  useUserStore: (selector: (state: { user: { role: string } }) => unknown) =>
    selector({ user: { role: mocks.role } }),
}));

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

vi.mock('../../shared/hooks/queries/useCampaigns', () => ({
  useListCampaigns: () => ({
    data: campaignsList,
    isLoading: false,
    isFetching: false,
    isError: false,
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

describe('Campaigns send gating', () => {
  it('hides the Activate (send) control for an operator, who lacks crm:campaigns:send', () => {
    mocks.role = 'operator';
    renderCampaigns();
    expect(screen.queryByRole('button', { name: /activate/i })).not.toBeInTheDocument();
    // Operator still has crm:campaigns, so edit/delete controls remain visible.
    expect(screen.getByRole('button', { name: /edit/i })).toBeInTheDocument();
  });

  it('shows the Activate (send) control for admin, who has crm:campaigns:send directly', () => {
    mocks.role = 'admin';
    renderCampaigns();
    expect(screen.getByRole('button', { name: /activate/i })).toBeInTheDocument();
  });

  it('hides all management controls for staff, who has no crm:campaigns access', () => {
    mocks.role = 'staff';
    renderCampaigns();
    expect(screen.queryByRole('button', { name: /activate/i })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /edit/i })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /delete campaign/i })).not.toBeInTheDocument();
  });
});
