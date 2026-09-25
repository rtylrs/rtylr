import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  createCampaign: vi.fn(),
  updateCampaign: vi.fn(),
  updateCampaignStatus: vi.fn(),
}));

const segments = vi.hoisted(() => [
  { id: 'segment-1', name: 'All customers', customerCount: 12, description: 'Everyone' },
]);

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mocks.navigate,
  useSearch: () => ({ id: undefined }),
  Link: ({ children }: { children?: React.ReactNode }) => <span>{children}</span>,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('@voxire/auth', () => ({
  useUserStore: (selector: (state: { user: { role: string } }) => unknown) =>
    selector({ user: { role: 'admin' } }),
}));

vi.mock('../../shared/hooks/queries/useCampaigns', () => ({
  useCreateCampaign: () => ({ mutate: mocks.createCampaign, isPending: false }),
  useUpdateCampaign: () => ({ mutate: mocks.updateCampaign, isPending: false }),
  useUpdateCampaignStatus: () => ({ mutate: mocks.updateCampaignStatus, isPending: false }),
  useGetCampaign: () => ({ data: undefined }),
}));

vi.mock('../../shared/hooks/queries/useSegments', () => ({
  useListSegments: () => ({ data: segments }),
}));

vi.mock('../../shared/hooks/queries/useCustomers', () => ({
  useListCustomers: () => ({ data: [] }),
}));

vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

import { toast } from 'sonner';
import SMSCampaign from './SMSCampaign';

interface SaveCallbacks {
  onSuccess: (result: { campaignId?: number; campaign_id?: number }) => void;
  onError: () => void;
}

interface StatusCallbacks {
  onSuccess: () => void;
  onError: () => void;
}

function renderEditor(): void {
  render(<SMSCampaign />);
}

function clickSaveAndActivate(): void {
  fireEvent.click(screen.getByRole('button', { name: /Save & activate/ }));
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('SMSCampaign activation retry (B56-R06)', () => {
  it('activates the campaign created by a failed attempt instead of creating a duplicate', () => {
    mocks.createCampaign.mockImplementation((_payload: unknown, callbacks: SaveCallbacks) => {
      callbacks.onSuccess({ campaignId: 7 });
    });
    mocks.updateCampaign.mockImplementation((_payload: unknown, callbacks: SaveCallbacks) => {
      callbacks.onSuccess({});
    });
    mocks.updateCampaignStatus
      .mockImplementationOnce((_payload: unknown, callbacks: StatusCallbacks) => {
        callbacks.onError();
      })
      .mockImplementationOnce((_payload: unknown, callbacks: StatusCallbacks) => {
        callbacks.onSuccess();
      });

    renderEditor();
    fireEvent.change(screen.getByLabelText('crm.campaignEditor.nameLabel'), {
      target: { value: 'Flash sale' },
    });

    clickSaveAndActivate();
    expect(toast.error).toHaveBeenCalledWith('crm.campaigns.savedNotActivated');

    // The campaign exists now; a second attempt must not create another one.
    clickSaveAndActivate();

    expect(mocks.createCampaign).toHaveBeenCalledTimes(1);
    expect(mocks.updateCampaign).toHaveBeenCalledTimes(1);
    expect(mocks.updateCampaign.mock.calls[0][0]).toMatchObject({ id: 7 });
    expect(mocks.updateCampaignStatus.mock.calls.map((call) => call[0])).toEqual([
      { id: 7, status: 'active' },
      { id: 7, status: 'active' },
    ]);
    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/campaigns' });
  });
});
