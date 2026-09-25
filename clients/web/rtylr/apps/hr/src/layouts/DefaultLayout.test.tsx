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
  pathname: '/staff',
}));

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, to, ...props }: React.PropsWithChildren<{ to: string }>) => (
    <a href={to} {...props}>
      {children}
    </a>
  ),
  Outlet: () => <div data-testid="outlet" />,
  useRouterState: ({ select }: { select: (s: { location: { pathname: string } }) => unknown }) =>
    select({ location: { pathname: mocks.pathname } }),
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
// `config: null` is production's real state, since nothing populates
// timeTrackingModel; hrConfig.value lets a test present the field as it would
// look once something does.
const hrConfig = vi.hoisted(() => ({
  value: null as { timeTrackingModel?: string } | null,
}));
vi.mock('../shared/stores/hrConfigStore', () => ({
  useHRConfigStore: (select: (s: { config: { timeTrackingModel?: string } | null }) => unknown) =>
    select({ config: hrConfig.value }),
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
    mocks.pathname = '/staff';
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

describe('DefaultLayout time tracking nav entries (#1068)', () => {
  const linkTo = (href: string): HTMLElement | null =>
    screen.queryAllByRole('link').find((a) => a.getAttribute('href') === href) ?? null;

  beforeEach(() => {
    mocks.role = 'admin';
    mocks.pathname = '/staff';
    hrConfig.value = null;
  });

  it('does not advertise Time when the config carries no tracking model', () => {
    // The route requires timeTrackingModel to be time_logging or both, and the
    // sidebar used to default it to 'both', so it offered a link that
    // redirected away for every user and every role.
    render(<DefaultLayout />);

    expect(linkTo('/time')).toBeNull();
  });

  it('does not advertise Time when the config exists but the field is unset', () => {
    hrConfig.value = {};

    render(<DefaultLayout />);

    expect(linkTo('/time')).toBeNull();
  });

  it.each(['time_logging', 'both'])('advertises Time when the model is %s', (model) => {
    hrConfig.value = { timeTrackingModel: model };

    render(<DefaultLayout />);

    // Guards the fix from over-reaching: once the backend populates the field,
    // the entry has to come back.
    expect(linkTo('/time')).not.toBeNull();
  });

  it('does not advertise Time for an attendance-only model', () => {
    hrConfig.value = { timeTrackingModel: 'attendance' };

    render(<DefaultLayout />);

    expect(linkTo('/time')).toBeNull();
  });

  it('keeps Attendance advertised, whose route has no config guard', () => {
    hrConfig.value = null;

    render(<DefaultLayout />);

    // Attendance uses `wrap`, not `configWrap`, so it is always reachable and
    // must keep its default. Removing the sidebar's fallback outright would
    // have hidden a working screen.
    expect(linkTo('/attendance')).not.toBeNull();
  });
});

describe('DefaultLayout sidebar active item', () => {
  const linkTo = (href: string): HTMLElement | null =>
    screen.queryAllByRole('link').find((a) => a.getAttribute('href') === href) ?? null;
  const isActive = (href: string): boolean => linkTo(href)?.style.color === 'var(--hr-text)';

  beforeEach(() => {
    mocks.role = 'admin';
    hrConfig.value = null;
    mocks.pathname = '/performance/goals';
  });

  it('marks only the most specific nav item active for a nested route', () => {
    render(<DefaultLayout />);

    expect(isActive('/performance/goals')).toBe(true);
    // `/performance` is a prefix of `/performance/goals`; both used to render
    // the active pill and the same `layoutId="sidebar-active"`.
    expect(isActive('/performance')).toBe(false);
  });

  it('keeps a section active for a detail route that has no nav item', () => {
    mocks.pathname = '/staff/123';

    render(<DefaultLayout />);

    expect(isActive('/staff')).toBe(true);
    expect(isActive('/staff/reporting-lines')).toBe(false);
  });

  it('marks the dashboard active only on the root path', () => {
    mocks.pathname = '/performance/goals';
    const { rerender } = render(<DefaultLayout />);
    expect(isActive('/')).toBe(false);

    mocks.pathname = '/';
    rerender(<DefaultLayout />);
    expect(isActive('/')).toBe(true);
    expect(isActive('/performance/goals')).toBe(false);
  });
});
