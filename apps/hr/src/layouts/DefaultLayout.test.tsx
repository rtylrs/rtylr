import React from 'react';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import DefaultLayout from './DefaultLayout';

/**
 * Reachability test for the reporting-lines screen.
 *
 * The screen, its route, its tests and all three locales shipped without any
 * link to it, so it was reachable only by typing the URL. That is the same
 * orphaned-screen shape as the deleted `features/profile/`, and nothing failed
 * when it happened - there was no test that asked whether a route is reachable.
 *
 * Both directions are asserted on purpose. Presence alone would still pass with
 * the `permission` key deleted; absence alone would still pass with the whole
 * nav item deleted.
 */

const mocks = vi.hoisted(() => ({
  role: 'admin' as string | null,
}));

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, to, ...props }: React.PropsWithChildren<{ to: string }>) => (
    <a href={to} {...props}>
      {children}
    </a>
  ),
  Outlet: () => <div data-testid="outlet" />,
  useRouterState: ({ select }: { select: (s: { location: { pathname: string } }) => unknown }) =>
    select({ location: { pathname: '/staff' } }),
}));

vi.mock('@voxire/auth', () => ({
  useUserStore: (select: (state: Record<string, unknown>) => unknown) =>
    select({
      user: { id: '1', name: 'Test User', email: 'test@example.com', role: mocks.role },
      role: mocks.role,
      setUser: vi.fn(),
    }),
}));
vi.mock('@voxire/components', () => ({ LicenseBanner: () => null }));
vi.mock('@voxire/config', () => ({ getAvatarUrl: () => '' }));
vi.mock('@voxire/caret', () => ({
  buildCaretRouteContext: () => ({}),
  CaretButton: () => null,
  CaretPanel: () => null,
  useCaretStore: () => ({ isOpen: false }),
}));
vi.mock('@voxire/ui', () => ({
  Avatar: ({ alt }: { alt: string }) => <span role="img" aria-label={alt} />,
}));
vi.mock('@voxire/utils', () => ({ useForesightLinks: () => undefined }));
vi.mock('../shared/lib/clients', () => ({
  licenseClient: {},
  userClient: { getMe: () => Promise.resolve({}) },
}));
vi.mock('../shared/stores/hrConfigStore', () => ({
  useHRConfigStore: (select: (s: { config: null }) => unknown) => select({ config: null }),
}));
vi.mock('../shared/lib/hrSignOut', () => ({ signOutFromHR: vi.fn() }));
vi.mock('../shared/components/CommandPalette', () => ({ default: () => null }));
vi.mock('../shared/components/GlobalShortcuts', () => ({ default: () => null }));
vi.mock('../shared/components/ui', () => ({ useMediaQuery: () => false }));
vi.mock('../shared/lib/payrollConfig', () => ({ isPayrollConfigured: () => false }));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en' } }),
}));

vi.mock('motion/react', () => {
  const passthrough = ({ children, ...rest }: React.PropsWithChildren<Record<string, unknown>>) => {
    const safe = Object.fromEntries(
      Object.entries(rest).filter(([k]) => k === 'className' || k === 'style' || k === 'ref')
    );
    return <div {...safe}>{children}</div>;
  };
  // A Proxy rather than a fixed list: motion.main was missing from the first
  // version and the failure surfaced as "element type is invalid", which names
  // no tag. Any motion.<tag> the layout reaches for resolves here.
  return {
    motion: new Proxy({}, { get: () => passthrough }),
    AnimatePresence: ({ children }: React.PropsWithChildren) => <>{children}</>,
  };
});

const reportingLink = (): HTMLElement | null =>
  screen.queryAllByRole('link').find((a) => a.getAttribute('href') === '/staff/reporting-lines') ??
  null;

describe('DefaultLayout reporting-lines nav entry', () => {
  beforeEach(() => {
    mocks.role = 'admin';
  });

  it('links to the reporting-lines screen for a role that can manage staff', () => {
    render(<DefaultLayout />);
    expect(reportingLink()).not.toBeNull();
  });

  it('hides it from a role that cannot manage staff', () => {
    mocks.role = 'viewer';
    render(<DefaultLayout />);
    expect(reportingLink()).toBeNull();
  });
});
