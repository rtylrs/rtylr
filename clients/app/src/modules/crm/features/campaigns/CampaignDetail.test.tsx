import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
}));

vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ id: '999' }),
  useNavigate: () => mocks.navigate,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('../../shared/hooks/queries/useCampaigns', () => ({
  useGetCampaign: () => ({ data: undefined, isLoading: false, error: new Error('not found') }),
  useCampaignStats: () => ({ data: undefined, isLoading: false, isError: false }),
}));

import CampaignDetail from './CampaignDetail';

function renderDetail(): void {
  render(<CampaignDetail />);
}

beforeEach(() => {
  mocks.navigate.mockClear();
});

describe('CampaignDetail not-found state', () => {
  it('shows a "back to campaigns" call-to-action when the campaign failed to load', () => {
    renderDetail();

    expect(screen.getByText('crm.campaignDetail.notFoundTitle')).toBeInTheDocument();
    expect(screen.getByText('crm.campaignDetail.notFoundDescription')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'crm.campaignDetail.backToCampaigns' }));

    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/crm/campaigns' });
  });
});
