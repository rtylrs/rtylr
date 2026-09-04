import { render, screen } from '@testing-library/react';
import React from 'react';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import VendorDashboard from './VendorDashboard';
import { useUserStore } from '../../../shared/stores/userStore';

const navigateMock = vi.hoisted(() => vi.fn());

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => navigateMock,
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/ui', () => ({
  ErrorBoundary: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('react-grid-layout', () => ({
  ResponsiveGridLayout: () => <div data-testid="grid-layout" />,
}));

vi.mock('react-dom', async () => {
  const actual = await vi.importActual<typeof import('react-dom')>('react-dom');
  return { ...actual, createPortal: (node: React.ReactNode) => node };
});

vi.mock('motion/react', () => ({
  motion: { div: ({ children }: { children: React.ReactNode }) => <div>{children}</div> },
  AnimatePresence: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

vi.mock('./widgets', () => ({
  getWidgetComponent: () => null,
  WIDGET_CATALOG: [],
  loadLayout: () => [],
  saveLayout: () => undefined,
  loadActiveWidgets: () => [],
  saveActiveWidgets: () => undefined,
  DEFAULT_LAYOUT: [],
}));

describe('VendorDashboard', () => {
  beforeEach(() => {
    navigateMock.mockClear();
    useUserStore.setState({ user: null });
  });

  it.each(['admin', 'owner', 'Admin', 'OWNER'])(
    'renders AdminDashboard for role "%s"',
    async (role) => {
      useUserStore.setState({ user: { id: 1, role } });

      render(<VendorDashboard />);

      expect(await screen.findByTitle('Customize dashboard')).toBeTruthy();
      expect(screen.queryByText(/Welcome to rtylr/i)).toBeNull();
    }
  );

  it('shows the custom role action and opens role creation from the dashboard', async () => {
    useUserStore.setState({ user: { id: 1, role: 'admin' } });

    render(<VendorDashboard />);

    const createRoleButton = await screen.findByRole('button', {
      name: 'dashboard.createCustomRole',
    });
    createRoleButton.click();

    expect(navigateMock).toHaveBeenCalledWith({ to: '/roles', search: { create: 1 } });
  });

  it.each(['operator', 'billing', 'menu_manager', 'store_manager', 'staff', 'viewer'])(
    'renders StaffDashboard, not AdminDashboard, for role "%s"',
    (role) => {
      useUserStore.setState({ user: { id: 2, role, apps: [] } });

      render(<VendorDashboard />);

      expect(screen.getByText(/Welcome to rtylr/i)).toBeTruthy();
      expect(screen.queryByTitle('Customize dashboard')).toBeNull();
    }
  );

  it('renders the restrictive StaffDashboard, not AdminDashboard, when role is absent', () => {
    useUserStore.setState({ user: { id: 3, apps: [] } });

    render(<VendorDashboard />);

    expect(screen.getByText(/Welcome to rtylr/i)).toBeTruthy();
    expect(screen.queryByTestId('grid-layout')).toBeNull();
  });

  it('renders the restrictive StaffDashboard, not AdminDashboard, when there is no user at all', () => {
    useUserStore.setState({ user: null });

    render(<VendorDashboard />);

    expect(screen.getByText(/Welcome to rtylr/i)).toBeTruthy();
    expect(screen.queryByTestId('grid-layout')).toBeNull();
  });
});

describe('StaffDashboard app links (#1069)', () => {
  const hrefs: string[] = [];

  beforeEach(() => {
    hrefs.length = 0;
    useUserStore.setState({ user: null });
    // window.location.href is assigned directly; replacing the object makes the
    // assignment observable and stops happy-dom attempting a navigation.
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: {
        hostname: 'localhost',
        get href(): string {
          return 'http://localhost:5175/';
        },
        set href(value: string) {
          hrefs.push(value);
        },
      },
    });
  });

  const asStaffWith = (apps: string[]): void => {
    useUserStore.setState({ user: { id: 2, role: 'staff', apps } });
  };

  it('redirects a single-app staff member to a resolved URL, not to rtylr.com', () => {
    asStaffWith(['flow']);

    render(<VendorDashboard />);

    expect(hrefs).toHaveLength(1);
    expect(hrefs[0]).not.toContain('rtylr.com');
    // APP_URLS is localhost-aware and window.location.hostname is localhost here.
    expect(hrefs[0]).toContain('localhost');
  });

  it('links each app through the resolver rather than a built hostname', () => {
    asStaffWith(['crm', 'hr']);

    render(<VendorDashboard />);

    const links = screen.getAllByRole('link');
    expect(links.length).toBeGreaterThan(0);
    for (const link of links) {
      expect(link.getAttribute('href')).not.toContain('rtylr.com');
      expect(link.getAttribute('href')).not.toBe('undefined');
    }
  });

  it('offers nothing, and redirects nowhere, for an app id it cannot resolve', () => {
    // APP_URLS has no entry for this, so resolving it would have produced
    // href="undefined". Previously it produced https://not-an-app.rtylr.com.
    asStaffWith(['not-an-app']);

    render(<VendorDashboard />);

    expect(hrefs).toHaveLength(0);
    expect(screen.queryAllByRole('link')).toHaveLength(0);
  });

  it('redirects nowhere when the staff member has no apps', () => {
    asStaffWith([]);

    render(<VendorDashboard />);

    expect(hrefs).toHaveLength(0);
  });
});
