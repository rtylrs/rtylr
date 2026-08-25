import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import FlowLayout from './FlowLayout';
import { useFlowStore } from '../shared/stores/flowStore';

// FlowLayout mounts the whole app shell (sidebar, header, realtime wiring,
// caret panel, command bar, etc.). This test only exercises the account menu
// (avatar/name trigger in the sidebar footer - issue #785: it didn't close on
// Escape or on an outside click), so every react-query-backed hook/component
// FlowLayout pulls in is stubbed out. Store hooks (userStore, appearanceStore,
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
  // Pulled in transitively: '@voxire/api' resolves its base URL at module
  // scope when the generated clients are constructed.
  resolveApiBase: () => 'https://api.test/',
}));

const providerProps = vi.hoisted(
  () =>
    ({ workspaceAccent: undefined }) as {
      workspaceAccent: string | undefined;
    }
);

vi.mock('@voxire/ui', () => ({
  Avatar: ({ alt }: { alt: string }) => <span role="img" aria-label={alt} />,
  // The real provider writes custom properties to documentElement, which this
  // test has no opinion about: the token values belong to the package's own
  // suite. What does belong here is the one prop this layout is responsible for
  // supplying, so the stub records it rather than discarding it.
  AppearanceProvider: ({
    children,
    workspaceAccent,
  }: React.PropsWithChildren<{ workspaceAccent?: string }>) => {
    providerProps.workspaceAccent = workspaceAccent;
    return <>{children}</>;
  },
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

// useForesightLinks is stubbed because it prefetches routes; useClickOutside is
// deliberately the REAL hook, since the outside-click cases below are what it
// exists to prove. Stubbing it would make those tests assert nothing.
vi.mock('@voxire/utils', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@voxire/utils')>()),
  useForesightLinks: () => undefined,
}));

vi.mock('@voxire/auth', () => ({
  signOut: mocks.signOut,
  useUserStore: (select: (state: { user: typeof mocks.authUser }) => unknown) =>
    select({ user: mocks.authUser }),
}));

vi.mock('@voxire/components', () => ({ LicenseBanner: () => null }));

// Sidebar preferences use React Query in production. The account-menu and
// appearance tests do not exercise that persistence path, so keep the shell
// isolated from its provider and network requirements here.
vi.mock('../shared/lib/sidebarPreferences', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../shared/lib/sidebarPreferences')>();
  const preferences = actual.createDefaultSidebarPreferences();
  const update = vi.fn();

  return {
    ...actual,
    useSidebarPreferences: () => ({
      preferences,
      update,
      isLoading: false,
      isSaving: false,
    }),
  };
});

vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

vi.mock('../shared/components/TrialBanner', () => ({ default: () => null }));
vi.mock('../shared/components/CommandBar', () => ({ default: () => null }));
vi.mock('../shared/components/QuickCreate', () => ({ default: () => null }));
vi.mock('../shared/components/BulkEditBar', () => ({ default: () => null }));
// The bin section runs its own react-query hook; this suite renders the layout
// without a QueryClientProvider, like the other data-fetching children above.
vi.mock('../features/projects-v2/RecycleBinSection', () => ({
  RecycleBinSection: () => null,
}));

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
  useFlowBusinessId: () => ({ data: '42' }),
}));
// FlowLayout reads real contracted hours per member (#868). Stubbed like every
// other react-query hook here: unmocked it needs a QueryClientProvider, and an
// empty map is the "no capacity known" case these tests do not exercise.
vi.mock('../shared/hooks/queries/useMemberCapacity', () => ({
  useMemberCapacityHours: () => ({ data: {} }),
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

// The 'FlowLayout workspace switcher' suite is deliberately gone.
//
// This PR removes workspace state from the layout: activeWorkspace is now a
// literal built from the business name, and there is no switcher to open, no
// list to enumerate and nothing to close. The four tests here asserted a control
// that no longer renders, so they were removed with it rather than left to fail
// or skipped as if the feature were coming back.
//
// If a workspace switcher returns, the assertions are recoverable from git
// history at this commit's parent.

describe('FlowLayout appearance wiring', () => {
  const renderLayout = () =>
    render(
      <FlowLayout>
        <div>content</div>
      </FlowLayout>
    );

  // A colour that is deliberately NOT ACCENT_DEFAULT, so the assertions below
  // can tell a live value from a constant. That mattered when the layout read a
  // real workspace: the fallback workspace's colour IS ACCENT_DEFAULT, so an
  // earlier version of this suite passed with the wiring replaced by a constant.
  const DISTINCT = '#a3d84b';

  beforeEach(() => {
    providerProps.workspaceAccent = undefined;
    useFlowStore.getState().setWorkspaces([]);
  });

  // This suite changed shape with the workspace removal, and the change is a
  // product decision worth reading rather than a test repair.
  //
  // The layout used to pass the active workspace's colour, which is what made
  // `accent: null` mean "follow the workspace" in the appearance panel. It now
  // builds activeWorkspace as a literal with `color: ACCENT_DEFAULT`, so the
  // accent the provider receives is constant.
  //
  // The consequence for the appearance feature: the user-accent-overrides-
  // workspace rule still holds, but its other half is gone. `accent: null` now
  // resolves to the brand teal rather than to anything workspace-derived. The
  // panel's reset control still works and the derivation is unaffected. Nothing
  // is broken; one half of a documented behaviour is simply no longer reachable.
  //
  // Asserting the constant rather than deleting the suite, so a future change
  // that reintroduces a real per-workspace accent has to come back here and say
  // so deliberately.
  it('hands the appearance provider a constant accent now that workspaces are gone', () => {
    useFlowStore.getState().setWorkspaces([{ id: 'ws-9', name: 'Studio', color: DISTINCT }]);
    useFlowStore.getState().setActiveWorkspace('ws-9');
    renderLayout();
    expect(providerProps.workspaceAccent).not.toBe(DISTINCT);
    expect(providerProps.workspaceAccent).toBeTruthy();
  });

  it('still supplies the provider an accent at all', () => {
    renderLayout();
    expect(providerProps.workspaceAccent).toBeTruthy();
  });
});
