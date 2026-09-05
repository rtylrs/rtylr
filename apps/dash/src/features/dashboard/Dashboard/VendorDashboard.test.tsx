import { render, screen, fireEvent } from '@testing-library/react';
import React from 'react';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import VendorDashboard from './VendorDashboard';
import { useUserStore } from '../../../shared/stores/userStore';
import { useCommandPalette } from '../../../shared/stores/commandPaletteStore';

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const mockNavigate = vi.fn();
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mockNavigate,
}));

describe('VendorDashboard', () => {
  beforeEach(() => {
    localStorage.clear();
    mockNavigate.mockClear();
    useUserStore.setState({ user: null });
    useCommandPalette.setState({ open: false });
  });

  it.each(['admin', 'owner', 'Admin', 'OWNER'])('renders zero-query home for role "%s"', (role) => {
    useUserStore.setState({
      user: { id: 1, role, businessName: 'Acme' } as never,
    });

    render(<VendorDashboard />);

    // Static home: greeting, search entry, nav columns — no widgets.
    expect(screen.getByText('Acme')).toBeTruthy();
    expect(screen.getByText(/Search apps, pages, actions/i)).toBeTruthy();
    expect(screen.getByText('Apps')).toBeTruthy();
    expect(screen.getByText('Shortcuts')).toBeTruthy();
    expect(screen.getByText('Recents')).toBeTruthy();
    expect(screen.queryByText(/Get the most out of rtylr/i)).toBeNull();
    expect(screen.queryByRole('checkbox')).toBeNull();
    expect(screen.queryByText(/Welcome to rtylr/i)).toBeNull();
  });

  it('opens the command palette from the big search button', () => {
    useUserStore.setState({ user: { id: 1, role: 'admin' } as never });

    render(<VendorDashboard />);

    fireEvent.click(screen.getByText(/Search apps, pages, actions/i));
    expect(useCommandPalette.getState().open).toBe(true);
  });

  it.each(['operator', 'billing', 'menu_manager', 'store_manager', 'staff', 'viewer'])(
    'renders StaffDashboard, not home, for role "%s"',
    (role) => {
      useUserStore.setState({ user: { id: 2, role, apps: [] } as never });

      render(<VendorDashboard />);

      expect(screen.getByText(/Welcome to rtylr/i)).toBeTruthy();
      expect(screen.queryByText(/Search apps, pages, actions/i)).toBeNull();
    }
  );

  it('renders the restrictive StaffDashboard, not home, when role is absent', () => {
    useUserStore.setState({ user: { id: 3, apps: [] } as never });

    render(<VendorDashboard />);

    expect(screen.getByText(/Welcome to rtylr/i)).toBeTruthy();
  });

  it('renders the restrictive StaffDashboard, not home, when there is no user at all', () => {
    useUserStore.setState({ user: null });

    render(<VendorDashboard />);

    expect(screen.getByText(/Welcome to rtylr/i)).toBeTruthy();
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
    useUserStore.setState({ user: { id: 2, role: 'staff', apps } as never });
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
