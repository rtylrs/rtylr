import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  createCampaign: vi.fn(),
  updateCampaign: vi.fn(),
  sendCampaign: vi.fn(),
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
  useGetCampaign: () => ({ data: undefined }),
  useCampaignAudience: () => ({
    data: { eligible: 3, total: 12 },
    isPending: false,
    isError: false,
  }),
  useSendCampaign: () => ({ mutate: mocks.sendCampaign, isPending: false }),
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
  fireEvent.change(screen.getByLabelText('crm.composer.body'), {
    target: { value: 'Hi {{first_name}}' },
  });
}

function clickSaveAndSend(): void {
  fireEvent.click(screen.getByRole('button', { name: /crm\.campaignEditor\.saveAndSend/ }));
}

beforeEach(() => {
  vi.clearAllMocks();
  mocks.effectiveAccess = createEffectiveAccess('crm', { isAdmin: true });
  segmentQuery.isError = false;
  segmentQuery.isLoading = false;
  mocks.createCampaign.mockImplementation((_payload: unknown, callbacks: SaveCallbacks) => {
    callbacks.onSuccess({ campaignId: 42 });
  });
  mocks.updateCampaign.mockImplementation((_payload: unknown, callbacks: SaveCallbacks) => {
    callbacks.onSuccess({});
  });
});

describe('EmailCampaign sending', () => {
  it('saves the campaign, then asks before sending it to the opted-in audience', () => {
    renderEditor();
    fillRequiredFields();

    clickSaveAndSend();

    // With no segment picked the campaign goes to every opted-in customer.
    expect(mocks.createCampaign.mock.calls[0][0]).toMatchObject({
      type: 'email',
      audienceSegmentId: null,
    });
    expect(mocks.sendCampaign).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole('button', { name: /crm\.campaignSend\.confirm/ }));
    expect(mocks.sendCampaign).toHaveBeenCalledWith(42, expect.anything());
  });

  it('sends the campaign it already created instead of creating a duplicate', () => {
    renderEditor();
    fillRequiredFields();

    clickSaveAndSend();
    fireEvent.click(screen.getByRole('button', { name: 'crmCommon.cancel' }));
    clickSaveAndSend();

    expect(mocks.createCampaign).toHaveBeenCalledTimes(1);
    expect(mocks.updateCampaign).toHaveBeenCalledTimes(1);
    expect(mocks.updateCampaign.mock.calls[0][0]).toMatchObject({ id: 42 });
  });

  it('will not send without a message', () => {
    renderEditor();
    fireEvent.change(screen.getByLabelText('crm.campaignEditor.nameLabel'), {
      target: { value: 'Spring promo' },
    });
    fireEvent.change(screen.getByLabelText('crm.composer.subject'), {
      target: { value: 'Spring is here' },
    });

    expect(screen.getByRole('button', { name: /crm\.campaignEditor\.saveAndSend/ })).toBeDisabled();
    expect(toast.error).not.toHaveBeenCalled();
  });
});
