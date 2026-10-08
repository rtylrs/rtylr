import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  createCampaign: vi.fn(),
  updateCampaign: vi.fn(),
  updateCampaignStatus: vi.fn(),
  effectiveAccess: null as unknown,
}));

const segments = vi.hoisted(() => [
  { id: 'segment-1', name: 'All customers', customerCount: 12, description: 'Everyone' },
]);
const segmentQuery = vi.hoisted(() => ({ isError: false, isLoading: false, refetch: vi.fn() }));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mocks.navigate,
  useSearch: () => ({ id: undefined }),
  Link: ({ children }: { children?: React.ReactNode }) => <span>{children}</span>,
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

vi.mock('../../shared/hooks/queries/useCampaigns', () => ({
  useCreateCampaign: () => ({ mutate: mocks.createCampaign, isPending: false }),
  useUpdateCampaign: () => ({ mutate: mocks.updateCampaign, isPending: false }),
  useUpdateCampaignStatus: () => ({ mutate: mocks.updateCampaignStatus, isPending: false }),
  useGetCampaign: () => ({ data: undefined }),
}));

vi.mock('../../shared/hooks/queries/useSegments', () => ({
  useListSegments: () => ({ data: segmentQuery.isError ? undefined : segments, ...segmentQuery }),
}));

vi.mock('../../shared/hooks/queries/useTemplates', () => ({
  useTemplates: () => ({ data: [], isError: false, refetch: vi.fn() }),
}));

vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));

import { toast } from 'sonner';
import EmailCampaign from './EmailCampaign';

interface SaveCallbacks {
  onSuccess: (result: { campaignId?: number; campaign_id?: number }) => void;
  onError: () => void;
}

interface StatusCallbacks {
  onSuccess: () => void;
  onError: () => void;
}

function renderEditor(): void {
  render(<EmailCampaign />);
}

function fillRequiredFields(): void {
  fireEvent.change(screen.getByLabelText('crm.campaignEditor.nameLabel'), {
    target: { value: 'Spring promo' },
  });
  fireEvent.change(screen.getByLabelText('crm.composer.subject'), {
    target: { value: 'Spring is here' },
  });
}

function clickSaveAndActivate(): void {
  fireEvent.click(screen.getByRole('button', { name: /Save & activate/ }));
}

beforeEach(() => {
  vi.clearAllMocks();
  mocks.effectiveAccess = createEffectiveAccess('crm', { isAdmin: true });
  segmentQuery.isError = false;
  segmentQuery.isLoading = false;
});

describe('EmailCampaign activation retry (B56-R05)', () => {
  it('prevents campaign activation when the audience list failed to load', () => {
    segmentQuery.isError = true;
    renderEditor();
    fillRequiredFields();

    expect(screen.getByRole('alert')).toHaveTextContent('crmErrors.generic');
    expect(screen.getByRole('button', { name: /Save & activate/ })).toBeDisabled();
  });

  it('activates the campaign created by a failed attempt instead of creating a duplicate', () => {
    mocks.createCampaign.mockImplementation((_payload: unknown, callbacks: SaveCallbacks) => {
      callbacks.onSuccess({ campaignId: 42 });
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
    fillRequiredFields();

    clickSaveAndActivate();
    expect(toast.error).toHaveBeenCalledWith('crm.campaigns.savedNotActivated');

    // The campaign exists now; a second attempt must not create another one.
    clickSaveAndActivate();

    expect(mocks.createCampaign).toHaveBeenCalledTimes(1);
    expect(mocks.updateCampaign).toHaveBeenCalledTimes(1);
    expect(mocks.updateCampaign.mock.calls[0][0]).toMatchObject({ id: 42 });
    expect(mocks.updateCampaignStatus.mock.calls.map((call) => call[0])).toEqual([
      { id: 42, status: 'active' },
      { id: 42, status: 'active' },
    ]);
    expect(mocks.navigate).toHaveBeenCalledWith({ to: '/crm/campaigns' });
  });
});
