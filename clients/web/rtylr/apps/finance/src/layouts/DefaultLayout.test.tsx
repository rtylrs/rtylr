import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi, beforeAll } from 'vitest';
import DefaultLayout from './DefaultLayout';
import { initTestI18n } from '../shared/testing/i18n';
import { createEffectiveAccess } from '../../../../test/effectiveAccess';

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  signOut: vi.fn(),
  router: { pathname: '/invoices', searchStr: '' },
  effectiveAccess: null as unknown,
}));

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, to, ...props }: React.PropsWithChildren<{ to: string }>) => (
    <a href={to} {...props}>
      {children}
    </a>
  ),
  Outlet: () => <div data-testid="outlet" />,
  useNavigate: () => mocks.navigate,
  useRouterState: ({ select }: { select: (state: { location: typeof mocks.router }) => unknown }) =>
    select({ location: mocks.router }),
}));

vi.mock('@voxire/auth', () => {
  const state = () => ({
    user: {
      id: '1',
      name: 'Test User',
      email: 'test@example.com',
      effectiveAccess: mocks.effectiveAccess,
    },
    effectiveAccess: mocks.effectiveAccess,
  });
  return {
    signOut: mocks.signOut,
    useUserStore: Object.assign(
      (select: (value: ReturnType<typeof state>) => unknown) => select(state()),
      { getState: state }
    ),
  };
});
vi.mock('@voxire/components', () => ({ LicenseBanner: () => null }));
vi.mock('@voxire/config', () => ({ getAvatarUrl: () => '' }));
vi.mock('@voxire/caret', () => ({
  buildCaretRouteContext: () => ({}),
  CaretButton: ({ className }: { className?: string }) => (
    <button type="button" className={className}>
      Caret
    </button>
  ),
  CaretPanel: () => null,
  useCaretStore: () => ({ isOpen: false }),
}));
vi.mock('@voxire/ui', () => ({
  Avatar: ({ alt }: { alt: string }) => <span role="img" aria-label={alt} />,
}));
vi.mock('@voxire/utils', () => ({ useForesightLinks: () => undefined }));
vi.mock('../shared/components/GlobalSearch', () => ({ default: () => null }));
vi.mock('../shared/components/TrialBanner', () => ({ default: () => null }));
vi.mock('../shared/lib/clients', () => ({
  licenseClient: { getLicenseState: () => Promise.resolve({ status: 'hosted', valid: true }) },
  userClient: { getMe: () => Promise.resolve({}) },
}));
vi.mock('motion/react', () => {
  const Motion = ({
    children,
    className,
    style,
    onClick,
  }: React.PropsWithChildren<{
    className?: string;
    style?: React.CSSProperties;
    onClick?: React.MouseEventHandler<HTMLDivElement>;
  }>) => (
    <div className={className} style={style} onClick={onClick}>
      {children}
    </div>
  );

  return {
    AnimatePresence: ({ children }: React.PropsWithChildren) => <>{children}</>,
    motion: { div: Motion, main: Motion, span: Motion },
  };
});

// i18next is not mocked here: the real locale files mean an assertion
// written in English also proves the key resolves (rtylr#1269).
beforeAll(async () => {
  await initTestI18n('en');
});

describe('Finance mobile navigation', () => {
  beforeEach(() => {
    mocks.router.pathname = '/invoices';
    mocks.navigate.mockClear();
    mocks.signOut.mockClear();
    mocks.effectiveAccess = createEffectiveAccess('finance', { isAdmin: true });
  });

  it('provides a mobile menu button and opens the sidebar drawer', () => {
    render(<DefaultLayout />);

    const sidebar = screen.getByRole('complementary', { name: 'Primary navigation' });
    expect(sidebar).toHaveClass('hidden', 'md:flex', 'w-56');

    const openButton = screen.getByRole('button', { name: 'Open navigation menu' });
    expect(openButton).toHaveClass('h-11', 'w-11');
    fireEvent.click(openButton);

    expect(sidebar).toHaveClass('fixed', 'inset-y-0', 'start-0', 'z-50', 'flex');
    expect(screen.getByRole('button', { name: 'Close navigation menu' })).toBeInTheDocument();
    expect(openButton).toHaveAttribute('aria-expanded', 'true');
  });

  it('closes the drawer when the close button or Escape is used', () => {
    render(<DefaultLayout />);

    fireEvent.click(screen.getByRole('button', { name: 'Open navigation menu' }));
    fireEvent.click(screen.getByRole('button', { name: 'Close navigation menu' }));
    expect(screen.getByRole('complementary', { name: 'Primary navigation' })).toHaveClass('hidden');

    fireEvent.click(screen.getByRole('button', { name: 'Open navigation menu' }));
    fireEvent.keyDown(window, { key: 'Escape' });
    expect(screen.getByRole('complementary', { name: 'Primary navigation' })).toHaveClass('hidden');
  });
});

describe('Finance sidebar permission filtering', () => {
  beforeEach(() => {
    mocks.router.pathname = '/invoices';
    mocks.navigate.mockClear();
    mocks.signOut.mockClear();
  });

  it('hides admin-tier nav items for a role without those permissions', () => {
    mocks.effectiveAccess = createEffectiveAccess('finance');
    render(<DefaultLayout />);

    expect(screen.queryByText('Bank Feeds')).not.toBeInTheDocument();
    expect(screen.queryByText('Reconcile')).not.toBeInTheDocument();
    expect(screen.queryByText('Accounts')).not.toBeInTheDocument();
    expect(screen.queryByText('Tax Config')).not.toBeInTheDocument();
    // Ungated items in the same groups remain visible.
    expect(screen.getByText('Journal')).toBeInTheDocument();
  });

  it('shows every nav item for admin', () => {
    mocks.effectiveAccess = createEffectiveAccess('finance', { isAdmin: true });
    render(<DefaultLayout />);

    expect(screen.getByText('Bank Feeds')).toBeInTheDocument();
    expect(screen.getByText('Reconcile')).toBeInTheDocument();
    expect(screen.getByText('Accounts')).toBeInTheDocument();
    expect(screen.getByText('Tax Config')).toBeInTheDocument();
  });

  it('shows every nav item for owner', () => {
    mocks.effectiveAccess = createEffectiveAccess('finance', { isOwner: true });
    render(<DefaultLayout />);

    expect(screen.getByText('Bank Feeds')).toBeInTheDocument();
    expect(screen.getByText('Reconcile')).toBeInTheDocument();
    expect(screen.getByText('Accounts')).toBeInTheDocument();
    expect(screen.getByText('Tax Config')).toBeInTheDocument();
  });
});
