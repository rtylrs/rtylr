import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

const navigateSpy = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => navigateSpy,
  Navigate: ({ to }: { to: string }) => <p>Redirected to {to}</p>,
}));

const pager = vi.hoisted(() => ({ getSettings: vi.fn(), updateSettings: vi.fn() }));
vi.mock('../../shared/lib/clients', () => ({ staffPagerClient: pager }));

const hub = vi.hoisted(() => ({
  getConfig: vi.fn(),
  updateConfig: vi.fn(),
  setSlug: vi.fn(),
}));
vi.mock('../../../../shared/lib/clients', () => ({
  configClient: { getConfig: hub.getConfig, updateConfig: hub.updateConfig },
  businessClient: { setSlug: hub.setSlug },
}));

import Setup from './Setup';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';
import { useUserStore } from '../../../../shared/stores/userStore';
import { initTestI18n } from '../../shared/testing/i18n';

function renderSetup() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <Setup />
    </QueryClientProvider>
  );
}

function signIn(access: { isOwner?: boolean; isAdmin?: boolean }) {
  useUserStore.setState({
    user: { businessName: 'Bean Bar', effectiveAccess: createEffectiveAccess(undefined, access) },
  });
}

beforeAll(async () => {
  await initTestI18n('en');
});

beforeEach(() => {
  navigateSpy.mockReset();
  Object.values(pager).forEach((mock) => mock.mockReset());
  Object.values(hub).forEach((mock) => mock.mockReset().mockResolvedValue({}));
  pager.getSettings.mockResolvedValue({});
  pager.updateSettings.mockResolvedValue({ enabled: true });
  hub.getConfig.mockResolvedValue({ business: { name: 'Bean Bar' } });
  signIn({ isOwner: true });
});

describe('pager setup', () => {
  it('saves the phone on the business, gives it a slug, switches the pager on and opens the board', async () => {
    renderSetup();
    expect(
      await screen.findByRole('heading', { name: 'Set up the pager for Bean Bar' })
    ).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText('Venue phone number'), {
      target: { value: '+961 3 123 456' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Switch on my pager' }));

    await waitFor(() => expect(navigateSpy).toHaveBeenCalledWith({ to: '/pager' }));
    expect(hub.updateConfig).toHaveBeenCalledWith({ contactNumber: '+961 3 123 456' });
    expect(hub.setSlug).toHaveBeenCalledWith({ slug: 'bean-bar' });
    expect(pager.updateSettings).toHaveBeenCalledWith({ enabled: true });
  });

  it('asks again for a phone number that is too short, without saving anything', async () => {
    renderSetup();
    fireEvent.change(await screen.findByLabelText('Venue phone number'), {
      target: { value: '123' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Switch on my pager' }));

    expect(await screen.findByText('Enter a phone number with 6 to 15 digits.')).toBeVisible();
    expect(hub.updateConfig).not.toHaveBeenCalled();
    expect(pager.updateSettings).not.toHaveBeenCalled();
  });

  it('goes straight to the board when the pager is already on', async () => {
    pager.getSettings.mockResolvedValue({ enabled: true });
    hub.getConfig.mockResolvedValue({ business: { name: 'Bean Bar', slug: 'beans' } });
    renderSetup();
    expect(await screen.findByText('Redirected to /pager')).toBeInTheDocument();
  });

  it('tells a member without admin rights who can set it up', async () => {
    signIn({});
    renderSetup();
    expect(
      await screen.findByText('Only the owner or an admin of this business can set up the pager.')
    ).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Switch on my pager' })).not.toBeInTheDocument();
  });
});
