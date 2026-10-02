import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import path from 'node:path';
import React from 'react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18nFrom } from '../../../test/i18nTestInstance';
import { BusinessRoute } from './BusinessRoute';
import { useUserStore } from '../stores/userStore';

const mockGetMe = vi.fn();
const mockGetConfig = vi.fn();
const routerState = vi.hoisted(() => ({ pathname: '/' }));

vi.mock('../lib/clients', () => ({
  userClient: { getMe: (...args: unknown[]) => mockGetMe(...args) },
  configClient: { getConfig: (...args: unknown[]) => mockGetConfig(...args) },
  accessClient: { requestAccess: vi.fn() },
}));

vi.mock('../../shell/PortalShell', () => ({
  default: ({ children }: React.PropsWithChildren) => <div>{children}</div>,
}));

vi.mock('@voxire/ui', () => ({ Avatar: () => <div data-testid="avatar" /> }));

vi.mock('@tanstack/react-router', () => ({
  Link: ({ to, children }: React.PropsWithChildren<{ to: string }>) => <a href={to}>{children}</a>,
  Outlet: () => <div data-testid="outlet">Outlet Content</div>,
  useRouterState: ({
    select,
  }: { select?: (s: { location: { pathname: string } }) => string } = {}) =>
    select ? select({ location: { pathname: routerState.pathname } }) : routerState.pathname,
}));

vi.mock('../../features/onboarding/CreateBusinessOnboarding', () => ({
  default: () => <div data-testid="onboarding">Create Business Onboarding</div>,
}));

type AppGrant = {
  appId: string;
  permissions: string[];
};

function accessDocument(
  options: {
    isOwner?: boolean;
    isAdmin?: boolean;
    dashPermissions?: string[];
    appGrants?: AppGrant[];
  } = {}
): Record<string, unknown> {
  const appGrants = options.appGrants ?? [];
  return {
    businessId: '10',
    dashPermissions: options.dashPermissions ?? [],
    subscribedApps: appGrants.map((grant) => grant.appId),
    licenses: [],
    apps: appGrants.map((grant) => ({
      appId: grant.appId,
      licensed: true,
      hasAccess: true,
      permissions: grant.permissions,
      permissionScopes: grant.permissions.map((permissionKey) => ({ permissionKey })),
      sources: ['explicit'],
    })),
    isOwner: options.isOwner ?? false,
    isAdmin: options.isAdmin ?? false,
    authorizationVersion: 1,
  };
}

describe('BusinessRoute authorization', () => {
  beforeAll(() => initTestI18nFrom(path.resolve(import.meta.dirname, '../../locales')));

  beforeEach(() => {
    vi.clearAllMocks();
    useUserStore.setState({ user: null, token: null });
    document.cookie = 'token=test-token; path=/';
    routerState.pathname = '/';
    mockGetConfig.mockResolvedValue({ business: { name: 'Acme Corp' } });
  });

  it('routes a user without a business into onboarding', async () => {
    mockGetMe.mockResolvedValue({ id: 1, email: 'test@example.com', businessId: 0 });

    render(<BusinessRoute />);

    await waitFor(() => expect(screen.getByTestId('onboarding')).toBeTruthy());
  });

  it('allows the owner to enter Dash', async () => {
    mockGetMe.mockResolvedValue({
      id: 1,
      email: 'owner@example.com',
      businessId: 10,
      effectiveAccess: accessDocument({
        isOwner: true,
        appGrants: [{ appId: 'flow', permissions: ['flow:tasks:view'] }],
      }),
    });

    render(<BusinessRoute />);

    await waitFor(() => expect(screen.getByTestId('outlet')).toBeTruthy());
  });

  it('shows a retryable error instead of cached member restrictions when account loading fails', async () => {
    useUserStore.setState({
      user: { id: 1, email: 'owner@example.com', fullName: 'Cached User', businessId: 10 },
    });
    mockGetMe.mockRejectedValueOnce(new Error('Authorization schema unavailable'));
    mockGetMe.mockResolvedValueOnce({
      id: 1,
      email: 'owner@example.com',
      businessId: 10,
      effectiveAccess: accessDocument({ isOwner: true }),
    });

    render(<BusinessRoute />);

    expect(await screen.findByRole('alert')).toHaveTextContent("Couldn't load your account");
    expect(screen.queryByText(/Welcome/i)).toBeNull();
    expect(screen.queryByTestId('outlet')).toBeNull();
    expect(mockGetConfig).not.toHaveBeenCalled();

    await userEvent.click(screen.getByRole('button', { name: 'Refresh' }));

    await waitFor(() => expect(screen.getByTestId('outlet')).toBeTruthy());
    expect(screen.queryByRole('alert')).toBeNull();
  });

  it('shows an account error when a business profile has no valid authorization document', async () => {
    mockGetMe.mockResolvedValue({
      id: 1,
      email: 'owner@example.com',
      businessId: 10,
    });

    render(<BusinessRoute />);

    expect(await screen.findByRole('alert')).toHaveTextContent("Couldn't load your account");
    expect(screen.queryByText(/Welcome/i)).toBeNull();
    expect(screen.queryByTestId('outlet')).toBeNull();
  });

  it('allows the admin to enter Dash without per-page grants', async () => {
    mockGetMe.mockResolvedValue({
      id: 2,
      email: 'admin@example.com',
      businessId: 10,
      effectiveAccess: accessDocument({ isAdmin: true }),
    });
    routerState.pathname = '/team';

    render(<BusinessRoute />);

    await waitFor(() => expect(screen.getByTestId('outlet')).toBeTruthy());
  });

  it('allows a member to open billing with the matching page permission', async () => {
    mockGetMe.mockResolvedValue({
      id: 3,
      email: 'billing@example.com',
      businessId: 10,
      effectiveAccess: accessDocument({ dashPermissions: ['hub:billing:view'] }),
    });
    routerState.pathname = '/settings/billing';

    render(<BusinessRoute />);

    await waitFor(() => expect(screen.getByTestId('outlet')).toBeTruthy());
  });

  it('allows an invite-only member into the invite page without team read permission', async () => {
    mockGetMe.mockResolvedValue({
      id: 3,
      email: 'inviter@example.test',
      businessId: 10,
      effectiveAccess: accessDocument({ dashPermissions: ['hub:team:invite'] }),
    });
    routerState.pathname = '/team/invite';

    render(<BusinessRoute />);

    await waitFor(() => expect(screen.getByTestId('outlet')).toBeTruthy());
  });

  it('allows a member to open a subscribed app page with its permission', async () => {
    mockGetMe.mockResolvedValue({
      id: 4,
      email: 'menu@example.com',
      businessId: 10,
      effectiveAccess: accessDocument({
        appGrants: [{ appId: 'menu', permissions: ['menu:branding:edit'] }],
      }),
    });
    routerState.pathname = '/design';

    render(<BusinessRoute />);

    await waitFor(() => expect(screen.getByTestId('outlet')).toBeTruthy());
  });

  it('allows every active member to enter Portal Home', async () => {
    mockGetMe.mockResolvedValue({
      id: 5,
      email: 'staff@example.com',
      fullName: 'Staff User',
      businessId: 10,
      effectiveAccess: accessDocument(),
    });

    render(<BusinessRoute />);

    await waitFor(() => expect(screen.getByTestId('outlet')).toBeTruthy());
  });

  it('does not infer owner privileges from a legacy role field', async () => {
    mockGetMe.mockResolvedValue({
      id: 6,
      email: 'stale@example.com',
      fullName: 'Stale Role User',
      businessId: 10,
      role: 'owner',
      effectiveAccess: accessDocument(),
    });

    routerState.pathname = '/team';
    render(<BusinessRoute />);

    await waitFor(() =>
      expect(screen.getByRole('alert')).toHaveTextContent('You don’t have access to this page')
    );
    expect(screen.queryByTestId('outlet')).toBeNull();
  });

  it('allows personal account settings without app permissions', async () => {
    mockGetMe.mockResolvedValue({
      id: 7,
      email: 'staff@example.com',
      businessId: 10,
      effectiveAccess: accessDocument(),
    });
    routerState.pathname = '/settings/account';

    render(<BusinessRoute />);

    await waitFor(() => expect(screen.getByTestId('outlet')).toBeTruthy());
  });
});
