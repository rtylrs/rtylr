import { render, screen, fireEvent } from '@testing-library/react';
import React from 'react';
import { describe, it, expect, beforeEach, vi } from 'vitest';
import VendorDashboard from './VendorDashboard';
import { useUserStore } from '../../../shared/stores/userStore';
import { useCommandPalette } from '../../../shared/stores/commandPaletteStore';
import { createEffectiveAccess } from '../../../../../../test/effectiveAccess';
import type { TenantAppId } from '@voxire/types';

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

  it.each([
    ['admin', createEffectiveAccess(undefined, { isAdmin: true })],
    ['owner', createEffectiveAccess(undefined, { isOwner: true })],
  ])('renders zero-query home for %s', (memberType, effectiveAccess) => {
    useUserStore.setState({
      user: { id: 1, businessName: 'Acme', effectiveAccess } as never,
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
    useUserStore.setState({
      user: {
        id: 1,
        effectiveAccess: createEffectiveAccess(undefined, { isAdmin: true }),
      } as never,
    });

    render(<VendorDashboard />);

    fireEvent.click(screen.getByText(/Search apps, pages, actions/i));
    expect(useCommandPalette.getState().open).toBe(true);
  });

  it('renders StaffDashboard, not home, for a member without dashboard permission', () => {
    useUserStore.setState({ user: { id: 2, effectiveAccess: createEffectiveAccess() } as never });

    render(<VendorDashboard />);

    expect(screen.getByText(/Welcome to rtylr/i)).toBeTruthy();
    expect(screen.queryByText(/Search apps, pages, actions/i)).toBeNull();
  });

  it('does not infer dashboard access from a legacy role field', () => {
    useUserStore.setState({ user: { id: 3, role: 'admin' } as never });

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

  const asStaffWith = (apps: TenantAppId[]): void => {
    useUserStore.setState({
      user: {
        id: 2,
        effectiveAccess: createEffectiveAccess(apps[0], { appIds: apps }),
      } as never,
    });
  };

  it('keeps a single-app staff member on Dash and offers an explicit app link', () => {
    asStaffWith(['flow']);

    render(<VendorDashboard />);

    expect(hrefs).toHaveLength(0);
    const link = screen.getByRole('link', { name: 'Open flow' });
    // APP_URLS is localhost-aware and window.location.hostname is localhost here.
    expect(link.getAttribute('href')).toContain('localhost');
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
    asStaffWith(['menu']);

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
