import { render, screen, waitFor } from '@testing-library/react';
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
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

vi.mock('@voxire/ui', () => ({
  Avatar: () => <div data-testid="avatar" />,
}));

vi.mock('lucide-react', () => ({
  MonitorSpeaker: () => <span data-testid="icon" />,
  Package: () => <span data-testid="icon" />,
  UserCheck: () => <span data-testid="icon" />,
  Users: () => <span data-testid="icon" />,
  CreditCard: () => <span data-testid="icon" />,
  Layers: () => <span data-testid="icon" />,
  Lock: () => <span data-testid="icon" />,
  BarChart3: () => <span data-testid="icon" />,
}));

vi.mock('@tanstack/react-router', () => ({
  Outlet: () => <div data-testid="outlet">Outlet Content</div>,
  useRouterState: ({ select }: { select?: (s: any) => any } = {}) =>
    select ? select({ location: { pathname: routerState.pathname } }) : routerState.pathname,
}));

vi.mock('../../features/onboarding/CreateBusinessOnboarding', () => ({
  default: () => <div data-testid="onboarding">Create Business Onboarding</div>,
}));

describe('BusinessRoute', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useUserStore.setState({ user: null, token: null });
    document.cookie = 'token=test-token; path=/';
    routerState.pathname = '/';
    mockGetConfig.mockResolvedValue({ business: {} });
  });

  it('renders onboarding if user has no businessId', async () => {
    mockGetMe.mockResolvedValue({
      id: 1,
      email: 'test@example.com',
      businessId: 0,
      isOwner: true,
    });

    render(<BusinessRoute />);

    await waitFor(() => {
      expect(screen.getByTestId('onboarding')).toBeTruthy();
    });
  });

  it('renders Outlet if user has businessId and is owner', async () => {
    mockGetMe.mockResolvedValue({
      id: 1,
      email: 'test@example.com',
      businessId: 10,
      isOwner: true,
      role: 'owner',
    });
    mockGetConfig.mockResolvedValue({ business: { name: 'Acme Corp' } });

    render(<BusinessRoute />);

    await waitFor(() => {
      expect(screen.getByTestId('outlet')).toBeTruthy();
    });
  });

  it('renders StaffView if non-owner staff has no management role', async () => {
    mockGetMe.mockResolvedValue({
      id: 2,
      email: 'staff@example.com',
      fullName: 'Staff User',
      businessId: 10,
      isOwner: false,
      role: 'staff',
      apps: [],
    });

    render(<BusinessRoute />);

    await waitFor(() => {
      expect(screen.getByText(/Welcome/i)).toBeTruthy();
    });
  });

  it('renders StaffView for a non-owner operator without normalized management access', async () => {
    mockGetMe.mockResolvedValue({
      id: 3,
      email: 'op@example.com',
      businessId: 10,
      isOwner: false,
      role: 'operator',
      apps: [],
    });
    routerState.pathname = '/team';

    render(<BusinessRoute />);

    await waitFor(() => {
      expect(screen.getByText(/Welcome/i)).toBeTruthy();
    });
  });

  it('renders StaffView for a non-owner billing user on an out-of-scope route', async () => {
    mockGetMe.mockResolvedValue({
      id: 4,
      email: 'billing@example.com',
      fullName: 'Billing User',
      businessId: 10,
      isOwner: false,
      role: 'billing',
      apps: [],
    });
    routerState.pathname = '/team';

    render(<BusinessRoute />);

    await waitFor(() => {
      expect(screen.getByText(/Welcome/i)).toBeTruthy();
    });
  });

  it('renders Outlet for a non-owner billing user on their scoped billing route', async () => {
    mockGetMe.mockResolvedValue({
      id: 5,
      email: 'billing@example.com',
      businessId: 10,
      isOwner: false,
      role: 'billing',
      apps: [],
    });
    routerState.pathname = '/settings/billing';

    render(<BusinessRoute />);

    await waitFor(() => {
      expect(screen.getByTestId('outlet')).toBeTruthy();
    });
  });

  it('renders StaffView for a non-owner menu_manager user on an out-of-scope route', async () => {
    mockGetMe.mockResolvedValue({
      id: 6,
      email: 'menu@example.com',
      fullName: 'Menu Manager User',
      businessId: 10,
      isOwner: false,
      role: 'menu_manager',
      apps: [],
    });
    routerState.pathname = '/settings/billing';

    render(<BusinessRoute />);

    await waitFor(() => {
      expect(screen.getByText(/Welcome/i)).toBeTruthy();
    });
  });

  it('renders Outlet for a non-owner menu_manager user on their scoped design route', async () => {
    mockGetMe.mockResolvedValue({
      id: 7,
      email: 'menu@example.com',
      businessId: 10,
      isOwner: false,
      role: 'menu_manager',
      apps: [],
    });
    routerState.pathname = '/design';

    render(<BusinessRoute />);

    await waitFor(() => {
      expect(screen.getByTestId('outlet')).toBeTruthy();
    });
  });

  it('renders StaffView for a non-owner store_manager user (POS-location-scoped, not global to Dash)', async () => {
    mockGetMe.mockResolvedValue({
      id: 8,
      email: 'store@example.com',
      fullName: 'Store Manager User',
      businessId: 10,
      isOwner: false,
      role: 'store_manager',
      apps: [],
    });
    routerState.pathname = '/team';

    render(<BusinessRoute />);

    await waitFor(() => {
      expect(screen.getByText(/Welcome/i)).toBeTruthy();
    });
  });
});
