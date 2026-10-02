import path from 'node:path';
import { render, screen, within, waitFor, cleanup } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider, useQuery } from '@tanstack/react-query';
import { beforeAll, beforeEach, afterEach, describe, expect, it, vi } from 'vitest';
import { initTestI18nFrom } from '../../../../../test/i18nTestInstance';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';
import { useUserStore } from '../../shared/stores/userStore';
import type { AddonData } from '../../shared/hooks/queries/useAddons';

const api = vi.hoisted(() => ({
  listAddons: vi.fn(),
  enable: vi.fn(),
  disable: vi.fn(),
  listIntegrations: vi.fn(),
  connect: vi.fn(),
  disconnect: vi.fn(),
  subscription: {
    data: { status: 'active' } as { status: string } | undefined,
    isError: false,
    refetch: vi.fn(),
  },
}));
const navigate = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({ useNavigate: () => navigate }));
vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock('../../shared/hooks/queries/useAddons', () => ({
  ADDONS_QUERY_KEY: ['addons'],
  useAddons: () => useQuery({ queryKey: ['addons'], queryFn: api.listAddons }),
  enableAddon: api.enable,
  disableAddon: api.disable,
}));
vi.mock('../../shared/lib/clients', () => ({
  integrationClient: {
    listIntegrations: api.listIntegrations,
    connectIntegration: api.connect,
    disconnectIntegration: api.disconnect,
  },
}));
vi.mock('../../shared/hooks/queries/useSubscription', () => ({
  useMySubscription: () => api.subscription,
  useBillingDetail: () => ({
    data: {
      status: 'active',
      lineItems: [{ type: 'addon', addonId: 'custom_domain', isGranted: true }],
    },
  }),
  centsToDisplay: (cents: number) => `$${(cents / 100).toFixed(2)}`,
}));
import Extensions from './Extensions';
let features: AddonData[];
let connections: { id: number; provider: string; status: string }[];
const locales = path.resolve(import.meta.dirname, '../../locales');
beforeAll(() => initTestI18nFrom(locales));
beforeEach(() => {
  vi.clearAllMocks();
  api.subscription.data = { status: 'active' };
  api.subscription.isError = false;
  useUserStore.setState({
    user: { effectiveAccess: createEffectiveAccess(undefined, { isOwner: true }) },
  });
  features = [
    {
      id: 'menu',
      name: 'Digital menu',
      description: 'Share your menu with customers.',
      icon: 'utensils-crossed',
      price: 0,
      isFree: true,
      isEnabled: false,
    },
    {
      id: 'custom_domain',
      name: 'Custom domain',
      description: 'Use your own domain.',
      icon: 'globe',
      price: 20,
      isFree: false,
      isEnabled: true,
    },
    {
      id: 'email_marketing',
      name: 'Email marketing',
      description: 'Reach your audience.',
      icon: 'mail',
      price: 10,
      isFree: false,
      isEnabled: false,
    },
  ];
  connections = [];
  api.listAddons.mockImplementation(async () => features.map((item) => ({ ...item })));
  api.listIntegrations.mockImplementation(async () => ({
    integrations: connections.map((item) => ({ ...item })),
  }));
  api.enable.mockImplementation(async (id: string) => {
    features = features.map((item) => (item.id === id ? { ...item, isEnabled: true } : item));
  });
  api.disable.mockImplementation(async (id: string) => {
    features = features.map((item) => (item.id === id ? { ...item, isEnabled: false } : item));
  });
  api.connect.mockImplementation(async ({ provider }: { provider: string }) => {
    connections.push({ id: 17, provider, status: 'connected' });
  });
  api.disconnect.mockImplementation(async ({ id }: { id: string }) => {
    connections = connections.filter((item) => String(item.id) !== id);
  });
});
afterEach(cleanup);
async function openExtensions() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  render(
    <QueryClientProvider client={client}>
      <Extensions />
    </QueryClientProvider>
  );
  await screen.findByRole('heading', { name: 'Digital menu' });
  return userEvent.setup();
}
function item(name: string) {
  const article = screen.getByRole('heading', { name, exact: true }).closest('article');
  if (!article) throw new Error(`Missing extension item ${name}`);
  return within(article);
}
describe('Unified Extensions', () => {
  it('excludes built-in apps from discovery and Added without disabling them', async () => {
    features.push(
      {
        id: 'insights',
        name: 'Insights',
        description: 'Business analytics',
        icon: 'bar-chart',
        price: 0,
        isFree: true,
        isEnabled: true,
      },
      {
        id: 'docs',
        name: 'Documentation',
        description: 'Public documentation',
        icon: 'file-text',
        price: 0,
        isFree: true,
        isEnabled: true,
      }
    );
    const ui = await openExtensions();
    expect(screen.queryByRole('heading', { name: 'Insights' })).toBeNull();
    expect(screen.queryByRole('heading', { name: 'Documentation' })).toBeNull();
    await ui.click(screen.getByRole('button', { name: 'Added', exact: true }));
    expect(screen.getByRole('status')).toHaveTextContent('1 item');
    expect(api.disable).not.toHaveBeenCalled();
  });
  it('searches features and connections together and resets empty results', async () => {
    const ui = await openExtensions();
    await ui.type(screen.getByRole('searchbox'), 'stripe');
    expect(screen.getByRole('heading', { name: 'Stripe' })).toBeVisible();
    expect(screen.queryByRole('heading', { name: 'Digital menu' })).toBeNull();
    await ui.clear(screen.getByRole('searchbox'));
    await ui.type(screen.getByRole('searchbox'), 'does-not-exist');
    expect(screen.getByText('No matching tools')).toBeVisible();
    await ui.click(screen.getByRole('button', { name: 'Show all tools' }));
    expect(screen.getByRole('heading', { name: 'Digital menu' })).toBeVisible();
  });
  it('filters by a shared category', async () => {
    const ui = await openExtensions();
    await ui.click(screen.getByRole('button', { name: 'Category' }));
    await ui.click(screen.getByRole('menuitem', { name: 'Payments' }));
    expect(screen.getByRole('heading', { name: 'Stripe' })).toBeVisible();
    expect(screen.queryByRole('heading', { name: 'Digital menu' })).toBeNull();
  });
  it('shows enabled features and connected providers together under Added', async () => {
    connections = [{ id: 17, provider: 'stripe', status: 'connected' }];
    const ui = await openExtensions();
    await ui.click(screen.getByRole('button', { name: 'Added', exact: true }));
    expect(screen.getByRole('heading', { name: 'Custom domain' })).toBeVisible();
    expect(item('Custom domain').getByText('Included')).toBeVisible();
    expect(screen.getAllByRole('heading', { name: 'Stripe' })).toHaveLength(1);
    expect(screen.queryByRole('heading', { name: 'Digital menu' })).toBeNull();
    await ui.click(item('Custom domain').getByRole('button', { name: 'Manage' }));
    expect(navigate).toHaveBeenCalledWith({ to: '/settings/domain' });
  });
  it('activates a feature after confirmation and refreshes Added', async () => {
    const ui = await openExtensions();
    await ui.click(item('Digital menu').getByRole('button', { name: 'Activate' }));
    expect(api.enable).not.toHaveBeenCalled();
    await ui.click(
      within(screen.getByRole('dialog')).getByRole('button', { name: 'Activate', exact: true })
    );
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
    expect(api.enable).toHaveBeenCalledWith('menu');
    expect(item('Digital menu').getByText('Added')).toBeVisible();
  });
  it('checks paid entitlements without claiming a purchase and keeps failures visible', async () => {
    api.enable.mockRejectedValue(new Error('Addon requires an active entitlement'));
    const ui = await openExtensions();
    await ui.click(item('Email marketing').getByRole('button', { name: 'Activate' }));
    const dialog = within(screen.getByRole('dialog'));
    expect(dialog.getByText(/Paid features require an active entitlement/)).toBeVisible();
    expect(dialog.getByText('Catalogue price: $10.00 per month.')).toBeVisible();
    await ui.click(dialog.getByRole('button', { name: 'Activate', exact: true }));
    expect(await dialog.findByRole('alert')).toHaveTextContent('Unable to activate');
    await ui.click(dialog.getByRole('button', { name: 'Cancel' }));
    expect(item('Email marketing').queryByText('Added')).toBeNull();
  });
  it('requires the item name before disabling and refreshes its status', async () => {
    const ui = await openExtensions();
    await ui.click(item('Custom domain').getByRole('button', { name: 'Remove' }));
    const dialog = within(screen.getByRole('dialog'));
    expect(dialog.getByRole('button', { name: 'Remove' })).toBeDisabled();
    await ui.type(dialog.getByRole('textbox'), 'Custom domain');
    await ui.click(dialog.getByRole('button', { name: 'Remove' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
    expect(api.disable).toHaveBeenCalledWith('custom_domain');
    expect(item('Custom domain').queryByText('Added')).toBeNull();
  });
  it('connects once through the existing API and reloads the connected state', async () => {
    const ui = await openExtensions();
    await ui.click(item('Stripe').getByRole('button', { name: 'Connect' }));
    await waitFor(() => expect(item('Stripe').getByText('Added')).toBeVisible());
    expect(api.connect).toHaveBeenCalledExactlyOnceWith({ provider: 'stripe' });
    expect(screen.getAllByRole('heading', { name: 'Stripe' })).toHaveLength(1);
  });
  it('removes a connection by server ID after confirmation', async () => {
    connections = [{ id: 17, provider: 'stripe', status: 'connected' }];
    const ui = await openExtensions();
    await ui.click(item('Stripe').getByRole('button', { name: 'Remove' }));
    const dialog = within(screen.getByRole('dialog'));
    await ui.type(dialog.getByRole('textbox'), 'Stripe');
    await ui.click(dialog.getByRole('button', { name: 'Remove' }));
    await waitFor(() =>
      expect(item('Stripe').getByRole('button', { name: 'Connect' })).toBeVisible()
    );
    expect(api.disconnect).toHaveBeenCalledExactlyOnceWith({ id: '17' });
  });
  it('keeps a connection when removal fails and allows retry', async () => {
    connections = [{ id: 17, provider: 'stripe', status: 'connected' }];
    api.disconnect.mockRejectedValue(new Error('offline'));
    const ui = await openExtensions();
    await ui.click(item('Stripe').getByRole('button', { name: 'Remove' }));
    const dialog = within(screen.getByRole('dialog'));
    await ui.type(dialog.getByRole('textbox'), 'Stripe');
    await ui.click(dialog.getByRole('button', { name: 'Remove' }));
    expect(await dialog.findByRole('alert')).toHaveTextContent('Unable to remove');
    await ui.click(dialog.getByRole('button', { name: 'Cancel' }));
    expect(item('Stripe').getByText('Added')).toBeVisible();
  });
  it('preserves features when connections fail and does not show a fake available catalogue', async () => {
    api.listIntegrations.mockRejectedValue(new Error('offline'));
    await openExtensions();
    expect(await screen.findByRole('alert')).toHaveTextContent('Connections could not be loaded');
    expect(screen.queryByRole('heading', { name: 'Stripe' })).toBeNull();
    expect(item('Digital menu').getByRole('button', { name: 'Activate' })).toBeVisible();
  });
  it('shows connections when features fail and can retry loading features', async () => {
    api.listAddons.mockRejectedValueOnce(new Error('offline'));
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={client}>
        <Extensions />
      </QueryClientProvider>
    );
    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Some business features could not be loaded'
    );
    expect(screen.getByRole('heading', { name: 'Stripe' })).toBeVisible();
    await userEvent.click(screen.getByRole('button', { name: 'Try again' }));
    expect(await screen.findByRole('heading', { name: 'Digital menu' })).toBeVisible();
  });
  it('allows members to browse without exposing management actions', async () => {
    useUserStore.setState({ user: { effectiveAccess: createEffectiveAccess() } });
    await openExtensions();
    expect(item('Stripe').getByRole('button', { name: 'Connect' })).toBeDisabled();
    expect(item('Digital menu').getByRole('button', { name: 'Activate' })).toBeDisabled();
    expect(screen.getByText('A business owner or admin can add and manage tools.')).toBeVisible();
  });
  it('blocks feature changes in trial mode while keeping connections available', async () => {
    api.subscription.data = { status: 'trial' };
    await openExtensions();
    expect(item('Digital menu').getByRole('button', { name: 'Activate' })).toBeDisabled();
    expect(item('Custom domain').getByRole('button', { name: 'Remove' })).toBeDisabled();
    expect(item('Stripe').getByRole('button', { name: 'Connect' })).toBeEnabled();
  });
  it('blocks feature changes if the subscription could not be verified', async () => {
    api.subscription.data = undefined;
    api.subscription.isError = true;
    await openExtensions();
    expect(screen.getByRole('alert')).toHaveTextContent('Unable to check your subscription');
    expect(item('Digital menu').getByRole('button', { name: 'Activate' })).toBeDisabled();
  });
});
