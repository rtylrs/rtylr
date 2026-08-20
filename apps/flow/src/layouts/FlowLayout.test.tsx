import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import FlowLayout from './FlowLayout';

// FlowLayout mounts the whole app shell (sidebar, header, realtime wiring,
// caret panel, command bar, etc.). This test only exercises the account menu
// (avatar/name trigger in the sidebar footer - issue #785: it didn't close on
// Escape or on an outside click), so every react-query-backed hook/component
// FlowLayout pulls in is stubbed out. Store hooks (userStore, themeStore,
// flowStore) and other pure/local hooks (useSidebarSections, useKeyboardShortcuts,
// useLifePreferences, useForesightLinks) are left real - none of them touch the
// network and they're already exercised by other passing tests in this app.
const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  signOut: vi.fn(),
  location: { pathname: '/inbox' },
  // A stable object reference: FlowLayout mirrors this into its local
  // userStore via a `useEffect` keyed on `[authUser, setUser]`. If the mock
  // returned a fresh object on every render, that effect would think the
  // dependency changed every time and call setUser in a loop.
  authUser: { id: '1', name: 'Test User', email: 'test@example.com' },
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, to, ...props }: React.PropsWithChildren<{ to: string }>) => (
    <a href={to} {...props}>
      {children}
    </a>
  ),
  useLocation: () => mocks.location,
  useNavigate: () => mocks.navigate,
  useParams: () => ({}),
  useRouterState: ({
    select,
  }: {
    select: (state: { location: { searchStr: string } }) => unknown;
  }) => select({ location: { searchStr: '' } }),
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
    onClick?: React.MouseEventHandler;
  }>) => (
    <div className={className} style={style} onClick={onClick}>
      {children}
    </div>
  );

  return {
    AnimatePresence: ({ children }: React.PropsWithChildren) => <>{children}</>,
    motion: { div: Motion, main: Motion, span: Motion, button: Motion },
  };
});

vi.mock('@voxire/config', () => ({
  useFeatureFlag: () => false,
  FLAG_KEYS: { APP_FLOW_PREVIEW_ROUTES: 'app.flow.previewRoutes' },
  getAvatarUrl: () => '',
  resolveAppUrl: () => '',
  // Pulled in transitively: '../shared/lib/api' calls createApiClient() at
  // module scope, which calls this to resolve the axios baseURL.
  resolveApiBase: () => 'https://api.test/',
}));

vi.mock('@voxire/ui', () => ({
  Avatar: ({ alt }: { alt: string }) => <span role="img" aria-label={alt} />,
}));

vi.mock('@voxire/caret', () => ({
  buildCaretRouteContext: () => ({}),
  CaretPanel: () => null,
  CaretButton: ({ className }: { className?: string }) => (
    <button type="button" className={className}>
      Caret
    </button>
  ),
  useCaretStore: () => ({ isOpen: false }),
}));

vi.mock('@voxire/utils', () => ({ useForesightLinks: () => undefined }));

vi.mock('@voxire/auth', () => ({
  signOut: mocks.signOut,
  useUserStore: (select: (state: { user: typeof mocks.authUser }) => unknown) =>
    select({ user: mocks.authUser }),
}));

vi.mock('@voxire/components', () => ({ LicenseBanner: () => null }));

vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

vi.mock('../shared/components/TrialBanner', () => ({ default: () => null }));
vi.mock('../shared/components/CommandBar', () => ({ default: () => null }));
vi.mock('../shared/components/QuickCreate', () => ({ default: () => null }));
vi.mock('../shared/components/BulkEditBar', () => ({ default: () => null }));

vi.mock('../features/workspace-dashboard/useWorkspaces', () => ({
  useWorkspaces: () => ({ data: undefined, isLoading: false }),
}));
vi.mock('../features/projects/useProjects', () => ({
  useProjects: () => ({ data: undefined }),
  useCreateProject: () => ({ mutate: vi.fn(), isPending: false }),
}));
vi.mock('../features/tasks/useMyTasksProject', () => ({
  useMyTasksProject: () => ({ data: undefined }),
}));
vi.mock('../features/inbox/useNotifications', () => ({
  useNotifications: () => ({ data: [] }),
}));
vi.mock('../shared/hooks/useFlowBusinessName', () => ({
  useFlowBusinessName: () => ({ data: null }),
}));
vi.mock('../shared/hooks/queries/useFlowTeam', () => ({
  useFlowTeam: () => ({ data: [] }),
}));
vi.mock('../shared/lib/useLayoutRealtime', () => ({
  useLayoutRealtime: () => undefined,
}));
vi.mock('../shared/lib/realtime', () => ({
  useRealtimeReconnectRefetch: () => undefined,
  useRealtimeStatus: () => 'disconnected',
}));
vi.mock('../shared/lib/realtimeRouteGate', () => ({
  routeNeedsRealtime: () => false,
}));

describe('FlowLayout account menu', () => {
  beforeEach(() => {
    mocks.navigate.mockClear();
    mocks.signOut.mockClear();
    mocks.location = { pathname: '/inbox' };
  });

  const renderLayout = () =>
    render(
      <FlowLayout>
        <div>content</div>
      </FlowLayout>
    );

  it('opens on trigger click and shows the sign-out action', () => {
    renderLayout();

    expect(screen.queryByText('flow.shell.signOut')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: /Test User/i }));

    expect(screen.getByText('flow.shell.signOut')).toBeInTheDocument();
  });

  it('closes on Escape', () => {
    renderLayout();

    fireEvent.click(screen.getByRole('button', { name: /Test User/i }));
    expect(screen.getByText('flow.shell.signOut')).toBeInTheDocument();

    fireEvent.keyDown(document, { key: 'Escape' });

    expect(screen.queryByText('flow.shell.signOut')).not.toBeInTheDocument();
  });

  it('closes when clicking outside the menu', () => {
    renderLayout();

    fireEvent.click(screen.getByRole('button', { name: /Test User/i }));
    expect(screen.getByText('flow.shell.signOut')).toBeInTheDocument();

    fireEvent.mouseDown(document.body);

    expect(screen.queryByText('flow.shell.signOut')).not.toBeInTheDocument();
  });

  it('stays open when clicking inside the menu', () => {
    renderLayout();

    fireEvent.click(screen.getByRole('button', { name: /Test User/i }));
    const signOutButton = screen.getByText('flow.shell.signOut');

    fireEvent.mouseDown(signOutButton);

    expect(screen.getByText('flow.shell.signOut')).toBeInTheDocument();
  });
});
